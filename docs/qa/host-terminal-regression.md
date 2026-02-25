# Host Terminal Regression Checklist

## Functional

1. Open `/hosts/terminal/:id` shows real host name/IP from API.
2. Connection check success -> status becomes connected.
3. Execute `pwd`/`whoami` returns real host output.
4. Wrong command returns stderr with non-zero exit semantics.
5. Quick commands execute correctly.
6. Reconnect can recover from disconnected state.

## Credential Paths

1. Password-auth host can execute commands.
2. SSH-key host can execute commands.
3. Encrypted SSH key host can execute commands with valid encryption key.

## UX

1. Output panel auto-scrolls on new lines.
2. Up/Down history navigation works.
3. Error/disconnected alerts are visible and actionable.

## Build Gate

- `go test ./...`
- `cd web && npm run build`
