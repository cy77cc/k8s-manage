import apiService from '../api';
import type { ApiResponse, PaginatedResponse } from '../api';

export interface ServiceItem {
  id: string;
  projectId: string;
  name: string;
  env: string;
  owner: string;
  status: 'running' | 'syncing' | 'deploying' | 'error';
  image: string;
  replicas: number;
  cpuLimit: number;
  memLimit: number;
  tags: string[];
  config: string;
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
}

export interface ServiceCreateParams {
  name: string;
  env: string;
  owner: string;
  status?: string;
  image: string;
  replicas?: number;
  cpuLimit?: number;
  memLimit?: number;
  tags?: string[];
  config?: string;
}

const parseTags = (input: unknown): string[] => {
  if (Array.isArray(input)) return input.map(String);
  if (typeof input === 'string') {
    try {
      const parsed = JSON.parse(input);
      if (Array.isArray(parsed)) return parsed.map(String);
    } catch {
      return input.split(',').map((v) => v.trim()).filter(Boolean);
    }
  }
  return [];
};

const mapService = (item: any): ServiceItem => ({
  id: String(item.id),
  projectId: String(item.project_id || item.projectId || ''),
  name: item.name,
  env: item.env || (Array.isArray(item.env_vars) ? (item.env_vars.find((x: any) => x.key === 'ENV')?.value || 'staging') : 'staging'),
  owner: item.owner || (Array.isArray(item.env_vars) ? (item.env_vars.find((x: any) => x.key === 'OWNER')?.value || 'system') : 'system'),
  status: item.status || 'running',
  image: item.image,
  replicas: item.replicas || 1,
  cpuLimit: item.cpu_limit || item.cpuLimit || 0,
  memLimit: item.mem_limit || item.memLimit || 0,
  tags: parseTags(item.tags),
  config: item.config || '',
  createdAt: item.created_at || item.createdAt,
  updatedAt: item.updated_at || item.updatedAt,
});

export const serviceApi = {
  async getList(params?: ServiceMgmtListParams): Promise<ApiResponse<PaginatedResponse<ServiceItem>>> {
    const response = await apiService.get<any[]>('/services', {
      params: { page: params?.page, page_size: params?.pageSize },
    });
    const list = (response.data || []).map(mapService);
    return {
      ...response,
      data: { list, total: response.total || list.length },
    };
  },

  async getDetail(id: string): Promise<ApiResponse<ServiceItem>> {
    const response = await apiService.get<any>(`/services/${id}`);
    return { ...response, data: mapService(response.data) };
  },

  async create(data: ServiceCreateParams): Promise<ApiResponse<ServiceItem>> {
    const projectId = Number(localStorage.getItem('projectId') || '1');
    const payload = {
      project_id: projectId,
      name: data.name,
      type: 'stateless',
      image: data.image,
      replicas: data.replicas || 1,
      service_port: 80,
      container_port: 8080,
      env_vars: [
        { key: 'ENV', value: data.env || 'staging' },
        { key: 'OWNER', value: data.owner || 'system' },
        { key: 'CONFIG', value: data.config || '' },
      ],
      resources: {
        limits: {
          cpu: `${Math.max(100, data.cpuLimit || 500)}m`,
          memory: `${Math.max(128, data.memLimit || 512)}Mi`,
        },
      },
    };
    const response = await apiService.post<any>('/services', payload);
    return { ...response, data: mapService(response.data) };
  },

  async update(id: string, data: Partial<ServiceCreateParams>): Promise<ApiResponse<ServiceItem>> {
    const projectId = Number(localStorage.getItem('projectId') || '1');
    const payload = {
      project_id: projectId,
      name: data.name,
      type: 'stateless',
      image: data.image,
      replicas: data.replicas || 1,
      service_port: 80,
      container_port: 8080,
      env_vars: [
        { key: 'ENV', value: data.env || 'staging' },
        { key: 'OWNER', value: data.owner || 'system' },
        { key: 'CONFIG', value: data.config || '' },
      ],
      resources: {
        limits: {
          cpu: `${Math.max(100, data.cpuLimit || 500)}m`,
          memory: `${Math.max(128, data.memLimit || 512)}Mi`,
        },
      },
    };
    const response = await apiService.put<any>(`/services/${id}`, payload);
    return { ...response, data: mapService(response.data) };
  },

  async remove(id: string): Promise<ApiResponse<void>> {
    return apiService.delete(`/services/${id}`);
  },

  async deploy(id: string, changeSet?: Record<string, any>): Promise<ApiResponse<void>> {
    const clusterId = Number(localStorage.getItem('clusterId') || '1');
    return apiService.post(`/services/${id}/deploy`, { cluster_id: clusterId, ...(changeSet || {}) });
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
    return { ...response, data: { list, total: response.total || list.length } };
  },

  async getQuota(): Promise<ApiResponse<ServiceQuota>> {
    return apiService.get('/services/quota');
  },
};
