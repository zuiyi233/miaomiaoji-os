
import { Project, Plugin, PluginActionResponse, Document, StoryEntity } from "../types";
import { apiRequest } from './apiClient';

export type PluginDTO = {
  id: number;
  name: string;
  version?: string;
  author?: string;
  description?: string;
  endpoint?: string;
  entry_point?: string;
  is_enabled?: boolean;
  status?: string;
  healthy?: boolean;
  latency_ms?: number;
  last_ping?: string;
  capabilities?: {
    id: number;
    cap_id: string;
    name: string;
    type: string;
    description?: string;
    icon?: string;
  }[];
  config?: Record<string, any>;
};

export type PluginListDTO = {
  plugins: PluginDTO[];
  total: number;
  page: number;
  page_size: number;
};

export function mapPluginFromApi(dto: PluginDTO): Plugin {
  return {
    id: String(dto.id),
    name: dto.name,
    version: dto.version || '1.0.0',
    author: dto.author || 'Unknown',
    description: dto.description || '',
    endpoint: dto.endpoint || '',
    isEnabled: !!dto.is_enabled,
    capabilities:
      dto.capabilities?.map((cap) => ({
        id: String(cap.id || cap.cap_id),
        name: cap.name,
        type: (cap.type as any) || 'text_processor',
        description: cap.description || '',
        icon: cap.icon || '',
      })) || [],
    status: dto.healthy ? 'online' : 'offline',
    latency: dto.latency_ms || undefined,
    lastPing: dto.last_ping ? Date.parse(dto.last_ping) : undefined,
    config: dto.config || {},
  };
}

export async function fetchPluginsApi(page: number = 1, pageSize: number = 20): Promise<PluginListDTO> {
  const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
  return apiRequest<PluginListDTO>(`/api/v1/plugins?${params.toString()}`, { method: 'GET' });
}

export async function enablePluginApi(id: string): Promise<void> {
  await apiRequest<void>(`/api/v1/plugins/${id}/enable`, { method: 'PUT' });
}

export async function disablePluginApi(id: string): Promise<void> {
  await apiRequest<void>(`/api/v1/plugins/${id}/disable`, { method: 'PUT' });
}

export async function pingPluginApi(id: string): Promise<void> {
  await apiRequest<void>(`/api/v1/plugins/${id}/ping`, { method: 'POST' });
}

export async function createPluginApi(payload: {
  name: string;
  version?: string;
  author?: string;
  description?: string;
  endpoint: string;
  entry_point?: string;
}): Promise<PluginDTO> {
  return apiRequest<PluginDTO>('/api/v1/plugins', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function deletePluginApi(id: string): Promise<void> {
  await apiRequest<void>(`/api/v1/plugins/${id}`, { method: 'DELETE' });
}

export interface PluginCallResult {
  success: boolean;
  actions: PluginActionResponse[];
  error?: string;
  latency?: number;
}

export const pingPlugin = async (endpoint: string): Promise<{ manifest: Partial<Plugin>, latency: number } | null> => {
  const start = Date.now();
  try {
    // Timeout handling for ping
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000);

    const response = await fetch(`${endpoint}/manifest`, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
      signal: controller.signal
    });
    
    clearTimeout(timeoutId);

    if (response.ok) {
      const manifest = await response.json();
      return { manifest, latency: Date.now() - start };
    }
  } catch (error) {
    console.error("Plugin Ping Failed:", error);
  }
  return null;
};

export const callPluginAction = async (
  plugin: Plugin,
  actionId: string,
  project: Project,
  activeDoc?: Document,
  payload?: any
): Promise<PluginCallResult> => {
  if (!plugin.isEnabled) {
    return { success: false, actions: [], error: "Plugin is currently disabled." };
  }

  const start = Date.now();
  try {
    const response = await fetch(`${plugin.endpoint}/action`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        actionId,
        pluginConfig: plugin.config || {},
        context: {
          project: {
            id: project.id,
            title: project.title,
            genre: project.genre,
            worldRules: project.worldRules,
            entities: project.entities,
          },
          activeDocument: activeDoc
        },
        payload
      })
    });

    const latency = Date.now() - start;

    if (!response.ok) {
      let errorMsg = `Server returned ${response.status}: ${response.statusText}`;
      try {
        const errData = await response.json();
        if (errData.message) errorMsg = errData.message;
      } catch (e) { /* ignore parse error */ }
      
      return { success: false, actions: [], error: errorMsg, latency };
    }

    const data = await response.json();
    return { 
      success: true,
      actions: Array.isArray(data) ? data : [data],
      latency
    };
  } catch (error: any) {
    let errorMsg = "Network error: Service unreachable or CORS blocked.";
    if (error.name === 'AbortError') errorMsg = "Request timed out.";
    
    return { 
      success: false, 
      actions: [], 
      error: errorMsg,
      latency: Date.now() - start 
    };
  }
};

export const executePluginActions = (
  actions: PluginActionResponse[],
  handlers: {
    updateDocument: (id: string, updates: Partial<Document>) => void,
    updateEntity: (id: string, updates: Partial<StoryEntity>) => void,
    activeDocId: string | null,
    onMessage: (msg: string, type: 'info' | 'error' | 'success') => void
  }
) => {
  actions.forEach(action => {
    try {
      switch (action.type) {
        case 'update_document':
          if (handlers.activeDocId && action.payload) {
            handlers.updateDocument(handlers.activeDocId, action.payload);
          }
          break;
        case 'update_entity':
          if (action.payload?.id) {
            handlers.updateEntity(action.payload.id, action.payload);
          }
          break;
        case 'show_message':
          if (action.payload?.text) {
            handlers.onMessage(action.payload.text, action.payload.type || 'info');
          }
          break;
        case 'add_log':
          console.log(`[Plugin Log]:`, action.payload);
          break;
        default:
          console.warn(`Plugin action type "${action.type}" is not supported by this version of NAO.`);
      }
    } catch (e) {
      console.error("Action execution failed:", e, action);
      handlers.onMessage(`Failed to execute action [${action.type}]: ${String(e)}`, 'error');
    }
  });
};
