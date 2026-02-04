import { apiRequest } from './apiClient';

export type ProviderConfigDTO = {
  provider: string;
  base_url: string;
  api_key: string;
};

export async function fetchProviderConfigApi(provider: string): Promise<ProviderConfigDTO> {
  const params = new URLSearchParams({ provider });
  return apiRequest<ProviderConfigDTO>(`/api/v1/ai/providers?${params.toString()}`, { method: 'GET' });
}

export async function updateProviderConfigApi(payload: {
  provider: string;
  base_url: string;
  api_key: string;
}): Promise<void> {
  await apiRequest('/api/v1/ai/providers', {
    method: 'PUT',
    body: JSON.stringify(payload)
  });
}

export async function testProviderConfigApi(provider: string): Promise<void> {
  await apiRequest('/api/v1/ai/providers/test', {
    method: 'POST',
    body: JSON.stringify({ provider })
  });
}
