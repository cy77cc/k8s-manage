# AI Multi-Domain Manual Test Checklist

## Preconditions

- Set `ai.use_multi_domain_arch: true` in `configs/config.yaml`
- Start the backend and frontend normally
- Use an account that can access the AI assistant UI

## Scenarios

### 1. Simple chat remains unchanged
- Send a simple greeting like `你好`
- Confirm the reply follows the existing simple-chat path
- Confirm no new approval or planner-specific UI appears

### 2. Agentic request enters multi-domain path
- Send `部署支付服务到生产环境并检查告警`
- Confirm the assistant responds through the new multi-domain path
- Confirm the response references multiple domains or a multi-domain plan summary

### 3. Toggle fallback works
- Set `ai.use_multi_domain_arch: false`
- Repeat the same agentic request
- Confirm the request uses the legacy agentic path instead of the multi-domain path

### 4. SSE compatibility remains intact
- Open browser devtools network tab for the AI stream request
- Confirm the stream still completes successfully
- Confirm the frontend does not regress in rendering or interaction

## Expected Outcome

- Simple chat behavior is unchanged
- Agentic requests switch behavior only when the rollout toggle is enabled
- Frontend streaming remains compatible in both modes
