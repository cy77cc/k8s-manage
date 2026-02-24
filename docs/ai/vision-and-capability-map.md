# AI Vision and Capability Map

## Goal

将 AI 从问答组件升级为平台级智能操作层，覆盖 Hosts/K8s/Services/RBAC/Projects 主链路。

## Capability Map

- Global Copilot: 全局 SSE 聊天、工具调用、审批与执行追踪。
- Hosts:
  - `os.get_cpu_mem`
  - `os.get_disk_fs`
  - `os.get_net_stat`
  - `os.get_process_top`
  - `os.get_journal_tail`
  - `host.ssh_exec_readonly`
- Kubernetes:
  - `k8s.list_resources`
  - `k8s.get_events`
  - `k8s.get_pod_logs`
- Services:
  - `service.get_detail`
  - `service.deploy_preview`
  - `service.deploy_apply` (approval required)

## Interaction Model

- SSE events:
  - `meta`
  - `delta`
  - `tool_call`
  - `tool_result`
  - `approval_required`
  - `done`
  - `error`

## Safety Baseline

- 默认只读。
- mutating 工具必须审批。
- 审批超时默认 10 分钟。
