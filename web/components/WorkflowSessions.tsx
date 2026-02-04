import React, { useEffect, useMemo, useState } from 'react';
import { Activity, CheckCircle2, Clock, Layers, Loader2, XCircle } from 'lucide-react';
import { listSessionsApi, listSessionsByProjectApi } from '../services/workflowApi';
import { useProject } from '../contexts/ProjectContext';

type SessionStatus = 'running' | 'success' | 'failed' | 'canceled';

interface SessionItem {
  id: string;
  title: string;
  mode: string;
  status: SessionStatus;
  updatedAt: string;
}

interface WorkflowSessionsProps {
  onSelect: (sessionId: string) => void;
}

const STATUS_META: Record<SessionStatus, { label: string; className: string; icon: React.ReactNode }> = {
  running: {
    label: '运行中',
    className: 'bg-brand-50 text-brand-700 dark:bg-brand-900/30 dark:text-brand-200',
    icon: <Activity className="w-3 h-3" />,
  },
  success: {
    label: '已完成',
    className: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-200',
    icon: <CheckCircle2 className="w-3 h-3" />,
  },
  failed: {
    label: '失败',
    className: 'bg-rose-50 text-rose-700 dark:bg-rose-900/30 dark:text-rose-200',
    icon: <XCircle className="w-3 h-3" />,
  },
  canceled: {
    label: '已取消',
    className: 'bg-ink-100 text-ink-600 dark:bg-zinc-800 dark:text-zinc-300',
    icon: <Clock className="w-3 h-3" />,
  },
};

export const WorkflowSessions: React.FC<WorkflowSessionsProps> = ({ onSelect }) => {
  const { activeProjectId } = useProject();
  const [sessions, setSessions] = useState<SessionItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchSessions = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const projectId = Number(activeProjectId);
        const result = Number.isFinite(projectId) && projectId > 0
          ? await listSessionsByProjectApi(projectId, 1, 50)
          : await listSessionsApi(1, 50);
        const items: SessionItem[] = (result.sessions || []).map((item) => {
          const updatedAt = item.updated_at || item.created_at;
          const updatedMs = updatedAt ? new Date(updatedAt).getTime() : 0;
          const isRecent = updatedMs > 0 && Date.now() - updatedMs < 5 * 60 * 1000;
          return {
            id: String(item.id),
            title: item.title || '未命名会话',
            mode: item.mode || 'Normal',
            status: isRecent ? 'running' : 'success',
            updatedAt: updatedAt || '--',
          };
        });
        setSessions(items);
      } catch (err: any) {
        setError(err?.message || '加载会话失败');
      } finally {
        setIsLoading(false);
      }
    };

    fetchSessions();
  }, [activeProjectId]);

  const items = useMemo(() => sessions, [sessions]);

  return (
    <div className="flex-1 overflow-hidden flex flex-col">
      <div className="px-4 sm:px-6 lg:px-8 pt-6 sm:pt-8">
        <div className="flex flex-col sm:flex-row sm:items-center gap-3">
          <div className="p-3 bg-brand-600 dark:bg-brand-500 rounded-2xl shadow-lg shadow-brand-100 dark:shadow-black/20">
            <Layers className="w-5 h-5 text-white" />
          </div>
          <div>
            <h2 className="text-2xl font-black text-ink-900 dark:text-zinc-100 font-serif">工作流会话</h2>
            <p className="text-[10px] text-ink-400 dark:text-zinc-500 font-black uppercase tracking-widest">Sessions · 运行与回放</p>
          </div>
        </div>
      </div>

      <div className="px-4 sm:px-6 lg:px-8 py-6 flex-1 overflow-y-auto custom-scrollbar">
        <div className="bg-white dark:bg-zinc-900 rounded-[2rem] border border-paper-200 dark:border-zinc-800 shadow-sm overflow-hidden">
          <div className="px-4 sm:px-6 py-4 border-b border-paper-100 dark:border-zinc-800 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
            <div className="flex items-center gap-2 text-xs font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500">
              <Activity className="w-3.5 h-3.5" /> 最近会话
            </div>
            <div className="text-[10px] text-ink-300 dark:text-zinc-600">
              {isLoading ? '加载中' : `${items.length} 条记录`}
            </div>
          </div>

          <div className="divide-y divide-paper-100 dark:divide-zinc-800">
            {isLoading ? (
              <div className="px-4 sm:px-6 py-12 text-center text-sm text-ink-400 dark:text-zinc-500 flex items-center justify-center gap-2">
                <Loader2 className="w-4 h-4 animate-spin" /> 加载会话中
              </div>
            ) : error ? (
              <div className="px-4 sm:px-6 py-12 text-center text-sm text-rose-500">{error}</div>
            ) : items.length === 0 ? (
              <div className="px-4 sm:px-6 py-12 text-center text-sm text-ink-400 dark:text-zinc-500">暂无会话记录</div>
            ) : (
              items.map((session) => {
                const meta = STATUS_META[session.status];
                return (
                  <button
                    key={session.id}
                    onClick={() => onSelect(session.id)}
                    className="w-full flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 px-4 sm:px-6 py-4 hover:bg-paper-50 dark:hover:bg-zinc-800/50 transition-colors text-left"
                  >
                    <div className="space-y-1">
                      <div className="flex flex-wrap items-center gap-3">
                        <h3 className="text-base font-bold text-ink-900 dark:text-zinc-100 break-all">{session.title}</h3>
                        <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-black uppercase tracking-widest ${meta.className}`}>
                          {meta.icon}
                          {meta.label}
                        </span>
                      </div>
                      <div className="text-xs text-ink-400 dark:text-zinc-500">模式：{session.mode} · 更新时间：{session.updatedAt}</div>
                    </div>
                    <div className="text-[10px] text-ink-300 dark:text-zinc-600">查看详情</div>
                  </button>
                );
              })
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
