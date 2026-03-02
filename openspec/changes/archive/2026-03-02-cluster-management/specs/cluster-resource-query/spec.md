## ADDED Requirements

### Requirement: Real-time Kubernetes resource query
The system MUST provide real-time resource visibility by querying the Kubernetes API on demand.

#### Scenario: List namespaces in cluster
- **WHEN** an authorized user requests namespace list for a cluster
- **THEN** the system MUST use the cluster's stored credentials to query the Kubernetes API and return all namespaces

#### Scenario: List workloads in namespace
- **WHEN** an authorized user requests workload list for a cluster and namespace
- **THEN** the system MUST return Deployments, StatefulSets, DaemonSets, Jobs, and CronJobs with status information
- **AND** each workload MUST include replica counts, health status, and age

#### Scenario: List pods with details
- **WHEN** an authorized user requests pod list for a cluster and namespace
- **THEN** the system MUST return pods with node assignment, status, IP, and container information

### Requirement: Service and configuration visibility
The system MUST allow viewing Kubernetes Services, Ingresses, ConfigMaps, and Secrets.

#### Scenario: List services and ingresses
- **WHEN** an authorized user requests service list for a cluster and namespace
- **THEN** the system MUST return Services with type, cluster IP, ports, and Ingresses with host and path rules

#### Scenario: List configuration resources
- **WHEN** an authorized user requests config list for a cluster and namespace
- **THEN** the system MUST return ConfigMaps and Secrets metadata (NOT secret values)
- **AND** Secrets MUST only display name, type, and creation timestamp

### Requirement: Storage resource visibility
The system MUST provide visibility into PersistentVolumes and PersistentVolumeClaims.

#### Scenario: List storage resources
- **WHEN** an authorized user requests storage list for a cluster
- **THEN** the system MUST return PVs with capacity, access modes, and status
- **AND** return PVCs with namespace, capacity, and bound PV reference

### Requirement: Deployment-centric service view
The system MUST allow viewing deployed services from the cluster perspective.

#### Scenario: List deployed services in cluster
- **WHEN** an authorized user requests service deployments for a cluster
- **THEN** the system MUST return services from the deployment table that are deployed to this cluster
- **AND** each service MUST include environment, last deployment time, and current status

### Requirement: Cluster credential caching
The system MUST cache Kubernetes client connections to reduce latency.

#### Scenario: Cache client connection
- **WHEN** the system establishes a Kubernetes client for a cluster
- **THEN** it MUST cache the client for subsequent requests
- **AND** invalidate cache on credential update or connection failure
