import React from 'react';
import { Coins, Receipt, ArrowUpRight, ArrowDownLeft } from 'lucide-react';

export const SettlementCenter: React.FC = () => {
  return (
    <div className="flex-1 h-full bg-paper-50 dark:bg-zinc-950 transition-colors duration-300 overflow-hidden flex flex-col">
      <div className="h-20 border-b border-paper-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 px-8 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-4">
          <div className="p-3 bg-brand-600 dark:bg-brand-500 rounded-2xl text-white shadow-lg">
            <Coins className="w-6 h-6" />
          </div>
          <div>
            <h1 className="text-xl font-black text-ink-900 dark:text-zinc-100 font-serif">积分结算</h1>
            <span className="text-[10px] font-black uppercase tracking-widest text-ink-300 dark:text-zinc-600">Settlements · Ledger</span>
          </div>
        </div>
        <div className="flex items-center gap-2 text-xs font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500">
          <Receipt className="w-4 h-4" /> 当前余额：--
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-10 custom-scrollbar">
        <div className="max-w-5xl mx-auto space-y-8">
          <div className="bg-white dark:bg-zinc-900 rounded-[2.5rem] p-8 border border-paper-200 dark:border-zinc-800 shadow-sm">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-sm font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500">最近流水</h3>
              <span className="text-[10px] text-ink-300 dark:text-zinc-600">骨架展示</span>
            </div>

            <div className="space-y-3">
              {[1, 2, 3].map((id) => (
                <div key={id} className="flex items-center gap-4 p-4 rounded-2xl border border-paper-100 dark:border-zinc-800 bg-paper-50 dark:bg-zinc-950">
                  <div className="w-10 h-10 rounded-xl bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center">
                    {id % 2 === 0 ? (
                      <ArrowDownLeft className="w-5 h-5 text-emerald-600 dark:text-emerald-300" />
                    ) : (
                      <ArrowUpRight className="w-5 h-5 text-rose-600 dark:text-rose-300" />
                    )}
                  </div>
                  <div className="flex-1">
                    <div className="text-sm font-bold text-ink-900 dark:text-zinc-100">示例流水 {id}</div>
                    <div className="text-[10px] text-ink-400 dark:text-zinc-500">2026-02-03 · 描述占位</div>
                  </div>
                  <div className="text-sm font-black text-ink-700 dark:text-zinc-200">{id % 2 === 0 ? '+120' : '-60'}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
