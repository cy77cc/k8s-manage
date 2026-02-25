# Service Management Regression Plan

## 1. Backend

- `preview`:
  - standard->k8s returns Deployment+Service YAML
  - standard->compose returns compose services block
  - custom invalid YAML returns diagnostics error
- `transform`:
  - standard->custom returns stable yaml and hash
- CRUD:
  - create/list/get/update/delete all return `code=1000`
  - list returns `data.list` and `data.total`
- deploy:
  - missing `service:deploy` denied
  - production env requires `service:approve`
- helm:
  - import success path
  - render with missing chart_ref returns diagnostics

## 2. Frontend

- provision page:
  - mode switch (`standard/custom`) works
  - preview updates after field changes
  - convert standard->custom populates editor
- list page:
  - filters work and data refreshes
- detail page:
  - deploy/rollback/events visible and callable
  - helm render result shown

## 3. Build Gates

- `go test ./...`
- `cd web && npm run build`
