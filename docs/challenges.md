# DevCloud 重难点

用 PostgreSQL 做持久化队列，用 Worker Pool 可靠地执行异步任务，并保证**状态一致、幂等、可重试、可恢复**。

```text
并发抢占 ──► 幂等入队 ──► 状态机（唯一事实来源）
    ▲                              │
    └──── 可靠性（租约/心跳/栅栏/恢复） ◄── 重试策略
```

---

## 1. 任务系统的并发与幂等

### 1.1 并发抢占：多个 Worker 不重复领取

**问题**：多个 worker 同时轮询任务表，如何保证同一条任务只被一个 worker 领走？

**方案**：`SELECT ... FOR UPDATE SKIP LOCKED`。在事务内锁定并跳过已被其他事务锁定的行：

```sql
SELECT id, resource_id, type
FROM task
WHERE status = 'PENDING' AND next_run_at <= now()
ORDER BY next_run_at
LIMIT 1
FOR UPDATE SKIP LOCKED;
```

拿到行后在同一事务内条件更新 `PENDING → RUNNING` 并提交。`SKIP LOCKED` 让并发 worker 各取一条、互不阻塞；`ORDER BY next_run_at` 保证先到期的先执行。

### 1.2 幂等（一）：同一操作不重复入队

**问题**：用户连点移除会产生多条移除任务；`Stop` 与 `Remove` 并发提交时，两个请求都读取到 `RUNNING`，各自建任务，最终状态互相覆盖。

**根因**：入队时只看状态、状态更新是无条件 `UPDATE`，存在 TOCTOU 窗口（Time-Of-Check to Time-Of-Use）——「检查」与「写入」之间状态可能已被改变。

**方案：入队即中间态 + 条件更新（CAS）**。把「稳定态 → 中间态」的迁移放进**入队事务**里（`controlResource`）：

1. 事务内按 `id + user_id` 读取资源（顺带做写路径的归属校验）；
2. `NextResourceStatus` 校验迁移合法性，非法 → `apperr.ErrInvalidState`；
3. **条件更新**：
   ```sql
   UPDATE resource SET status = :next
   WHERE id = :id AND status = :current;
   ```
   - 影响 0 行 → 期间被并发请求改过 → `apperr.ErrConflict`；
   - 影响 1 行 → 本次请求赢得迁移，继续；
4. 同事务创建 PENDING 任务，提交。

**效果**：并发下只有一个请求能把资源推进到中间态，天然互斥；连点/并发直接得到 409。无需应用层锁。

**代码**：`repo/repos/task.go` `controlResource`；错误语义 `apperr` + `web/err_handler.go`（409）。

### 1.3 幂等（二）：任务不重复产生副作用

1.2 只保证「同一操作不重复入队」，但执行阶段仍是 **at-least-once**：重试、worker 崩溃恢复、以及「任务被回收后旧 worker 才跑完」，都可能让同一个 Docker 操作被执行或收尾多次。不追求 exactly-once，而是让重复执行安全：

- **状态机 + 条件写**：每次执行基于数据库当前状态计算迁移，状态已变则迁移非法，重复执行不会重复推进状态。
- **栅栏（fencing）**：终态写入带 `WHERE status='RUNNING' AND worker_id=?`，任务被回收转交后旧 worker 匹配 0 行，无法覆盖（`repo/repos/task.go` 的 `SetSuccess/SetFailed/Retry`）。
- **资源级互斥**：入队即中间态（1.2），同一资源同一时刻只有一条在途操作，不同操作不会交错。

Provider 层面，Docker 的 start/stop/remove 重复调用基本无副作用；**例外是 Create**——用「仅允许 `CREATING`」守卫 + `SetCreateSuccess` 条件写（只有一次能成功）+ 同名容器冲突三重保护，避免重复创建（`worker/executor.go`、`repo/repos/resource.go`）。

> 结论：接口层前置检查挡大部分请求，最终正确性靠**数据库条件写（CAS / `FOR UPDATE`）与持有者栅栏**。

---

## 2. 状态机设计

### 2.1 单一事实来源

所有资源状态迁移集中在一张表 `transition/resource.go`，键为 `(当前状态, 事件)`、值为目标状态；表中不存在即非法迁移。业务代码不允许「随意 SetStatus」，必须先过 `NextResourceStatus`。

### 2.2 两段式事件（核心设计）

把事件分成两类，明确「谁在什么时候推进状态」：

- **进入事件**（`create/start/stop/restart/remove`）：由 **API 入队时**触发，稳定态 → 中间态；
- **完成事件**（`*_success / *_failed`）：由 **worker 执行结束后**触发，中间态 → 终态/回滚。

```text
STOPPED --start--> STARTING --start_success--> RUNNING
                            --start_failed---> STOPPED
```

好处：
- 入队后资源立刻处于「操作中」，别的请求看到中间态会被 CAS 拒绝 → 与第 1 节配合实现互斥；
- worker 只做「中间态 → 终态」，语义单纯，失败有明确的回滚目标。

**踩过的坑**：早期 executor 先写中间态、再用**读进来的旧 `r.Status`** 触发完成迁移，导致完成迁移永远非法、资源卡在中间态。修正为：迁移始终基于「数据库当前值」，且两段迁移职责分离（收敛到 `runOperation`）。

### 2.3 失败与清理路径

- 创建失败进入 `FAILED`；同时开放 `FAILED → REMOVING`，保证失败资源能被移除（不会永久堆积）。
- 移除未真正生成容器的资源（创建失败、无 `runtime_id`）时直接标记 `REMOVED`，不调用 Docker。
- 失败回滚语义：`start_failed → STOPPED`、`stop_failed → RUNNING`、`restart_failed → RUNNING`、`remove_failed → STOPPED`。

### 2.4 create 的特例

`create` 没有「前置稳定态」，初始即 `CREATING`。落地时：
- `SetCreateSuccess` 仅在 `status=CREATING` 时写 `runtime_id` 并置 `STOPPED`；
- `SetCreateFailed` 同理只允许从 `CREATING` 置 `FAILED`。

目标为 `STOPPED` 的原因：Docker `Create` 只创建容器、不启动，与真实状态一致。

---

## 3. 重试策略

### 3.1 何时重试

只重试**暂时性失败**（Docker/网络超时、临时不可用）；参数错误、非法状态等不应重试。当前按「执行报错即重试，直到次数耗尽」处理，后续可按错误类型细分。

### 3.2 退避与次数

- 字段：`attempts / max_attempts / next_run_at`；领取任务时 `attempts+1` 并写 `started_at`。
- 失败时 `TaskManager.fail` 判断：`attempts < max_attempts` → `Retry`（回到 `PENDING`，`next_run_at = now + backoff`，记录错误）；否则 → `FAILED`（写 `finished_at / error_message`）。
- 退避：`1s · 2^(n-1)`，上限 30s，带 50%~100% **抖动**，避免大量任务同时重试形成尖峰。

### 3.3 「重试」与「恢复」的区别

- **重试**：任务执行了但失败 → 由业务失败路径按退避重排；
- **恢复**：任务可能还在执行，只是 worker 失联 → 由心跳超时判定后重新入队（见第 4 节）。

两者分别用 `next_run_at` 和 `heartbeat_at` 表达，互不干扰。

---

## 4. 任务可靠性（故障恢复）

### 4.1 问题：如何判断 worker「死了」

最初的朴素做法是「`started_at` 超过固定时长即视为卡死」。它有两个坑：

1. **长任务被误杀**：合法地跑了很久的任务会被回收，可能被重复执行；
2. **旧 worker 覆盖**：任务被回收转交新 worker 后，旧 worker 复活并调用 `SetSuccess/SetFailed`，无条件更新会把新状态覆盖掉。

根因是缺少「谁在持有任务」以及「它还活着吗」的信息。

### 4.2 方案：租约 + 心跳 + 栅栏

- **租约（lease）**：领取时写入 `worker_id / started_at / heartbeat_at`，标识持有者与最后活跃时间。
- **心跳（heartbeat）**：执行期间 worker 每 5s 刷新：
  ```sql
  UPDATE task SET heartbeat_at = now()
  WHERE id = ? AND status = 'RUNNING' AND worker_id = ?;
  ```
- **栅栏（fencing）**：`SetSuccess / SetFailed / Retry` 都带 `WHERE status='RUNNING' AND worker_id=?`，只有当前持有者能收尾；被回收后旧 worker 的写入匹配 0 行而失效。
- **超时恢复**：恢复协程定期扫描 `heartbeat_at` 超时（`heartbeat_at < now - 30s` 或为空）的 `RUNNING` 任务，重新入队（未超次数）或终结（超次数）。**更新时再次校验心跳仍超时**——若持有者期间刷新过心跳则跳过，不会误回收活着的 worker。

### 4.3 参数与不变式

- 参数：心跳间隔 5s / 心跳超时 30s / 扫描间隔 10s / 单批上限 100。
- 三个必须遵守的不变式：
  1. **领取必须在事务内建立租约**（RUNNING + worker_id + heartbeat_at 同时写入）；
  2. **终态写入必须带 `worker_id` 条件**；
  3. **回收更新必须复查心跳仍超时**。

**代码**：`repo/repos/task.go`（`ClaimPendingTask / Heartbeat / SetSuccess / SetFailed / Retry / RecoverStuck`）、`worker/pool.go`（`process` 起心跳协程、`recoverLoop` 回收）。

### 4.4 Worker 轮询与优雅关闭

- 有任务 → 立即继续；无任务（`apperr.ErrNoPendingTask`）→ 视为空闲，按 200ms→2s 指数退避 + 抖动轮询，避免多 worker 同时空转打库。
- 关闭时若执行因 `context.Canceled/DeadlineExceeded` 中断，**不算业务失败、不重试**，保留租约交由心跳超时恢复，避免把正在做的任务误标失败。

---

## 5. 总结设计准则

1. **状态迁移只有一份事实来源**（迁移表），非法迁移直接拒绝、不写库。
2. **入队即中间态**，用数据库当前状态决定「能否操作」，用**条件更新（CAS）**保证并发唯一赢家。
3. **终态写入带持有者栅栏**，靠**心跳**而非固定超时判断存活。
4. **失败能重试、崩溃能恢复、重复要幂等**——分别对应 backoff、心跳租约、CAS/幂等。
