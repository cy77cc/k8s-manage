# Service Studio Regression Plan (Phase-1)

## Backend API

1. `render/preview`:
   - standard/custom 均可渲染
   - 传入 variables 可替换模板变量
   - unresolved_vars 正确返回

2. `variables`:
   - extract 返回完整变量列表
   - values upsert/get 按 env 生效

3. `revisions`:
   - create 后 list 可见新增 revision

4. `deploy-target`:
   - 保存默认 target 后 deploy/preview 可复用

5. `deploy`:
   - 返回 `release_record_id`
   - 失败时 release record 状态为 failed

6. `releases`:
   - list 可查询最新发布记录

## Frontend

1. 创建页为双栏编辑预览布局
2. 变量面板能展示 detected vars，输入后右侧预览实时变化
3. 详情页可保存 deploy target 和 env variable set
4. deploy preview 与 deploy apply 可执行
5. revisions/releases tab 数据可正常加载

## Commands

- `go test ./...`
- `cd web && npm run build`
