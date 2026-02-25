# Service Config Contract

## 1. API Contract

- `POST /api/v1/services/render/preview`
  - request: `{ mode, target, service_name, service_type, standard_config?, custom_yaml? }`
  - response: `{ rendered_yaml, diagnostics[], normalized_config }`
- `POST /api/v1/services/transform`
  - request: `{ standard_config, target, service_name, service_type }`
  - response: `{ custom_yaml, source_hash }`
- `POST /api/v1/services`
  - new fields: `service_kind, runtime_type, team_id, env, labels, config_mode, standard_config, custom_yaml, render_target`
- `PUT /api/v1/services/:id`
  - same shape as create, supports ownership/tag/config updates
- `GET /api/v1/services`
  - query: `project_id, team_id, runtime_type, env, label_selector, q`
  - response: `data: { list, total }`
- `POST /api/v1/services/:id/deploy`
  - request: `{ deploy_target, cluster_id, approval_token }`
- Helm:
  - `POST /api/v1/services/helm/import`
  - `POST /api/v1/services/helm/render`
  - `POST /api/v1/services/:id/deploy/helm`

## 2. Data Mapping

- `runtime_type`: `k8s | compose | helm`
- `config_mode`: `standard | custom`
- `labels`: `[{key,value}]` -> `services.labels_json`
- `standard_config` -> `services.standard_config_json`
- `custom_yaml` -> `services.custom_yaml`
- rendered output -> `services.yaml_content`

## 3. Compatibility

- Legacy fields (`image/replicas/service_port/container_port/env_vars/resources/yaml_content`) still accepted.
- Backend normalizes legacy payload into `standard_config` when `config_mode=standard`.
- Frontend list/detail consume new fields but can read legacy `yaml_content` fallback.
