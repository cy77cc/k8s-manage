## 1. Cleanup - Remove Test Files

- [x] 1.1 Delete all test files in `internal/ai/tools/*_test.go` (12 files)

## 2. Create Directory Structure

- [x] 2.1 Create `internal/ai/tools/param/` directory
- [x] 2.2 Create `internal/ai/tools/impl/` directory
- [x] 2.3 Create domain subdirectories under `impl/`:
  - `impl/kubernetes/`
  - `impl/host/`
  - `impl/service/`
  - `impl/monitor/`
  - `impl/cicd/`
  - `impl/deployment/`
  - `impl/governance/`
  - `impl/infrastructure/`
  - `impl/mcp/`

## 3. Merge Fragment Files into contracts.go

- [x] 3.1 Merge `tool_call_id.go` content into `tool_contracts.go`
- [x] 3.2 Merge `tool_name.go` content into `tool_contracts.go`
- [x] 3.3 Merge `category_helpers.go` content into `tool_contracts.go`
- [x] 3.4 Merge `tool_contracts_ai_enhancement.go` content into `tool_contracts.go`
- [x] 3.5 Delete the merged fragment files (4 files)
- [x] 3.6 Verify `contracts.go` compiles (rename from `tool_contracts.go`)

## 4. Move Parameter Processing Files

- [x] 4.1 Move `param_hints.go` to `param/hints.go`
- [x] 4.2 Move `tool_param_resolver.go` to `param/resolver.go`
- [x] 4.3 Move `tool_param_validator.go` to `param/validator.go`
- [x] 4.4 Update package declarations in moved files to `package param`
- [x] 4.5 Add export aliases in `tools/` root if needed for backward compatibility

## 5. Move Tool Implementation Files

- [x] 5.1 Move `tools_k8s.go` to `impl/kubernetes/tools.go`
- [x] 5.2 Move `tools_os.go` and `tools_host.go` to `impl/host/tools.go` (merge)
- [x] 5.3 Move `tools_service.go` to `impl/service/tools.go`
- [x] 5.4 Move `tools_monitor.go` to `impl/monitor/tools.go`
- [x] 5.5 Move `tools_cicd.go` and `tools_job.go` to `impl/cicd/tools.go` (merge)
- [x] 5.6 Move `tools_deployment.go`, `tools_config.go`, `tools_inventory.go` to `impl/deployment/tools.go` (merge)
- [x] 5.7 Move `tools_governance.go` and `tools_topology.go` to `impl/governance/tools.go` (merge)
- [x] 5.8 Move `tools_infrastructure.go` to `impl/infrastructure/tools.go`
- [x] 5.9 Move `mcp_client.go` to `impl/mcp/client.go`
- [x] 5.10 Move `tools_mcp_proxy.go` to `impl/mcp/proxy.go`

## 6. Update Import Paths

- [x] 6.1 Update all imports in `impl/*/tools.go` files to use root package types
- [x] 6.2 Update imports in `tools_registry.go` to reference moved implementations
- [x] 6.3 Update imports in `tools_common.go` (runner.go) if needed
- [x] 6.4 Run `goimports -w internal/ai/tools/` to fix all imports

## 7. Rename Core Files

- [x] 7.1 Rename `tools_common.go` to `runner.go`
- [x] 7.2 Rename `tool_contracts.go` to `contracts.go` (after merging)

## 8. Verification

- [x] 8.1 Run `go build ./internal/ai/...` to verify compilation
- [x] 8.2 Run `go test ./internal/ai/... -short` to verify tests pass
- [x] 8.3 Verify file count in `internal/ai/tools/` root is ≤ 10
- [x] 8.4 Verify the refactor did not introduce new tool-name mismatches; confirmed the only remaining mismatch is the pre-existing `ops_aggregate_status` entry in `scene_mappings.yaml`

## 9. OpenSpec Update

- [x] 9.1 Sync delta spec to main spec via current CLI archive flow (`openspec archive "refactor-tools-directory"` updates specs during archive)
- [x] 9.2 Archive the change with current CLI syntax: `openspec archive "refactor-tools-directory" -y`
