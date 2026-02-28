# Rollout Plan: Enterprise DevOps Console UI Redesign

## Phase 1
- Enable dark-first design tokens and shell layout.
- Verify sidebar collapse, top health bar, and role-aware menu visibility.

## Phase 2
- Roll out monitoring-first dashboard and monitor tables with skeleton loading.
- Enable search/filter/sort on core operational tables.

## Phase 3
- Expand to service management table UX and feedback consistency.
- Execute regression and contrast checklist before full promotion.

## Rollback
- Set `VITE_UI_THEME_LEGACY=true` to revert to legacy visual shell.
- Keep route mappings unchanged to avoid navigation disruption.
