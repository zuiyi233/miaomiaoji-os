// SSE 客户端 - Stream Adapter 模式
// 支持 EventSource (GET) 和 Fetch Stream (POST/需要 Header)

import { getToken, getApiBaseUrl } from './apiClient';

// SSE 事件类型（与后端 pkg/sse/sse.go 对应）
export type SSEEventType =
  | 'step.appended'
  | 'quality.checked'
  | 'export.ready'
  | 'progress.updated'
  | 'workflow.done'
  | 'job.created'
  | 'job.started'
  | 'job.progress'
  | 'job.failed'
  | 'job.succeeded'
  | 'job.canceled'
  | 'error';

// SSE 事件数据结构
export interface SSEEventData<T = unknown> {
  type: SSEEventType;
  data: T;
  timestamp: string;
}

// 连接状态
export type ConnectionState = 'connecting' | 'connected' | 'disconnected' | 'reconnecting';

// SSE 客户端配置
export interface SSEClientConfig {
  sessionId: string | number;
  onEvent?: (event: SSEEventData) => void;
  onStepAppended?: (data: StepAppendedData) => void;
  onQualityChecked?: (data: QualityCheckedData) => void;
  onExportReady?: (data: ExportReadyData) => void;
  onProgressUpdated?: (data: ProgressUpdatedData) => void;
  onWorkflowDone?: (data: WorkflowDoneData) => void;
  onError?: (error: SSEErrorData) => void;
  onStateChange?: (state: ConnectionState) => void;
  // 重连配置
  maxRetries?: number;
  baseRetryDelay?: number;
  maxRetryDelay?: number;
  // 超时配置
  idleTimeout?: number;
}

// 事件数据类型
export interface StepAppendedData {
  step_id: number;
  title: string;
  content: string;
  job_uuid?: string;
  plugin_id?: number;
  timestamp: string;
}

export interface QualityCheckedData {
  step_id: number;
  passed: boolean;
  score: number;
  issues: string[];
  timestamp: string;
}

export interface ExportReadyData {
  export_id: string;
  format: string;
  file_url: string;
  timestamp: string;
}

export interface ProgressUpdatedData {
  progress: number;
  message?: string;
  timestamp: string;
}

export interface WorkflowDoneData {
  mode: string;
  document_id?: number;
  timestamp: string;
}

export interface SSEErrorData {
  message: string;
  error: string;
}

// 解析 SSE 行数据
function parseSSELine(line: string): { field: string; value: string } | null {
  if (!line || line.startsWith(':')) return null;
  const colonIndex = line.indexOf(':');
  if (colonIndex === -1) return { field: line, value: '' };
  const field = line.slice(0, colonIndex);
  let value = line.slice(colonIndex + 1);
  if (value.startsWith(' ')) value = value.slice(1);
  return { field, value };
}

// 计算指数退避延迟（带 jitter）
function getRetryDelay(attempt: number, baseDelay: number, maxDelay: number): number {
  const delay = Math.min(baseDelay * Math.pow(2, attempt), maxDelay);
  const jitter = delay * 0.2 * Math.random();
  return delay + jitter;
}

/**
 * SSE 客户端类
 * 使用 Fetch + ReadableStream 实现，支持自定义 Authorization header
 */
export class SSEClient {
  private config: Required<SSEClientConfig>;
  private abortController: AbortController | null = null;
  private state: ConnectionState = 'disconnected';
  private retryCount = 0;
  private retryTimeoutId: ReturnType<typeof setTimeout> | null = null;
  private idleTimeoutId: ReturnType<typeof setTimeout> | null = null;
  private accumulatedContent: Map<number, string> = new Map();

  constructor(config: SSEClientConfig) {
    this.config = {
      sessionId: config.sessionId,
      onEvent: config.onEvent || (() => {}),
      onStepAppended: config.onStepAppended || (() => {}),
      onQualityChecked: config.onQualityChecked || (() => {}),
      onExportReady: config.onExportReady || (() => {}),
      onProgressUpdated: config.onProgressUpdated || (() => {}),
      onWorkflowDone: config.onWorkflowDone || (() => {}),
      onError: config.onError || (() => {}),
      onStateChange: config.onStateChange || (() => {}),
      maxRetries: config.maxRetries ?? 10,
      baseRetryDelay: config.baseRetryDelay ?? 1000,
      maxRetryDelay: config.maxRetryDelay ?? 30000,
      idleTimeout: config.idleTimeout ?? 60000,
    };
  }

  // 获取累积内容（用于断线恢复）
  getAccumulatedContent(stepId: number): string {
    return this.accumulatedContent.get(stepId) || '';
  }

  // 追加内容到累积缓存
  private appendContent(stepId: number, content: string): void {
    const existing = this.accumulatedContent.get(stepId) || '';
    this.accumulatedContent.set(stepId, existing + content);
  }

  // 清空累积内容
  clearAccumulatedContent(): void {
    this.accumulatedContent.clear();
  }

  // 更新连接状态
  private setState(newState: ConnectionState): void {
    if (this.state !== newState) {
      this.state = newState;
      this.config.onStateChange(newState);
    }
  }

  // 重置空闲超时
  private resetIdleTimeout(): void {
    if (this.idleTimeoutId) {
      clearTimeout(this.idleTimeoutId);
    }
    this.idleTimeoutId = setTimeout(() => {
      console.warn('[SSE] Idle timeout - no events received');
      this.reconnect();
    }, this.config.idleTimeout);
  }

  // 处理接收到的事件
  private handleEvent(eventType: string, eventData: string): void {
    this.resetIdleTimeout();
    this.retryCount = 0;

    try {
      const parsed: SSEEventData = JSON.parse(eventData);
      if (!parsed.type && eventType) {
        parsed.type = eventType as SSEEventType;
      }
      this.config.onEvent(parsed);

      switch (parsed.type) {
        case 'step.appended':
          const stepData = parsed.data as StepAppendedData;
          if (stepData.step_id && stepData.content) {
            this.appendContent(stepData.step_id, stepData.content);
          }
          this.config.onStepAppended(stepData);
          break;
        case 'quality.checked':
          this.config.onQualityChecked(parsed.data as QualityCheckedData);
          break;
        case 'export.ready':
          this.config.onExportReady(parsed.data as ExportReadyData);
          break;
        case 'progress.updated':
          this.config.onProgressUpdated(parsed.data as ProgressUpdatedData);
          break;
        case 'workflow.done':
          this.config.onWorkflowDone(parsed.data as WorkflowDoneData);
          break;
        case 'error':
          this.config.onError(parsed.data as SSEErrorData);
          break;
      }
    } catch (e) {
      console.error('[SSE] Failed to parse event data:', e, eventData);
    }
  }

  // 连接到 SSE 流
  async connect(): Promise<void> {
    if (this.state === 'connected' || this.state === 'connecting') {
      return;
    }

    this.setState('connecting');
    this.abortController = new AbortController();

    const token = getToken();
    const baseUrl = getApiBaseUrl();
    const url = `${baseUrl}/api/v1/sse/stream?session_id=${this.config.sessionId}`;

    try {
      const response = await fetch(url, {
        method: 'GET',
        headers: {
          'Accept': 'text/event-stream',
          'Cache-Control': 'no-cache',
          ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
        },
        signal: this.abortController.signal,
      });

      if (!response.ok) {
        throw new Error(`SSE connection failed: ${response.status} ${response.statusText}`);
      }

      if (!response.body) {
        throw new Error('SSE response has no body');
      }

      this.setState('connected');
      this.resetIdleTimeout();

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      let currentEventType = '';
      let currentEventData = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (line === '') {
            // 空行表示事件结束
            if (currentEventData) {
              this.handleEvent(currentEventType, currentEventData);
            }
            currentEventType = '';
            currentEventData = '';
            continue;
          }

          const parsed = parseSSELine(line);
          if (!parsed) continue;

          switch (parsed.field) {
            case 'event':
              currentEventType = parsed.value;
              break;
            case 'data':
              currentEventData += (currentEventData ? '\n' : '') + parsed.value;
              break;
          }
        }
      }
    } catch (error: unknown) {
      if (error instanceof Error && error.name === 'AbortError') {
        this.setState('disconnected');
        return;
      }
      console.error('[SSE] Connection error:', error);
      this.reconnect();
    }
  }

  // 重连逻辑
  private reconnect(): void {
    if (this.retryCount >= this.config.maxRetries) {
      console.error('[SSE] Max retries reached, giving up');
      this.setState('disconnected');
      this.config.onError({ message: '连接失败，已达最大重试次数', error: 'MAX_RETRIES_EXCEEDED' });
      return;
    }

    this.setState('reconnecting');
    const delay = getRetryDelay(this.retryCount, this.config.baseRetryDelay, this.config.maxRetryDelay);
    this.retryCount++;

    console.log(`[SSE] Reconnecting in ${Math.round(delay)}ms (attempt ${this.retryCount}/${this.config.maxRetries})`);

    this.retryTimeoutId = setTimeout(() => {
      this.connect();
    }, delay);
  }

  // 断开连接
  disconnect(): void {
    if (this.abortController) {
      this.abortController.abort();
      this.abortController = null;
    }
    if (this.retryTimeoutId) {
      clearTimeout(this.retryTimeoutId);
      this.retryTimeoutId = null;
    }
    if (this.idleTimeoutId) {
      clearTimeout(this.idleTimeoutId);
      this.idleTimeoutId = null;
    }
    this.setState('disconnected');
  }

  // 获取当前状态
  getState(): ConnectionState {
    return this.state;
  }
}

/**
 * EventSource 适配器（用于同域/Cookie 认证场景）
 * 注意：EventSource 不支持自定义 header，仅适用于 Cookie 认证
 */
export class EventSourceAdapter {
  private eventSource: EventSource | null = null;
  private config: Required<SSEClientConfig>;
  private state: ConnectionState = 'disconnected';
  private retryCount = 0;
  private retryTimeoutId: ReturnType<typeof setTimeout> | null = null;

  constructor(config: SSEClientConfig) {
    this.config = {
      sessionId: config.sessionId,
      onEvent: config.onEvent || (() => {}),
      onStepAppended: config.onStepAppended || (() => {}),
      onQualityChecked: config.onQualityChecked || (() => {}),
      onExportReady: config.onExportReady || (() => {}),
      onProgressUpdated: config.onProgressUpdated || (() => {}),
      onWorkflowDone: config.onWorkflowDone || (() => {}),
      onError: config.onError || (() => {}),
      onStateChange: config.onStateChange || (() => {}),
      maxRetries: config.maxRetries ?? 10,
      baseRetryDelay: config.baseRetryDelay ?? 1000,
      maxRetryDelay: config.maxRetryDelay ?? 30000,
      idleTimeout: config.idleTimeout ?? 60000,
    };
  }

  private setState(newState: ConnectionState): void {
    if (this.state !== newState) {
      this.state = newState;
      this.config.onStateChange(newState);
    }
  }

  connect(): void {
    if (this.state === 'connected' || this.state === 'connecting') {
      return;
    }

    this.setState('connecting');
    const baseUrl = getApiBaseUrl();
    const url = `${baseUrl}/api/v1/sse/stream?session_id=${this.config.sessionId}`;

    this.eventSource = new EventSource(url, { withCredentials: true });

    this.eventSource.onopen = () => {
      this.setState('connected');
      this.retryCount = 0;
    };

    this.eventSource.onerror = () => {
      this.eventSource?.close();
      this.reconnect();
    };

    // 监听各类事件
    const eventTypes: SSEEventType[] = [
      'step.appended',
      'quality.checked',
      'export.ready',
      'progress.updated',
      'workflow.done',
      'job.created',
      'job.started',
      'job.progress',
      'job.failed',
      'job.succeeded',
      'job.canceled',
      'error',
    ];
    eventTypes.forEach(type => {
      this.eventSource?.addEventListener(type, (e: MessageEvent) => {
        try {
          const parsed: SSEEventData = JSON.parse(e.data);
          this.config.onEvent(parsed);

          switch (type) {
            case 'step.appended':
              this.config.onStepAppended(parsed.data as StepAppendedData);
              break;
            case 'quality.checked':
              this.config.onQualityChecked(parsed.data as QualityCheckedData);
              break;
            case 'export.ready':
              this.config.onExportReady(parsed.data as ExportReadyData);
              break;
            case 'progress.updated':
              this.config.onProgressUpdated(parsed.data as ProgressUpdatedData);
              break;
            case 'workflow.done':
              this.config.onWorkflowDone(parsed.data as WorkflowDoneData);
              break;
            // job.* 事件目前仅透传给 onEvent，由业务组件按需订阅
            case 'job.created':
            case 'job.started':
            case 'job.progress':
            case 'job.failed':
            case 'job.succeeded':
            case 'job.canceled':
              break;
            case 'error':
              this.config.onError(parsed.data as SSEErrorData);
              break;
          }
        } catch (err) {
          console.error('[EventSource] Failed to parse event:', err);
        }
      });
    });
  }

  private reconnect(): void {
    if (this.retryCount >= this.config.maxRetries) {
      this.setState('disconnected');
      this.config.onError({ message: '连接失败，已达最大重试次数', error: 'MAX_RETRIES_EXCEEDED' });
      return;
    }

    this.setState('reconnecting');
    const delay = getRetryDelay(this.retryCount, this.config.baseRetryDelay, this.config.maxRetryDelay);
    this.retryCount++;

    this.retryTimeoutId = setTimeout(() => {
      this.connect();
    }, delay);
  }

  disconnect(): void {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
    if (this.retryTimeoutId) {
      clearTimeout(this.retryTimeoutId);
      this.retryTimeoutId = null;
    }
    this.setState('disconnected');
  }

  getState(): ConnectionState {
    return this.state;
  }
}

/**
 * 创建 SSE 客户端的工厂函数
 * 默认使用 Fetch Stream（支持 Authorization header）
 */
export function createSSEClient(config: SSEClientConfig): SSEClient {
  return new SSEClient(config);
}

/**
 * React Hook: 使用 SSE 连接
 */
// 说明：SSEClient 需要在组件中自行管理实例与生命周期
