# Frontend/Backend Separate Dev Design

**Date:** 2026-03-08

## Goal

将本地开发模式调整为标准前后端分离：

- 前端使用 Vite dev server 独立启动
- 后端开发模式下只提供 API 和 WebSocket
- 生产构建与发布继续保留当前 `web/dist` embed 到 Go 二进制的模式

## Current State

- 前端位于 `web/`，使用 Vite，`web/package.json` 已提供 `npm run dev`
- `web/vite.config.ts` 已配置 `/api` 和 `/ws` 代理到 `http://127.0.0.1:8080`
- 后端在 [`internal/service/service.go`](/root/project/k8s-manage/internal/service/service.go) 中总是调用 `registerWebStaticRoutes`
- 静态资源来自 `web/dist` 的 embed 文件系统，当前开发流程需要先构建前端再运行后端

## Chosen Approach

采用显式开发模式切换：

- 后端增加一个明确的开发模式判断
- 当运行在开发模式时，不注册前端静态资源和 SPA fallback
- 当运行在非开发模式时，保持现有 embed 静态资源加载逻辑不变

## Why This Approach

### Pros

- 开发与生产职责清晰
- 本地调试不再依赖反复执行前端构建
- 不影响现有生产构建、部署和 embed 交付方式
- 与当前 Vite proxy 配置天然兼容

### Rejected Alternatives

- 使用 `SERVE_WEB=false` 一类布尔开关
  - 能实现，但语义不如显式 `dev/prod` 清晰
- 自动探测 `localhost:5173`
  - 行为隐式，容易产生“为什么这次走 embed、那次没走”的调试问题

## Design

### 1. Backend Runtime Mode

后端引入一个显式的开发模式配置，推荐以环境变量实现，例如：

- `APP_ENV=development`
- 或 `APP_MODE=dev`

判断逻辑需要集中在后端配置/启动路径中，而不是散落在 handler 内部。

### 2. Static Route Registration

[`internal/service/service.go`](/root/project/k8s-manage/internal/service/service.go) 中的静态资源注册改为条件式：

- 开发模式：
  - 注册 `/api/*`
  - 注册 `/ws/*`
  - 不注册 `/assets/*`
  - 不注册 `vite.svg` / `favicon.ico` 的 embed 路由
  - 不注册 SPA `NoRoute` fallback
- 非开发模式：
  - 保持当前 embed 逻辑不变

### 3. Developer Workflow

新增明确的开发命令：

- `make dev-backend`
- `make dev-frontend`

可选新增：

- `make dev` 同时启动两者，适合常用本地开发

开发时访问：

- 前端：`http://127.0.0.1:5173`
- 后端：`http://127.0.0.1:8080`

前端 API 通过 Vite proxy 转发到后端。

### 4. Production Workflow

保持不变：

1. `make web-build`
2. `make build`
3. 运行后端二进制

生产模式下仍由后端直接服务 embed 的前端静态资源。

## Error Handling

- 如果生产模式下 `web/dist` 未构建，仍保留当前错误行为：访问前端路由时返回 `frontend not built`
- 如果开发模式下用户访问后端 `/`，应返回普通 404，而不是尝试回退到 embed 前端

## Testing

### Backend

至少覆盖以下行为：

- 开发模式下不注册静态前端路由
- 生产模式下仍注册静态路由
- 开发模式下 `/api/health` 仍正常

### Docs / Workflow

- README 增加“开发模式（前后端分离）”说明
- README 保留“生产构建（embed）”说明

## Compatibility

- 对外 API 不变
- WebSocket 路由不变
- 生产部署方式不变
- 仅调整本地开发默认工作流

## Success Criteria

- 本地开发不需要先执行 `make web-build`
- `go run main.go` 在开发模式下只提供 API / WS
- `cd web && npm run dev` 可独立访问页面并正常调用后端 API
- 生产构建仍能通过 embed 提供完整前端页面
