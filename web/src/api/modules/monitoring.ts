import apiService from '../api';
import type { ApiResponse, PaginatedResponse } from '../api';

// 告警数据结构
export interface Alert {
  id: string;
  title: string;
  severity: string;
  source: string;
  status: string;
  createdAt: string;
}

// 告警规则数据结构
export interface AlertRule {
  id: string;
  name: string;
  condition: string;
  severity: string;
  enabled: boolean;
  channels: string[];
  createdAt: string;
}

// 监控指标数据结构
export interface MetricData {
  timestamp: string;
  value: number;
}

// 告警列表请求参数
export interface AlertListParams {
  page?: number;
  pageSize?: number;
  severity?: string;
  status?: string;
}

// 告警规则列表请求参数
export interface AlertRuleListParams {
  page?: number;
  pageSize?: number;
  severity?: string;
  enabled?: boolean;
}

// 监控指标请求参数
export interface MetricParams {
  metric: string;
  startTime: string;
  endTime: string;
  interval?: string;
}

// 监控告警API
export const monitoringApi = {
  // 获取告警列表
  async getAlertList(params?: AlertListParams): Promise<ApiResponse<PaginatedResponse<Alert>>> {
    const response = await apiService.get<Alert[]>('/alerts', {
      params: {
        page: params?.page,
        page_size: params?.pageSize,
        severity: params?.severity,
        status: params?.status,
      },
    });
    const list = (response.data || []).map((item: any) => ({
      id: String(item.id),
      title: item.message || item.title || '',
      severity: item.severity,
      source: item.metric || item.source || '',
      status: item.status,
      createdAt: item.created_at || item.createdAt,
    }));
    return {
      ...response,
      data: {
        list,
        total: response.total || 0,
      },
    };
  },

  // 获取告警规则列表
  async getAlertRuleList(params?: AlertRuleListParams): Promise<ApiResponse<PaginatedResponse<AlertRule>>> {
    const response = await apiService.get<AlertRule[]>('/alert-rules', {
      params: {
        page: params?.page,
        page_size: params?.pageSize,
      },
    });
    const list = (response.data || []).map((item: any) => ({
      id: String(item.id),
      name: item.name,
      condition: `${item.metric} ${item.operator} ${item.threshold}`,
      severity: item.severity,
      enabled: item.enabled,
      channels: item.channels || [],
      createdAt: item.created_at || item.createdAt,
      metric: item.metric,
      operator: item.operator,
      threshold: item.threshold,
    }));
    return {
      ...response,
      data: {
        list,
        total: response.total || 0,
      },
    };
  },

  // 获取监控指标
  async getMetrics(params: MetricParams): Promise<ApiResponse<MetricData[]>> {
    return apiService.get('/metrics', {
      params: {
        metric: params.metric,
        start_time: params.startTime,
        end_time: params.endTime,
      },
    });
  },
};
