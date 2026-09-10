# DevCloud

> 轻量级云资源管理与异步编排平台 —— 位于用户与基础设施（Docker）之间的资源/任务编排层。

DevCloud 通过统一的 Resource 模型管理 Docker 资源，基于 PostgreSQL + Worker Pool 实现可靠的**异步任务执行、状态机、幂等、重试、超时与故障恢复**。

- 设计/需求文档：[`docs/design.md`](./docs/design.md)
- 重难点与解决过程：[`docs/challenges.md`](./docs/challenges.md)
- 前端接口约定：[`docs/frontend-spec.md`](./docs/frontend-spec.md)

---

## 功能特性

- 用户登录/登出、JWT 访问令牌 + Refresh Token 轮换（bcrypt 存储密码、HttpOnly Cookie）
- Project 化的资源管理（当前实现为用户维度的资源归属与校验）
- Docker 资源生命周期：创建 / 启动 / 停止 / 重启 / 移除，全部异步执行
- **资源状态机**：非法转换拒绝、入队即进入中间态、条件更新防并发覆盖
- **任务系统**：PostgreSQL 持久化队列、`FOR UPDATE SKIP LOCKED` 抢占、指数退避重试
- **可靠性**：任务租约（lease）+ 心跳（heartbeat）+ 栅栏（fencing）、worker 崩溃超时恢复
- 就绪探针 `/ready`（PostgreSQL / Docker），请求体大小限制，统一错误码映射

## 技术栈

| 领域 | 选型 |
| --- | --- |
| 语言 | Go 1.26 |
| Web | Echo v5、go-playground/validator、ByteDance sonic |
| 数据库 | PostgreSQL（ent ORM + lib/pq） |
| 基础设施 | Docker（moby/moby client） |
| 认证 | golang-jwt/jwt v5、golang.org/x/crypto/bcrypt |
| 配置 | YAML（go.yaml.in/yaml/v4）+ 环境变量覆盖 |

> 说明：早期设计中的 Redis（分布式锁/缓存/限流）目前未接入，已从代码与依赖中移除，待第二阶段按需引入。

## 架构

```text
                       HTTP (Echo)
                           │
             ┌─────────────┴─────────────┐
             ▼                           ▼
      handler / middleware          health / ready
             │
             ▼
          service
             │
             ▼
      repo (ent + tx, 条件更新/抢占)
             │
      ┌──────┴───────┐
      ▼              ▼
  PostgreSQL      Worker Pool ──► Provider (Docker)
   (tasks/resource)   │
                      ├─ ClaimPendingTask (SKIP LOCKED + 租约)
                      ├─ Heartbeat / Recovery
                      └─ Executor → Docker 容器操作
```

## 目录结构

```text
.
├── main.go                 # 组装依赖、启动 HTTP 服务与 Worker Pool
├── apperr/                 # 跨层共享的业务错误（InvalidState/Conflict/NoPendingTask）
├── config/                 # 配置加载与环境变量覆盖
├── ent/                    # ent 生成代码
│   └── schema/             # 数据模型定义（resource/task/user/refreshtoken）
├── provider/               # 基础设施抽象
│   └── docker/             # Docker Provider 实现
├── repo/repos/             # 数据访问层（事务、条件更新、任务抢占/恢复）
├── storage/                # PostgreSQL 连接
├── transition/             # 资源状态机（迁移表）
├── utils/                  # 哈希/随机串等工具
├── web/                    # HTTP 层：server/handler/service/middleware/dto/errmsg
└── worker/                 # 任务执行：pool / task_manager / executor
```

## 快速开始

### 1. 准备配置

复制样例并按需修改（`config.yaml` 已被 `.gitignore` 忽略）：

```bash
cp config.yaml.sample config.yaml
```

关键配置项：

```yaml
server_addr: :8080
postgres: host=127.0.0.1 port=5432 user=postgres dbname=dev-cloud password=123456
jwt_secret: "change-me-to-a-random-secret"
token_salt: "change-me-to-a-random-salt"
access_ttl: 900        # access token 秒
refresh_ttl: 604800    # refresh token 秒
cookie_secure: false
worker_pool: 4         # 省略时默认 CPU*2
docker:
  host: unix:///var/run/docker.sock
  tls:
    enabled: false
    ca: ""
    cert: ""
    key: ""
```

敏感项支持用环境变量覆盖（优先级高于文件）：

```bash
export POSTGRES_DSN="host=... user=... password=... dbname=..."
export JWT_SECRET="..."
export TOKEN_SALT="..."
```

### 2. 运行

开发模式（`ENV=development` 会自动创建/迁移数据库表结构）：

```bash
ENV=development go run . -config ./config.yaml
```

其他方式（`Makefile` 已提供常用目标）：

```bash
make run     # 本地开发运行（ENV=development，自动建表）
make build   # 编译到 build/devcloud
make cross   # 交叉编译 linux/amd64、darwin/amd64、darwin/arm64
make vet     # go vet ./...
make fmt     # gofmt -w .
make tidy    # go mod tidy
make test    # go test ./...
make clean   # 清理 build/

# 或不通过 Makefile：
go build ./...
go vet ./...
```

### 3. 健康检查

```bash
curl localhost:8080/health   # 存活：OK
curl localhost:8080/ready    # 就绪：依赖可用性（postgres/docker），异常返回 503
```

## API 一览

| 方法 | 路径 | 说明 | 鉴权 |
| --- | --- | --- | --- |
| GET | `/health` | 存活探针 | 否 |
| GET | `/ready` | 就绪探针 | 否 |
| POST | `/api/login` | 登录（限流） | 否 |
| POST | `/api/refresh-token` | 刷新令牌（Cookie） | 否 |
| POST | `/api/logout` | 登出 | 是 |
| POST | `/api/user/password/update` | 修改密码 | 是 |
| GET | `/api/user` | 当前用户信息 | 是 |
| GET | `/api/resources` | 资源列表（分页） | 是 |
| GET | `/api/resource/:id` | 资源详情 | 是 |
| POST | `/api/resource` | 创建资源（异步） | 是 |
| POST | `/api/resource/:id/start` | 启动（异步） | 是 |
| POST | `/api/resource/:id/stop` | 停止（异步） | 是 |
| POST | `/api/resource/:id/restart` | 重启（异步） | 是 |
| POST | `/api/resource/:id/remove` | 移除（异步） | 是 |
| GET | `/api/tasks/:id` | 任务列表（分页） | 是 |
| GET | `/api/task/:id` | 任务详情 | 是 |

除登录/刷新/健康检查外均需 `Authorization: Bearer <access_token>`。

## 资源状态机

采用「两段式」：**入队**时由 API 把资源推进到中间态（事务内 + 条件更新），**执行**结束时由 worker 推到终态或回滚。迁移表见 [`transition/resource.go`](./transition/resource.go)。

```text
CREATING ──create_success──► STOPPED ──start──► STARTING ──start_success──► RUNNING
    │                                          │
    ├──create_failed──► FAILED ◄───────────────┘（start_failed 回退 STOPPED）
                          │
                          └──remove──► REMOVING ──remove_success──► REMOVED

RUNNING ──restart──► RESTARTING ──restart_success──► RUNNING（restart_failed 回 RUNNING）
RUNNING ──stop────► STOPPING  ──stop_success──────► STOPPED（stop_failed 回 RUNNING）
STOPPED/FAILED ──remove──► REMOVING ──remove_success──► REMOVED（remove_failed 回 STOPPED）
```

## 任务与 Worker

任务状态：`PENDING → RUNNING → SUCCESS / FAILED`，失败未超限时回到 `PENDING`（指数退避）。

可靠性机制：

- **抢占**：`SELECT ... WHERE status='PENDING' AND next_run_at<=now ORDER BY next_run_at FOR UPDATE SKIP LOCKED`，多 worker 不重复领取。
- **租约 + 心跳**：领取时写入 `worker_id / started_at / heartbeat_at`；执行期间 worker 每 5s 刷新心跳。
- **栅栏（fencing）**：`SetSuccess/SetFailed/Retry` 带 `WHERE status='RUNNING' AND worker_id=?`，旧持有者无法覆盖已被回收的任务。
- **超时恢复**：恢复协程每 10s 扫描 `heartbeat_at` 超过 30s 未更新的 `RUNNING` 任务，重新入队或终结。
- **重试**：失败未超 `max_attempts` 时按 `1s·2^(n-1)`（上限 30s，带抖动）重排。

核心实现：[`repo/repos/task.go`](./repo/repos/task.go)、[`worker/pool.go`](./worker/pool.go)、[`worker/executor.go`](./worker/executor.go)。

## 后续规划

- RBAC / Tenant / API Key / Audit Log
- Credential 生成、加密存储与密码重置
- 资源规格（端口、环境变量、CPU/内存）在 Docker Provider 中落地
- 引入 Redis 做分布式锁/缓存/限流（按需）
- 补充状态机与任务系统的自动化测试
