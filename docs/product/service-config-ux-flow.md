# Service Config UX Flow

## 1. Create Flow

1. Fill basic metadata: project/team/name/env/owner/type/labels.
2. Choose config mode:
   - `standard`: fill image/replicas/ports/resources/env
   - `custom`: edit YAML directly
3. Use real-time preview (`k8s`/`compose`) and diagnostics.
4. Optional: convert standard to custom with one click.
5. Submit service in draft status.

## 2. List Flow

- filter by `team/runtime/env/label_selector/query`.
- show runtime, config mode, tags, owner, status.

## 3. Detail Flow

- inspect ownership, labels, config mode, rendered output.
- deploy with target selection (`k8s|compose|helm`).
- helm import/render/deploy available in detail page.

## 4. Error UX

- preview/render errors returned as `diagnostics[]`.
- deployment permission failures return explicit backend message (`service:deploy`, `service:approve`).
