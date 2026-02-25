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
  }): Promise<ApiResponse<{ resolved_manifest: string; checks: Array<{ code: string; message: string; level: string }>; warnings: Array<{ code: string; message: string; level: string }>; runtime: string }>> {
    return apiService.post('/deploy/releases/preview', payload);
  },
  applyRelease(payload: {
    service_id: number;
    target_id: number;
    env?: string;
    strategy?: string;
    variables?: Record<string, string>;
  }): Promise<ApiResponse<{ release_id: number; status: string }>> {
    return apiService.post('/deploy/releases/apply', payload);
  },
  rollbackRelease(id: number): Promise<ApiResponse<{ release_id: number; status: string }>> {
    return apiService.post(`/deploy/releases/${id}/rollback`);
  },
  getReleases(params?: { service_id?: number; target_id?: number }): Promise<ApiResponse<PaginatedResponse<DeployRelease>>> {
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
};

