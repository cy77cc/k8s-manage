# AI Release Checklist

## Functional

- [ ] `GET /ai/capabilities` 返回权限过滤后的工具列表
- [ ] readonly 工具可直接 preview + execute
- [ ] mutating 工具未审批时返回 approval required
- [ ] 审批通过后 mutating 工具可执行
- [ ] `GET /ai/executions/:id` 可查执行记录

## Security

- [ ] 未登录无法调用 `/api/v1/ai/*`
- [ ] 非权限用户无法执行工具
- [ ] host readonly 命令白名单生效
- [ ] target 白名单生效（localhost 或已登记节点）

## Stability

- [ ] SSE 聊天连续输出无中断
- [ ] 工具超时返回结构化错误
- [ ] `go test ./...` 通过
- [ ] `npm run build` 通过
