# Proposal: 添加后端 Jobs API 路由

## 问题描述

前端任务中心页面 (`/tasks`) 调用 `/api/v1/jobs` API，但后端没有实现该路由，导致 404 错误。

## 错误信息

```
GET /api/v1/jobs?page=1&page_size=100 404 (Not Found)
```

## 影响范围

- 前端: `web/src/pages/Tasks/TasksPage.tsx`
- 前端 API: `web/src/api/modules/tasks.ts`
- 后端: 需要新建 `internal/service/jobs/` 模块

## 解决方案

创建新的 `jobs` 服务模块，实现前端所需的 API 端点。

### API 端点设计

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/jobs | 获取任务列表 |
| POST | /api/v1/jobs | 创建任务 |
| GET | /api/v1/jobs/:id | 获取任务详情 |
| PUT | /api/v1/jobs/:id | 更新任务 |
| DELETE | /api/v1/jobs/:id | 删除任务 |
| POST | /api/v1/jobs/:id/start | 启动任务 |
| POST | /api/v1/jobs/:id/stop | 停止任务 |
| GET | /api/v1/jobs/:id/executions | 获取任务执行记录 |
| GET | /api/v1/jobs/:id/logs | 获取任务日志 |

### 数据模型

```go
type Job struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    Name        string    `json:"name"`
    Type        string    `json:"type"`           // shell, script
    Command     string    `json:"command"`
    HostIDs     string    `json:"host_ids"`
    Cron        string    `json:"cron"`
    Status      string    `json:"status"`         // pending, running, success, failed
    Timeout     int       `json:"timeout"`
    Priority    int       `json:"priority"`
    Description string    `json:"description"`
    LastRun     *time.Time `json:"last_run"`
    NextRun     *time.Time `json:"next_run"`
    CreatedBy   uint      `json:"created_by"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type JobExecution struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    JobID     uint      `json:"job_id"`
    HostID    uint      `json:"host_id"`
    HostIP    string    `json:"host_ip"`
    Status    string    `json:"status"`
    ExitCode  int       `json:"exit_code"`
    Output    string    `json:"output"`
    StartTime time.Time `json:"start_time"`
    EndTime   *time.Time `json:"end_time"`
    CreatedAt time.Time `json:"created_at"`
}

type JobLog struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    JobID       uint      `json:"job_id"`
    ExecutionID uint      `json:"execution_id"`
    Level       string    `json:"level"`
    Message     string    `json:"message"`
    CreatedAt   time.Time `json:"created_at"`
}
```

## 文件结构

```
internal/service/jobs/
├── routes.go      # 路由注册
├── handler.go     # HTTP 处理器
├── logic.go       # 业务逻辑
└── types.go       # 请求/响应类型

internal/model/
└── job.go         # 数据模型 (新增)
```

## 风险评估

- **低风险** - 新增模块，不影响现有功能
- **需数据库迁移** - 需要创建 jobs、job_executions、job_logs 表

## 验收标准

- [ ] GET /api/v1/jobs 返回任务列表
- [ ] POST /api/v1/jobs 创建任务成功
- [ ] GET /api/v1/jobs/:id 返回任务详情
- [ ] PUT /api/v1/jobs/:id 更新任务成功
- [ ] DELETE /api/v1/jobs/:id 删除任务成功
- [ ] POST /api/v1/jobs/:id/start 启动任务成功
- [ ] POST /api/v1/jobs/:id/stop 停止任务成功
- [ ] 前端任务中心页面正常加载
