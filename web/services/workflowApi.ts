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

export type ChapterWriteBack = {
  mode?: string;
  set_status?: string;
  set_summary?: boolean;
};

export type ChapterGeneratePayload = {
  project_id: number;
  session_id?: number;
  document_id?: number;
  volume_id?: number;
  title?: string;
  order_index?: number;
  provider: string;
  path: string;
  body: string;
  write_back?: ChapterWriteBack;
};

export type ChapterAnalyzePayload = {
  project_id: number;
  session_id?: number;
  document_id: number;
  provider: string;
  path: string;
  body: string;
  write_back?: ChapterWriteBack;
};

export type ChapterRewritePayload = {
  project_id: number;
  session_id?: number;
  document_id: number;
  rewrite_mode?: string;
  provider: string;
  path: string;
  body: string;
  write_back?: ChapterWriteBack;
};

export type ChapterBatchItem = {
  client_document_id?: string;
  title: string;
  order_index?: number;
  outline: string;
};

export type ChapterBatchPayload = {
  project_id: number;
  session_id?: number;
  volume_id?: number;
  items: ChapterBatchItem[];
  provider: string;
  path: string;
  body_template: string;
  write_back?: ChapterWriteBack;
};

export type RunWorkflowResponse = {
  session: SessionDTO;
  step: SessionStepDTO;
  content: string;
  raw: any;
};

export type ChapterGenerateResponse = {
  session: SessionDTO;
  document: any;
  steps: SessionStepDTO[];
  content: string;
  raw: any;
};

export type ChapterAnalyzeResponse = {
  session: SessionDTO;
  document: any;
  content: string;
  raw: any;
};

export type ChapterRewriteResponse = {
  session: SessionDTO;
  document: any;
  content: string;
  raw: any;
};

export type ChapterBatchResponse = {
  session: SessionDTO;
  documents: any[];
  results?: Array<{ client_document_id: string; document: any }>;
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

export async function runWizardWorldWorkflowApi(payload: RunWorkflowPayload): Promise<RunWorkflowResponse> {
	return apiRequest<RunWorkflowResponse>('/api/v1/workflows/wizard/world', {
		method: 'POST',
		body: JSON.stringify(payload),
	});
}

export async function runWizardCharactersWorkflowApi(payload: RunWorkflowPayload): Promise<RunWorkflowResponse> {
	return apiRequest<RunWorkflowResponse>('/api/v1/workflows/wizard/characters', {
		method: 'POST',
		body: JSON.stringify(payload),
	});
}

export async function runWizardOutlineWorkflowApi(payload: RunWorkflowPayload): Promise<RunWorkflowResponse> {
	return apiRequest<RunWorkflowResponse>('/api/v1/workflows/wizard/outline', {
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

export async function runChapterGenerateApi(payload: ChapterGeneratePayload): Promise<ChapterGenerateResponse> {
  return apiRequest<ChapterGenerateResponse>('/api/v1/workflows/chapters/generate', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function runChapterAnalyzeApi(payload: ChapterAnalyzePayload): Promise<ChapterAnalyzeResponse> {
  return apiRequest<ChapterAnalyzeResponse>('/api/v1/workflows/chapters/analyze', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function runChapterRewriteApi(payload: ChapterRewritePayload): Promise<ChapterRewriteResponse> {
  return apiRequest<ChapterRewriteResponse>('/api/v1/workflows/chapters/rewrite', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function runChapterBatchApi(payload: ChapterBatchPayload): Promise<ChapterBatchResponse> {
  return apiRequest<ChapterBatchResponse>('/api/v1/workflows/chapters/batch', {
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
