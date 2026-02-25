# Service Studio UX Spec (Phase-1)

## Layout

- 双栏布局：
  - 左侧：服务编辑器（基础信息 + 配置模式 + 变量面板）
  - 右侧：实时渲染预览（K8s / Compose / Helm tabs）
- 顶部动作：
  - 创建服务
  - 刷新预览

## User Flow

1. 填写基础信息（项目、团队、服务名、环境、运行时）
2. 选择配置模式：
   - standard：镜像、副本、端口、资源
   - custom：直接编辑 YAML
3. 系统实时渲染预览并展示 diagnostics
4. 若识别出模板变量（`{{var}}`），在变量面板填写值
5. 提交创建，进入详情页继续做 deploy target、变量集与发布

## Detail Flow

- 配置默认部署目标（cluster + namespace）
- 管理环境变量集（按 env 存储）
- 先执行 deploy preview，再执行 deploy apply
- 查看 revisions 与 release records

## Error Handling

- unresolved vars：提示并标红变量名
- preview diagnostics：warning/error 分级显示
- deploy 失败：显示后端错误信息，并在 release records 记录
