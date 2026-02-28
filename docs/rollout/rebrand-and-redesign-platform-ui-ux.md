# Rebrand and UI/UX Redesign Rollout Plan

## Phase 1: Foundation

- Merge brand assets, brand config, global design tokens, and app shell updates.
- Enable new UI theme by default (`data-ui-theme=nebula`).
- Keep rollback switch available via `VITE_UI_THEME_LEGACY=true`.

## Phase 2: Governance and High-Frequency Flows

- Confirm explicit action patterns on users/roles/permissions pages.
- Validate role-aware visibility in grouped navigation.
- Run targeted regression suites for layout and governance page interactions.

## Phase 3: Polish and Stabilization

- Complete visual checklist and accessibility spot check.
- Collect stakeholder feedback on navigation IA and branding.
- Freeze palette/token APIs for downstream page migrations.

## Rollback Strategy

- Immediate rollback: set env `VITE_UI_THEME_LEGACY=true` and redeploy frontend.
- If navigation issues occur, temporarily disable governance menu grouping via existing permission-driven rendering and route protection.
- Keep all route mappings unchanged, ensuring users can continue task access paths after rollback.
