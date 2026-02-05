import { apiRequest } from './apiClient';

export type DocumentDTO = {
  id: number;
  title: string;
  content: string;
  summary: string;
  status: string;
  order_index: number;
  time_node: string;
  duration: string;
  target_word_count: number;
  chapter_goal: string;
  core_plot: string;
  hook: string;
  cause_effect: string;
  foreshadowing_details: string;
  project_id: number;
  volume_id: number;
  created_at: string;
  updated_at: string;
};

export type CreateDocumentPayload = {
  title: string;
  content?: string;
  summary?: string;
  status?: string;
  order_index: number;
  time_node?: string;
  duration?: string;
  target_word_count?: number;
  chapter_goal?: string;
  core_plot?: string;
  hook?: string;
  cause_effect?: string;
  foreshadowing_details?: string;
  volume_id?: number;
};

export async function createDocumentApi(projectId: number, payload: CreateDocumentPayload): Promise<DocumentDTO> {
  return apiRequest<DocumentDTO>(`/api/v1/projects/${projectId}/documents`, {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}
