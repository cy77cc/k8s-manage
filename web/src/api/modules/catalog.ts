import apiService from '../api';
import type { ApiResponse, PaginatedResponse } from '../api';

export type CatalogTemplateStatus = 'draft' | 'pending_review' | 'published' | 'rejected';
export type CatalogTemplateVisibility = 'private' | 'public';
export type CatalogDeployTarget = 'k8s' | 'compose';

export interface CatalogVariableSchema {
  name: string;
  type: 'string' | 'number' | 'password' | 'boolean' | 'select' | 'textarea';
  default?: string | number | boolean;
  required: boolean;
  description?: string;
  options?: string[];
}

export interface CatalogCategory {
  id: number;
  name: string;
  display_name: string;
  icon?: string;
  description?: string;
  sort_order: number;
  is_system: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface CatalogTemplate {
  id: number;
  name: string;
  display_name: string;
  description?: string;
  icon?: string;
  category_id: number;
  version: string;
  owner_id: number;
  visibility: CatalogTemplateVisibility;
  status: CatalogTemplateStatus;
  k8s_template?: string;
  compose_template?: string;
  variables_schema: CatalogVariableSchema[];
  readme?: string;
  tags: string[];
  deploy_count: number;
  review_note?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CategoryCreatePayload {
  name: string;
  display_name: string;
  icon?: string;
  description?: string;
  sort_order?: number;
}

export interface TemplateCreatePayload {
  name: string;
  display_name: string;
  description?: string;
  icon?: string;
  category_id: number;
  version?: string;
  visibility?: CatalogTemplateVisibility;
  k8s_template?: string;
  compose_template?: string;
  variables_schema: CatalogVariableSchema[];
  readme?: string;
  tags?: string[];
}

export interface CatalogPreviewPayload {
  template_id: number;
  target: CatalogDeployTarget;
  variables: Record<string, string | number | boolean>;
}

export interface CatalogPreviewResponse {
  rendered_yaml: string;
  unresolved_vars: string[];
}

export interface CatalogDeployPayload {
  template_id: number;
  target: CatalogDeployTarget;
  project_id: number;
  team_id?: number;
  service_name: string;
  namespace?: string;
  cluster_id?: number;
  environment?: string;
  variables: Record<string, string | number | boolean>;
  deploy_now?: boolean;
}

export interface CatalogDeployResponse {
  service_id: number;
  template_id: number;
  deploy_count: number;
}

export interface CatalogTemplateListParams {
  category_id?: number;
  status?: CatalogTemplateStatus;
  visibility?: CatalogTemplateVisibility;
  q?: string;
  mine?: boolean;
}

export const catalogApi = {
  async listCategories(): Promise<ApiResponse<PaginatedResponse<CatalogCategory>>> {
    const response = await apiService.get<any>('/catalog/categories');
    return {
      ...response,
      data: {
        list: Array.isArray(response.data?.list) ? response.data.list : [],
        total: Number(response.data?.total || 0),
      },
    };
  },

  async createCategory(payload: CategoryCreatePayload): Promise<ApiResponse<CatalogCategory>> {
    return apiService.post('/catalog/categories', payload);
  },

  async updateCategory(id: number, payload: Partial<CategoryCreatePayload>): Promise<ApiResponse<CatalogCategory>> {
    return apiService.put(`/catalog/categories/${id}`, payload);
  },

  async deleteCategory(id: number): Promise<ApiResponse<void>> {
    return apiService.delete(`/catalog/categories/${id}`);
  },

  async listTemplates(params?: CatalogTemplateListParams): Promise<ApiResponse<PaginatedResponse<CatalogTemplate>>> {
    const response = await apiService.get<any>('/catalog/templates', { params });
    return {
      ...response,
      data: {
        list: Array.isArray(response.data?.list) ? response.data.list : [],
        total: Number(response.data?.total || 0),
      },
    };
  },

  async getTemplate(id: number): Promise<ApiResponse<CatalogTemplate>> {
    return apiService.get(`/catalog/templates/${id}`);
  },

  async createTemplate(payload: TemplateCreatePayload): Promise<ApiResponse<CatalogTemplate>> {
    return apiService.post('/catalog/templates', payload);
  },

  async updateTemplate(id: number, payload: Partial<TemplateCreatePayload>): Promise<ApiResponse<CatalogTemplate>> {
    return apiService.put(`/catalog/templates/${id}`, payload);
  },

  async deleteTemplate(id: number): Promise<ApiResponse<void>> {
    return apiService.delete(`/catalog/templates/${id}`);
  },

  async submitTemplate(id: number): Promise<ApiResponse<CatalogTemplate>> {
    return apiService.post(`/catalog/templates/${id}/submit`);
  },

  async publishTemplate(id: number): Promise<ApiResponse<CatalogTemplate>> {
    return apiService.post(`/catalog/templates/${id}/publish`);
  },

  async rejectTemplate(id: number, reason: string): Promise<ApiResponse<CatalogTemplate>> {
    return apiService.post(`/catalog/templates/${id}/reject`, { reason });
  },

  async preview(payload: CatalogPreviewPayload): Promise<ApiResponse<CatalogPreviewResponse>> {
    return apiService.post('/catalog/preview', payload);
  },

  async deploy(payload: CatalogDeployPayload): Promise<ApiResponse<CatalogDeployResponse>> {
    return apiService.post('/catalog/deploy', payload);
  },
};

export default catalogApi;
