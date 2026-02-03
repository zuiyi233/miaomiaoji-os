import { apiRequest } from './apiClient';

export type ProfileDTO = {
  id: number;
  username: string;
  nickname?: string;
  email?: string;
  role: 'user' | 'admin';
  points?: number;
  check_in_streak?: number;
};

export async function fetchProfileApi(): Promise<ProfileDTO> {
  // 注意：后端 Gin 路由注册为 GET /api/v1/users/profile（无末尾 /），此处保持一致
  return apiRequest<ProfileDTO>('/api/v1/users/profile', {
    method: 'GET',
  });
}
