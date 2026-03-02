## ADDED Requirements

### Requirement: Cluster creation with kubeadm automation
The system MUST provide fully automated Kubernetes cluster bootstrap using kubeadm on selected hosts via SSH.

#### Scenario: Bootstrap cluster on selected hosts
- **WHEN** an authorized operator submits a cluster bootstrap request with control plane host, worker hosts, K8s version, and CNI selection
- **THEN** the system MUST execute preflight checks, install containerd, install kubeadm/kubelet/kubectl, initialize control plane, install CNI, join workers, and store kubeconfig
- **AND** the system MUST report step-by-step progress and allow cancellation

#### Scenario: Bootstrap fails mid-process
- **WHEN** a bootstrap step fails after partial installation
- **THEN** the system MUST execute rollback scripts for completed steps and mark the task as failed with diagnostic output
- **AND** the system MUST preserve step-level logs for troubleshooting

### Requirement: External cluster import via kubeconfig
The system MUST allow importing existing Kubernetes clusters by providing kubeconfig credentials.

#### Scenario: Import cluster with valid kubeconfig
- **WHEN** an authorized user submits a kubeconfig with cluster admin access
- **THEN** the system MUST validate connectivity, extract cluster metadata, create Cluster and Credential records, and sync initial node information
- **AND** the system MUST mark the cluster source as `external_managed`

#### Scenario: Import with invalid kubeconfig
- **WHEN** a user submits a malformed or unauthorized kubeconfig
- **THEN** the system MUST reject the import with specific validation errors
- **AND** the system MUST NOT create any cluster records

### Requirement: Cluster source distinction
The system MUST distinguish between platform-managed and externally-managed clusters with appropriate metadata and lifecycle controls.

#### Scenario: Platform-managed cluster lifecycle
- **WHEN** a cluster is created via bootstrap workflow
- **THEN** the system MUST set `source` to `platform_managed` and enable full lifecycle operations (upgrade, scale, delete)

#### Scenario: External-managed cluster limitations
- **WHEN** a cluster is imported from external kubeconfig
- **THEN** the system MUST set `source` to `external_managed` and disable destructive operations that require infrastructure access

### Requirement: Cluster node tracking
The system MUST maintain a separate `cluster_nodes` table tracking node status synced from Kubernetes API.

#### Scenario: Sync nodes after cluster creation
- **WHEN** a cluster bootstrap completes successfully
- **THEN** the system MUST query the Kubernetes API for node information and populate `cluster_nodes` records
- **AND** each node MUST be linked to its host record if applicable

#### Scenario: Node status refresh
- **WHEN** an authorized user requests node list for a cluster
- **THEN** the system MUST query the Kubernetes API and update `cluster_nodes` with current status
