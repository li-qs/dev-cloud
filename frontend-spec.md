# DevCloud Frontend Specification

## 1. 项目简介

DevCloud 是一个轻量级云资源管理平台。

用户可以通过 Web 控制台创建、启动、停止、重启和删除资源。

当前基础设施 Provider 只有 Docker，但前端设计不能与 Docker API 强耦合，Provider 应作为资源的一个属性展示。

DevCloud 后端使用 Go + Echo + Ent + PostgreSQL + Redis。

资源操作采用异步任务模型：

    用户操作
       ↓
    API 创建 Task
       ↓
    Task PENDING
       ↓
    Worker Claim
       ↓
    Task RUNNING
       ↓
    Provider 执行
       ↓
    Task SUCCESS / FAILED
       ↓
    Resource 状态更新

前端需要重点体现 Resource 和 Task 的异步关系。

---

# 2. 技术栈

使用：

- React
- TypeScript
- Vite
- React Router
- 一个成熟的 React UI 组件库
- Fetch 或 Axios
- CSS / UI 组件库自身的样式系统

推荐优先使用：

- shadcn/ui

如果项目已经使用其他成熟 UI 组件库，可以保持一致，不要为了更换 UI 库重构项目。

不要引入没有必要的复杂状态管理框架。

优先：

- React state
- React Context（只有确实需要时）
- API 层封装

暂时不要：

- Redux
- WebSocket
- SSE
- GraphQL
- 微前端
- Kubernetes UI
- 复杂权限系统

---

# 3. UI 定位

这是一个现代化的 Cloud Console / SaaS 管理控制台。

不要做成传统的企业 CRUD 后台。

整体风格：

- 简洁
- 专业
- 信息密度适中
- 状态清晰
- 操作明确
- 类似现代云平台控制台

不要：

- 大量渐变
- 大量动画
- 花哨 Landing Page
- 无意义的数据图表
- 过度装饰

---

# 4. 页面结构

需要实现以下页面：

    /login

    /dashboard

    /resources

    /resources/new

    /resources/:id

    /tasks

    /tasks/:id

    /settings

整体 Layout：

    ┌──────────────────────────────────────────────┐
    │ Header                                       │
    ├──────────────┬───────────────────────────────┤
    │              │                               │
    │ Sidebar      │ Main Content                  │
    │              │                               │
    │ Dashboard    │                               │
    │ Resources    │                               │
    │ Tasks        │                               │
    │              │                               │
    │ Settings     │                               │
    │              │                               │
    └──────────────┴───────────────────────────────┘

---

# 5. 登录页面

路径：

    /login

功能：

- 用户名/邮箱
- 密码
- 登录按钮
- 登录失败提示
- 登录成功后进入 Dashboard

如果后端使用 JWT：

- access token 保存方式按照后端 API 实际约定实现
- refresh token 如果由 HttpOnly Cookie 管理，前端不要读取 Cookie
- API 请求需要自动携带认证信息

如果当前后端登录 API 尚未完全确定，可以先定义 API client interface，不要把 API URL 写死在组件内部。

---

# 6. Dashboard

路径：

    /dashboard

Dashboard 用于展示系统概况。

顶部统计卡片：

    Total Resources
    Running
    Creating
    Failed

例如：

    ┌────────────┐
    │ Resources  │
    │    12      │
    └────────────┘

    ┌────────────┐
    │ Running    │
    │     8      │
    └────────────┘

    ┌────────────┐
    │ Creating   │
    │     2      │
    └────────────┘

    ┌────────────┐
    │ Failed     │
    │     2      │
    └────────────┘

下面显示：

## Recent Resources

字段：

- Name
- Provider
- Image
- Status
- Created At

点击资源进入：

    /resources/:id

---

## Recent Tasks

字段：

- ID
- Resource
- Type
- Status
- Created At

点击 Task 进入：

    /tasks/:id

Dashboard 不需要复杂图表。

---

# 7. Resources 页面

路径：

    /resources

页面标题：

    Resources

右上角：

    + Create Resource

提供：

- 搜索
- Status 筛选
- Provider 筛选

资源表格：

    Name
    Image
    Provider
    Status
    Created At
    Actions

例如：

    ubuntu-dev
    ubuntu:24.04
    Docker
    RUNNING
    2 minutes ago

---

# 8. Resource 状态

Resource 状态：

    CREATING
    RUNNING
    STOPPING
    STOPPED
    STARTING
    RESTARTING
    DELETING
    DELETED
    FAILED

前端需要统一封装 ResourceStatus 组件。

例如：

    RUNNING
    → 正常状态

    CREATING
    → loading 状态

    STARTING
    → loading 状态

    STOPPING
    → loading 状态

    RESTARTING
    → loading 状态

    DELETING
    → loading 状态

    STOPPED
    → 非运行状态

    FAILED
    → 错误状态

不要在不同页面重复实现状态 UI。

应该统一：

    <ResourceStatus status={resource.status} />

---

# 9. Resource Actions

根据 Resource 状态决定允许的操作。

## RUNNING

允许：

    Stop
    Restart
    Delete
    View Details

## STOPPED

允许：

    Start
    Delete
    View Details

## CREATING

允许：

    View Details
    View Task

不允许：

    Start
    Stop
    Restart

## STARTING

不允许重复操作。

## STOPPING

不允许重复操作。

## RESTARTING

不允许重复操作。

## DELETING

不允许重复操作。

## FAILED

允许：

    View Details

是否允许 Retry 根据后端 API 决定。

前端不能自己假设 FAILED 一定可以 Retry。

---

# 10. Create Resource

路径：

    /resources/new

表单：

    Name
    Provider
    Image
    CPU
    Memory
    Configuration

当前 Provider：

    Docker

例如：

    Name
    [ ubuntu-dev ]

    Provider
    [ Docker ]

    Image
    [ ubuntu:24.04 ]

    CPU
    [ 2 ]

    Memory
    [ 2048 MB ]

按钮：

    Cancel
    Create Resource

---

# 11. 创建资源的异步交互

创建资源不是同步等待 Docker 创建完成。

API 返回后，前端应该立即进入资源详情页面。

后端可能返回：

    {
        "resource_id": 123,
        "task_id": 456
    }

前端显示：

    Resource
    ubuntu-dev

    Status
    CREATING

    Task
    #456

    Task Status
    RUNNING

然后通过 Task API 查询任务状态。

推荐暂时使用 polling：

    GET /api/tasks/:id

每 1~2 秒请求一次。

当 Task：

    SUCCESS

停止 polling，并重新获取 Resource。

最终：

    Resource
    RUNNING

如果 Task：

    FAILED

停止 polling，并显示：

    Resource creation failed

    Error:
    <error_message>

不要实现 WebSocket 或 SSE。

---

# 12. Resource Detail

路径：

    /resources/:id

页面结构：

    ← Resources

    ubuntu-dev

    Docker · ubuntu:24.04

    ● RUNNING

    [ Stop ] [ Restart ] [ Delete ]

---

## Overview

显示：

    Name
    Provider
    Image
    Runtime ID
    Status
    Created At
    Updated At

Runtime ID 是基础设施 Runtime ID。

例如 Docker Container ID。

---

## Configuration

显示：

    CPU
    Memory
    Network
    其他 config

根据后端实际返回的数据展示。

不要在前端硬编码 Docker 特有字段。

---

## Recent Tasks

显示该 Resource 最近的 Task：

    ID
    Type
    Status
    Retry Count
    Created At

点击进入：

    /tasks/:id

---

# 13. Resource Detail 的异步状态

如果 Resource 当前处于：

    CREATING
    STARTING
    STOPPING
    RESTARTING
    DELETING

页面应该明显展示：

    operation in progress

并显示对应 Task。

例如：

    ┌─────────────────────────────────┐
    │ Restarting resource...          │
    │                                 │
    │ Task #1024                      │
    │ RUNNING                         │
    └─────────────────────────────────┘

页面可以 polling Resource 或 Task。

优先 polling Task。

---

# 14. Delete Resource

点击 Delete 时显示确认弹窗：

    Delete Resource?

    ubuntu-dev

    This action cannot be undone.

    [Cancel] [Delete]

删除操作成功提交后：

    Resource → DELETING

然后等待对应 Task。

Task 成功：

    Resource → DELETED

然后跳转回：

    /resources

---

# 15. Tasks 页面

路径：

    /tasks

Task 页面是 DevCloud 的核心页面之一。

需要突出异步任务系统。

提供：

- 搜索
- Status 筛选
- Type 筛选

表格：

    ID
    Resource
    Type
    Status
    Retry Count
    Created At
    Updated At

Task Type：

    CREATE_RESOURCE
    START_RESOURCE
    STOP_RESOURCE
    RESTART_RESOURCE
    DELETE_RESOURCE

---

# 16. Task 状态

Task 状态：

    PENDING
    RUNNING
    SUCCESS
    FAILED

状态含义：

    PENDING
    等待 Worker 执行

    RUNNING
    Worker 已经 Claim，正在执行

    SUCCESS
    执行成功

    FAILED
    执行失败

前端统一封装：

    <TaskStatus status={task.status} />

---

# 17. Task Detail

路径：

    /tasks/:id

页面：

    Task #1024

    CREATE_RESOURCE

    Status
    SUCCESS

显示：

    Task ID
    Resource
    Type
    Status
    Retry Count
    Created At
    Updated At

如果失败：

    Status
    FAILED

    Error
    failed to start container:
    context deadline exceeded

错误信息需要明显展示，但不要直接暴露敏感信息。

---

# 18. Task 与 Resource 的关系

前端必须体现：

    Task
      ↓
    Resource

例如：

    CREATE_RESOURCE
          ↓
    ubuntu-dev

Task Detail 可以点击 Resource：

    View Resource

Resource Detail 可以点击 Task：

    View Task

形成双向导航。

---

# 19. API Client

不要在 React 页面组件里直接写：

    fetch("/api/resources")

应该统一封装 API。

推荐：

    src/
    ├── api/
    │   ├── client.ts
    │   ├── auth.ts
    │   ├── resource.ts
    │   └── task.ts

例如：

    resourceApi.list()
    resourceApi.get(id)
    resourceApi.create(data)
    resourceApi.start(id)
    resourceApi.stop(id)
    resourceApi.restart(id)
    resourceApi.delete(id)

Task：

    taskApi.list()
    taskApi.get(id)

---

# 20. TypeScript Types

推荐：

    src/types/resource.ts
    src/types/task.ts
    src/types/user.ts

Resource：

    interface Resource {
        id: number
        user_id: number
        name: string
        provider: string
        image: string
        runtime_id?: string
        status: ResourceStatus
        config?: Record<string, unknown>
        error_message?: string
        created_at: string
        updated_at: string
        deleted_at?: string
    }

ResourceStatus：

    type ResourceStatus =
        | "CREATING"
        | "RUNNING"
        | "STOPPING"
        | "STOPPED"
        | "STARTING"
        | "RESTARTING"
        | "DELETING"
        | "DELETED"
        | "FAILED"

Task：

    interface Task {
        id: number
        user_id: number
        resource_id: number
        type: TaskType
        status: TaskStatus
        payload?: Record<string, unknown>
        retry_count: number
        error_message?: string
        next_run_at?: string
        created_at: string
        updated_at: string
    }

TaskType：

    type TaskType =
        | "CREATE_RESOURCE"
        | "START_RESOURCE"
        | "STOP_RESOURCE"
        | "RESTART_RESOURCE"
        | "DELETE_RESOURCE"

TaskStatus：

    type TaskStatus =
        | "PENDING"
        | "RUNNING"
        | "SUCCESS"
        | "FAILED"

如果后端实际 JSON 字段不同，以后端 API 为准。

---

# 21. Error Handling

统一处理 API 错误。

页面不要到处：

    alert(error)

应该有统一错误展示机制。

错误类型：

    400
    Bad Request

    401
    Unauthorized

    403
    Forbidden

    404
    Not Found

    409
    Conflict

    500
    Internal Server Error

特别注意：

409 很可能表示资源当前状态不允许执行操作。

例如：

    Resource is already stopping

应该给用户明确提示。

---

# 22. Loading

所有 API 请求都需要正确处理 loading。

例如：

    Loading resources...

按钮执行操作时：

    Stop
    ↓
    Stopping...

防止用户连续点击。

对于：

    CREATING
    STARTING
    STOPPING
    RESTARTING
    DELETING

应该显示明显的进行中状态。

---

# 23. Empty State

Resources 没有数据：

    No resources yet.

    Create your first resource.

    [Create Resource]

Tasks 没有数据：

    No tasks found.

不要显示空白页面。

---

# 24. Responsive

优先桌面端。

需要支持：

- 1440px
- 1280px
- 1024px

移动端不需要作为第一优先级。

---

# 25. Mock API

在后端 API 完成之前，可以使用 mock data。

但是：

**Mock API 与真实 API 必须通过同一套 API interface 使用。**

页面组件不应该知道当前数据来自 Mock 还是真实 API。

例如：

    ResourcePage
        ↓
    resourceApi.list()
        ↓
    Mock / Real API

切换数据源时不要修改页面组件。

---

# 26. Polling

只在需要的时候 polling。

例如 Task：

    PENDING
    RUNNING

进行 polling。

当：

    SUCCESS
    FAILED

停止 polling。

不要让整个 Dashboard 每秒刷新。

推荐：

    interval = 2000ms

页面离开后停止 polling。

---

# 27. Authentication

所有需要登录的页面：

    /dashboard
    /resources
    /tasks
    /settings

都需要认证。

未登录：

    → /login

登录页面不能访问需要认证的 API。

如果 API 返回：

    401

统一清理前端认证状态并跳转：

    /login

---

# 28. Settings

当前只需要一个简单页面。

可以显示：

    Account
    Username
    Email

以及：

    Logout

不要实现复杂的系统配置。

---

# 29. 不要实现的功能

当前版本明确不实现：

- Kubernetes
- Docker Swarm
- Container Registry
- Billing
- IAM
- Multi-region
- VPC
- Load Balancer
- WebSocket
- SSE
- Prometheus Dashboard
- Grafana
- Kubernetes Dashboard
- 复杂 RBAC
- 多租户
- 实时监控大盘

这些可以作为未来扩展，不属于当前版本。

---

# 30. 视觉重点

前端最重要的不是展示很多功能，而是让用户一眼理解：

    Resource

    当前状态

    当前正在执行的操作

    对应 Task

    Task 执行结果

例如：

    ubuntu-dev

    ● RUNNING

    ─────────────────────────

    Recent Task

    #1024
    RESTART_RESOURCE
    ✓ SUCCESS

这是整个 UI 的核心信息结构。

---

# 31. 核心用户流程

## 创建资源

    Resources
        ↓
    Create Resource
        ↓
    Submit
        ↓
    Resource CREATING
        ↓
    Task PENDING
        ↓
    Task RUNNING
        ↓
    Task SUCCESS
        ↓
    Resource RUNNING

---

## 停止资源

    Resource RUNNING
        ↓
    Click Stop
        ↓
    Task PENDING
        ↓
    Task RUNNING
        ↓
    Resource STOPPING
        ↓
    Task SUCCESS
        ↓
    Resource STOPPED

---

## 启动资源

    Resource STOPPED
        ↓
    Click Start
        ↓
    Resource STARTING
        ↓
    Task RUNNING
        ↓
    Task SUCCESS
        ↓
    Resource RUNNING

---

## 失败

    Task RUNNING
        ↓
    Provider Error
        ↓
    Retry
        ↓
    Task FAILED
        ↓
    Resource FAILED

前端需要展示：

    Error Message

---

# 32. Component Design

建议至少抽象：

    Layout
    Sidebar
    Header

    ResourceStatus
    ResourceActions
    ResourceTable
    ResourceCard

    TaskStatus
    TaskType
    TaskTable

    LoadingState
    EmptyState
    ErrorState
    ConfirmDialog

不要为了抽象而抽象。

如果组件只使用一次，不需要强行抽象。

---

# 33. 最终目标

完成后的 DevCloud 前端应该让用户能够：

1. 登录
2. 查看 Dashboard
3. 创建 Resource
4. 查看 Resource 状态
5. 启动 Resource
6. 停止 Resource
7. 重启 Resource
8. 删除 Resource
9. 查看 Task
10. 查看 Task 执行结果
11. 查看失败原因

最重要的是：

**用户的每一个 Resource 操作都能够与一个异步 Task 对应起来。**

UI 应该围绕：

    Resource + Task + Status

三个核心概念设计。

不要把项目做成普通的后台 CRUD 系统。
