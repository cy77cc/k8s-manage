# Release Notes: NebulaOps Brand and UI Redesign

## Highlights

- Platform canonical name is now `NebulaOps`.
- New logo system added with primary, simplified, and monochrome variants.
- App shell redesigned with grouped task-oriented navigation and refreshed visual language.

## User-Facing Changes

- Updated login and application shell branding.
- Navigation reorganized into:
  - 资源与资产
  - 交付与变更
  - 可观测与工具
  - 访问治理 (permission-based)
- Governance pages keep explicit action entry points (`详情`, `编辑`, `编辑权限`, `复制`) for discoverability.

## Technical Notes

- New design tokens added under `web/src/styles/design-tokens.css`.
- Ant Design and Tailwind theme values now consume centralized token variables.
- Rollback toggle supported: set `VITE_UI_THEME_LEGACY=true`.

## Verification

- Layout permission visibility tests passed.
- Governance users interaction tests passed.
- OpenSpec artifacts and specs remain valid for apply/archive flow.
