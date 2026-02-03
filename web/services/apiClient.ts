export type ApiError = {
  code?: number;
  message: string;
  status?: number;
};

export type ApiEnvelope<T> = {
  code: number;
  message: string;
  data: T;
};

const TOKEN_KEY = 'nao_jwt_token_v1';

export function getApiBaseUrl(): string {
  const raw = (import.meta as any).env?.VITE_API_BASE_URL as string | undefined;
  return (raw || '').replace(/\/$/, '');
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string | null) {
  if (!token) {
    localStorage.removeItem(TOKEN_KEY);
    return;
  }
  localStorage.setItem(TOKEN_KEY, token);
}

function buildUrl(path: string): string {
  const base = getApiBaseUrl();
  if (!path.startsWith('/')) path = `/${path}`;
  return base ? `${base}${path}` : path;
}

async function parseJsonSafe(res: Response): Promise<any> {
  try {
    return await res.json();
  } catch {
    return null;
  }
}

let refreshPromise: Promise<string | null> | null = null;

async function refreshToken(): Promise<string | null> {
  if (!refreshPromise) {
    refreshPromise = (async () => {
      const token = getToken();
      if (!token) return null;

      const res = await fetch(buildUrl('/api/v1/auth/refresh'), {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      const body = await parseJsonSafe(res);
      if (!res.ok) return null;
      if (!body || typeof body.code !== 'number') return null;
      if (body.code !== 0) return null;

      const newToken = body.data?.token;
      if (typeof newToken !== 'string' || !newToken) return null;

      setToken(newToken);
      return newToken;
    })().finally(() => {
      refreshPromise = null;
    });
  }

  return refreshPromise;
}

export async function apiRequest<T>(
  path: string,
  init: RequestInit & { skipAuth?: boolean } = {}
): Promise<T> {
  const url = buildUrl(path);
  if (url.includes('/api/v1/projects/')) {
    throw new Error('API 路径不应包含末尾斜杠: /api/v1/projects/');
  }

  const headers = new Headers(init.headers || undefined);

  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  if (!init.skipAuth) {
    const token = getToken();
    if (token) headers.set('Authorization', `Bearer ${token}`);
  }

  const doFetch = async (): Promise<Response> => {
    return fetch(url, {
      ...init,
      headers,
    });
  };

  let res = await doFetch();

  if (res.status === 401 && !init.skipAuth) {
    const newToken = await refreshToken();
    if (newToken) {
      headers.set('Authorization', `Bearer ${newToken}`);
      res = await doFetch();
    }
  }

  const body = await parseJsonSafe(res);

  if (!res.ok) {
    const msg = body?.message || res.statusText || '请求失败';
    const err: ApiError = { status: res.status, message: msg, code: body?.code };
    throw err;
  }

  if (!body || typeof body.code !== 'number') {
    throw { message: '响应格式错误', status: res.status } as ApiError;
  }

  if (body.code !== 0) {
    throw { message: body.message || '请求失败', code: body.code, status: res.status } as ApiError;
  }

  return body.data as T;
}
