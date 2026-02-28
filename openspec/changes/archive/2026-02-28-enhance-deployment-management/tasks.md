## 1. Data model and contract extension

- [x] 1.1 Add migrations for environment install jobs, step logs, credential-source metadata, and encrypted secret references in `storage/migrations`
- [x] 1.2 Extend API contracts in `api/*/v1` for environment bootstrap, job status query, external credential import, and connectivity test responses
- [x] 1.3 Implement backward-compatible DTO mapping for existing deployment targets that lack credential-source fields

## 2. Environment bootstrap engine (SSH + binary)

- [x] 2.1 Implement SSH execution pipeline and job state machine (`queued/running/succeeded/failed/rolled_back`) in deployment/cluster service logic
- [x] 2.2 Implement runtime package resolver using `script/runtime/<runtime>/<version>/` manifest + checksum validation
- [x] 2.3 Implement runtime adapters for `k8s` and `compose` install/uninstall/preflight/health-check steps with structured diagnostics

## 3. Cluster credential ingestion and security

- [x] 3.1 Implement platform-managed cluster credential registration and binding path for deployment targets
- [x] 3.2 Implement external credential import flow supporting kubeconfig and certificate bundle with schema and connectivity validation
- [x] 3.3 Enforce Casbin policies for credential import/test/read operations and ensure plaintext secrets are never returned by default APIs

## 4. Deployment flow integration and frontend support

- [x] 4.1 Integrate bootstrap readiness checks into deployment target creation and release preflight gates
- [x] 4.2 Add frontend flows in Deployment pages and `web/src/api/modules` for environment creation, import credentials, bootstrap progress, and failure diagnostics
- [x] 4.3 Align AI command entry with the same environment bootstrap and credential governance path used by Deployment UI

## 5. Validation, scripts, and rollout

- [x] 5.1 Define and add `script/` artifact layout (runtime package, checksum, install/uninstall scripts, manifest examples) and document required files
- [x] 5.2 Add regression tests covering SSH bootstrap failures, checksum mismatch, invalid kubeconfig/cert import, RBAC denial, and target readiness gating
- [x] 5.3 Run `openspec validate --json` and update `docs/roadmap.md` + `docs/progress.md` for environment deployment phase status
