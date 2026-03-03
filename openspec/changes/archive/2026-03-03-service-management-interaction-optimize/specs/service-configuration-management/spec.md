# service-configuration-management Specification (Delta)

## ADDED Requirements

### Requirement: Service List View Mode Toggle
The system SHALL provide card and list view modes for the service list page with automatic default selection.

#### Scenario: Default view mode based on service count
- **WHEN** user navigates to the service list page
- **AND** the service count is less than or equal to 8
- **THEN** the system MUST display services in card view by default

#### Scenario: Default list view for many services
- **WHEN** user navigates to the service list page
- **AND** the service count is greater than 8
- **THEN** the system MUST display services in list view by default

#### Scenario: User toggles view mode
- **WHEN** user clicks the view mode toggle (卡片/list)
- **THEN** the system MUST switch to the selected view mode
- **AND** persist the user's preference to localStorage

#### Scenario: View mode preference restored
- **WHEN** user returns to the service list page
- **AND** user has previously selected a view mode
- **THEN** the system MUST restore the user's preferred view mode
- **AND** ignore the automatic default logic

### Requirement: Service Card Clickable Area
The system SHALL make the entire service card clickable for navigation to detail page.

#### Scenario: Click on card body navigates to detail
- **WHEN** user clicks anywhere on a service card
- **AND** the click is not on checkbox, dropdown, or other interactive element
- **THEN** the system MUST navigate to the service detail page `/services/:id`

#### Scenario: Checkbox selection does not trigger navigation
- **WHEN** user clicks on the checkbox within a service card
- **THEN** the system MUST toggle the checkbox selection
- **AND** MUST NOT navigate to the detail page

#### Scenario: Dropdown menu does not trigger navigation
- **WHEN** user clicks on the "更多" dropdown button
- **THEN** the system MUST open the dropdown menu
- **AND** MUST NOT navigate to the detail page

### Requirement: Service List Table View
The system SHALL provide a table view for services with sortable columns and row actions.

#### Scenario: Table view displays service information
- **WHEN** user selects list view mode
- **THEN** the system MUST display a table with columns: 服务名, 状态, 环境, 运行时, 负责人, 标签, 操作

#### Scenario: Click service name in table
- **WHEN** user clicks on a service name in the table
- **THEN** the system MUST navigate to the service detail page

#### Scenario: Table row actions
- **WHEN** user views the table
- **THEN** each row MUST display action buttons: 启动, 停止, 删除

### Requirement: Project Selection in Service Creation
The system SHALL display project name instead of ID and control editability based on user permissions.

#### Scenario: Display project name in form
- **WHEN** user opens the service creation page
- **THEN** the system MUST display the current project name
- **AND** MUST NOT display the project ID

#### Scenario: User with project switch permission
- **WHEN** user has permission to switch projects
- **THEN** the system MUST display a project dropdown selector
- **AND** allow the user to select from available projects

#### Scenario: User without project switch permission
- **WHEN** user does NOT have permission to switch projects
- **THEN** the system MUST display the current project name in read-only mode
- **AND** MUST NOT allow changing the project

#### Scenario: Team ID field hidden
- **WHEN** user opens the service creation page
- **THEN** the system MUST NOT display the team ID field
- **AND** the system MUST automatically fill team_id from context

### Requirement: Service Creation Page Layout
The system SHALL display the back button at the top of the page with optimized button styling.

#### Scenario: Back button at page top
- **WHEN** user opens the service creation page
- **THEN** the system MUST display a "返回" button at the top-left of the page
- **AND** the button MUST be outside the main content card

#### Scenario: Back button navigation
- **WHEN** user clicks the "返回" button
- **THEN** the system MUST navigate to the service list page

#### Scenario: Button styling optimization
- **WHEN** user views the service creation form
- **THEN** the "创建服务" button MUST use standard button styling
- **AND** MUST NOT use oversized primary button style

### Requirement: Chinese Localization for Service Creation
The system SHALL display all UI text in Chinese for the service creation page.

#### Scenario: Page title in Chinese
- **WHEN** user opens the service creation page
- **THEN** the system MUST display "服务工作室" as the page title

#### Scenario: Editor section labels in Chinese
- **WHEN** user views the editor section
- **THEN** the system MUST display:
  - "编辑器" for Editor
  - "预览" for Preview
  - "模板变量" for Template Variables
  - "诊断信息" for Diagnostics

#### Scenario: Form labels in Chinese
- **WHEN** user views the service creation form
- **THEN** the system MUST display:
  - "服务端口" for Service Port
  - "容器端口" for Container Port
  - "内存" for Memory

#### Scenario: Status tags in Chinese
- **WHEN** user views preview status tags
- **THEN** the system MUST display:
  - "渲染中" for rendering
  - "就绪" for ready
  - "未解析" for unresolved
  - "必填" for required

#### Scenario: Tab labels in Chinese
- **WHEN** user views the preview tabs
- **THEN** the system MUST display:
  - "K8s 配置" for K8s YAML
  - "Compose 配置" for Compose YAML

## MODIFIED Requirements

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

#### Scenario: Dropdown menu does not have redundant view detail option
- **WHEN** user opens the service card dropdown menu
- **THEN** the system MUST NOT display "查看详情" option
- **AND** the system MUST display only: "编辑配置", "启动服务", "停止服务", "删除服务"

## REMOVED Requirements

None.
