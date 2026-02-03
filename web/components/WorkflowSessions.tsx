import React, { useMemo } from 'react';
import { Activity, CheckCircle2, Clock, Layers, XCircle } from 'lucide-react';

type SessionStatus = 'running' | 'success' | 'failed' | 'canceled';

interface SessionItem {
  id: string;
  title: string;
  mode: string;
  status: SessionStatus;
  updatedAt: string;
}

interface WorkflowSessionsProps {
  sessions: SessionItem[];
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

export const WorkflowSessions: React.FC<WorkflowSessionsProps> = ({ sessions, onSelect }) => {
  const items = useMemo(() => sessions, [sessions]);

  return (
    <div className="flex-1 overflow-hidden flex flex-col">
      <div className="px-8 pt-8">
        <div className="flex items-center gap-3">
          <div className="p-3 bg-brand-600 dark:bg-brand-500 rounded-2xl shadow-lg shadow-brand-100 dark:shadow-black/20">
            <Layers className="w-5 h-5 text-white" />
          </div>
          <div>
            <h2 className="text-2xl font-black text-ink-900 dark:text-zinc-100 font-serif">工作流会话</h2>
            <p className="text-[10px] text-ink-400 dark:text-zinc-500 font-black uppercase tracking-widest">Sessions · 运行与回放</p>
          </div>
        </div>
      </div>

      <div className="px-8 py-6 flex-1 overflow-y-auto custom-scrollbar">
        <div className="bg-white dark:bg-zinc-900 rounded-[2rem] border border-paper-200 dark:border-zinc-800 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-paper-100 dark:border-zinc-800 flex items-center justify-between">
            <div className="flex items-center gap-2 text-xs font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500">
              <Activity className="w-3.5 h-3.5" /> 最近会话
            </div>
            <div className="text-[10px] text-ink-300 dark:text-zinc-600">仅展示骨架</div>
          </div>

          <div className="divide-y divide-paper-100 dark:divide-zinc-800">
            {items.length === 0 ? (
              <div className="px-6 py-12 text-center text-sm text-ink-400 dark:text-zinc-500">暂无会话记录</div>
            ) : (
              items.map((session) => {
                const meta = STATUS_META[session.status];
                return (
                  <button
                    key={session.id}
                    onClick={() => onSelect(session.id)}
                    className="w-full flex items-center justify-between px-6 py-4 hover:bg-paper-50 dark:hover:bg-zinc-800/50 transition-colors text-left"
                  >
                    <div className="space-y-1">
                      <div className="flex items-center gap-3">
                        <h3 className="text-base font-bold text-ink-900 dark:text-zinc-100">{session.title}</h3>
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
