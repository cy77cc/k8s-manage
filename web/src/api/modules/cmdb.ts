import apiService from '../api';
import type { ApiResponse } from '../api';

export interface CMDBAsset {
  id: string;
  assetType: string;
  source: string;
  name: string;
  status: string;
  owner: string;
  attrsJson?: string;
}

export interface CMDBRelation {
  id: string;
  fromAssetId: string;
  toAssetId: string;
  relationType: string;
}

export const cmdbApi = {
  async listAssets(params?: { assetType?: string; status?: string; keyword?: string }): Promise<ApiResponse<CMDBAsset[]>> {
    const res = await apiService.get<any[]>('/cmdb/assets', {
      params: {
        asset_type: params?.assetType,
        status: params?.status,
        keyword: params?.keyword,
      },
    });
    return {
      ...res,
      data: (res.data || []).map((x: any) => ({
        id: String(x.id),
        assetType: x.asset_type,
        source: x.source,
        name: x.name,
        status: x.status,
        owner: x.owner,
        attrsJson: x.attrs_json,
      })),
    };
  },

  async createAsset(payload: {
    assetType: string;
    name: string;
    source?: string;
    status?: string;
    owner?: string;
    attrsJson?: string;
  }): Promise<ApiResponse<CMDBAsset>> {
    return apiService.post('/cmdb/assets', {
      asset_type: payload.assetType,
      name: payload.name,
      source: payload.source,
      status: payload.status,
      owner: payload.owner,
      attrs_json: payload.attrsJson,
    });
  },

  async listRelations(): Promise<ApiResponse<CMDBRelation[]>> {
    const res = await apiService.get<any[]>('/cmdb/relations');
    return {
      ...res,
      data: (res.data || []).map((x: any) => ({
        id: String(x.id),
        fromAssetId: String(x.from_asset_id),
        toAssetId: String(x.to_asset_id),
        relationType: x.relation_type,
      })),
    };
  },

  async createRelation(payload: { fromAssetId: string; toAssetId: string; relationType: string }): Promise<ApiResponse<CMDBRelation>> {
    return apiService.post('/cmdb/relations', {
      from_asset_id: Number(payload.fromAssetId),
      to_asset_id: Number(payload.toAssetId),
      relation_type: payload.relationType,
    });
  },

  async listChanges(params?: { assetId?: string }): Promise<ApiResponse<any[]>> {
    return apiService.get('/cmdb/changes', { params: { asset_id: params?.assetId } });
  },
};
