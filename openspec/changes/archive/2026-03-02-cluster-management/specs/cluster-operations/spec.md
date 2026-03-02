## ADDED Requirements

### Requirement: Node lifecycle management
The system MUST support adding and removing nodes from platform-managed clusters.

#### Scenario: Add worker node to cluster
- **WHEN** an authorized operator requests adding a new worker node to a platform-managed cluster
- **THEN** the system MUST execute kubeadm join on the target host
- **AND** update cluster_nodes and sync node status

#### Scenario: Remove node from cluster
- **WHEN** an authorized operator requests removing a node from a cluster
- **THEN** the system MUST cordon and drain the node, execute kubeadm reset, and delete the node record
- **AND** update host record to remove cluster association

#### Scenario: Node operation on external cluster
- **WHEN** an operator attempts node lifecycle operations on an externally-managed cluster
- **THEN** the system MUST reject the operation with an appropriate error message

### Requirement: Cluster connectivity testing
The system MUST provide cluster connectivity testing with detailed diagnostics.

#### Scenario: Test cluster connection
- **WHEN** an authorized user requests connectivity test for a cluster
- **THEN** the system MUST attempt to connect using stored credentials
- **AND** return connection status, latency, and Kubernetes version

#### Scenario: Test with stale credentials
- **WHEN** connectivity test fails due to expired or invalid credentials
- **THEN** the system MUST update credential status and return specific error guidance

### Requirement: Cluster update and deletion
The system MUST support updating cluster metadata and safe deletion.

#### Scenario: Update cluster metadata
- **WHEN** an authorized user updates cluster name, description, or labels
- **THEN** the system MUST persist changes and update the audit log

#### Scenario: Delete platform-managed cluster
- **WHEN** an authorized user deletes a platform-managed cluster
- **THEN** the system MUST execute kubeadm reset on all nodes, delete credentials, and remove cluster records
- **AND** require confirmation for destructive operation

#### Scenario: Delete external cluster reference
- **WHEN** an authorized user deletes an externally-managed cluster
- **THEN** the system MUST only remove platform records without affecting the actual cluster
- **AND** preserve audit trail of the deletion

### Requirement: Cluster event visibility
The system MUST provide access to Kubernetes cluster events.

#### Scenario: Query cluster events
- **WHEN** an authorized user requests events for a cluster
- **THEN** the system MUST return Kubernetes events sorted by creation timestamp
- **AND** support filtering by namespace, resource type, and severity
