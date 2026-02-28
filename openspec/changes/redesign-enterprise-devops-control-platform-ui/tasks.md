## 1. Design Tokens and Theme Foundation

- [x] 1.1 Define dark-first enterprise design tokens (deep blue/cool gray, accent blue, radius 10px, light shadow) in `web/src/styles` and shared theme config
- [x] 1.2 Wire tokens into Ant Design ConfigProvider and Tailwind extensions to remove page-local hardcoded palette drift
- [x] 1.3 Add/adjust shared primitives for card containers, status chips, and shell surfaces using unified token APIs
- [x] 1.4 Add visual baseline checklist for semantic status colors and contrast in dark mode

## 2. App Shell and Information Architecture

- [x] 2.1 Refactor app shell to support collapsible left sidebar, top health status bar, and stable content viewport structure
- [x] 2.2 Reorganize navigation into task-oriented groups aligned to monitoring and incident handling workflows
- [x] 2.3 Ensure route mapping compatibility and preserve role-aware menu visibility in both expanded/collapsed states
- [x] 2.4 Add/update tests for governance navigation visibility and unauthorized route guard behavior after IA changes

## 3. Monitoring-First Main Content Experience

- [x] 3.1 Implement 12-column responsive dashboard layout prioritizing monitoring charts in upper region
- [x] 3.2 Place service/alert operational tables in lower region with clear visual hierarchy and action density control
- [x] 3.3 Introduce Skeleton loading states for key dashboard widgets, tables, and chart containers
- [x] 3.4 Add empty/error state patterns consistent with the redesigned dark UI language

## 4. Operational UX Enhancements

- [x] 4.1 Standardize table capabilities for core operational pages (search, filter, sorting)
- [x] 4.2 Enforce secondary confirmation for dangerous actions in service/governance/ops mutation flows
- [x] 4.3 Standardize toast feedback for success/failure outcomes across high-frequency operations
- [x] 4.4 Implement controlled real-time refresh strategy (interval + visibility-aware behavior) for monitor-centric pages

## 5. Responsiveness, Accessibility, and Validation

- [x] 5.1 Tune responsive breakpoints for shell, dashboard grid, and operational tables across desktop/tablet/mobile widths
- [x] 5.2 Verify keyboard navigation order and focus visibility in shell, tables, and critical action controls
- [x] 5.3 Validate color contrast/readability under dark theme for primary and secondary content layers
- [x] 5.4 Run frontend regression tests and OpenSpec validation, then document rollout and rollback notes for release
