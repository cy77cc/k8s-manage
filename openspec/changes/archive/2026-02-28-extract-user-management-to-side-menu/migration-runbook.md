# Governance Menu Migration Runbook

## Rollout (Task 5.2)
1. Backend-first deploy
- Deploy version containing `/api/v1/rbac/migration/events` and enhanced RBAC deny logging.
- Verify endpoint auth and 200 response with an authorized account.

2. Frontend rollout
- Enable `VITE_FEATURE_GOVERNANCE_MENU=true`.
- Verify sidebar shows `访问治理` for users with `rbac:read`.
- Verify `/settings/users|roles|permissions` redirects to `/governance/*`.

3. 30-day observation window
- Keep legacy redirect routes active for at least 30 days.
- Review metrics daily using queries below.

## Metrics Verification (Task 5.3)

Assumes app logs are centralized and searchable.

### 1) 403 rate (governance RBAC denied)
Search pattern:
- `rbac deny actor=`

Example shell query:
```bash
rg "rbac deny actor=" /var/log/k8s-manage/app.log | wc -l
```

### 2) Old-path traffic (`/settings/*` legacy entry)
Search pattern:
- `rbac migration event=legacy_redirect`

Example shell query:
```bash
rg "rbac migration event=legacy_redirect" /var/log/k8s-manage/app.log | wc -l
```

### 3) Governance task completion time
Search pattern:
- `rbac migration event=governance_task`
- extract `duration_ms=<n>` for percentile calculation

Example shell query:
```bash
rg "rbac migration event=governance_task" /var/log/k8s-manage/app.log \
  | sed -n 's/.*duration_ms=\([0-9]\+\).*/\1/p'
```

## Cutover Criteria (Task 5.4 gate)
- Legacy redirect traffic is near zero for 7 consecutive days.
- 403 rate stable within expected baseline.
- Governance task duration p95 within acceptable UX SLO.
- Product/security sign-off is recorded.

After criteria pass, remove legacy `/settings/{users,roles,permissions}` redirects.
