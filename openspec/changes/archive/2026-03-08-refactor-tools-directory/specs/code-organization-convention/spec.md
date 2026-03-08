## ADDED Requirements

### Requirement: AI Tools Directory SHALL Be Organized By Domain

AI 工具目录 SHALL 按领域域分组，每个领域一个子目录，根目录只保留核心类型和注册入口。

**Threshold Rules:**
| Category | Max Files Per Directory | Action When Exceeded |
|----------|------------------------|----------------------|
| Tools Root | 10 | Group implementations into `impl/` subdirectory |
| Param Utilities | 8 | Keep in `param/` subdirectory |

**Recommended Structure:**
```
internal/ai/tools/
├── contracts.go           # Core types (ToolMeta, ToolResult, errors)
├── registry.go            # Tool registration (BuildLocalTools)
├── runner.go              # Execution logic (runWithPolicyAndEvent)
├── wrapper.go             # Risk-based wrappers
├── builder.go             # Tool building utilities
├── category.go            # Scene-based tool filtering
├── param/                 # Parameter handling
│   ├── hints.go
│   ├── resolver.go
│   └── validator.go
└── impl/                  # Tool implementations by domain
    ├── kubernetes/
    │   └── tools.go
    ├── host/
    │   └── tools.go
    ├── service/
    │   └── tools.go
    ├── monitor/
    │   └── tools.go
    ├── cicd/
    │   └── tools.go
    ├── deployment/
    │   └── tools.go
    ├── governance/
    │   └── tools.go
    ├── infrastructure/
    │   └── tools.go
    └── mcp/
        ├── client.go
        └── proxy.go
```

#### Scenario: tools directory exceeds threshold

- **GIVEN** `internal/ai/tools/` contains more than 10 Go files at root level
- **WHEN** maintainers organize the codebase
- **THEN** tool implementations SHALL be moved to `impl/<domain>/` subdirectories
- **AND** each subdirectory SHALL contain related tools for one domain

#### Scenario: new domain tools are added

- **GIVEN** a new domain of AI tools needs to be implemented
- **WHEN** adding the tool implementation
- **THEN** a new subdirectory SHALL be created under `impl/` if the domain doesn't exist
- **AND** the tool file SHALL be placed in the corresponding domain subdirectory

---

### Requirement: Tool Implementation Files SHALL Follow Naming Convention

工具实现文件 SHALL 遵循命名规范，文件名反映工具领域。

**Naming Rules:**
- Implementation files: `tools.go` within each `impl/<domain>/` directory
- Core files: descriptive names (`contracts.go`, `registry.go`, `runner.go`)
- Parameter utilities: `param/<purpose>.go`

#### Scenario: naming a new tool file

- **GIVEN** a developer is creating a new tool implementation
- **WHEN** saving the file
- **THEN** the file SHALL be named `tools.go` and placed in the appropriate `impl/<domain>/` directory
- **AND** the package SHALL be named after the domain (e.g., `package kubernetes`)

---

### Requirement: Core Tool Types SHALL Remain in Root Package

核心工具类型 SHALL 保留在根包，确保实现包可以依赖统一类型定义。

**Types in Root Package:**
- `ToolMeta`, `ToolMode`, `ToolRisk`
- `ToolResult`, `ToolExecutionError`
- `ApprovalRequiredError`, `ConfirmationRequiredError`
- `PlatformDeps`, `RegisteredTool`
- Context helpers (`WithToolPolicyChecker`, `EmitToolEvent`)

#### Scenario: implementation package imports core types

- **GIVEN** a tool implementation in `impl/kubernetes/tools.go`
- **WHEN** the implementation needs `ToolMeta` or `ToolResult`
- **THEN** it SHALL import from `github.com/cy77cc/k8s-manage/internal/ai/tools`
- **AND** use the types directly without redefining
