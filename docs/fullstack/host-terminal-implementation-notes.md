# Host Terminal Implementation Notes

## Frontend

- Replaced `web/src/pages/Hosts/HostTerminalPage.tsx` mock logic with real API flow.
- Data sources:
  - `Api.hosts.getHostDetail(id)`
  - `Api.hosts.sshCheck(id)`
  - `Api.hosts.sshExec(id, command)`
- Added:
  - status machine (`connecting/connected/disconnected/error`)
  - line renderer (`input/output/error/system`)
  - command history navigation (up/down)
  - quick commands

## API Type Fix

- `web/src/api/modules/hosts.ts`
  - `sshCheck` return type changed from `ApiResponse<void>` to `ApiResponse<{reachable:boolean;message?:string}>`.

## Backend

- `internal/service/host/handler/host_exec.go`
  - Added `loadNodePrivateKey` helper.
  - Unified password/key auth for check/exec/batch.
