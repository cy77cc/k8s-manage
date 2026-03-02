# Capability: deployment-target-management

## Purpose
This capability covers the creation, configuration, and management of deployment targets, including runtime selection, resource binding, and environment bootstrap.

## Requirements

### Requirement: Deployment Target List
The system SHALL display all deployment targets grouped by environment with their runtime type, status, and readiness.

#### Scenario: View all deployment targets
- **WHEN** user navigates to deployment targets page
- **THEN** system displays targets grouped by environment (Production, Staging, Development) with runtime type and status

#### Scenario: Filter by environment
- **WHEN** user selects an environment filter
- **THEN** system displays only targets for the selected environment

#### Scenario: Filter by runtime type
- **WHEN** user selects a runtime filter (K8s, Compose)
- **THEN** system displays only targets matching the selected runtime type

### Requirement: Create Deployment Target
The system SHALL provide a wizard for creating deployment targets with runtime selection and resource binding.

#### Scenario: Start target creation
- **WHEN** user clicks "Create Target" button
- **THEN** system displays a wizard with steps: basic info, runtime selection, resource binding, environment bootstrap

#### Scenario: Configure basic information
- **WHEN** user is on basic info step
- **THEN** system prompts for target name, environment (Development/Staging/Production), project, and team

#### Scenario: Select K8s runtime
- **WHEN** user selects "K8s" as runtime type
- **THEN** system prompts for cluster selection or credential selection

#### Scenario: Select Compose runtime
- **WHEN** user selects "Compose" as runtime type
- **THEN** system prompts for host node selection (at least one required)

#### Scenario: Bind platform-managed cluster
- **WHEN** user selects a platform-managed cluster for K8s target
- **THEN** system sets cluster_source to "platform_managed" and binds the cluster ID

#### Scenario: Bind external cluster via credential
- **WHEN** user selects an external credential for K8s target
- **THEN** system sets cluster_source to "external_managed" and binds the credential ID

#### Scenario: Bind hosts for Compose target
- **WHEN** user selects host nodes for Compose target
- **THEN** system validates hosts are available and have valid IP addresses

#### Scenario: Submit target creation
- **WHEN** user completes all wizard steps and clicks "Create"
- **THEN** system creates the deployment target with readiness_status "unknown"

### Requirement: Environment Bootstrap
The system SHALL provide automated environment initialization for deployment targets with runtime package installation.

#### Scenario: Start environment bootstrap
- **WHEN** user initiates bootstrap for a deployment target
- **THEN** system displays bootstrap wizard with runtime package version selection

#### Scenario: Select runtime package version
- **WHEN** user is selecting runtime package
- **THEN** system displays available versions with descriptions (e.g., "Kubernetes v1.28.0 with Calico CNI")

#### Scenario: Validate package manifest
- **WHEN** user selects a runtime package version
- **THEN** system validates the manifest file exists and checksum matches

#### Scenario: Execute preflight checks
- **WHEN** bootstrap process starts
- **THEN** system executes preflight checks on all target hosts and displays results

#### Scenario: Preflight check success
- **WHEN** all preflight checks pass
- **THEN** system proceeds to installation phase

#### Scenario: Preflight check failure
- **WHEN** any preflight check fails
- **THEN** system displays error details and halts bootstrap process

#### Scenario: Install runtime packages
- **WHEN** preflight checks pass
- **THEN** system installs runtime packages on all target hosts and displays real-time progress

#### Scenario: Installation progress tracking
- **WHEN** installation is in progress
- **THEN** system displays current phase, progress percentage, and live logs for each host

#### Scenario: Verify installation
- **WHEN** installation completes
- **THEN** system executes verification checks to ensure runtime is properly installed

#### Scenario: Bootstrap success
- **WHEN** all bootstrap phases complete successfully
- **THEN** system updates target readiness_status to "ready" and stores bootstrap job ID

#### Scenario: Bootstrap failure with rollback
- **WHEN** installation or verification fails
- **THEN** system executes rollback/uninstall scripts and updates status to "failed"

### Requirement: Target Detail View
The system SHALL display detailed information about a deployment target including configuration, bound resources, and deployment history.

#### Scenario: View target details
- **WHEN** user clicks on a deployment target
- **THEN** system displays target name, runtime type, environment, status, readiness status, and creation date

#### Scenario: View bound cluster for K8s target
- **WHEN** user views a K8s deployment target
- **THEN** system displays the bound cluster name, endpoint, and cluster source (platform/external)

#### Scenario: View bound hosts for Compose target
- **WHEN** user views a Compose deployment target
- **THEN** system displays all bound host nodes with IP, role, weight, and status

#### Scenario: View deployment history
- **WHEN** user views target details
- **THEN** system displays recent deployments to this target with service name, version, status, and timestamp

#### Scenario: View resource usage
- **WHEN** user views target details
- **THEN** system displays resource utilization (CPU, memory, storage) for the target

### Requirement: Target Readiness Status
The system SHALL track and display the readiness status of deployment targets.

#### Scenario: Unknown readiness
- **WHEN** target is created without bootstrap
- **THEN** system sets readiness_status to "unknown"

#### Scenario: Bootstrap pending
- **WHEN** target has a bootstrap job ID but job is not completed
- **THEN** system sets readiness_status to "bootstrap_pending"

#### Scenario: Ready status
- **WHEN** target's bootstrap job completes successfully
- **THEN** system sets readiness_status to "ready"

#### Scenario: Not ready status
- **WHEN** target fails readiness checks
- **THEN** system sets readiness_status to "not_ready" with error details

#### Scenario: Prevent deployment to non-ready targets
- **WHEN** user attempts to deploy to a target with readiness_status not "ready" or "unknown"
- **THEN** system rejects the deployment with error message

### Requirement: Target Node Management
The system SHALL allow management of host nodes bound to deployment targets.

#### Scenario: Add nodes to target
- **WHEN** user adds host nodes to an existing Compose target
- **THEN** system validates hosts and creates target_node bindings

#### Scenario: Remove nodes from target
- **WHEN** user removes host nodes from a target
- **THEN** system deletes the target_node bindings

#### Scenario: Update node role
- **WHEN** user changes a node's role (manager/worker)
- **THEN** system updates the target_node record

#### Scenario: Update node weight
- **WHEN** user changes a node's weight for load balancing
- **THEN** system updates the target_node weight value

### Requirement: Target Validation
The system SHALL validate deployment target configuration before creation and updates.

#### Scenario: K8s target requires cluster or credential
- **WHEN** user creates K8s target without cluster_id or credential_id
- **THEN** system returns validation error

#### Scenario: Compose target requires hosts
- **WHEN** user creates Compose target without any host nodes
- **THEN** system returns validation error

#### Scenario: Validate host availability
- **WHEN** user binds a host to Compose target
- **THEN** system validates host status is not "offline", "error", or "inactive"

#### Scenario: Validate host IP exists
- **WHEN** user binds a host to target
- **THEN** system validates host has a non-empty IP address

### Requirement: Bootstrap Job Tracking
The system SHALL track environment bootstrap jobs with detailed step-by-step progress.

#### Scenario: Create bootstrap job
- **WHEN** environment bootstrap starts
- **THEN** system creates a job record with unique ID, status "queued", and target information

#### Scenario: Track bootstrap steps
- **WHEN** bootstrap executes on each host
- **THEN** system creates step records for each phase (preflight, install, verify) per host

#### Scenario: Store step output
- **WHEN** each bootstrap step completes
- **THEN** system stores the command output (truncated to 2000 chars) and error messages

#### Scenario: Query bootstrap job status
- **WHEN** user requests bootstrap job details
- **THEN** system returns job status, all step records, and completion timestamps

### Requirement: RBAC Protection
The system SHALL require appropriate permissions for deployment target management operations.

#### Scenario: Unauthorized target creation
- **WHEN** user without deployment target management permission attempts to create a target
- **THEN** system returns 403 Forbidden error

#### Scenario: Unauthorized bootstrap initiation
- **WHEN** user without environment bootstrap permission attempts to start bootstrap
- **THEN** system returns 403 Forbidden error

#### Scenario: Authorized operations
- **WHEN** user with appropriate permissions performs target operations
- **THEN** system allows the operation and logs the action
