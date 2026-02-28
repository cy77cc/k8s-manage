## 1. Brand Naming and Logo Foundations

- [x] 1.1 Finalize candidate platform names and select one canonical name for product surfaces
- [x] 1.2 Create logo asset pack (primary, simplified, monochrome) and add source/export files to frontend assets
- [x] 1.3 Integrate canonical brand name and logo assets into core entry points (login, top nav, page title, app shell)
- [x] 1.4 Add brand usage constraints (minimum size, clear space, aspect ratio handling) in shared UI utilities or component wrappers

## 2. Design Token and Theme System

- [x] 2.1 Define centralized design tokens for color, typography, spacing, radius, shadow, and semantic states in `web/src/styles`
- [x] 2.2 Wire tokens into Ant Design/Tailwind theme configuration used by shared layouts and components
- [x] 2.3 Refactor common components to consume tokens instead of page-local hardcoded styles
- [x] 2.4 Add visual regression checklist for key semantic states (success, warning, error, info)

## 3. Global Layout and Navigation Redesign

- [x] 3.1 Redesign app shell layout (header, sidebar/navigation, content container) with the new visual language
- [x] 3.2 Reorganize navigation into task-oriented groups while preserving existing route mapping
- [x] 3.3 Implement role-aware navigation rendering against effective roles/permissions for new IA structure
- [x] 3.4 Validate unauthorized navigation entries and protected actions remain hidden for non-privileged users

## 4. Governance UX Alignment

- [x] 4.1 Update Users, Roles, and Permissions pages to explicit and discoverable action patterns under the new design system
- [x] 4.2 Ensure role list retains explicit `View Details` and `Edit Permissions` controls in redesigned table layouts
- [x] 4.3 Align governance page feedback flows (loading/empty/error/success) with global interaction standards
- [x] 4.4 Verify keyboard navigation and contrast readability on redesigned governance workflows

## 5. Validation, Rollout, and Fallback

- [x] 5.1 Execute frontend regression checks for core flows and role-aware visibility matrix
- [x] 5.2 Perform OpenSpec validation and resolve any schema/spec consistency issues
- [x] 5.3 Prepare phased rollout plan (foundation -> core pages -> polish) with rollback toggle to previous shell/theme
- [x] 5.4 Publish release notes summarizing brand change, navigation updates, and user-facing interaction differences
