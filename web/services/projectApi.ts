import { apiRequest } from './apiClient';
import type { Project } from '../types';

type ProjectDTO = {
  id: number;
  external_id?: string;
  title: string;
  genre?: string;
  tags?: string[];
  core_conflict?: string;
  character_arc?: string;
  ultimate_value?: string;
  world_rules?: string;
  ai_settings?: Record<string, any>;
  snapshot?: Record<string, any>;
  user_id: number;
  created_at: string;
  updated_at: string;
};

type SnapshotPayload = {
  external_id: string;
  title?: string;
  ai_settings?: Record<string, any>;
  snapshot: Record<string, any>;
};

type ProjectListResponse = {
  list: ProjectDTO[];
  page_info: { page: number; size: number; total: number };
};

export function projectFromSnapshot(dto: ProjectDTO): Project | null {
  if (!dto.snapshot) return null;
  return dto.snapshot as Project;
}

export async function listProjectsApi(page = 1, size = 100): Promise<ProjectListResponse> {
  return apiRequest<ProjectListResponse>(`/api/v1/projects?page=${page}&size=${size}`, {
    method: 'GET',
  });
}

export async function upsertProjectSnapshotApi(payload: SnapshotPayload): Promise<ProjectDTO> {
  return apiRequest<ProjectDTO>('/api/v1/projects/snapshot', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function backupProjectSnapshotApi(payload: SnapshotPayload): Promise<{ file_id: number; file_name: string; storage_key: string; project_id: number }> {
  return apiRequest<{ file_id: number; file_name: string; storage_key: string; project_id: number }>(
    '/api/v1/projects/backup',
    {
      method: 'POST',
      body: JSON.stringify(payload),
    }
  );
}

export async function getLatestBackupApi(projectId: number): Promise<{ id: number; file_name: string; storage_key: string; created_at: string }> {
  return apiRequest<{ id: number; file_name: string; storage_key: string; created_at: string }>(
    `/api/v1/projects/${projectId}/backup/latest`,
    { method: 'GET' }
  );
}

export async function listProjectBackupsApi(projectId: number, page = 1, pageSize = 20) {
  return apiRequest<{ files: Array<{ id: number; file_name: string; created_at: string; size_bytes: number }>; total: number; page: number; page_size: number }>(
    `/api/v1/files/project/${projectId}?page=${page}&page_size=${pageSize}&file_type=backup`,
    { method: 'GET' }
  );
}

export function getBackupDownloadUrl(fileId: number) {
  return `/api/v1/files/${fileId}/download`;
}

export async function downloadBackupFile(fileId: number) {
  const url = getBackupDownloadUrl(fileId);
  const res = await fetch(url, {
    headers: {
      Authorization: `Bearer ${localStorage.getItem('nao_jwt_token_v1') || ''}`
    }
  });
  if (!res.ok) throw new Error('下载失败');
  return res.json();
}

export function toSnapshotPayload(project: Project, externalId: string): SnapshotPayload {
  return {
    external_id: externalId,
    title: project.title,
    ai_settings: project.aiSettings as Record<string, any>,
    snapshot: project as unknown as Record<string, any>,
  };
}
