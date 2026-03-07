## 1. Define control-plane contracts

- [x] 1.1 Define structured task-lifecycle contracts in `internal/ai` for objective, plan, step, execution record, evidence, replan decision, and next actions
- [x] 1.2 Introduce an orchestration state model that supports planning, execution, interruption, replanning, and final outcome transitions
- [x] 1.3 Align existing approval, confirmation, preview, session, and execution state dependencies with the new control-plane contracts

## 2. Introduce planner, executor, and replanner runtime

- [x] 2.1 Add explicit Planner runtime under `internal/ai` that emits domain-level operational steps instead of direct tool-call plans
- [x] 2.2 Add explicit Executor runtime under `internal/ai` that consumes plan steps and produces execution records, evidence, and issues
- [x] 2.3 Add explicit Replanner runtime under `internal/ai` that turns execution outcomes into bounded control-plane decisions
- [x] 2.4 Rework the current orchestration entrypoint so chat and resume flows run through the new control-plane runtime instead of a single implicit platform-agent loop

## 3. Add domain executor routing

- [x] 3.1 Introduce a domain executor router in `internal/ai` for step-to-domain dispatch
- [x] 3.2 Implement Host executor boundaries and route host-related plan steps through them
- [x] 3.3 Implement K8s executor boundaries and route k8s-related plan steps through them
- [x] 3.4 Implement Service executor boundaries and route service-related plan steps through them
- [x] 3.5 Implement Monitor executor boundaries and route monitor-related plan steps through them

## 4. Introduce card-oriented platform events

- [x] 4.1 Define the backend platform event family for `plan_created`, `step_status`, `evidence`, `ask_user`, `approval_required`, `replan_decision`, `summary`, and `next_actions`
- [x] 4.2 Add event projection logic in `internal/ai` so orchestration state is emitted as structured platform events
- [x] 4.3 Keep SSE compatibility projection in `internal/service/ai` while exposing the richer event semantics from the AI core
- [x] 4.4 Update gateway-facing stream handling so current consumers continue to function during the transition to application-card events

## 5. Integrate gateway compatibility and control-plane dependencies

- [x] 5.1 Rewire `internal/service/ai` handlers to delegate all AI task lifecycle execution through the new control-plane runtime
- [x] 5.2 Preserve `/api/v1/ai` route compatibility and approval/resume behavior while internal orchestration changes
- [x] 5.3 Ensure tool preview, execution query, session management, and approval endpoints continue to work against the refactored control-plane runtime

## 6. Verify AIOps control-plane behavior

- [x] 6.1 Add backend tests for planner/executor/replanner contracts and task-state transitions
- [x] 6.2 Add backend tests for Host, K8s, Service, and Monitor executor routing
- [x] 6.3 Add backend tests for approval, confirmation, interruption, and resume flows through the new control-plane runtime
- [x] 6.4 Add backend tests for platform event projection and SSE compatibility behavior
- [x] 6.5 Verify the resulting architecture leaves `internal/ai` as the clear host for future supervisor and deeper domain workflow extensions
