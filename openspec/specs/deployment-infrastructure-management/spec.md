# Capability: deployment-infrastructure-management

## Purpose
This capability covers the management of infrastructure resources required for deployment operations, including Kubernetes clusters, credentials, and host pools.

## Requirements

### Requirement: Cluster List and Overview
The system SHALL display a list of all Kubernetes clusters with their current status, resource utilization, and management mode.

#### Scenario: View cluster list
- **WHEN** user navigates to the cluster management page
- **THEN** system displays all clusters with name, status, node count, and management mode

#### Scenario: Filter clusters by status
- **WHEN** user selects a status filter (active, inactive, error)
- **THEN** system displays only clusters matching the selected status

### Requirement: Cluster Creation Wizard
The system SHALL provide a multi-step wizard for creating new Kubernetes clusters through automated bootstrap.

#### Scenario: Start cluster creation
- **WHEN** user clicks "Create Cluster" button
- **THEN** system displays a wizard with steps: basic info, control plane selection, worker node selection, CNI configuration

#### Scenario: Select control plane host
- **WHEN** user is on the control plane selection step
- **THEN** system displays available hosts with resource information and allows selection of one host

#### Scenario: Select worker nodes
- **WHEN** user is on the worker node selection step
- **THEN** system displays available hosts and allows selection of multiple worker nodes

#### Scenario: Configure CNI plugin
- **WHEN** user is on the CNI configuration step
- **THEN** system provides options for CNI plugins (Calico, Flannel, etc.) with descriptions

#### Scenario: Submit cluster creation
- **WHEN** user completes all wizard steps and clicks "Create"
- **THEN** system initiates cluster bootstrap process and displays progress

### Requirement: Cluster Bootstrap Progress Tracking
The system SHALL display real-time progress of cluster bootstrap operations with detailed phase information.

#### Scenario: View bootstrap progress
- **WHEN** cluster bootstrap is in progress
- **THEN** system displays current phase (preflight, install, verify) with progress percentage

#### Scenario: View bootstrap logs
- **WHEN** user views bootstrap progress
- **THEN** system displays real-time logs from the bootstrap process

#### Scenario: Bootstrap completion
- **WHEN** cluster bootstrap completes successfully
- **THEN** system updates cluster status to "active" and displays success message

#### Scenario: Bootstrap failure
- **WHEN** cluster bootstrap fails
- **THEN** system displays error details and provides rollback option

### Requirement: Cluster Detail View
The system SHALL display detailed information about a specific cluster including nodes, resources, and deployment history.

#### Scenario: View cluster details
- **WHEN** user clicks on a cluster from the list
- **THEN** system displays cluster details including endpoint, nodes, resource usage, and recent deployments

#### Scenario: View cluster nodes
- **WHEN** user views cluster details
- **THEN** system displays all nodes with their roles (control-plane, worker), status, and resource allocation

### Requirement: Credential Management
The system SHALL allow users to manage cluster credentials for both platform-managed and external clusters.

#### Scenario: List credentials
- **WHEN** user navigates to credential management page
- **THEN** system displays all credentials with name, type, source, and last test status

#### Scenario: Register platform credential
- **WHEN** user selects a platform-managed cluster and clicks "Register Credential"
- **THEN** system automatically extracts kubeconfig from the cluster and creates a credential record

#### Scenario: Import external credential via kubeconfig
- **WHEN** user clicks "Import External Credential" and provides kubeconfig file
- **THEN** system validates the kubeconfig and creates a credential record with source "external_managed"

#### Scenario: Import external credential via certificates
- **WHEN** user clicks "Import External Credential" and provides endpoint, CA cert, client cert, and client key
- **THEN** system validates the certificates and creates a credential record

#### Scenario: Test credential connectivity
- **WHEN** user clicks "Test Connection" on a credential
- **THEN** system attempts to connect to the cluster and displays result with latency

#### Scenario: Successful connectivity test
- **WHEN** credential connectivity test succeeds
- **THEN** system updates last_test_status to "ok" and displays success message with latency

#### Scenario: Failed connectivity test
- **WHEN** credential connectivity test fails
- **THEN** system updates last_test_status to "failed" and displays error message

### Requirement: Credential Security
The system SHALL encrypt all sensitive credential data (kubeconfig, certificates, tokens) before storage.

#### Scenario: Store encrypted credentials
- **WHEN** user creates or imports a credential
- **THEN** system encrypts all sensitive fields using the configured encryption key before database storage

#### Scenario: Retrieve credentials for use
- **WHEN** system needs to use a credential for cluster operations
- **THEN** system decrypts the credential data using the encryption key

### Requirement: Host Pool Management
The system SHALL display and manage the pool of available hosts for cluster and deployment target operations.

#### Scenario: View host pool
- **WHEN** user navigates to host management page
- **THEN** system displays all hosts with IP, status, resource capacity, and current assignments

#### Scenario: Filter available hosts
- **WHEN** user filters hosts by status "available"
- **THEN** system displays only hosts that are not currently assigned to any cluster or target

#### Scenario: View host assignments
- **WHEN** user views a host's details
- **THEN** system displays which clusters or deployment targets the host is assigned to

### Requirement: RBAC Protection
The system SHALL require appropriate permissions for all infrastructure management operations.

#### Scenario: Unauthorized cluster creation
- **WHEN** user without cluster management permission attempts to create a cluster
- **THEN** system returns 403 Forbidden error

#### Scenario: Unauthorized credential access
- **WHEN** user without credential management permission attempts to view credentials
- **THEN** system returns 403 Forbidden error

#### Scenario: Authorized operations
- **WHEN** user with appropriate permissions performs infrastructure operations
- **THEN** system allows the operation and logs the action
