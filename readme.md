# DevCloud —— 云资源管理与异步编排平台

> 项目类型：企业级开发者资源管理平台
> 项目定位：面向企业开发人员的轻量级资源管理与编排平台
> 核心技术：Go + PostgreSQL + Redis + Docker
> 当前阶段：V1 核心功能设计
> 文档版本：v1.0

---

# 1. 项目概述

## 1.1 项目背景

在企业开发环境中，应用、数据库、Redis 等基础资源通常运行在 Docker 等容器环境中。

在资源数量较少时，开发人员可以直接通过 Docker CLI 或服务器命令管理资源，例如：

```bash
docker run
docker stop
docker start
docker rm
```

随着项目和开发人员数量增加，直接操作服务器会产生以下问题：

* 资源缺乏统一管理
* 不同用户权限难以控制
* 资源状态无法统一维护
* 长时间操作阻塞 HTTP 请求
* 操作失败后缺乏自动重试机制
* Worker 或服务异常后任务无法恢复
* 用户重复提交可能导致资源重复创建
* 缺乏统一的操作记录

因此，本项目计划实现一个轻量级的资源管理与异步编排平台。

---

# 2. 项目定位

DevCloud 不负责重新实现 Docker 或 Kubernetes 等底层基础设施能力。

系统定位为：

> **位于用户和基础设施之间的资源管理与编排层。**

整体结构：

```text
用户
 │
 ▼
DevCloud
 │
 ├── 用户管理
 ├── Project
 ├── Resource
 ├── Task
 ├── Worker
 ├── Credential
 └── 权限控制
 │
 ▼
Docker
```

用户不需要直接操作服务器，而是通过 DevCloud 的 Web 控制台或 API 管理资源。

---

# 3. 项目目标

## 3.1 核心目标

实现一个完整的资源生命周期管理系统：

```text
创建资源
   ↓
异步执行
   ↓
资源创建
   ↓
运行
   ↓
停止 / 启动 / 重启
   ↓
删除
```

同时保证资源操作具备：

* 异步执行
* 状态管理
* 失败重试
* 幂等性
* 超时处理
* Worker 故障恢复
* 并发控制

---

## 3.2 技术目标

重点实践以下后端能力：

### Go

* Goroutine
* Channel
* Context
* Worker Pool
* 并发控制
* Graceful Shutdown

### 数据库

* PostgreSQL
* 事务
* 索引
* 行级锁
* `FOR UPDATE SKIP LOCKED`
* 数据一致性

### Redis

* 分布式锁
* 缓存
* 限流

### Docker

* Docker API
* Container 生命周期管理

### 系统设计

* 状态机
* 异步任务
* 幂等
* Retry
* Backoff
* Timeout
* 故障恢复

---

# 4. 用户角色

V1 暂时定义三种角色。

## 4.1 Owner

项目所有者。

权限：

* 创建资源
* 删除资源
* 启动/停止资源
* 查看资源
* 查看任务
* 管理项目成员

---

## 4.2 Developer

开发人员。

权限：

* 创建资源
* 启动/停止资源
* 查看资源
* 查看任务

不能：

* 删除项目
* 管理项目成员

---

## 4.3 Viewer

只读用户。

权限：

* 查看项目
* 查看资源
* 查看任务

不能：

* 创建资源
* 操作资源
* 删除资源

---

# 5. 核心概念

系统主要包含以下实体：

```text
User
 │
 ▼
Project
 │
 ▼
Resource
 │
 ├──── Credential
 │
 └──── Task
          │
          ▼
       Worker
          │
          ▼
        Docker
```

---

# 6. User 用户

用户是系统的基础身份实体。

## 6.1 功能

* 用户注册
* 用户登录
* 获取当前用户信息
* 修改密码

---

# 7. Project 项目

Project 用于组织资源。

例如：

```text
ecommerce
├── redis
├── mysql
└── web
```

## 7.1 功能

* 创建 Project
* 查询 Project
* 修改 Project
* 删除 Project
* 查看 Project 下的资源
* 管理 Project 成员

---

# 8. Resource 资源

Resource 是系统最核心的业务对象。

V1 暂时只支持：

> Docker App

---

## 8.1 创建资源

用户填写：

```text
资源名称
Docker Image
CPU
Memory
Port
Environment Variables
```

例如：

```text
Name:
my-nginx

Image:
nginx:1.27

CPU:
1 Core

Memory:
512 MB

Port:
8080 → 80

Environment:
APP_ENV=production
```

---

## 8.2 创建流程

```text
用户
 │
 ▼
POST /resources
 │
 ▼
权限检查
 │
 ▼
参数校验
 │
 ▼
创建 Resource
status = CREATING
 │
 ▼
创建 Task
status = PENDING
 │
 ▼
返回用户
```

API 不等待 Docker 创建完成。

---

# 9. Resource 状态机

Resource 使用状态机管理生命周期。

```text
                 ┌──────────────┐
                 │   CREATING   │
                 └──────┬───────┘
                        │
                  创建成功/失败
                  ┌─────┴─────┐
                  ▼           ▼
             RUNNING       FAILED
                │
        ┌───────┼────────┐
        ▼       ▼        ▼
     STOPPING RESTARTING DELETING
        │       │        │
        ▼       ▼        ▼
     STOPPED  RUNNING   DELETED
        │
        └──── STARTING ───→ RUNNING
```

系统必须限制非法状态转换。

例如：

```text
CREATING → STOPPING
```

属于非法操作。

---

# 10. Resource 操作

V1 支持：

## Create

```text
创建 Docker Container
```

## Start

```text
STOPPED
   ↓
STARTING
   ↓
RUNNING
```

## Stop

```text
RUNNING
   ↓
STOPPING
   ↓
STOPPED
```

## Restart

```text
RUNNING
   ↓
RESTARTING
   ↓
RUNNING
```

## Delete

```text
RUNNING
   ↓
DELETING
   ↓
DELETED
```

所有操作均通过 Task 异步执行。

---

# 11. Task 任务系统

Task 是项目的核心技术模块。

所有可能耗时的资源操作都转换为 Task：

```text
CREATE_RESOURCE
START_RESOURCE
STOP_RESOURCE
RESTART_RESOURCE
DELETE_RESOURCE
```

---

# 12. Task 状态

```text
PENDING
   ↓
RUNNING
   ↓
SUCCESS
```

失败：

```text
RUNNING
   ↓
FAILED
```

需要重试：

```text
FAILED
   ↓
RETRY
   ↓
PENDING
```

---

# 13. Task 执行流程

```text
用户请求
   │
   ▼
创建 Task
   │
   ▼
PENDING
   │
   ▼
Worker 获取 Task
   │
   ▼
RUNNING
   │
   ▼
执行 Docker 操作
   │
   ├──── 成功 ────→ SUCCESS
   │
   └──── 失败
           │
           ▼
         Retry
           │
      ┌────┴────┐
      ▼         ▼
   可重试      达到最大次数
      │         │
      ▼         ▼
   PENDING    FAILED
```

---

# 14. Worker Pool

系统内部维护 Worker Pool。

例如：

```text
Task Queue
    │
    ├──── Worker 1
    ├──── Worker 2
    ├──── Worker 3
    ├──── Worker 4
    └──── Worker 5
```

V1 固定：

```text
Worker = 5
```

Worker 负责：

1. 获取 Task
2. 抢占 Task
3. 执行 Task
4. 更新 Task 状态
5. 处理错误
6. 进行重试
7. 记录执行信息

---

# 15. Task 获取机制

V1 不使用 Kafka、RabbitMQ 等 MQ。

直接使用 PostgreSQL 作为任务持久化和任务协调基础。

Worker 从 `tasks` 表获取任务。

核心思路：

```sql
SELECT *
FROM tasks
WHERE status = 'PENDING'
ORDER BY created_at
LIMIT 1
FOR UPDATE SKIP LOCKED;
```

多个 Worker 可以同时获取不同 Task。

例如：

```text
             PostgreSQL
                  │
                tasks
                  │
        ┌─────────┼─────────┐
        ▼         ▼         ▼
     Worker1   Worker2   Worker3
       Task1      Task2      Task3
```

避免多个 Worker 获取同一个任务。

---

# 16. Task 重试

部分错误可以重试。

例如：

```text
Docker API timeout
Network timeout
Temporary unavailable
```

V1 默认：

```text
最大重试次数：3
```

使用指数退避：

```text
第一次失败
 ↓
等待 1 秒

第二次失败
 ↓
等待 2 秒

第三次失败
 ↓
FAILED
```

---

# 17. Task 幂等

系统必须避免同一个操作重复执行。

例如用户连续点击：

```text
Delete
Delete
Delete
```

不能产生：

```text
Task 1
Task 2
Task 3
```

然后三个 Worker 同时删除资源。

因此需要设计：

```text
Idempotency Key
```

或者基于：

```text
Resource + Operation + State
```

判断是否存在相同的执行任务。

---

# 18. Worker 故障恢复

Worker 执行任务过程中可能崩溃。

例如：

```text
Task
 ↓
RUNNING
 ↓
Worker Crash
```

如果没有恢复机制：

```text
Task 永远 RUNNING
```

因此 Task 需要记录：

```text
started_at
heartbeat_at
worker_id
```

Worker 定期发送 heartbeat。

如果：

```text
当前时间 - heartbeat_at > timeout
```

系统认为 Worker 已经失效。

然后：

```text
RUNNING
   ↓
TIMEOUT
   ↓
RETRY
   ↓
PENDING
```

重新交给其他 Worker 执行。

---

# 19. Credential 凭证管理

部分资源需要访问凭证。

例如 Redis：

```text
Host
Port
Username
Password
```

MySQL：

```text
Host
Port
Database
Username
Password
```

V1 可以在创建资源时生成随机密码。

例如：

```text
Username:
default

Password:
随机生成
```

密码不能明文存储。

可以使用：

```text
AES-GCM
```

加密后保存。

---

# 20. Credential 使用流程

```text
创建资源
   │
   ▼
生成随机密码
   │
   ▼
加密密码
   │
   ▼
保存 Credential
   │
   ▼
创建 Docker Resource
   │
   ▼
RUNNING
   │
   ▼
用户查看连接信息
```

---

# 21. 密码重置

用户可以：

```text
Reset Password
```

系统：

```text
生成新密码
   ↓
创建 RESET_CREDENTIAL Task
   ↓
Worker 执行
   ↓
更新资源密码
   ↓
更新 Credential
   ↓
SUCCESS
```

旧密码失效。

---

# 22. API Key

除了 Web 控制台，系统提供 API。

用户可以创建：

```text
API Key
```

例如：

```text
dev-key
```

权限：

```text
resource:read
resource:create
resource:operate
```

调用：

```http
Authorization: Bearer dc_xxxxxxxxx
```

API Key 支持：

* 创建
* 删除
* 禁用
* 设置权限
* 设置过期时间

---

# 23. RBAC 权限控制

所有敏感操作必须进行权限检查。

例如：

```text
POST /resources
        │
        ▼
当前用户
        │
        ▼
Project Member
        │
        ▼
Role
        │
        ▼
Permission
```

例如：

```text
Owner
 ├── resource:create
 ├── resource:delete
 ├── resource:operate
 └── member:manage

Developer
 ├── resource:create
 ├── resource:operate
 └── resource:read

Viewer
 └── resource:read
```

---

# 24. 多租户

V1 可以加入基础的 Tenant 概念。

结构：

```text
Tenant
 │
 ├── Users
 │
 └── Projects
       │
       └── Resources
```

所有资源必须属于某个 Tenant。

例如：

```text
Tenant A
└── Project A
    └── Redis A

Tenant B
└── Project B
    └── Redis B
```

Tenant A 的用户不能访问 Tenant B 的资源。

---

# 25. 审计日志

所有重要操作记录 Audit Log。

例如：

```text
User:
YunShn

Action:
DELETE_RESOURCE

Resource:
redis-prod

Result:
SUCCESS

Time:
2026-09-01 20:15:32
```

需要记录：

* 操作用户
* Project
* Resource
* 操作类型
* 操作结果
* IP
* 时间
* 错误信息

---

# 26. 核心数据模型

V1 主要包含：

```text
users
tenants
projects
project_members

resources
credentials

tasks

api_keys

audit_logs
```

关系：

```text
Tenant
 │
 ├── Users
 │
 └── Projects
       │
       ├── Members
       │
       └── Resources
             │
             ├── Credential
             │
             └── Tasks
```

---

# 27. 技术架构

```text
                         React
                           │
                           ▼
                    Go API Server
                           │
              ┌────────────┼────────────┐
              │            │            │
              ▼            ▼            ▼
           Auth/RBAC    Resource      Task API
                           │            │
                           │            ▼
                           │        PostgreSQL
                           │            │
                           │            ▼
                           │       Worker Pool
                           │            │
                           └──────┬─────┘
                                  ▼
                                Docker

                         Redis
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
            Cache        Lock         Rate Limit
```

---

# 28. PostgreSQL 职责

PostgreSQL 保存核心业务数据：

```text
User
Tenant
Project
Resource
Task
Credential
API Key
Audit Log
```

同时承担：

* 事务
* Task 持久化
* Task 抢占
* 状态管理
* 数据一致性

---

# 29. Redis 职责

Redis 暂时不承担核心业务数据。

主要用于：

### 分布式锁

```text
lock:resource:{id}
```

### 缓存

```text
user:{id}:permissions
```

### API 限流

```text
rate_limit:{api_key}
```

Redis 不可用时，核心业务数据仍然保存在 PostgreSQL。

---

# 30. Docker Provider

系统不应该让业务代码直接调用 Docker API。

定义统一接口：

```text
ResourceProvider

Create()
Start()
Stop()
Restart()
Delete()
```

V1 实现：

```text
DockerProvider
```

未来如果需要支持 Kubernetes，可以新增：

```text
KubernetesProvider
```

而无需修改上层 Resource Service。

V1 暂不实现 Kubernetes。

---

# 31. 核心创建流程

完整流程：

```text
用户
 │
 │ Create Resource
 ▼
API Server
 │
 ├── Auth
 │
 ├── Permission Check
 │
 ├── Validate
 │
 └── Transaction
       │
       ├── Resource
       │     CREATING
       │
       ├── Credential
       │
       └── Task
             PENDING
               │
               ▼
             Worker
               │
               ▼
        DockerProvider
               │
               ▼
          Docker API
               │
         ┌─────┴─────┐
         ▼           ▼
       Success      Failed
         │           │
         ▼           ▼
      RUNNING       Retry
                     │
               ┌─────┴─────┐
               ▼           ▼
            Success      Failed
               │           │
               ▼           ▼
            RUNNING      FAILED
```

---

# 32. 删除资源流程

```text
用户点击 Delete
       │
       ▼
权限检查
       │
       ▼
检查 Resource 状态
       │
       ▼
创建 DELETE Task
       │
       ▼
Resource = DELETING
       │
       ▼
Worker
       │
       ▼
Docker Delete
       │
   ┌───┴───┐
   ▼       ▼
成功      失败
   │       │
   ▼       ▼
DELETED   Retry
```

---

# 33. Worker 异常流程

```text
Task
 │
 ▼
RUNNING
 │
 ▼
Worker Crash
 │
 ▼
Heartbeat Timeout
 │
 ▼
Task Recovery
 │
 ▼
Retry
 │
 ▼
PENDING
 │
 ▼
New Worker
```

---

# 34. 非功能需求

## 34.1 一致性

需要保证：

```text
Resource
Task
Credential
```

之间的数据一致性。

例如创建 Resource 时：

```text
Resource 创建成功
Task 创建成功
Credential 创建成功
```

必须通过事务保证。

---

## 34.2 幂等性

相同操作重复提交不能产生错误的副作用。

---

## 34.3 可恢复性

Worker、API Server 重启后：

* 未完成 Task 不丢失
* 超时 Task 可以恢复
* Resource 状态最终能够恢复

---

## 34.4 并发安全

需要避免：

```text
同一 Resource
Stop
Delete
```

同时执行。

---

# 35. V1 功能范围

## 必须实现

```text
[x] 用户注册/登录
[x] Project
[x] Docker Resource
[x] Resource CRUD
[x] Start / Stop / Restart
[x] Task
[x] Worker Pool
[x] Task Retry
[x] Task Timeout
[x] Task Recovery
[x] 基础幂等
[x] Docker Provider
[x] Credential
```

## 第二阶段

```text
[ ] RBAC
[ ] Tenant
[ ] API Key
[ ] Redis Lock
[ ] Rate Limit
[ ] Audit Log
```

## 暂不实现

```text
[ ] Kafka
[ ] RabbitMQ
[ ] Kubernetes
[ ] Prometheus
[ ] Loki
[ ] Grafana
[ ] 计费
[ ] 自动扩缩容
[ ] 多云
```

---

# 36. V1 开发顺序

建议严格按照下面顺序开发。

```text
Phase 1
项目基础架构
    ↓
User
    ↓
Project

Phase 2
Resource
    ↓
Docker Provider
    ↓
Create Resource

Phase 3
Task
    ↓
Worker Pool
    ↓
异步 Create

Phase 4
Start / Stop / Restart / Delete
    ↓
Resource State Machine

Phase 5
Retry
    ↓
Timeout
    ↓
Idempotency
    ↓
Worker Recovery

Phase 6
Credential
    ↓
Password
    ↓
Reset Password

Phase 7
RBAC
    ↓
Tenant
    ↓
API Key

Phase 8
Redis
    ↓
Distributed Lock
    ↓
Rate Limit
```

---

# 37. 项目最终核心能力

完成 V1 后，系统应该能够完成：

```text
用户
 │
 ▼
创建 Project
 │
 ▼
创建 Docker Resource
 │
 ▼
系统生成 Credential
 │
 ▼
创建 Task
 │
 ▼
Worker 异步执行
 │
 ▼
Docker 创建 Container
 │
 ▼
Resource RUNNING
 │
 ├── Stop
 ├── Start
 ├── Restart
 ├── Reset Password
 └── Delete
```

并且能够处理：

```text
任务失败
任务重试
任务超时
Worker 崩溃
重复提交
并发操作
```

---

# 38. 项目面试定位

项目不以“实现一个云平台”为目标，而以：

> **设计并实现一个可靠的资源异步编排系统**

为核心。

面试重点可以围绕：

```text
为什么需要 Task？

为什么不能同步调用 Docker？

Worker 如何获取任务？

多个 Worker 如何避免重复获取？

如何保证幂等？

Task 执行到一半 Worker 挂了怎么办？

如何实现 Retry？

为什么使用指数退避？

Resource 状态和 Task 状态有什么区别？

数据库事务如何保证一致性？

Redis 在系统中解决什么问题？

为什么不使用 MQ？

为什么 Docker 不直接暴露给业务层？
```

这些问题构成项目的主要技术深度。

---

# 39. 项目一句话介绍

> **DevCloud 是一个基于 Go 开发的轻量级云资源管理与异步编排平台，通过统一的 Resource 模型管理 Docker 资源，并基于 PostgreSQL + Worker Pool 实现可靠的异步任务执行、状态管理、幂等、重试和故障恢复。**
