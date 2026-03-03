# Service Studio API Contract (Phase-1)

## Render and Template Vars

- `POST /api/v1/services/render/preview`
  - request: `{ mode, target, service_name, service_type, standard_config?, custom_yaml?, variables?, validate_only? }`
  - response: `{ rendered_yaml, resolved_yaml, diagnostics, unresolved_vars, detected_vars, ast_summary }`

- `POST /api/v1/services/transform`
  - request: `{ standard_config, target, service_name, service_type }`
  - response: `{ custom_yaml, source_hash, detected_vars }`

- `POST /api/v1/services/variables/extract`
  - request: `{ standard_config?, custom_yaml?, render_target, service_name?, service_type? }`
  - response: `{ vars: [{ name, required, default, description, source_path }] }`

## Variable Sets

- `GET /api/v1/services/:id/variables/schema`
  - response: `{ vars: TemplateVar[] }`

- `GET /api/v1/services/:id/variables/values?env=staging`
  - response: `{ service_id, env, values, secret_keys, updated_at }`

- `PUT /api/v1/services/:id/variables/values`
  - request: `{ env, values, secret_keys }`
  - response: same as get values

## Revision and Release

- `GET /api/v1/services/:id/revisions`
  - response: `{ list, total }`

- `POST /api/v1/services/:id/revisions`
  - request: `{ config_mode, render_target, standard_config?, custom_yaml?, variable_schema? }`
  - response: `ServiceRevision`

- `PUT /api/v1/services/:id/deploy-target`
  - request: `{ cluster_id, namespace, deploy_target, policy }`
  - response: `ServiceDeployTarget`
  - **DEPRECATED**: 部署目标不再在服务配置中存储，部署时直接指定

- `POST /api/v1/services/:id/deploy/preview`
  - request: `{ env, cluster_id, namespace?, variables?, deploy_target? }`
  - response: `{ resolved_yaml, checks, warnings, target }`
  - **注意**: `cluster_id` 为必填参数

- `POST /api/v1/services/:id/deploy`
  - request: `{ cluster_id, namespace?, env?, variables?, deploy_target?, approval_token? }`
  - response: `{ release_record_id }`
  - **注意**:
    - `cluster_id` 为必填参数
    - 部署时校验 `Service.env` 与 `Cluster.env_type` 是否匹配
    - 不匹配时返回 `400 ENV_MISMATCH` 错误

- `GET /api/v1/services/:id/releases`
  - response: `{ list, total }`
