# AI Module Rewrite - Tasks

## Phase 1: Core Framework

### 1.1 IntentRouter

- [x] Create `internal/ai/router/` directory structure
- [x] Implement `IntentClassifier` with domain classification
- [x] Implement `IntentRouter` using `compose.Graph`
- [x] Add domain routing configuration
- [x] Write unit tests for classification

### 1.2 ActionGraph

- [x] Create `internal/ai/graph/` directory structure
- [x] Implement `ActionGraph` using `compose.Workflow`
- [x] Implement `sanitizeNode` Lambda
- [x] Implement `reasoningNode` with ChatModel
- [x] Implement `validationNode` with K8s OpenAPI validation
- [x] Implement `executionNode` with ToolsNode
- [x] Define `GraphState` and state handlers
- [x] Write unit tests for workflow

### 1.3 SecurityAspect

- [x] Create `internal/ai/aspect/` directory structure
- [x] Implement `SecurityAspect` with Eino Callbacks
- [x] Implement `PermissionChecker` interface
- [x] Implement `InterruptHandler` for risk operations
- [x] Implement `AuditLogger` for operation logging
- [x] Wire aspect into ToolsNode middleware
- [x] Write unit tests for aspect

### 1.4 State Management

- [x] Create `internal/ai/state/` directory structure
- [x] Implement `SessionState` with Redis backend
- [x] Implement `CheckpointStore` using `compose.CheckPointStore`
- [x] Integrate with ActionGraph workflow
- [x] Write unit tests for state management

### 1.5 Main Entry Point

- [x] Create `internal/ai/agent.go`
- [x] Implement `AIAgent` struct with router + graph composition
- [x] Implement `Query(ctx, sessionID, message)` method
- [x] Implement `Resume(ctx, sessionID, response)` method
- [x] Update `internal/service/ai/handler/` to use new agent
- [x] Fix compilation errors in handlers

## Phase 2: Approval System

### 2.1 Approval Task Model

- [x] Create `internal/model/ai/approval_task.go`
- [x] Add database migration
- [x] Implement CRUD operations

### 2.2 Task Generator

- [x] Create `internal/ai/approval/generator.go`
- [x] Implement LLM-based task detail generation
- [x] Implement `TaskDetail` struct with steps, risks, rollback
- [x] Write unit tests for generator

### 2.3 Approval Router

- [x] Create `internal/ai/approval/router.go`
- [x] Implement `ApprovalRouter` interface
- [x] Implement `ResourceOwnerRouter` based on resource type
- [x] Write unit tests for router

### 2.4 Approval Executor

- [x] Create `internal/ai/approval/executor.go`
- [x] Implement `ApprovalExecutor` for post-approval execution
- [x] Add execution result notification
- [x] Write unit tests for executor

### 2.5 Approval API

- [x] Update `internal/service/ai/routes.go`
- [x] Implement approval CRUD handlers
- [x] Implement approve/reject handlers
- [x] Add SSE notification for task updates
- [x] Update frontend API calls (if needed)

## Phase 3: RAG Enhancement

### 3.1 Knowledge Indexer

- [x] Create `internal/ai/rag/indexer.go`
- [x] Implement `Indexer` interface with Milvus backend
- [x] Support user-input knowledge entries
- [x] Add multi-tenant namespace filtering

### 3.2 Knowledge Retriever

- [x] Create `internal/ai/rag/retriever.go`
- [x] Implement `Retriever` interface
- [x] Integrate with ActionGraph reasoning node
- [x] Add context augmentation to prompts

### 3.3 Feedback Collector

- [x] Create `internal/ai/rag/feedback.go`
- [x] Implement `FeedbackCollector` interface
- [x] Implement Q&A extraction from session
- [x] Implement automatic indexing on positive feedback
- [x] Add feedback API endpoint

### 3.4 RAG Integration

- [x] Wire RAG into ActionGraph
- [x] Update prompts to include RAG context
- [x] Test retrieval quality

## Phase 4: Testing & Documentation

### 4.1 Integration Tests

- [x] Test full workflow: router → graph → execution
- [x] Test interrupt/resume flow
- [x] Test approval flow (both tracks)
- [x] Test RAG feedback flow

### 4.2 Documentation

- [x] Update `MEMORY.md` with new architecture
- [x] Create API documentation
- [x] Create troubleshooting guide

## Dependencies

- Phase 1 must complete before Phase 2
- Phase 2.1-2.3 must complete before Phase 2.4-2.5
- Phase 3 can run in parallel with Phase 2
- Phase 4 starts after Phase 1-3 complete
