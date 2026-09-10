package worker

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/resource"
	"devcloud/ent/task"
	"devcloud/provider"
	"devcloud/repo"
	"devcloud/transition"
	"fmt"
	"log/slog"
)

type Executor struct {
	repo     *repo.Repo
	registry *provider.Registry
}

func newExecutor(
	repo *repo.Repo,
	registry *provider.Registry,
) *Executor {
	return &Executor{
		repo:     repo,
		registry: registry,
	}
}

// execute 根据任务类型分发到对应的资源操作。
// 注意：进入中间态（如 RUNNING→STOPPING）已由入队时的 controlResource 完成，
// 这里只负责「调用 provider + 推到终态/回滚」。
func (e *Executor) execute(ctx context.Context, t *ent.Task) error {
	r, err := e.repo.Resource.Get(ctx, t.ResourceID)
	if err != nil {
		return err
	}

	switch t.Type {
	case task.TypeCREATE_RESOURCE:
		return e.createResource(ctx, r)

	case task.TypeSTART_RESOURCE:
		return e.startResource(ctx, r)

	case task.TypeSTOP_RESOURCE:
		return e.stopResource(ctx, r)

	case task.TypeRESTART_RESOURCE:
		return e.restartResource(ctx, r)

	case task.TypeREMOVE_RESOURCE:
		return e.removeResource(ctx, r)

	default:
		return fmt.Errorf("unsupported task type: %s", t.Type)
	}
}

// runOperation 执行停止/重启/删除这类「有中间态」的操作：
//   - 按当前中间态计算成功/失败后的目标状态（非法则直接返回，不调用 provider）；
//   - 调用 provider；成功推进到终态，失败则回滚到操作前的稳定态。
func (e *Executor) runOperation(
	ctx context.Context,
	r *ent.Resource,
	successEvent transition.ResourceEvent,
	failedEvent transition.ResourceEvent,
	op func(ctx context.Context, runtimeID string) error,
) error {
	successStatus, err := transition.NextResourceStatus(r.Status, successEvent)
	if err != nil {
		return err
	}

	if err := op(ctx, r.RuntimeID); err != nil {
		failedStatus, terr := transition.NextResourceStatus(r.Status, failedEvent)
		if terr != nil {
			failedStatus = r.Status
		}
		_ = e.repo.Resource.SetStatus(ctx, r.ID, failedStatus)
		return err
	}

	return e.repo.Resource.SetStatus(ctx, r.ID, successStatus)
}

// createResource 创建容器。仅允许 CREATING 状态；成功写 runtime_id 并置 STOPPED，
// 失败置 FAILED。若成功写库时状态已被并发改变，则放弃写入以避免重复创建。
func (e *Executor) createResource(ctx context.Context, r *ent.Resource) error {
	if r.Status != resource.StatusCREATING {
		return fmt.Errorf("resource %d is not in CREATING status: %s", r.ID, r.Status)
	}

	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	runtimeID, err := p.Create(ctx, provider.ResourceSpec{
		Name:       r.Name,
		Image:      r.Image,
		Config:     r.Config,
		Credential: provider.Credential{}, // TODO：r.Credential 解密后填入
	})
	if err != nil {
		_ = e.repo.Resource.SetCreateFailed(ctx, r.ID)
		return err
	}

	if err := e.repo.Resource.SetCreateSuccess(ctx, r.ID, runtimeID); err != nil {
		slog.Warn("set create success", "resource_id", r.ID, "error", err)
		// 状态已被其他操作推进：不再重试，避免重复创建容器。
		if ent.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

// startResource 启动容器：STARTING → RUNNING，失败回退 STOPPED。
func (e *Executor) startResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	return e.runOperation(
		ctx,
		r,
		transition.ResourceStartSuccess,
		transition.ResourceStartFailed,
		p.Start,
	)
}

// restartResource 重启容器：RESTARTING → RUNNING，失败回退 RUNNING。
func (e *Executor) restartResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	return e.runOperation(
		ctx,
		r,
		transition.ResourceRestartSuccess,
		transition.ResourceRestartFailed,
		p.Restart,
	)
}

// stopResource 停止容器：STOPPING → STOPPED，失败回退 RUNNING。
func (e *Executor) stopResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	return e.runOperation(
		ctx,
		r,
		transition.ResourceStopSuccess,
		transition.ResourceStopFailed,
		p.Stop,
	)
}

// removeResource 删除容器：REMOVING → REMOVED，失败回退 STOPPED。
// 未生成容器的资源（创建失败，无 runtime_id）直接标记删除完成，不调用 provider。
func (e *Executor) removeResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	if r.RuntimeID == "" {
		removedStatus, err := transition.NextResourceStatus(r.Status, transition.ResourceRemoveSuccess)
		if err != nil {
			return err
		}
		return e.repo.Resource.SetStatus(ctx, r.ID, removedStatus)
	}

	return e.runOperation(
		ctx,
		r,
		transition.ResourceRemoveSuccess,
		transition.ResourceRemoveFailed,
		p.Remove,
	)
}
