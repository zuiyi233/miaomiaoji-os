import React from 'react';
import { ArrowLeft, Activity, ListChecks, Terminal, AlertCircle } from 'lucide-react';
import { ConnectionState } from '../services/sseClient';

interface WorkflowDetailProps {
  sessionId: string;
  onBack: () => void;
  connectionState: ConnectionState;
}

const STATE_META: Record<ConnectionState, { label: string; className: string }> = {
  connecting: { label: '连接中', className: 'bg-brand-50 text-brand-700 dark:bg-brand-900/30 dark:text-brand-200' },
  connected: { label: '已连接', className: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-200' },
  disconnected: { label: '已断开', className: 'bg-rose-50 text-rose-700 dark:bg-rose-900/30 dark:text-rose-200' },
  reconnecting: { label: '重连中', className: 'bg-amber-50 text-amber-700 dark:bg-amber-900/30 dark:text-amber-200' },
};

export const WorkflowDetail: React.FC<WorkflowDetailProps> = ({ sessionId, onBack, connectionState }) => {
  const stateMeta = STATE_META[connectionState];

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
            <p className="text-[10px] text-ink-400 dark:text-zinc-500 font-black uppercase tracking-widest">Session ID · {sessionId}</p>
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
          <div className="flex-1 p-6 text-sm text-ink-400 dark:text-zinc-500">
            这里显示 Steps 树结构（骨架）
          </div>
        </div>

        <div className="bg-white dark:bg-zinc-900 rounded-[2rem] border border-paper-200 dark:border-zinc-800 shadow-sm overflow-hidden flex flex-col">
          <div className="px-6 py-4 border-b border-paper-100 dark:border-zinc-800 flex items-center gap-2 text-xs font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500">
            <Terminal className="w-3.5 h-3.5" /> 输出面板
          </div>
          <div className="flex-1 p-6 text-sm text-ink-400 dark:text-zinc-500">
            <div className="flex items-center gap-2">
              <AlertCircle className="w-4 h-4" />
              流式输出区域（骨架）
            </div>
            <div className="mt-4 text-xs text-ink-300 dark:text-zinc-600">后续接入 SSE 增量输出</div>
          </div>
        </div>
      </div>
    </div>
  );
};
