# Release Notes: Enterprise DevOps Console UI Redesign

## What changed
- Default dark-mode enterprise shell with low-saturation deep blue/cool gray palette.
- Sidebar + top health status bar + monitoring-first dashboard hierarchy.
- Services and monitor tables now support search/filter/sort consistently.
- Added skeleton loading and improved empty-state handling for key pages.
- Dangerous workflows continue to require explicit confirmation.

## Ops impact
- No backend API contract changes.
- Feature remains compatible with existing RBAC route protections.
- Legacy rollback available via `VITE_UI_THEME_LEGACY=true`.
