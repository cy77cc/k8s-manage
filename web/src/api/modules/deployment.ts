import apiService from '../api';
import type { ApiResponse, PaginatedResponse } from '../api';

export interface DeployTargetNode {
  host_id: number;
  role: string;
  weight: number;
}

export interface DeployTarget {
  id: number;
  name: string;
  target_type: 'k8s' | 'compose';
  runtime_type: 'k8s' | 'compose';
  cluster_id: number;
  project_id: number;
  team_id: number;
  env: string;
  status: string;
  nodes?: Array<{
    host_id: number;
    name: string;
    ip: string;
    status: string;
    role: string;
    weight: number;
  }>;
  created_at: string;
  updated_at: string;
}

export interface DeployRelease {
  id: number;
  service_id: number;
  target_id: number;
  namespace_or_project: string;
  runtime_type: 'k8s' | 'compose';
  strategy: string;
  revision_id: number;
  status: string;
  lifecycle_state?: string;
  diagnostics_json?: string;
  verification_json?: string;
  source_release_id?: number;
  target_revision?: string;
  created_at: string;
}

export interface DeployReleaseTimelineEvent {
  id: number;
  release_id: number;
  action: string;
  actor: number;
  detail: Record<string, any> | null;
  created_at: string;
}

export interface Inspection {
  id: number;
  release_id: number;
  target_id: number;
  service_id: number;
  stage: string;
  summary: string;
  status: string;
  created_at: string;
}

export interface ClusterBootstrapTask {
  id: string;
  name: string;
  control_plane_host_id: number;
  worker_ids_json: string;
  cni: string;
  status: string;
  result_json?: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export const deploymentApi = {
  getTargets(): Promise<ApiResponse<PaginatedResponse<DeployTarget>>> {
    return apiService.get('/deploy/targets');
  },
  createTarget(payload: {
    name: string;
    target_type: 'k8s' | 'compose';
    cluster_id?: number;
    project_id?: number;
    team_id?: number;
    env?: string;
    nodes?: DeployTargetNode[];
  }): Promise<ApiResponse<DeployTarget>> {
    return apiService.post('/deploy/targets', payload);
  },
  updateTarget(id: number, payload: Partial<{
    name: string;
    target_type: 'k8s' | 'compose';
    cluster_id: number;
    project_id: number;
    team_id: number;
    env: string;
    nodes: DeployTargetNode[];
  }>): Promise<ApiResponse<DeployTarget>> {
    return apiService.put(`/deploy/targets/${id}`, payload);
  },
  deleteTarget(id: number): Promise<ApiResponse<void>> {
    return apiService.delete(`/deploy/targets/${id}`);
  },
  previewRelease(payload: {
    service_id: number;
    target_id: number;
    env?: string;
    strategy?: string;
    variables?: Record<string, string>;
  }): Promise<ApiResponse<{ resolved_manifest: string; checks: Array<{ code: string; message: string; level: string }>; warnings: Array<{ code: string; message: string; level: string }>; runtime: string; preview_token?: string; preview_expires_at?: string }>> {
    return apiService.post('/deploy/releases/preview', payload);
  },
  applyRelease(payload: {
    service_id: number;
    target_id: number;
    env?: string;
    strategy?: string;
    variables?: Record<string, string>;
    preview_token?: string;
  }): Promise<ApiResponse<{ release_id: number; status: string; runtime_type: string; approval_required?: boolean; approval_ticket?: string; lifecycle_state?: string; reason_code?: string }>> {
    return apiService.post('/deploy/releases/apply', payload);
  },
  approveRelease(id: number, payload?: { comment?: string }): Promise<ApiResponse<{ release_id: number; status: string; runtime_type: string; lifecycle_state?: string }>> {
    return apiService.post(`/deploy/releases/${id}/approve`, payload || {});
  },
  rejectRelease(id: number, payload?: { comment?: string }): Promise<ApiResponse<{ release_id: number; status: string; runtime_type: string; lifecycle_state?: string }>> {
    return apiService.post(`/deploy/releases/${id}/reject`, payload || {});
  },
  rollbackRelease(id: number): Promise<ApiResponse<{ release_id: number; status: string }>> {
    return apiService.post(`/deploy/releases/${id}/rollback`);
  },
  getReleases(params?: { service_id?: number; target_id?: number }): Promise<ApiResponse<PaginatedResponse<DeployRelease>>> {
    return apiService.get('/deploy/releases', { params });
  },
  getReleaseDetail(id: number): Promise<ApiResponse<DeployRelease>> {
    return apiService.get(`/deploy/releases/${id}`);
  },
  getReleaseTimeline(id: number): Promise<ApiResponse<PaginatedResponse<DeployReleaseTimelineEvent>>> {
    return apiService.get(`/deploy/releases/${id}/timeline`);
  },
  getReleasesByRuntime(params?: { service_id?: number; target_id?: number; runtime_type?: 'k8s' | 'compose' }): Promise<ApiResponse<PaginatedResponse<DeployRelease>>> {
    return apiService.get('/deploy/releases', { params });
  },
  getGovernance(serviceId: number, env?: string): Promise<ApiResponse<any>> {
    return apiService.get(`/services/${serviceId}/governance`, { params: { env } });
  },
  putGovernance(serviceId: number, payload: any): Promise<ApiResponse<any>> {
    return apiService.put(`/services/${serviceId}/governance`, payload);
  },
  runInspection(payload: { release_id?: number; target_id?: number; service_id?: number; stage: 'pre' | 'post' | 'periodic' }): Promise<ApiResponse<Inspection>> {
    return apiService.post('/aiops/inspections/run', payload);
  },
  listInspections(params?: { service_id?: number; target_id?: number }): Promise<ApiResponse<PaginatedResponse<Inspection>>> {
    return apiService.get('/aiops/inspections', { params });
  },
  previewClusterBootstrap(payload: {
    name: string;
    control_plane_host_id: number;
    worker_host_ids?: number[];
    cni?: string;
  }): Promise<ApiResponse<{ name: string; control_plane_host_id: number; worker_host_ids: number[]; cni: string; steps: string[]; expected_endpoint: string }>> {
    return apiService.post('/deploy/clusters/bootstrap/preview', payload);
  },
  applyClusterBootstrap(payload: {
    name: string;
    control_plane_host_id: number;
    worker_host_ids?: number[];
    cni?: string;
  }): Promise<ApiResponse<{ task_id: string; status: string; cluster_id?: number; target_id?: number }>> {
    return apiService.post('/deploy/clusters/bootstrap/apply', payload);
  },
  getClusterBootstrapTask(taskId: string): Promise<ApiResponse<ClusterBootstrapTask>> {
    return apiService.get(`/deploy/clusters/bootstrap/${encodeURIComponent(taskId)}`);
  },
};
