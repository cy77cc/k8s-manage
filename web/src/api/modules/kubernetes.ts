import apiService from '../api';
import type { ApiResponse, PaginatedResponse } from '../api';

// 集群数据结构
export interface Cluster {
  id: string;
  name: string;
  version: string;
  status: string;
  server?: string;
  credentialRef?: string;
  lastConnectMessage?: string;
  nodes: number;
  pods: number;
  cpu: number;
  memory: number;
  createdAt: string;
}

// 节点数据结构
export interface Node {
  id: string;
  name: string;
  ip?: string;
  role: string;
  status: string;
  cpu: number;
  memory: number;
  pods: number;
  labels: Record<string, string>;
}

// Pod数据结构
export interface Pod {
  id: string;
  name: string;
  namespace: string;
  status: string;
  phase: string;
  node: string;
  cpu: number;
  memory: number;
  restarts: number;
  age: string;
  labels: Record<string, string>;
  containers: Container[];
  qosClass: string;
  createdAt?: string;
  startTime: string;
}

// 容器数据结构
export interface Container {
  name: string;
  image: string;
  status: string;
  cpu: number;
  memory: number;
  restarts: number;
}

// 服务数据结构
export interface Service {
  id: string;
  name: string;
  namespace: string;
  type: string;
  clusterIP: string;
  externalIP?: string;
  ports: ServicePort[];
  selector: Record<string, string>;
  age: string;
}

// 服务端口数据结构
export interface ServicePort {
  port: number;
  targetPort: number;
  protocol: string;
}

// Ingress数据结构
export interface Ingress {
  id: string;
  name: string;
  namespace: string;
  host: string;
  path: string;
  service: string;
  port: number;
  tls: boolean;
}

// 集群列表请求参数
export interface ClusterListParams {
  page?: number;
  pageSize?: number;
  status?: string;
}

export interface CreateClusterPayload {
  name: string;
  server: string;
  kubeconfig?: string;
  credential_ref?: string;
  description?: string;
}

// 节点列表请求参数
export interface NodeListParams {
  page?: number;
  pageSize?: number;
  status?: string;
}

// Pod列表请求参数
export interface PodListParams {
  page?: number;
  pageSize?: number;
  namespace?: string;
  status?: string;
}

// 服务列表请求参数
export interface ServiceListParams {
  page?: number;
  pageSize?: number;
  namespace?: string;
  type?: string;
}

// Ingress列表请求参数
export interface IngressListParams {
  page?: number;
  pageSize?: number;
  namespace?: string;
}

// Kubernetes管理API
export const kubernetesApi = {
  async createCluster(payload: CreateClusterPayload): Promise<ApiResponse<Cluster>> {
    const response = await apiService.post<any>('/clusters', payload);
    const item = response.data || {};
    return {
      ...response,
      data: {
        id: String(item.id),
        name: item.name,
        version: item.version,
        status: item.status,
        server: item.server,
        credentialRef: item.credential_ref,
        lastConnectMessage: item.last_connect_message,
        nodes: item.nodes ?? 0,
        pods: item.pods ?? 0,
        cpu: item.cpu ?? 0,
        memory: item.memory ?? 0,
        createdAt: item.created_at ?? item.createdAt,
      },
    };
  },

  // 获取集群列表
  async getClusterList(params?: ClusterListParams): Promise<ApiResponse<PaginatedResponse<Cluster>>> {
    const response = await apiService.get<Cluster[]>('/clusters', {
      params: {
        page: params?.page,
        page_size: params?.pageSize,
      },
    });
    const list = (response.data || []).map((item: any) => ({
      id: String(item.id),
      name: item.name,
      version: item.version,
      status: item.status,
      server: item.server,
      credentialRef: item.credential_ref,
      lastConnectMessage: item.last_connect_message,
      nodes: item.nodes ?? 0,
      pods: item.pods ?? 0,
      cpu: item.cpu ?? 0,
      memory: item.memory ?? 0,
      createdAt: item.created_at ?? item.createdAt,
    }));
    return {
      ...response,
      data: {
        list,
        total: response.total || 0,
      },
    };
  },

  // 获取集群详情
  async getClusterDetail(id: string): Promise<ApiResponse<Cluster>> {
    return apiService.get(`/clusters/${id}`);
  },

  // 获取集群节点列表
  async getClusterNodes(clusterId: string, params?: NodeListParams): Promise<ApiResponse<PaginatedResponse<Node>>> {
    const response = await apiService.get<Node[]>(`/clusters/${clusterId}/nodes`, {
      params: {
        page: params?.page,
        page_size: params?.pageSize,
      },
    });
    const list = (response.data || []).map((item: any) => ({
      id: String(item.id),
      name: item.name,
      role: item.roles || item.role || '',
      status: (item.status || '').toLowerCase(),
      cpu: item.cpu_cores ?? item.cpu ?? 0,
      memory: item.memory ?? 0,
      pods: item.pods ?? 0,
      labels: item.labels || {},
      ip: item.ip,
    }));
    return {
      ...response,
      data: {
        list,
        total: response.total || 0,
      },
    };
  },

  // 后端当前提供 namespaces / deployments
  async getClusterNamespaces(clusterId: string): Promise<ApiResponse<any[]>> {
    return apiService.get(`/clusters/${clusterId}/namespaces`);
  },
  async getClusterDeployments(clusterId: string, namespace?: string): Promise<ApiResponse<any[]>> {
    return apiService.get(`/clusters/${clusterId}/deployments`, { params: { namespace } });
  },

  async getClusterPods(clusterId: string, namespace?: string): Promise<ApiResponse<any[]>> {
    return apiService.get(`/clusters/${clusterId}/pods`, { params: { namespace } });
  },

  async testClusterConnect(clusterId: string): Promise<ApiResponse<void>> {
    return apiService.post(`/clusters/${clusterId}/connect/test`);
  },

  async getClusterEvents(clusterId: string, namespace?: string): Promise<ApiResponse<any[]>> {
    return apiService.get(`/clusters/${clusterId}/events`, { params: { namespace } });
  },

  async getPodLogs(clusterId: string, pod: string, container?: string, namespace?: string): Promise<ApiResponse<{ logs: string }>> {
    return apiService.get(`/clusters/${clusterId}/logs`, { params: { pod, container, namespace } });
  },

  async getClusterServices(clusterId: string, namespace?: string): Promise<ApiResponse<any[]>> {
    const response = await apiService.get<any[]>(`/clusters/${clusterId}/services`, { params: { namespace } });
    const list = (response.data || []).map((item: any) => ({
      ...item,
      ports: typeof item.ports === 'string' ? JSON.parse(item.ports || '[]') : item.ports,
    }));
    return { ...response, data: list };
  },

  async getClusterIngresses(clusterId: string, namespace?: string): Promise<ApiResponse<any[]>> {
    return apiService.get(`/clusters/${clusterId}/ingresses`, { params: { namespace } });
  },

  async previewDeploy(clusterId: string, payload: { namespace: string; name: string; image: string; replicas: number }): Promise<ApiResponse<any>> {
    return apiService.post(`/clusters/${clusterId}/deploy/preview`, payload);
  },

  async applyDeploy(clusterId: string, payload: { namespace: string; name: string; image: string; replicas: number }): Promise<ApiResponse<void>> {
    return apiService.post(`/clusters/${clusterId}/deploy/apply`, payload);
  },
};
