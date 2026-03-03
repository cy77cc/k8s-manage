## ADDED Requirements

### Requirement: Service Configuration Inline Editing
The system SHALL provide inline editing for service configuration without requiring a modal dialog.

#### Scenario: View service configuration
- **WHEN** user navigates to service detail page and selects the "配置" tab
- **THEN** the system MUST display configuration cards in read-only mode
- **AND** each card MUST have an "编辑" button to enable editing

#### Scenario: Edit service configuration inline
- **WHEN** user clicks the "编辑" button on a configuration card
- **THEN** the system MUST switch that card to edit mode
- **AND** display form fields for the configuration content
- **AND** display "保存" and "取消" buttons

#### Scenario: Save inline configuration changes
- **WHEN** user modifies configuration and clicks "保存"
- **THEN** the system MUST validate the configuration
- **AND** persist changes to the database
- **AND** switch the card back to read-only mode
- **AND** display a success message

#### Scenario: Cancel inline configuration changes
- **WHEN** user clicks "取消" after modifying configuration
- **THEN** the system MUST discard unsaved changes
- **AND** restore the original values
- **AND** switch the card back to read-only mode

### Requirement: Service Configuration Tab Structure
The system SHALL organize service configuration into logical sections without deployment target settings.

#### Scenario: Configuration tab content
- **WHEN** user views the "配置" tab of a service
- **THEN** the system MUST display:
  - 服务配置卡片 (basic info, runtime, labels, config content)
  - 环境变量集卡片 (environment variables)
  - 渲染输出预览 (rendered YAML preview)
- **AND** the system MUST NOT display any "部署目标" or cluster selection UI

### Requirement: Service List Edit Configuration Navigation
The system SHALL navigate to service detail configuration tab when editing from the list page.

#### Scenario: Click edit configuration from list
- **WHEN** user clicks "编辑配置" from the service list dropdown menu
- **THEN** the system MUST navigate to `/services/:id?tab=config`
- **AND** activate the "配置" tab on the detail page

#### Scenario: Direct URL access to configuration tab
- **WHEN** user navigates directly to `/services/:id?tab=config`
- **THEN** the system MUST load the service detail page
- **AND** automatically activate the "配置" tab

## REMOVED Requirements

### Requirement: Default Deployment Target Storage
**Reason**: Deployment targets should be selected at deploy time, not stored as service configuration.

**Migration**: Existing `default_target_id` and `default_deployment_target_id` fields in Service table will be retained for backward compatibility but will no longer be used. The `ServiceDeployTarget` table remains for historical data queries.

#### Scenario: Existing default target ignored
- **WHEN** a service has `default_target_id` set
- **THEN** the system MUST ignore this value during deployment
- **AND** require explicit `cluster_id` parameter in deploy request
