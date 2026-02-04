import React, { useEffect, useMemo, useState } from 'react';
import { ArrowLeft, Activity, ListChecks, Terminal, AlertCircle, Loader2 } from 'lucide-react';
import { ConnectionState, createSSEClient, StepAppendedData } from '../services/sseClient';
import { getSessionApi, listStepsApi, SessionDTO, SessionStepDTO } from '../services/workflowApi';
import { useParams } from 'react-router-dom';

interface WorkflowDetailProps {
  sessionId?: string;
  onBack: () => void;
}

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

  const sessionNumericId = Number(resolvedSessionId);
  const stateMeta = STATE_META[connectionState];

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
      onStateChange: (state) => setConnectionState(state),
      onStepAppended: (data: StepAppendedData) => {
        setSteps((prev) => {
          const existing = prev.find((item) => item.id === data.step_id);
          if (existing) {
            return prev.map((item) =>
              item.id === data.step_id
                ? { ...item, content: item.content + data.content }
                : item
            );
          }
          const next: SessionStepDTO = {
            id: data.step_id,
            title: data.title || 'AI 输出',
            content: data.content,
            format_type: session?.mode || 'workflow',
            order_index: prev.length + 1,
            session_id: sessionNumericId,
            metadata: null,
            created_at: data.timestamp,
            updated_at: data.timestamp,
          };
          return [...prev, next];
        });
        setActiveStepId((prev) => prev ?? data.step_id);
      },
    });
    client.connect();
    return () => client.disconnect();
  }, [sessionNumericId, session?.mode]);

  const activeStep = useMemo(() => steps.find((item) => item.id === activeStepId) || null, [steps, activeStepId]);

  return (
    <div className="flex-1 overflow-hidden flex flex-col">
      <div className="px-8 pt-8">
        <button
          onClick={onBack}
          className="flex items-center gap-2 text-xs font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500 hover:text-ink-900 dark:hover:text-zinc-100 transition-colors"
        >
          <ArrowLeft className="w-3.5 h-3.5" /> 返回会话列表
        </button>
        <div className="mt-4 flex items-center gap-3">
          <div className="p-3 bg-ink-900 dark:bg-zinc-100 rounded-2xl">
            <Terminal className="w-5 h-5 text-white dark:text-zinc-900" />
          </div>
          <div>
            <h2 className="text-2xl font-black text-ink-900 dark:text-zinc-100 font-serif">会话详情</h2>
            <p className="text-[10px] text-ink-400 dark:text-zinc-500 font-black uppercase tracking-widest">
              Session ID · {resolvedSessionId || '--'}
            </p>
          </div>
          <span className={`ml-auto inline-flex items-center gap-2 px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-widest ${stateMeta.className}`}>
            <Activity className="w-3 h-3" /> {stateMeta.label}
          </span>
        </div>
      </div>

      <div className="px-8 py-6 flex-1 overflow-hidden grid grid-cols-1 xl:grid-cols-[320px_1fr] gap-6">
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
              <div className="space-y-2">
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
                <div className="text-xs font-black uppercase tracking-widest text-ink-300 dark:text-zinc-600">
                  {activeStep.title}
                </div>
                <pre className="whitespace-pre-wrap font-sans text-sm text-ink-800 dark:text-zinc-200 leading-relaxed">
                  {activeStep.content || '内容为空'}
                </pre>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
