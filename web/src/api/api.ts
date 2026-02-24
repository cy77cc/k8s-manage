import axios from 'axios';
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';

// 响应数据结构
export interface ApiResponse<T = unknown> {
  success: boolean;
  data: T;
  message?: string;
  messageKey?: string;
  dataSource?: string;
  total?: number;
  error?: {
    code: string;
    message: string;
  };
}

// 分页响应结构
export interface PaginatedResponse<T> {
  total: number;
  list: T[];
}

// API服务类
class ApiService {
  private instance: AxiosInstance;

  constructor() {
    this.instance = axios.create({
      baseURL: import.meta.env.VITE_API_BASE || '/api/v1',
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // 请求拦截器
    this.instance.interceptors.request.use(
      (config) => {
        // 添加认证token
        const token = localStorage.getItem('token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        const projectId = localStorage.getItem('projectId');
        if (projectId) {
          config.headers['X-Project-ID'] = projectId;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // 响应拦截器
    this.instance.interceptors.response.use(
      (response: AxiosResponse<any>) => {
        const payload = response.data;
        // 兼容后端统一结构：{ code, msg/message, data, total }
        if (typeof payload?.code === 'number') {
          if (payload.code !== 1000 && payload.code !== 200) {
            return Promise.reject(new Error(payload.msg || payload.message || '请求失败'));
          }
          response.data = {
            success: true,
            message: payload.msg || payload.message,
            messageKey: payload.message_key,
            dataSource: payload.data_source,
            data: payload.data,
            total: payload.total,
          } as ApiResponse;
          return response;
        }

        if (!payload?.success) {
          return Promise.reject(new Error(payload?.error?.message || payload?.message || '请求失败'));
        }
        return response;
      },
      (error) => {
        const message = error.response?.data?.message || error.response?.data?.error?.message || error.message || '网络错误';
        return Promise.reject(new Error(message));
      }
    );
  }

  // GET请求
  async get<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.instance.get<ApiResponse<T>>(url, config);
    return response.data;
  }

  // POST请求
  async post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.instance.post<ApiResponse<T>>(url, data, config);
    return response.data;
  }

  // PUT请求
  async put<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.instance.put<ApiResponse<T>>(url, data, config);
    return response.data;
  }

  // DELETE请求
  async delete<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.instance.delete<ApiResponse<T>>(url, config);
    return response.data;
  }
}

export const apiService = new ApiService();
export default apiService;
