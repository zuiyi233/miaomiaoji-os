import React from 'react';
import { Database, Upload, Search, BookOpen } from 'lucide-react';

export const CorpusCenter: React.FC = () => {
  return (
    <div className="flex-1 h-full bg-paper-50 dark:bg-zinc-950 transition-colors duration-300 overflow-hidden flex flex-col">
      <div className="h-20 border-b border-paper-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 px-8 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-4">
          <div className="p-3 bg-brand-600 dark:bg-brand-500 rounded-2xl text-white shadow-lg">
            <Database className="w-6 h-6" />
          </div>
          <div>
            <h1 className="text-xl font-black text-ink-900 dark:text-zinc-100 font-serif">语料库</h1>
            <span className="text-[10px] font-black uppercase tracking-widest text-ink-300 dark:text-zinc-600">Corpus · Import / Search</span>
          </div>
        </div>
        <button className="flex items-center gap-2 px-4 py-2 bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-xl text-xs font-bold uppercase tracking-widest shadow-md hover:opacity-90 transition-all">
          <Upload className="w-4 h-4" /> 导入语料
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-10 custom-scrollbar">
        <div className="max-w-5xl mx-auto space-y-8">
          <div className="bg-white dark:bg-zinc-900 rounded-[2.5rem] p-8 border border-paper-200 dark:border-zinc-800 shadow-sm">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-sm font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500">语料检索</h3>
              <span className="text-[10px] text-ink-300 dark:text-zinc-600">骨架功能</span>
            </div>

            <div className="flex gap-4">
              <div className="flex-1 relative">
                <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-ink-300 dark:text-zinc-600" />
                <input
                  placeholder="关键词 / 题材 / 标签"
                  className="w-full pl-12 pr-4 py-3 bg-paper-50 dark:bg-zinc-950 rounded-2xl border-none focus:ring-2 focus:ring-brand-200 dark:focus:ring-brand-500 text-sm font-bold text-ink-900 dark:text-zinc-100 outline-none transition-all"
                />
              </div>
              <button className="px-6 py-3 bg-brand-600 text-white rounded-2xl text-xs font-black uppercase tracking-widest shadow-md">搜索</button>
            </div>
          </div>

          <div className="bg-white dark:bg-zinc-900 rounded-[2.5rem] p-8 border border-paper-200 dark:border-zinc-800 shadow-sm">
            <div className="flex items-center gap-3 mb-6">
              <BookOpen className="w-5 h-5 text-brand-500" />
              <h3 className="text-sm font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500">最新语料</h3>
            </div>

            <div className="grid md:grid-cols-2 gap-4">
              {[1, 2, 3, 4].map((id) => (
                <div key={id} className="p-5 rounded-2xl border border-paper-100 dark:border-zinc-800 bg-paper-50 dark:bg-zinc-950">
                  <div className="text-sm font-bold text-ink-900 dark:text-zinc-100">示例语料条目 {id}</div>
                  <div className="text-[10px] text-ink-400 dark:text-zinc-500 mt-1">类型：玄幻 · 2026-02-03</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
