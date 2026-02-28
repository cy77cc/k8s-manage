# Visual Regression Checklist (Dark Mode)

## Semantic States

- Success state uses `--color-success` consistently in tags, cards, and feedback.
- Warning state uses `--color-warning` consistently in tags, cards, and feedback.
- Error state uses `--color-error` consistently in tags, cards, and feedback.
- Info state uses `--color-info` consistently in tags, cards, and feedback.

## Contrast and Readability

- Primary text on dark surfaces remains readable in cards, tables, and modals.
- Secondary text remains readable without reducing critical information contrast.
- Interactive focus ring remains visible in dark table rows and action controls.

## Shell and Layout

- Sidebar collapse/expand keeps navigation labels and active route feedback clear.
- Top health status bar remains readable across viewport sizes.
- Dashboard uses monitoring-first 12-grid hierarchy with no overlap at tablet widths.

## Table and Feedback UX

- Search/filter/sort controls are usable in services and monitor tables.
- Empty table states use explicit placeholders, not blank space.
- Skeleton loading appears during async data fetch in dashboard/monitor/services.
- Dangerous-action confirmations appear before mutation-triggering flows.
