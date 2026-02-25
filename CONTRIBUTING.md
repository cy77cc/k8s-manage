# Contributing

## OpenSpec-First Rule

For any PR that introduces or changes product/platform capability behavior, you MUST update OpenSpec in the same PR.

Required actions:

1. Create or update a change under `openspec/changes/<change-name>/`.
2. Keep artifacts aligned (`proposal.md`, `design.md` when needed, `specs/*/spec.md`, `tasks.md`).
3. Mark completed tasks in `tasks.md` (`- [x]`) and keep pending tasks as `- [ ]`.
4. Run validation before merge:
   - `openspec validate --json`

Allowed exception:

- Pure refactor/chore/docs formatting with no behavior/capability impact may skip OpenSpec updates, but PR must explicitly state the reason.
