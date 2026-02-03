import { apiRequest } from './apiClient';

export type ProjectListItemDTO = {
  id: number;
  title: string;
  genre?: string;
  tags?: string[];
  core_conflict?: string;
  character_arc?: string;
  ultimate_value?: string;
  world_rules?: string;
  ai_settings?: any;
  user_id?: number;
  created_at?: string;
  updated_at?: string;
};

export type PageInfoDTO = {
  page: number;
  size: number;
  total: number;
};

export type ProjectListDTO = {
  list: ProjectListItemDTO[];
  page_info: PageInfoDTO;
};

export async function fetchProjectsApi(page: number = 1, size: number = 50): Promise<ProjectListDTO> {
  const params = new URLSearchParams({ page: String(page), size: String(size) });
  return apiRequest<ProjectListDTO>(`/api/v1/projects?${params.toString()}`, {
    method: 'GET',
  });
}
