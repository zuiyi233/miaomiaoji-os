import { apiRequest } from './apiClient';
import { listProjectsApi } from './projectApi';

export type SessionDTO = {
  id: number;
  title: string;
  mode: string;
  project_id: number;
  user_id: number;
  created_at: string;
  updated_at: string;
};

export type SessionStepDTO = {
  id: number;
  title: string;
  content: string;
  format_type: string;
  order_index: number;
  session_id: number;
  metadata?: Record<string, any> | null;
  created_at: string;
  updated_at: string;
};

export type SessionListResponse = {
  sessions: SessionDTO[];
  total: number;
  page: number;
  page_size: number;
};

export type RunWorkflowPayload = {
  project_id: number;
  session_id?: number;
  title?: string;
  step_title?: string;
  provider: string;
  path: string;
  body: string;
};

export type RunWorkflowResponse = {
  session: SessionDTO;
  step: SessionStepDTO;
  content: string;
  raw: any;
};

export async function listSessionsApi(page = 1, pageSize = 20): Promise<SessionListResponse> {
  return apiRequest<SessionListResponse>(`/api/v1/sessions?page=${page}&page_size=${pageSize}`, {
    method: 'GET',
  });
}

export async function listSessionsByProjectApi(projectId: number, page = 1, pageSize = 20): Promise<SessionListResponse> {
  return apiRequest<SessionListResponse>(
    `/api/v1/sessions/projects/${projectId}?page=${page}&page_size=${pageSize}`,
    { method: 'GET' }
  );
}

export async function getSessionApi(sessionId: number): Promise<SessionDTO> {
  return apiRequest<SessionDTO>(`/api/v1/sessions/${sessionId}`, { method: 'GET' });
}

export async function listStepsApi(sessionId: number): Promise<SessionStepDTO[]> {
  return apiRequest<SessionStepDTO[]>(`/api/v1/sessions/${sessionId}/steps`, { method: 'GET' });
}

export async function runWorldWorkflowApi(payload: RunWorkflowPayload): Promise<RunWorkflowResponse> {
  return apiRequest<RunWorkflowResponse>('/api/v1/workflows/world', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function runPolishWorkflowApi(payload: RunWorkflowPayload): Promise<RunWorkflowResponse> {
  return apiRequest<RunWorkflowResponse>('/api/v1/workflows/polish', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function resolveProjectIdByExternalId(externalId: string): Promise<number | null> {
  if (!externalId) return null;
  const result = await listProjectsApi(1, 200);
  const match = (result.list || []).find((item) => item.external_id === externalId);
  return match?.id || null;
}
