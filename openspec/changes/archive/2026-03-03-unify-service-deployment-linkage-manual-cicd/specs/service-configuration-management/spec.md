## MODIFIED Requirements

### Requirement: Service Configuration Tab Structure
The system SHALL organize service configuration into logical sections without embedding deployment target selection, and SHALL provide manual deployment actions by creating unified release requests instead of running an independent service-side deployment state machine.

#### Scenario: Configuration tab content
- **WHEN** user views the "配置" tab of a service
- **THEN** the system MUST display:
  - 服务配置卡片 (basic info, runtime, labels, config content)
  - 环境变量集卡片 (environment variables)
  - 渲染输出预览 (rendered YAML preview)
- **AND** the system MUST NOT display any "部署目标" or cluster selection UI in this tab

#### Scenario: Manual deploy action from service context
- **WHEN** user triggers manual deployment from service-related entry
- **THEN** the system MUST create a unified release request and return a release identifier for lifecycle tracking
- **AND** the system MUST NOT create a separate service-only deployment lifecycle record as the primary execution state
