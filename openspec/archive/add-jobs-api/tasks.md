# Tasks: 添加后端 Jobs API 路由

## Task List

- [x] **Task 1**: 创建数据模型 `internal/model/job.go`
- [x] **Task 2**: 创建 jobs 服务模块 `internal/service/jobs/`
- [x] **Task 3**: 注册 jobs 服务到 `internal/service/service.go`
- [x] **Task 4**: 更新数据库迁移配置

---

## Task Details

### Task 1: 创建数据模型

**文件:** `internal/model/job.go`

创建 Job、JobExecution、JobLog 三个模型。

---

### Task 2: 创建 jobs 服务模块

**文件:**
- `internal/service/jobs/routes.go` - 路由注册
- `internal/service/jobs/handler.go` - HTTP 处理器
- `internal/service/jobs/logic.go` - 业务逻辑
- `internal/service/jobs/types.go` - 请求/响应类型

API 端点:
- `GET /jobs` - 列表
- `POST /jobs` - 创建
- `GET /jobs/:id` - 详情
- `PUT /jobs/:id` - 更新
- `DELETE /jobs/:id` - 删除
- `POST /jobs/:id/start` - 启动
- `POST /jobs/:id/stop` - 停止
- `GET /jobs/:id/executions` - 执行记录
- `GET /jobs/:id/logs` - 日志

---

### Task 3: 注册服务

**文件:** `internal/service/service.go`

添加:
```go
import "github.com/cy77cc/k8s-manage/internal/service/jobs"
// ...
jobs.RegisterJobsHandlers(v1, serverCtx)
```

---

### Task 4: 更新迁移配置

**文件:** `internal/testutil/integration.go` (测试用)

添加模型到自动迁移列表。

---

## 执行顺序

1. Task 1 → Task 2 → Task 3 → Task 4

## 验证步骤

1. 启动后端服务
2. 调用 `GET /api/v1/jobs` 验证返回空列表
3. 启动前端，访问任务中心页面验证无 404 错误
