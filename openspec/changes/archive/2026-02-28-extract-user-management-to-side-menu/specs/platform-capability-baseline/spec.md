## MODIFIED Requirements

### Requirement: API V1 Domain Surface SHALL Be Baseline Documented
The platform SHALL maintain a baseline capability specification for the `/api/v1` domains currently wired by the backend service, including user/auth, host, cluster, service, deployment, RBAC, AI, AIOPS, monitoring, project, and node compatibility routes, and SHALL explicitly classify user/role/permission governance as an independent access-governance navigation domain rather than a system-settings subcategory.

#### Scenario: Route group baseline exists
- **WHEN** reviewers compare OpenSpec with backend router registration
- **THEN** the documented baseline SHALL include all currently registered domain groups under `internal/service/service.go`

#### Scenario: Access-governance baseline classification exists
- **WHEN** reviewers inspect baseline documentation for management navigation responsibilities
- **THEN** user, role, and permission management SHALL be documented as a dedicated governance domain with role-aware visibility expectations
