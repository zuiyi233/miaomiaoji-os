import { apiRequest, setToken } from './apiClient';

export type LoginRequest = {
  username: string;
  password: string;
};

export type RegisterRequest = {
  username: string;
  password: string;
  email?: string;
  nickname?: string;
};

export type AuthUserDTO = {
  id: number;
  username: string;
  nickname?: string;
  email?: string;
  role: 'user' | 'admin';
  points?: number;
  check_in_streak?: number;
};

export type AuthTokenResponseDTO = {
  token: string;
  expires_in: number;
  user: AuthUserDTO;
};

export async function loginApi(payload: LoginRequest): Promise<AuthTokenResponseDTO> {
  const data = await apiRequest<AuthTokenResponseDTO>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify(payload),
    skipAuth: true,
  });

  if (data?.token) {
    setToken(data.token);
  }

  return data;
}

export async function registerApi(payload: RegisterRequest): Promise<AuthTokenResponseDTO> {
  const data = await apiRequest<AuthTokenResponseDTO>('/api/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify(payload),
    skipAuth: true,
  });

  if (data?.token) {
    setToken(data.token);
  }

  return data;
}

export async function logoutApi(): Promise<void> {
  await apiRequest<void>('/api/v1/auth/logout', {
    method: 'POST',
  });
  setToken(null);
}
