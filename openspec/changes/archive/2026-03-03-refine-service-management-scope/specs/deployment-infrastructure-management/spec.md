## ADDED Requirements

### Requirement: Cluster Environment Type
The system SHALL maintain an environment type for each cluster to constrain deployment targets.

#### Scenario: Cluster with environment type
- **WHEN** a cluster is created or imported
- **THEN** the system MUST set `env_type` to one of: `development`, `staging`, `production`
- **AND** the default value MUST be `development`

#### Scenario: Update cluster environment type
- **WHEN** an authorized user updates a cluster's environment type
- **THEN** the system MUST persist the change
- **AND** the system MUST log the change in the audit trail

### Requirement: Deployment Environment Matching
The system SHALL validate environment matching when deploying a service to a cluster.

#### Scenario: Deploy to matching environment
- **WHEN** a deployment is requested with `service_id` and `cluster_id`
- **AND** the service's `env` matches the cluster's `env_type`
- **THEN** the system MUST allow the deployment to proceed

#### Scenario: Reject deployment to mismatched environment
- **WHEN** a deployment is requested with `service_id` and `cluster_id`
- **AND** the service's `env` does NOT match the cluster's `env_type`
- **THEN** the system MUST reject the deployment with error code `ENV_MISMATCH`
- **AND** the error message MUST indicate the expected and actual environment types

#### Scenario: Deployment without cluster_id
- **WHEN** a deployment is requested without `cluster_id`
- **THEN** the system MUST reject the request with validation error
- **AND** the error message MUST indicate `cluster_id` is required
