package repos

import (
	"context"
	"devcloud/apperr"
	"devcloud/ent"
	"devcloud/ent/resource"
	"devcloud/ent/task"
	"devcloud/transition"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
)

type Task struct {
	db *ent.Client
}

func NewTask(db *ent.Client) *Task {
	return &Task{db: db}
}

// List 分页返回指定用户的任务。
func (r *Task) List(ctx context.Context, userID, offset, limit int) ([]*ent.Task, error) {
	return r.db.Task.
		Query().
		Where(
			task.UserID(userID),
		).
		Offset(offset).
		Limit(limit).
		All(ctx)
}

// Count 统计指定用户的任务总数。
func (r *Task) Count(ctx context.Context, userID int) (int, error) {
	return r.db.Task.
		Query().
		Where(
			task.UserID(userID),
		).
		Count(ctx)
}

// Get 按 id + 用户查询单个任务；非本人任务返回 NotFound。
func (r *Task) Get(ctx context.Context, userID, id int) (*ent.Task, error) {
	return r.db.Task.
		Query().
		Where(
			task.ID(id),
			task.UserID(userID),
		).
		Only(ctx)
}

// Create 在同一事务内创建资源(status=CREATING)与 CREATE_RESOURCE 任务(PENDING)，
// 保证二者要么同时成功、要么同时回滚，避免出现没有任务的孤儿资源。
func (r *Task) Create(
	ctx context.Context,
	userID int,
	name string,
	provider resource.Provider,
	image string,
	config map[string]any,
	payload map[string]any,
) (*ent.Task, *ent.Resource, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()

	rsc, err := tx.Resource.
		Create().
		SetUserID(userID).
		SetName(name).
		SetProvider(provider).
		SetImage(image).
		SetStatus(resource.StatusCREATING).
		SetConfig(config).
		Save(ctx)
	if err != nil {
		return nil, nil, err
	}

	t, err := tx.Task.
		Create().
		SetUserID(userID).
		SetResourceID(rsc.ID).
		SetType(task.TypeCREATE_RESOURCE).
		SetStatus(task.StatusPENDING).
		SetPayload(payload).
		Save(ctx)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return t, rsc, nil
}

// ClaimPendingTask 抢占一条到期任务并建立租约。
//
//   - 用 `FOR UPDATE SKIP LOCKED` 选出最小的到期任务（PENDING 且 next_run_at<=now），
//     多个 worker 并发时不会领到同一条；
//   - 在同一事务内写 RUNNING、worker_id、started_at、heartbeat_at，并自增 attempts；
//   - 无任务时返回 apperr.ErrNoPendingTask（正常空闲）；
//   - 若重试次数已耗尽，先将其终结为 FAILED 再返回 ErrNoPendingTask。
func (r *Task) ClaimPendingTask(ctx context.Context, workerID string) (*ent.Task, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	t, err := tx.Task.
		Query().
		Where(
			task.StatusEQ(task.StatusPENDING),
			task.NextRunAtLTE(time.Now()),
		).
		Order(ent.Asc(task.FieldNextRunAt)).
		ForUpdate(
			sql.WithLockAction(sql.SkipLocked),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperr.ErrNoPendingTask
		}
		return nil, err
	}

	now := time.Now()

	// 兜底：重试次数已耗尽，直接终结该任务。
	if t.Attempts >= t.MaxAttempts {
		if err := tx.Task.
			UpdateOneID(t.ID).
			Where(task.StatusEQ(task.StatusPENDING)).
			SetStatus(task.StatusFAILED).
			SetErrorMessage("max attempts reached").
			SetFinishedAt(now).
			Exec(ctx); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return nil, apperr.ErrNoPendingTask
	}

	// 领取即建立租约：记录持有者、开始时间与心跳时间。
	err = tx.Task.
		UpdateOneID(t.ID).
		Where(task.StatusEQ(task.StatusPENDING)).
		SetStatus(task.StatusRUNNING).
		SetWorkerID(workerID).
		SetAttempts(t.Attempts + 1).
		SetStartedAt(now).
		SetHeartbeatAt(now).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	t.Status = task.StatusRUNNING
	t.WorkerID = workerID
	t.Attempts += 1
	t.StartedAt = &now
	t.HeartbeatAt = &now
	return t, nil
}

// Heartbeat 刷新任务心跳，仅当任务仍由该 worker 持有且处于 RUNNING。
func (r *Task) Heartbeat(ctx context.Context, id int, workerID string) error {
	return r.db.Task.
		UpdateOneID(id).
		Where(
			task.StatusEQ(task.StatusRUNNING),
			task.WorkerIDEQ(workerID),
		).
		SetHeartbeatAt(time.Now()).
		Exec(ctx)
}

// SetSuccess 完成持有中的任务；workerID 作为栅栏，旧持有者无法覆盖。
func (r *Task) SetSuccess(ctx context.Context, id int, workerID string) error {
	return r.db.Task.
		UpdateOneID(id).
		Where(
			task.StatusEQ(task.StatusRUNNING),
			task.WorkerIDEQ(workerID),
		).
		SetStatus(task.StatusSUCCESS).
		SetFinishedAt(time.Now()).
		Exec(ctx)
}

// SetFailed 将持有中的任务终结为 FAILED，并记录错误与完成时间；同样受 workerID 栅栏保护。
func (r *Task) SetFailed(ctx context.Context, id int, workerID string, errMsg string) error {
	return r.db.Task.
		UpdateOneID(id).
		Where(
			task.StatusEQ(task.StatusRUNNING),
			task.WorkerIDEQ(workerID),
		).
		SetStatus(task.StatusFAILED).
		SetErrorMessage(errMsg).
		SetFinishedAt(time.Now()).
		Exec(ctx)
}

// Retry 将未超限的失败任务重新置为 PENDING，并推迟到 nextRunAt 再次执行；
// 同时归还租约（清空 worker_id/heartbeat_at）。
func (r *Task) Retry(ctx context.Context, id int, workerID string, errMsg string, nextRunAt time.Time) error {
	return r.db.Task.
		UpdateOneID(id).
		Where(
			task.StatusEQ(task.StatusRUNNING),
			task.WorkerIDEQ(workerID),
		).
		SetStatus(task.StatusPENDING).
		SetNextRunAt(nextRunAt).
		SetErrorMessage(errMsg).
		ClearWorkerID().
		ClearHeartbeatAt().
		Exec(ctx)
}

// RecoverStuck 回收心跳超时（持有者疑似失效）且仍为 RUNNING 的任务；
// 重试次数已耗尽的直接终结为 FAILED。返回处理的条数。
//
// 判定与更新都以 heartbeat_at 为准：若持有者期间刷新了心跳，条件更新会匹配 0 行而跳过，
// 因此不会误回收仍存活的 worker（对比之前基于 started_at 的粗粒度猜测）。
func (r *Task) RecoverStuck(ctx context.Context, timeout time.Duration, limit int) (int, error) {
	deadline := time.Now().Add(-timeout)
	stale := task.Or(
		task.HeartbeatAtIsNil(),
		task.HeartbeatAtLTE(deadline),
	)

	tasks, err := r.db.Task.
		Query().
		Where(
			task.StatusEQ(task.StatusRUNNING),
			stale,
		).
		Order(ent.Asc(task.FieldHeartbeatAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return 0, err
	}

	recovered := 0
	for _, t := range tasks {
		now := time.Now()

		var execErr error
		if t.Attempts >= t.MaxAttempts {
			execErr = r.db.Task.
				UpdateOneID(t.ID).
				Where(task.StatusEQ(task.StatusRUNNING), stale).
				SetStatus(task.StatusFAILED).
				SetErrorMessage("recovered: max attempts reached").
				SetFinishedAt(now).
				ClearWorkerID().
				Exec(ctx)
		} else {
			execErr = r.db.Task.
				UpdateOneID(t.ID).
				Where(task.StatusEQ(task.StatusRUNNING), stale).
				SetStatus(task.StatusPENDING).
				SetNextRunAt(now).
				SetErrorMessage("recovered: heartbeat timeout").
				ClearWorkerID().
				ClearHeartbeatAt().
				Exec(ctx)
		}
		if execErr != nil {
			if ent.IsNotFound(execErr) {
				// 期间持有者已刷新心跳或任务已被处理，跳过。
				continue
			}
			return recovered, execErr
		}
		recovered++
	}

	return recovered, nil
}

// controlResource 在事务内完成控制类操作（start/stop/restart/remove）的入队：
//  1. 读取并校验资源归属（非本人返回 NotFound）；
//  2. 按事件做状态机迁移（稳定态→中间态），条件更新保证并发下只有一个请求成功；
//  3. 创建对应的 PENDING 任务。
//
// 状态非法返回 apperr.ErrInvalidState，并发冲突返回 apperr.ErrConflict。
func (r *Task) controlResource(
	ctx context.Context,
	userID,
	resourceID int,
	taskType task.Type,
	payload map[string]any,
) (*ent.Task, error) {
	var event transition.ResourceEvent
	switch taskType {
	case task.TypeSTART_RESOURCE:
		event = transition.ResourceStart
	case task.TypeSTOP_RESOURCE:
		event = transition.ResourceStop
	case task.TypeRESTART_RESOURCE:
		event = transition.ResourceRestart
	case task.TypeREMOVE_RESOURCE:
		event = transition.ResourceRemove
	default:
		return nil, fmt.Errorf("invalid task type: %s", taskType)
	}

	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	rsc, err := tx.Resource.
		Query().
		Where(
			resource.ID(resourceID),
			resource.UserID(userID),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	nextStatus, err := transition.NextResourceStatus(rsc.Status, event)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", apperr.ErrInvalidState, err)
	}

	// 条件更新：仅当状态仍为读取到的值时才迁移。0 行代表期间被并发操作改变。
	err = tx.Resource.
		UpdateOneID(resourceID).
		Where(
			resource.StatusEQ(rsc.Status),
		).
		SetStatus(nextStatus).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperr.ErrConflict
		}
		return nil, err
	}

	t, err := tx.Task.
		Create().
		SetUserID(userID).
		SetResourceID(resourceID).
		SetType(taskType).
		SetStatus(task.StatusPENDING).
		SetPayload(payload).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return t, nil
}

// RemoveResource 入队删除资源任务（STOPPED/FAILED → REMOVING）。
func (r *Task) RemoveResource(
	ctx context.Context,
	userID,
	resourceID int,
	payload map[string]any,
) (*ent.Task, error) {
	return r.controlResource(
		ctx,
		userID,
		resourceID,
		task.TypeREMOVE_RESOURCE,
		payload,
	)
}

// StartResource 入队启动资源任务（STOPPED → STARTING）。
func (r *Task) StartResource(
	ctx context.Context,
	userID,
	resourceID int,
	payload map[string]any,
) (*ent.Task, error) {
	return r.controlResource(
		ctx,
		userID,
		resourceID,
		task.TypeSTART_RESOURCE,
		payload,
	)
}

// StopResource 入队停止资源任务（RUNNING → STOPPING）。
func (r *Task) StopResource(
	ctx context.Context,
	userID,
	resourceID int,
	payload map[string]any,
) (*ent.Task, error) {
	return r.controlResource(
		ctx,
		userID,
		resourceID,
		task.TypeSTOP_RESOURCE,
		payload,
	)
}

// RestartResource 入队重启资源任务（RUNNING → RESTARTING）。
func (r *Task) RestartResource(
	ctx context.Context,
	userID,
	resourceID int,
	payload map[string]any,
) (*ent.Task, error) {
	return r.controlResource(
		ctx,
		userID,
		resourceID,
		task.TypeRESTART_RESOURCE,
		payload,
	)
}
