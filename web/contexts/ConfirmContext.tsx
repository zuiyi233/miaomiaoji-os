import React, { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { AlertTriangle } from 'lucide-react';

type ConfirmOptions = {
  title?: string;
  description?: string;
  confirmText?: string;
  cancelText?: string;
  tone?: 'default' | 'danger';
};

type ConfirmState = ConfirmOptions & {
  isOpen: boolean;
  resolve?: (value: boolean) => void;
};

type ConfirmContextValue = {
  confirm: (options: ConfirmOptions) => Promise<boolean>;
};

const ConfirmContext = createContext<ConfirmContextValue | null>(null);

export const useConfirm = () => {
  const ctx = useContext(ConfirmContext);
  if (!ctx) throw new Error('useConfirm 必须在 ConfirmProvider 内使用');
  return ctx;
};

export const ConfirmButton: React.FC<
  React.ButtonHTMLAttributes<HTMLButtonElement> & ConfirmOptions & { onConfirm: () => void }
> = ({ onConfirm, title, description, confirmText, cancelText, tone, onClick, ...rest }) => {
  const { confirm } = useConfirm();
  return (
    <button
      {...rest}
      onClick={async (e) => {
        onClick?.(e);
        const ok = await confirm({ title, description, confirmText, cancelText, tone });
        if (ok) onConfirm();
      }}
    />
  );
};

export const ConfirmProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [state, setState] = useState<ConfirmState>({ isOpen: false });
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (state.isOpen) {
      wrapperRef.current?.focus();
    }
  }, [state.isOpen]);

  const close = useCallback((result: boolean) => {
    setState((prev) => {
      prev.resolve?.(result);
      return { isOpen: false };
    });
  }, []);

  const confirm = useCallback((options: ConfirmOptions) => {
    return new Promise<boolean>((resolve) => {
      setState({
        isOpen: true,
        resolve,
        title: options.title || '请确认操作',
        description: options.description || '该操作无法撤销，请确认是否继续。',
        confirmText: options.confirmText || '确认',
        cancelText: options.cancelText || '取消',
        tone: options.tone || 'default',
      });
    });
  }, []);

  const value = useMemo(() => ({ confirm }), [confirm]);

  const danger = state.tone === 'danger';

  return (
    <ConfirmContext.Provider value={value}>
      {children}
      {state.isOpen && (
        <div
          className="fixed inset-0 z-[220] flex items-center justify-center px-6"
          onKeyDown={(e) => {
            if (e.key === 'Escape') close(false);
          }}
          tabIndex={-1}
          ref={wrapperRef}
        >
          <div className="absolute inset-0 bg-ink-900/40 dark:bg-black/60 backdrop-blur-sm" onClick={() => close(false)} />
          <div className="relative w-full max-w-md bg-white/90 dark:bg-zinc-900/90 backdrop-blur-2xl border border-paper-200 dark:border-white/10 rounded-[2.5rem] shadow-[0_25px_80px_rgba(0,0,0,0.18)] overflow-hidden">
            <div className="p-7">
              <div className="flex items-start gap-4">
                <div className={`w-12 h-12 rounded-2xl flex items-center justify-center ${danger ? 'bg-rose-100 text-rose-600 dark:bg-rose-900/30 dark:text-rose-300' : 'bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-300'}`}>
                  <AlertTriangle className="w-5 h-5" />
                </div>
                <div className="flex-1">
                  <h3 className="text-lg font-black text-ink-900 dark:text-zinc-100 tracking-tight">{state.title}</h3>
                  <p className="text-xs text-ink-500 dark:text-zinc-400 mt-2 leading-relaxed">{state.description}</p>
                </div>
                <button
                  onClick={() => close(false)}
                  className="p-2 rounded-xl text-ink-300 hover:text-ink-600 dark:text-zinc-600 dark:hover:text-zinc-300 transition-colors"
                  aria-label="取消"
                >
                  <span className="text-xs font-black">×</span>
                </button>
              </div>
            </div>
            <div className="px-7 pb-7 flex items-center gap-3">
              <button
                onClick={() => close(false)}
                className="flex-1 py-3 rounded-2xl text-[11px] font-black uppercase tracking-widest bg-paper-100 dark:bg-zinc-800 text-ink-600 dark:text-zinc-300 hover:bg-paper-200 dark:hover:bg-zinc-700 transition-all border border-paper-200 dark:border-zinc-700"
              >
                {state.cancelText}
              </button>
              <button
                onClick={() => close(true)}
                className={`flex-1 py-3 rounded-2xl text-[11px] font-black uppercase tracking-widest text-white transition-all ${danger ? 'bg-rose-600 hover:bg-rose-500 shadow-[0_12px_30px_rgba(244,63,94,0.35)]' : 'bg-ink-900 hover:bg-black shadow-[0_12px_30px_rgba(15,23,42,0.35)]'}`}
              >
                {state.confirmText}
              </button>
            </div>
          </div>
        </div>
      )}
    </ConfirmContext.Provider>
  );
};
