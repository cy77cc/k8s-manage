import apiService from '../api';
import type { ApiResponse, PaginatedResponse } from '../api';

export type ServiceRuntimeType = 'k8s' | 'compose' | 'helm';
export type ServiceConfigMode = 'standard' | 'custom';

export interface LabelKV {
  key: string;
  value: string;
}

export interface EnvKV {
  key: string;
  value: string;
}

export interface PortConfig {
  name?: string;
  protocol?: string;
  container_port: number;
  service_port: number;
}

export interface StandardServiceConfig {
  image: string;
  replicas: number;
  ports: PortConfig[];
  envs: EnvKV[];
  resources?: Record<string, string>;
}

export interface ServiceItem {
  id: string;
  projectId: string;
  teamId: string;
  name: string;
  env: string;
  owner: string;
  runtimeType: ServiceRuntimeType;
  configMode: ServiceConfigMode;
  serviceKind: string;
  status: string;
  labels: LabelKV[];
  standardConfig?: StandardServiceConfig;
  customYaml?: string;
  renderedYaml?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface ServiceEvent {
  id: string;
  serviceId: string;
  type: string;
  level: string;
  message: string;
  createdAt: string;
}

export interface ServiceQuota {
  cpuLimit: number;
  memoryLimit: number;
  cpuUsed: number;
  memoryUsed: number;
}

export interface ServiceMgmtListParams {
  page?: number;
  pageSize?: number;
  projectId?: string;
  teamId?: string;
  runtimeType?: ServiceRuntimeType | 'all';
  env?: string;
  labelSelector?: string;
  q?: string;
}

export interface ServiceCreateParams {
  project_id?: number;
  team_id?: number;
  name: string;
  env: string;
  owner: string;
  runtime_type: ServiceRuntimeType;
  config_mode: ServiceConfigMode;
  service_kind: string;
  service_type: 'stateless' | 'stateful';
  render_target: 'k8s' | 'compose';
  labels?: LabelKV[];
  standard_config?: StandardServiceConfig;
  custom_yaml?: string;
  source_template_version?: string;
  status?: string;
}

export interface RenderPreviewReq {
  mode: ServiceConfigMode;
  target: 'k8s' | 'compose';
  service_name: string;
  service_type: 'stateless' | 'stateful';
  standard_config?: StandardServiceConfig;
  custom_yaml?: string;
}

export interface RenderPreviewResp {
  rendered_yaml: string;
  diagnostics: Array<{ level: string; code: string; message: string }>;
  normalized_config?: StandardServiceConfig;
}

const mapService = (item: any): ServiceItem => ({
  id: String(item.id),
  projectId: String(item.project_id || ''),
  teamId: String(item.team_id || ''),
  name: item.name || '',
  env: item.env || 'staging',
  owner: item.owner || 'system',
  runtimeType: (item.runtime_type || 'k8s') as ServiceRuntimeType,
  configMode: (item.config_mode || 'standard') as ServiceConfigMode,
  serviceKind: item.service_kind || 'web',
  status: item.status || 'draft',
  labels: Array.isArray(item.labels) ? item.labels.map((x: any) => ({ key: x.key || '', value: x.value || '' })) : [],
  standardConfig: item.standard_config,
  customYaml: item.custom_yaml,
  renderedYaml: item.rendered_yaml || item.yaml_content,
  createdAt: item.created_at,
  updatedAt: item.updated_at,
});

export const serviceApi = {
  async preview(data: RenderPreviewReq): Promise<ApiResponse<RenderPreviewResp>> {
    return apiService.post('/services/render/preview', data);
  },

  async transform(data: { standard_config: StandardServiceConfig; target: 'k8s' | 'compose'; service_name: string; service_type: 'stateless' | 'stateful' }): Promise<ApiResponse<{ custom_yaml: string; source_hash: string }>> {
    return apiService.post('/services/transform', data);
  },

  async getList(params?: ServiceMgmtListParams): Promise<ApiResponse<PaginatedResponse<ServiceItem>>> {
    const response = await apiService.get<any>('/services', {
      params: {
        page: params?.page,
        page_size: params?.pageSize,
        project_id: params?.projectId,
        team_id: params?.teamId,
        runtime_type: params?.runtimeType && params.runtimeType !== 'all' ? params.runtimeType : undefined,
        env: params?.env && params.env !== 'all' ? params.env : undefined,
        label_selector: params?.labelSelector,
        q: params?.q,
      },
    });
    const payload = response.data || {};
    const list = (payload.list || []).map(mapService);
    return {
      ...response,
      data: { list, total: Number(payload.total || list.length) },
    };
  },

  async getDetail(id: string): Promise<ApiResponse<ServiceItem>> {
    const response = await apiService.get<any>(`/services/${id}`);
    return { ...response, data: mapService(response.data) };
  },

  async create(data: ServiceCreateParams): Promise<ApiResponse<ServiceItem>> {
    const response = await apiService.post<any>('/services', {
      project_id: data.project_id || Number(localStorage.getItem('projectId') || '1'),
      team_id: data.team_id || Number(localStorage.getItem('teamId') || '1'),
      ...data,
    });
    return { ...response, data: mapService(response.data) };
  },

  async update(id: string, data: Partial<ServiceCreateParams>): Promise<ApiResponse<ServiceItem>> {
    const response = await apiService.put<any>(`/services/${id}`, data);
    return { ...response, data: mapService(response.data) };
  },

  async remove(id: string): Promise<ApiResponse<void>> {
    return apiService.delete(`/services/${id}`);
  },

  async deploy(id: string, payload?: { deploy_target?: ServiceRuntimeType; cluster_id?: number; approval_token?: string }): Promise<ApiResponse<void>> {
    return apiService.post(`/services/${id}/deploy`, payload || {});
  },

  async rollback(id: string): Promise<ApiResponse<void>> {
    return apiService.post(`/services/${id}/rollback`);
  },

  async getEvents(id: string): Promise<ApiResponse<PaginatedResponse<ServiceEvent>>> {
    const response = await apiService.get<any[]>(`/services/${id}/events`);
    const list = (response.data || []).map((item: any) => ({
      id: String(item.id),
      serviceId: String(item.service_id || item.serviceId),
      type: item.type,
      level: item.level,
      message: item.message,
      createdAt: item.created_at || item.createdAt,
    }));
    return { ...response, data: { list, total: list.length } };
  },

  async getQuota(): Promise<ApiResponse<ServiceQuota>> {
    return apiService.get('/services/quota');
  },

  async helmImport(payload: { service_id: number; chart_name: string; chart_version?: string; chart_ref?: string; values_yaml?: string; rendered_yaml?: string }): Promise<ApiResponse<any>> {
    return apiService.post('/services/helm/import', payload);
  },

  async helmRender(payload: { release_id?: number; chart_name?: string; chart_ref?: string; values_yaml?: string; rendered_yaml?: string }): Promise<ApiResponse<{ rendered_yaml: string; diagnostics: Array<{ level: string; code: string; message: string }> }>> {
    return apiService.post('/services/helm/render', payload);
  },

  async deployHelm(id: string): Promise<ApiResponse<void>> {
    return apiService.post(`/services/${id}/deploy/helm`);
  },
};
