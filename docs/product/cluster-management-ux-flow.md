# Cluster Management UX Flow (Phase-1)

## Main Entry

`K8sPage` uses cluster detail drawer with modular tabs:

- Overview
- Namespaces
- Rollouts
- HPA
- Quotas
- Existing Nodes/Workloads/Network/Events tabs remain for compatibility.

## Namespace Governance

- Create namespace
- View namespace-to-team bindings
- Update team bindings by team id and namespace list

## Release Strategy

- Rollout wizard supports `rolling | blue-green | canary`
- Preview manifest before apply
- Per rollout actions: promote / abort / rollback

## Elasticity and Quota

- HPA editor for CPU/MEM thresholds
- Quota/LimitRange editors for resource governance

## Permission UX

- Backend enforces permission and returns structured message.
- Frontend keeps action buttons but surfaces server-side denial reason.
