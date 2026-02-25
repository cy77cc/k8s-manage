# Host Terminal Real Session Architecture

## Decision

- Terminal page must execute real commands against managed hosts via backend `/api/v1/hosts/:id/ssh/exec`.
- Authentication supports both password and ssh key.
- UI session is command-oriented (request/response), not PTY streaming in this phase.

## Backend Changes

- `internal/service/host/handler/host_exec.go`
  - Added private key loading path from `ssh_keys` when `node.ssh_key_id` exists.
  - Added encrypted key decryption (`utils.DecryptText`) using `security.encryption_key`.
  - `SSHCheck/SSHExec/BatchExec` now all support password + key auth.

## Security Notes

- Private key never returned to frontend.
- Decryption happens server-side only for runtime SSH dial.
- Existing API remains under JWT middleware; RBAC remains host write scope.

## Next Phase

- Upgrade to websocket PTY stream for interactive shells (vim/top/cursor navigation).
- Add command audit trail with operator + host + command hash + timestamp.
