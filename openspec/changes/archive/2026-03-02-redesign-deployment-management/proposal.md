## Why

The current deployment management UI exposes less than 20% of the backend's capabilities. While the backend implements a complete PaaS deployment system (cluster bootstrap, environment initialization, deployment targets, release lifecycle, approval workflows, credential management), the frontend only provides basic release history viewing and a minimal deployment wizard. As a PaaS platform, this makes the deployment functionality essentially unusable for real-world scenarios. Users cannot provision clusters, initialize environments, manage deployment targets, or leverage the sophisticated release management features that already exist in the backend.

## What Changes

### New Frontend Modules (4 major areas)

**1. Infrastructure Management**
- Cluster management pages (list, create wizard, detail, operations)
- Credential management pages (list, register platform credentials, import external credentials, connectivity testing)
- Host management enhancements (host pool, status monitoring, SSH connection management)

**2. Deployment Target Management**
- Deployment target list page (grouped by environment)
- Create target wizard (runtime selection, cluster/host binding, environment bootstrap)
- Target detail page (basic info, bound nodes, deployment history, resource usage)
- Environment initialization wizard (runtime package selection, preflight checks, real-time installation progress, verification results)

**3. Release Management Enhancements**
- Deployment overview dashboard (multi-environment status, pending approvals, in-progress deployments, recent failures)
- Improved release creation wizard (5-step flow with enhanced preview and strategy selection)
- Real-time deployment progress tracking (pod status, health checks, live logs)
- Approval center (pending approvals, initiated by me, approval history)
- Enhanced release detail page (state flow diagram, real-time logs, manifest diff, approval workflow, verification results)

**4. Governance & Observability**
- Deployment topology visualization (service dependency graph, environment distribution, real-time status)
- Policy management UI (traffic, resilience, access, SLO policies)
- Enhanced audit logs (operation records, change tracking, compliance reports)
- AIOps insights integration (deployment risk assessment, anomaly detection, optimization suggestions)

### API Integration
- Connect all existing backend APIs that are currently unused by the frontend
- Add real-time progress polling for long-running operations (bootstrap, deployment)
- Implement WebSocket or SSE for live log streaming (optional enhancement)

### Component Library Additions
- `BootstrapProgressTracker` - Real-time installation progress with phase indicators
- `DeploymentTopology` - Interactive service dependency and environment distribution graph
- `ReleaseStateFlow` - Visual state machine for release lifecycle
- `CredentialTestResult` - Connectivity test results with latency metrics
- `EnvironmentStatusCard` - Multi-environment health overview
- `ApprovalWorkflow` - Approval request and decision UI

## Capabilities

### New Capabilities
- `deployment-infrastructure-management`: Cluster provisioning, credential management, and host pool operations for PaaS infrastructure
- `deployment-target-management`: Deployment target lifecycle (create, configure, bootstrap, monitor) and environment initialization workflows
- `deployment-release-management`: Enhanced release creation, real-time progress tracking, approval workflows, and rollback operations
- `deployment-observability`: Deployment topology visualization, audit logs, policy management, and AIOps insights

### Modified Capabilities
<!-- No existing capabilities are being modified - this is purely additive -->

## Impact

### Frontend
- **New Pages**: 15+ new page components across 4 modules
- **New Components**: 20+ reusable components for deployment workflows
- **API Modules**: Extensive use of existing `web/src/api/modules/deployment.ts` APIs
- **Routes**: New route structure under `/deployment/*` with nested navigation
- **State Management**: Complex state for real-time progress tracking and multi-step wizards

### Backend
- **No Changes Required**: All necessary APIs already exist in `internal/service/deployment/`
- **API Coverage**: Full utilization of existing endpoints (targets, releases, approvals, bootstrap, credentials)
- **Models**: All required models already defined in `internal/model/deployment.go`

### User Experience
- **Before**: Deployment management is essentially non-functional (only release history viewing)
- **After**: Complete self-service PaaS deployment platform with cluster provisioning, environment management, and sophisticated release workflows
- **Learning Curve**: Progressive disclosure design minimizes complexity for simple use cases while exposing advanced features when needed

### Dependencies
- No new external dependencies required
- Leverages existing Ant Design components and Tailwind utilities
- May add `react-flow` or similar for topology visualization (optional)

### Migration
- No breaking changes - purely additive functionality
- Existing deployment history and data remain accessible
- Users can continue using existing workflows while new features are adopted incrementally
