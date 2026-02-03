import React from 'react';
import { FileText, UploadCloud, Download, Eye, File } from 'lucide-react';

export const FileCenter: React.FC = () => {
  return (
    <div className="flex-1 h-full bg-paper-50 dark:bg-zinc-950 transition-colors duration-300 overflow-hidden flex flex-col">
      <div className="h-20 border-b border-paper-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 px-8 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-4">
          <div className="p-3 bg-brand-600 dark:bg-brand-500 rounded-2xl text-white shadow-lg">
            <FileText className="w-6 h-6" />
          </div>
          <div>
            <h1 className="text-xl font-black text-ink-900 dark:text-zinc-100 font-serif">文件中心</h1>
            <span className="text-[10px] font-black uppercase tracking-widest text-ink-300 dark:text-zinc-600">Files · Upload / Download</span>
          </div>
        </div>
        <button className="flex items-center gap-2 px-4 py-2 bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-xl text-xs font-bold uppercase tracking-widest shadow-md hover:opacity-90 transition-all">
          <UploadCloud className="w-4 h-4" /> 上传文件
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-10 custom-scrollbar">
        <div className="max-w-5xl mx-auto space-y-8">
          <div className="bg-white dark:bg-zinc-900 rounded-[2.5rem] p-8 border border-paper-200 dark:border-zinc-800 shadow-sm">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-sm font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500">最近文件</h3>
              <span className="text-[10px] text-ink-300 dark:text-zinc-600">骨架列表</span>
            </div>

            <div className="space-y-3">
              {[1, 2, 3].map((id) => (
                <div key={id} className="flex items-center gap-4 p-4 rounded-2xl border border-paper-100 dark:border-zinc-800 bg-paper-50 dark:bg-zinc-950">
                  <div className="w-10 h-10 rounded-xl bg-brand-100 dark:bg-brand-900/30 flex items-center justify-center">
                    <File className="w-5 h-5 text-brand-600 dark:text-brand-300" />
                  </div>
                  <div className="flex-1">
                    <div className="text-sm font-bold text-ink-900 dark:text-zinc-100">示例文件_{id}.txt</div>
                    <div className="text-[10px] text-ink-400 dark:text-zinc-500">2.1MB · 2026-02-03</div>
                  </div>
                  <div className="flex items-center gap-2">
                    <button className="p-2 rounded-xl bg-white dark:bg-zinc-800 border border-paper-200 dark:border-zinc-700 text-ink-400 hover:text-ink-900 dark:hover:text-zinc-100">
                      <Eye className="w-4 h-4" />
                    </button>
                    <button className="p-2 rounded-xl bg-white dark:bg-zinc-800 border border-paper-200 dark:border-zinc-700 text-ink-400 hover:text-ink-900 dark:hover:text-zinc-100">
                      <Download className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
