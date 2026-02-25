# Host Terminal UX Redesign (Phase-1)

## Goal

Replace mock terminal with production-facing real host terminal while keeping operation simple.

## Interaction Model

- Open terminal page -> run real SSH connectivity check -> show live status.
- User enters command and presses Enter -> backend executes command -> stdout/stderr appended.
- Command history supports ArrowUp/ArrowDown.
- Quick command panel provides common diagnostics.

## Visual Design

- Dark ops-console style.
- Left: large terminal output area.
- Right: quick commands and safety hints.
- Top: host identity + connection badge + reconnect/disconnect actions.
- Mid metrics: session duration, latency, command count, online/offline.

## Scope Boundary

- This phase is not full PTY, therefore no interactive TUI behavior.
- `clear` and `exit` are UI-level session actions.
