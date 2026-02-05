import React, { useEffect, useMemo, useState } from 'react';
import { ArrowLeft, Activity, ListChecks, Terminal, AlertCircle, Loader2, Clock, Wrench } from 'lucide-react';
import { ConnectionState, createSSEClient, ProgressUpdatedData, StepAppendedData, WorkflowDoneData, SSEEventData } from '../services/sseClient';
import { getSessionApi, listStepsApi, SessionDTO, SessionStepDTO } from '../services/workflowApi';
import { useParams } from 'react-router-dom';

interface WorkflowDetailProps {
  sessionId?: string;
  onBack: () => void;
}

type TimelineItem = {
  id: string;
  type: string;
  title: string;
  time: string;
  level?: 'info' | 'warn' | 'error';
  payload?: any;
};

const STATE_META: Record<ConnectionState, { label: string; className: string }> = {
  connecting: { label: '连接中', className: 'bg-brand-50 text-brand-700 dark:bg-brand-900/30 dark:text-brand-200' },
  connected: { label: '已连接', className: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-200' },
  disconnected: { label: '已断开', className: 'bg-rose-50 text-rose-700 dark:bg-rose-900/30 dark:text-rose-200' },
  reconnecting: { label: '重连中', className: 'bg-amber-50 text-amber-700 dark:bg-amber-900/30 dark:text-amber-200' },
};

export const WorkflowDetail: React.FC<WorkflowDetailProps> = ({ sessionId, onBack }) => {
  const params = useParams();
  const resolvedSessionId = sessionId && sessionId !== '未知' ? sessionId : params.sessionId || '';
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected');
  const [session, setSession] = useState<SessionDTO | null>(null);
  const [steps, setSteps] = useState<SessionStepDTO[]>([]);
  const [activeStepId, setActiveStepId] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [progress, setProgress] = useState<number | null>(null);
  const [doneInfo, setDoneInfo] = useState<{ mode: string; documentId?: number } | null>(null);
  const [timeline, setTimeline] = useState<TimelineItem[]>([]);

  const sessionNumericId = Number(resolvedSessionId);
  const stateMeta = STATE_META[connectionState];

  const pushTimeline = (item: TimelineItem) => {
    setTimeline((prev) => {
      const next = [...prev, item];
      return next.length > 50 ? next.slice(next.length - 50) : next;
    });
  };

  const formatTime = (value?: string) => {
    if (!value) return '--:--:--';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleTimeString('zh-CN', { hour12: false });
  };

  useEffect(() => {
    if (!resolvedSessionId || !Number.isFinite(sessionNumericId) || sessionNumericId <= 0) {
      setError('无效的会话 ID');
      return;
    }

    const fetchData = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const [sessionData, stepList] = await Promise.all([
          getSessionApi(sessionNumericId),
          listStepsApi(sessionNumericId),
        ]);
        setSession(sessionData);
        const ordered = [...stepList].sort((a, b) => a.order_index - b.order_index);
        setSteps(ordered);
        if (ordered.length > 0) {
          setActiveStepId((prev) => prev ?? ordered[ordered.length - 1].id);
        }
      } catch (err: any) {
        setError(err?.message || '加载会话失败');
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [resolvedSessionId, sessionNumericId]);

  useEffect(() => {
    if (!Number.isFinite(sessionNumericId) || sessionNumericId <= 0) return;
    const client = createSSEClient({
      sessionId: sessionNumericId,
      onEvent: (event: SSEEventData) => {
        if (!event?.type) return;
        if (event.type.startsWith('job.')) {
          pushTimeline({
            id: `${event.type}-${event.timestamp}-${Math.random()}`,
            type: event.type,
            title: `作业事件 · ${event.type.replace('job.', '')}`,
            time: formatTime(event.timestamp),
            level: event.type === 'job.failed' ? 'error' : 'info',
            payload: event.data,
          });
        }
        if (event.type === 'error') {
          pushTimeline({
            id: `${event.type}-${event.timestamp}-${Math.random()}`,
            type: event.type,
            title: '事件错误',
            time: formatTime(event.timestamp),
            level: 'error',
            payload: event.data,
          });
        }
      },
      onStateChange: (state) => setConnectionState(state),
      onStepAppended: (data: StepAppendedData) => {
        setSteps((prev) => {
          const existing = prev.find((item) => item.id === data.step_id);
          if (existing) {
            return prev.map((item) =>
              item.id === data.step_id
                ? {
                    ...item,
                    content:
                      typeof data.content === 'string'
                        ? `${item.content ?? ''}${data.content}`
                        : item.content,
                  }
                : item
            );
          }
          const metadata: Record<string, any> = {};
          if (data.job_uuid) metadata.job_uuid = data.job_uuid;
          if (typeof data.plugin_id === 'number') metadata.plugin_id = data.plugin_id;
          const next: SessionStepDTO = {
            id: data.step_id,
            title: data.title || 'AI 输出',
            content: data.content,
            format_type: 'workflow',
            order_index: prev.length + 1,
            session_id: sessionNumericId,
            metadata: Object.keys(metadata).length > 0 ? metadata : null,
            created_at: data.timestamp,
            updated_at: data.timestamp,
          };
          return [...prev, next];
        });
        setActiveStepId((prev) => prev ?? data.step_id);
        pushTimeline({
          id: `step.appended-${data.step_id}-${data.timestamp}`,
          type: 'step.appended',
          title: data.title || '新增步骤',
          time: formatTime(data.timestamp),
          level: 'info',
          payload: data,
        });
      },
      onProgressUpdated: (data: ProgressUpdatedData) => {
        if (typeof data.progress === 'number') {
          const safe = Math.max(0, Math.min(100, Math.round(data.progress)));
          setProgress(safe);
        }
        pushTimeline({
          id: `progress.updated-${data.timestamp}-${Math.random()}`,
          type: 'progress.updated',
          title: data.message || '进度更新',
          time: formatTime(data.timestamp),
          level: 'info',
          payload: data,
        });
      },
      onWorkflowDone: (data: WorkflowDoneData) => {
        setDoneInfo({ mode: data.mode || 'workflow', documentId: data.document_id });
        setProgress(100);
        pushTimeline({
          id: `workflow.done-${data.timestamp}`,
          type: 'workflow.done',
          title: `流程完成 · ${data.mode || 'workflow'}`,
          time: formatTime(data.timestamp),
          level: 'info',
          payload: data,
        });
      },
      onError: (data) => {
        setError(data?.message || 'SSE 事件错误');
        pushTimeline({
          id: `error-${Date.now()}`,
          type: 'error',
          title: data?.message || 'SSE 错误',
          time: formatTime(new Date().toISOString()),
          level: 'error',
          payload: data,
        });
      },
    });
    client.connect();
    return () => client.disconnect();
  }, [sessionNumericId, session?.mode]);

  const activeStep = useMemo(() => steps.find((item) => item.id === activeStepId) || null, [steps, activeStepId]);
  const recentTimeline = useMemo(() => [...timeline].slice(-8).reverse(), [timeline]);

  const toolCalls = useMemo(() => {
    if (!activeStep || activeStep.format_type !== 'tool.calls') return null;
    if (typeof activeStep.content !== 'string') return null;
    try {
      const parsed = JSON.parse(activeStep.content);
      return Array.isArray(parsed) ? parsed : null;
    } catch {
      return null;
    }
  }, [activeStep]);

  const pluginResultText = useMemo(() => {
    if (!activeStep || activeStep.format_type !== 'plugin_result') return null;
    if (typeof activeStep.content !== 'string') return String(activeStep.content ?? '');
    try {
      const parsed = JSON.parse(activeStep.content);
      return JSON.stringify(parsed, null, 2);
    } catch {
      return activeStep.content;
    }
  }, [activeStep]);

  return (
    <div className="flex-1 overflow-hidden flex flex-col">
      <div className="px-4 sm:px-6 lg:px-8 pt-6 sm:pt-8">
        <button
          onClick={onBack}
          className="flex items-center gap-2 text-xs font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500 hover:text-ink-900 dark:hover:text-zinc-100 transition-colors"
        >
          <ArrowLeft className="w-3.5 h-3.5" /> 返回会话列表
        </button>
        <div className="mt-4 flex flex-col sm:flex-row sm:items-center gap-3">
          <div className="p-3 bg-ink-900 dark:bg-zinc-100 rounded-2xl">
            <Terminal className="w-5 h-5 text-white dark:text-zinc-900" />
          </div>
          <div>
            <h2 className="text-2xl font-black text-ink-900 dark:text-zinc-100 font-serif">会话详情</h2>
            <p className="text-[10px] text-ink-400 dark:text-zinc-500 font-black uppercase tracking-widest">
              Session ID · {resolvedSessionId || '--'}
            </p>
          </div>
          <span className={`sm:ml-auto inline-flex items-center gap-2 px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-widest ${stateMeta.className}`}>
            <Activity className="w-3 h-3" /> {stateMeta.label}
          </span>
        </div>

        {(progress !== null || doneInfo) && (
          <div className="mt-3 flex flex-col gap-2">
            <div className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500">
              <span>进度</span>
              <span className="tabular-nums">{progress ?? 0}%</span>
              {doneInfo && (
                <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-200">
                  已完成 · {doneInfo.mode}
                </span>
              )}
            </div>
            <div className="h-2 w-full bg-paper-100 dark:bg-zinc-800 rounded-full overflow-hidden border border-paper-200 dark:border-zinc-700">
              <div
                className="h-full bg-brand-600 transition-all duration-300"
                style={{ width: `${progress ?? 0}%` }}
              />
            </div>
          </div>
        )}
      </div>

        <div className="px-4 sm:px-6 lg:px-8 py-6 flex-1 overflow-hidden grid grid-cols-1 xl:grid-cols-[320px_1fr] gap-6">
          <div className="bg-white dark:bg-zinc-900 rounded-[2rem] border border-paper-200 dark:border-zinc-800 shadow-sm overflow-hidden flex flex-col">
            <div className="px-6 py-4 border-b border-paper-100 dark:border-zinc-800 flex items-center gap-2 text-xs font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500">
              <ListChecks className="w-3.5 h-3.5" /> 步骤列表
            </div>
            <div className="flex-1 p-3 overflow-y-auto custom-scrollbar">
              {isLoading ? (
              <div className="h-full flex items-center justify-center text-sm text-ink-400 dark:text-zinc-500 gap-2">
                <Loader2 className="w-4 h-4 animate-spin" /> 加载步骤中
              </div>
              ) : error ? (
                <div className="px-3 py-6 text-sm text-rose-500">{error}</div>
              ) : steps.length === 0 ? (
                <div className="px-3 py-6 text-sm text-ink-400 dark:text-zinc-500">暂无步骤</div>
              ) : (
                <div className="space-y-4">
                  <div className="rounded-2xl border border-paper-100 dark:border-zinc-800 bg-paper-50 dark:bg-zinc-950/40 px-4 py-3">
                    <div className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500">
                      <Clock className="w-3 h-3" /> 最近时间线
                    </div>
                    {recentTimeline.length === 0 ? (
                      <div className="mt-2 text-xs text-ink-400 dark:text-zinc-600">暂无事件</div>
                    ) : (
                      <div className="mt-3 space-y-2">
                        {recentTimeline.map((item) => (
                          <div key={item.id} className="flex items-start gap-2 text-[11px]">
                            <span className="mt-0.5 h-1.5 w-1.5 rounded-full bg-ink-300 dark:bg-zinc-600" />
                            <div className="flex-1 min-w-0">
                              <div className={`font-semibold ${item.level === 'error' ? 'text-rose-500' : 'text-ink-700 dark:text-zinc-200'}`}>
                                {item.title}
                              </div>
                              <div className="text-[10px] text-ink-400 dark:text-zinc-600">{item.time}</div>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                  {steps.map((step) => (
                    <button
                      key={step.id}
                      onClick={() => setActiveStepId(step.id)}
                      className={`w-full text-left px-4 py-3 rounded-2xl border transition-all ${
                      step.id === activeStepId
                        ? 'border-brand-400 bg-brand-50 text-ink-900 dark:bg-brand-900/20 dark:text-zinc-100'
                        : 'border-paper-100 dark:border-zinc-800 text-ink-400 dark:text-zinc-500 hover:bg-paper-50 dark:hover:bg-zinc-800/40'
                    }`}
                  >
                    <div className="text-xs font-bold">{step.title || '步骤'}</div>
                    <div className="text-[10px] mt-1 opacity-70">#{step.order_index || step.id}</div>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>

          <div className="bg-white dark:bg-zinc-900 rounded-[2rem] border border-paper-200 dark:border-zinc-800 shadow-sm overflow-hidden flex flex-col">
            <div className="px-6 py-4 border-b border-paper-100 dark:border-zinc-800 flex items-center gap-2 text-xs font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500">
              <Terminal className="w-3.5 h-3.5" /> 输出面板
            </div>
            <div className="flex-1 p-6 text-sm text-ink-400 dark:text-zinc-500 overflow-y-auto custom-scrollbar">
              {!activeStep ? (
                <div className="flex items-center gap-2">
                  <AlertCircle className="w-4 h-4" />
                  暂无输出内容
                </div>
              ) : (
                <div className="space-y-4">
                  <div className="flex flex-wrap items-center gap-2 text-xs font-black uppercase tracking-widest text-ink-300 dark:text-zinc-600">
                    <span>{activeStep.title}</span>
                    {activeStep.metadata?.job_uuid && (
                      <span className="inline-flex items-center gap-1 rounded-full bg-amber-50 text-amber-700 dark:bg-amber-900/30 dark:text-amber-200 px-2 py-0.5 text-[10px] font-bold uppercase">
                        <Wrench className="w-3 h-3" /> Job: {activeStep.metadata.job_uuid}
                      </span>
                    )}
                    {typeof activeStep.metadata?.plugin_id === 'number' && (
                      <span className="inline-flex items-center gap-1 rounded-full bg-brand-50 text-brand-700 dark:bg-brand-900/30 dark:text-brand-200 px-2 py-0.5 text-[10px] font-bold uppercase">
                        Plugin: {activeStep.metadata.plugin_id}
                      </span>
                    )}
                  </div>
                  {toolCalls ? (
                    <div className="space-y-3">
                      {toolCalls.length === 0 && <div className="text-ink-400 dark:text-zinc-500">暂无工具调用</div>}
                      {toolCalls.map((call, index) => {
                        const name = call?.name || call?.Name || `Tool #${index + 1}`;
                        const args = call?.arguments || call?.Arguments || call?.args || {};
                        return (
                          <div key={`${name}-${index}`} className="rounded-2xl border border-paper-100 dark:border-zinc-800 bg-paper-50 dark:bg-zinc-950/40 p-4">
                            <div className="text-[11px] font-black uppercase tracking-widest text-ink-600 dark:text-zinc-300">
                              {name}
                            </div>
                            <pre className="mt-2 whitespace-pre-wrap font-mono text-xs text-ink-800 dark:text-zinc-200 leading-relaxed">
                              {JSON.stringify(args, null, 2)}
                            </pre>
                          </div>
                        );
                      })}
                    </div>
                  ) : activeStep.format_type === 'plugin_result' ? (
                    <pre className="whitespace-pre-wrap font-mono text-xs text-ink-800 dark:text-zinc-200 leading-relaxed">
                      {pluginResultText || '内容为空'}
                    </pre>
                  ) : (
                    <pre className="whitespace-pre-wrap font-sans text-sm text-ink-800 dark:text-zinc-200 leading-relaxed">
                      {typeof activeStep.content === 'string' ? activeStep.content : '内容为空'}
                    </pre>
                  )}
                </div>
              )}
            </div>
          </div>
      </div>
    </div>
  );
};
