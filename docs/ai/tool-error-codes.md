# Tool Error Codes

## Input Validation
- `missing_param`: required field is missing.
- `invalid_param`: field value/type/enum is invalid.
- `param_conflict`: mutually exclusive or conflicting fields.

## Policy / Permission
- `policy_denied`: rejected by RBAC/policy/approval checks.

## Runtime
- `tool_error`: tool execution failed for non-input reasons.

## Notes
- Resolver may retry once automatically when error is `missing_param`.
- SSE trace includes `retry` and `param_resolution` for diagnostics.
