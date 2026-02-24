import apiService from '../api';
import type { ApiResponse } from '../api';

export const cicdApi = {
  async createPipeline(payload: {
    service_name: string;
    repo_url: string;
    branch?: string;
    build_cmd?: string;
    dockerfile_path?: string;
    image_repo?: string;
    cluster_id?: number;
    namespace?: string;
  }): Promise<ApiResponse<any>> {
    return apiService.post('/cicd/pipelines', payload);
  },
  async getPipeline(id: string): Promise<ApiResponse<any>> {
    return apiService.get(`/cicd/pipelines/${id}`);
  },
  async runPipeline(id: string): Promise<ApiResponse<any>> {
    return apiService.post(`/cicd/pipelines/${id}/run`);
  },
  async getRunLogs(id: string): Promise<ApiResponse<any[]>> {
    return apiService.get(`/cicd/runs/${id}/logs`);
  },
  async rollback(id: string): Promise<ApiResponse<void>> {
    return apiService.post(`/cicd/runs/${id}/rollback`);
  },
};

