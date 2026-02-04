import { apiRequest } from './apiClient';

export type RedeemResponse = {
  code: string;
  duration_days: number;
  ai_access_until: string;
  used_count: number;
  status: string;
};

export type RedemptionCodeDTO = {
  code: string;
  status: string;
  expires_at?: string;
  max_uses: number;
  used_count: number;
  duration_days: number;
  created_by: number;
  batch_id?: string;
  prefix?: string;
  tags?: string[];
  note?: string;
  source?: string;
  created_at?: string;
};

export type CodeListDTO = {
  list: RedemptionCodeDTO[];
  page_info: { page: number; size: number; total: number };
};

export async function redeemCodeApi(payload: {
  request_id: string;
  idempotency_key?: string;
  code: string;
  device_id?: string;
  client_time?: string;
  app_id?: string;
  platform?: string;
  app_version?: string;
  result_status?: string;
  result_error_code?: string;
  entitlement_delta?: Record<string, any>;
}): Promise<RedeemResponse> {
  return apiRequest<RedeemResponse>('/api/v1/codes/redeem', {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

export async function fetchCodesApi(page = 1, size = 50, status = 'all', search = '', sort = 'desc'): Promise<CodeListDTO> {
  const params = new URLSearchParams({
    page: String(page),
    size: String(size),
    status,
    search,
    sort
  });
  return apiRequest<CodeListDTO>(`/api/v1/codes?${params.toString()}`, { method: 'GET' });
}

export async function generateCodesApi(payload: {
  prefix: string;
  length: number;
  count: number;
  validity_days: number;
  max_uses: number;
  tags: string[];
  note: string;
  source: string;
}): Promise<{ list: RedemptionCodeDTO[] }> {
  return apiRequest<{ list: RedemptionCodeDTO[] }>('/api/v1/codes/generate', {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

export async function batchUpdateCodesApi(payload: {
  codes: string[];
  action: 'disable' | 'enable' | 'delete' | 'renew';
  value?: number;
}): Promise<void> {
  await apiRequest('/api/v1/codes/batch', {
    method: 'PUT',
    body: JSON.stringify(payload)
  });
}

export function getExportCodesUrl(status = 'all', search = ''): string {
  const params = new URLSearchParams({ status, search });
  return `/api/v1/codes/export?${params.toString()}`;
}
