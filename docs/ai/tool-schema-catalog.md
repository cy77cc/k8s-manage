# Tool Schema Catalog

## Local Tools

## `os.get_cpu_mem`
- required: none
- defaults: `target=localhost`

## `os.get_disk_fs`
- required: none
- defaults: `target=localhost`

## `os.get_net_stat`
- required: none
- defaults: `target=localhost`

## `os.get_process_top`
- required: none
- defaults: `target=localhost`, `limit=10`

## `os.get_journal_tail`
- required: `service`
- defaults: `target=localhost`, `lines=200`

## `os.get_container_runtime`
- required: none
- defaults: `target=localhost`

## `k8s.list_resources`
- required: `resource`
- defaults: `namespace=default`, `limit=50`

## `k8s.get_events`
- required: none
- defaults: `namespace=default`, `limit=50`

## `k8s.get_pod_logs`
- required: `pod`
- defaults: `namespace=default`, `tail_lines=200`

## `service.get_detail`
- required: `service_id`

## `service.deploy_preview`
- required: `service_id`, `cluster_id`

## `service.deploy_apply`
- required: `service_id`, `cluster_id`
- mode: `mutating` (approval required)

## `host.ssh_exec_readonly`
- required: `host_id`, `command`
- command must be allow-listed readonly command

## `host.list_inventory`
- required: none
- defaults: `limit=50`

## `host.batch_exec_preview`
- required: `host_ids`, `command`
- readonly preview (no command execution)

## `host.batch_exec_apply`
- required: `host_ids`, `command`
- mode: `mutating` (approval required)
- dangerous commands are blocked

## `host.batch_status_update`
- required: `host_ids`, `action`
- action: `online|offline|maintenance`
- mode: `mutating` (approval required)
