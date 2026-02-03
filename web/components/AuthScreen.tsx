
import React, { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { Sparkles, User, Lock, ArrowRight, Loader2, BookOpen, Key } from 'lucide-react';

export const AuthScreen: React.FC = () => {
  const { login, signup } = useAuth();
  const [isLogin, setIsLogin] = useState(true);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [inviteCode, setInviteCode] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username || !password) return;
    if (!isLogin && !inviteCode) {
      setError('请输入兑换码');
      return;
    }

    setIsSubmitting(true);
    setError('');
    
    try {
      if (isLogin) {
        const success = await login(username, password);
        if (!success) setError('用户名或密码错误 (默认管理员: admin / admin)');
      } else {
        const result = await signup(username, password, inviteCode);
        if (!result.success) {
          setError(result.message || '注册失败');
        }
      }
    } catch (err) {
      setError('系统异常，请重试');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-paper-50 dark:bg-zinc-950 flex items-center justify-center p-6 relative overflow-hidden transition-colors duration-300">
      {/* Background Ambience */}
      <div className="absolute top-0 left-0 w-full h-full opacity-10 dark:opacity-20 pointer-events-none">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-brand-400 dark:bg-brand-600 rounded-full blur-[128px]" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-indigo-400 dark:bg-indigo-600 rounded-full blur-[128px]" />
      </div>

      <div className="relative z-10 w-full max-w-md animate-in fade-in zoom-in-95 duration-700">
        <div className="text-center mb-10 space-y-4">
          <div className="inline-flex items-center justify-center p-4 bg-white dark:bg-zinc-900 rounded-3xl shadow-xl ring-1 ring-gray-200 dark:ring-zinc-800 mb-4">
            <BookOpen className="w-10 h-10 text-ink-900 dark:text-zinc-100" />
          </div>
          <h1 className="text-4xl font-black text-ink-900 dark:text-zinc-100 font-serif tracking-tight">Novel Agent OS</h1>
          <p className="text-ink-500 dark:text-zinc-400 font-medium">进入你的文学宇宙</p>
        </div>

        <form onSubmit={handleSubmit} className="bg-white dark:bg-zinc-900 p-8 rounded-[2.5rem] shadow-2xl border border-gray-100 dark:border-zinc-800 space-y-6">
          <div className="space-y-4">
            <div className="space-y-1">
              <label className="text-[10px] font-black text-gray-400 dark:text-zinc-600 uppercase tracking-widest pl-2">用户名</label>
              <div className="relative">
                <User className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-300" />
                <input 
                  value={username}
                  onChange={e => setUsername(e.target.value)}
                  className="w-full pl-12 pr-4 py-4 bg-gray-50 dark:bg-zinc-950 rounded-2xl border-none outline-none focus:ring-2 focus:ring-brand-200 dark:focus:ring-brand-900/30 text-ink-900 dark:text-zinc-100 font-bold transition-all"
                  placeholder="请输入用户名"
                />
              </div>
            </div>
            <div className="space-y-1">
              <label className="text-[10px] font-black text-gray-400 dark:text-zinc-600 uppercase tracking-widest pl-2">密码</label>
              <div className="relative">
                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-300" />
                <input 
                  type="password"
                  value={password}
                  onChange={e => setPassword(e.target.value)}
                  className="w-full pl-12 pr-4 py-4 bg-gray-50 dark:bg-zinc-950 rounded-2xl border-none outline-none focus:ring-2 focus:ring-brand-200 dark:focus:ring-brand-900/30 text-ink-900 dark:text-zinc-100 font-bold transition-all"
                  placeholder="请输入密码"
                />
              </div>
            </div>

            {!isLogin && (
              <div className="space-y-1 animate-in slide-in-from-top-2">
                <label className="text-[10px] font-black text-brand-600 dark:text-brand-400 uppercase tracking-widest pl-2">邀请兑换码 (Required)</label>
                <div className="relative">
                  <Key className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-brand-500" />
                  <input 
                    value={inviteCode}
                    onChange={e => setInviteCode(e.target.value.toUpperCase())}
                    className="w-full pl-12 pr-4 py-4 bg-brand-50/50 dark:bg-zinc-950 rounded-2xl border border-brand-100 dark:border-zinc-800 outline-none focus:ring-2 focus:ring-brand-200 dark:focus:ring-brand-900/30 text-ink-900 dark:text-zinc-100 font-bold transition-all placeholder:text-brand-300"
                    placeholder="XXXX-XXXX"
                  />
                </div>
              </div>
            )}
          </div>

          {error && (
            <div className="p-3 bg-rose-50 dark:bg-rose-950/20 border border-rose-100 dark:border-rose-900/30 rounded-xl text-xs font-bold text-rose-600 dark:text-rose-400 text-center animate-in fade-in zoom-in-95">
              {error}
            </div>
          )}

          <button 
            type="submit"
            disabled={isSubmitting}
            className="w-full py-5 bg-ink-900 dark:bg-white text-white dark:text-zinc-900 rounded-2xl font-black text-sm uppercase tracking-widest flex items-center justify-center gap-3 shadow-xl hover:bg-black dark:hover:bg-zinc-100 transition-all active:scale-[0.98] disabled:opacity-50"
          >
            {isSubmitting ? <Loader2 className="w-5 h-5 animate-spin" /> : (
              <>
                {isLogin ? '登录工作台' : '验证并注册'} <ArrowRight className="w-4 h-4" />
              </>
            )}
          </button>

          <div className="text-center pt-4 border-t border-gray-50 dark:border-zinc-800">
             <button 
               type="button"
               onClick={() => { setIsLogin(!isLogin); setError(''); }}
               className="text-xs font-bold text-gray-400 hover:text-ink-900 dark:hover:text-zinc-100 transition-colors"
             >
               {isLogin ? '使用兑换码注册' : '返回账号登录'}
             </button>
          </div>
        </form>
        
        <div className="mt-8 text-center text-[10px] font-black text-gray-300 dark:text-zinc-700 uppercase tracking-widest flex items-center justify-center gap-2">
           <Sparkles className="w-3 h-3" /> Powered by Gemini Agentic Engine
        </div>
      </div>
    </div>
  );
};
