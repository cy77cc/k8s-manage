# AI Tool Call UX Rules (Product)

## User-facing Goals
- Tool invocation should feel reliable and explainable.
- Missing parameter failures must become recoverable guidance.

## UX Rules
- Show tool trace with:
  - selected tool
  - resolved params
  - whether retry happened
- When tool fails for missing params, assistant should provide next-step hint:
  - example: "缺少 host_id，请先选择目标主机。"
- Keep mutating actions behind approval UI.

## Copy Guidelines
- Prefer actionable, short Chinese messages.
- Do not expose raw stack traces.
