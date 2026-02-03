
import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useProject } from '../contexts/ProjectContext';
import { useAuth } from '../contexts/AuthContext';
import { ConfirmButton, useConfirm } from '../contexts/ConfirmContext';
import { fetchInternetInspiration, DEFAULT_AI_SETTINGS } from '../services/aiService';
import { 
  Book, Plus, Trash2, ArrowRight, Sparkles, TrendingUp, 
  Calendar, Clock, Target, FileText, PenTool, 
  X, Save, Hourglass, Zap, ChevronRight, Bookmark,
  Lightbulb, CheckCircle2, Circle, Trophy, Download, 
  Award, Share2, MoreHorizontal, Layout, Users, 
  CheckCircle, ListChecks, Star, Quote, RefreshCw, ExternalLink,
  Loader2, Settings, Shield, User as UserIcon, Lock, Gift, Coins, Copy,
  Wallet, Ticket
} from 'lucide-react';
import { Project, ViewMode } from '../types';

interface ProjectDashboardProps {
  onCreateNew: () => void;
}

// --- Helper: Real-time Clock & Date ---
const useChineseTime = () => {
  const [timeStr, setTimeStr] = useState('');
  const [dateStr, setDateStr] = useState('');
  const [period, setPeriod] = useState('');

  useEffect(() => {
    const update = () => {
      const now = new Date();
      setTimeStr(`${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}`);
      
      let p = '';
      const h = now.getHours();
      if (h >= 23 || h < 1) p = '子时·半夜';
      else if (h >= 1 && h < 5) p = '寅丑·凌晨';
      else if (h >= 5 && h < 9) p = '卯辰·清晨';
      else if (h >= 9 && h < 12) p = '巳午·上午';
      else if (h >= 12 && h < 14) p = '未时·正午';
      else if (h >= 14 && h < 18) p = '申酉·下午';
      else if (h >= 18 && h < 21) p = '戌时·黄昏';
      else p = '亥时·人定';
      setPeriod(p);

      setDateStr(now.toLocaleDateString('zh-CN', { 
        year: 'numeric',
        month: 'long', 
        day: 'numeric', 
        weekday: 'long' 
      }));
    };
    
    update();
    const timer = setInterval(update, 1000 * 60); 
    return () => clearInterval(timer);
  }, []);

  return { timeStr, dateStr, period };
};

// --- Daily Progress Tracking ---
const useDailyProgress = (currentTotalWords: number) => {
  const [dailyTarget, setDailyTarget] = useState(() => Number(localStorage.getItem('nao_daily_target')) || 2000);
  const [todayWords, setTodayWords] = useState(0);
  const [isCheckedIn, setIsCheckedIn] = useState(() => localStorage.getItem('nao_last_checkin') === new Date().toISOString().split('T')[0]);
  const { awardPoints } = useAuth(); // Award points for goal completion

  useEffect(() => {
    localStorage.setItem('nao_daily_target', String(dailyTarget));
  }, [dailyTarget]);

  useEffect(() => {
    const todayStr = new Date().toISOString().split('T')[0];
    const storedData = localStorage.getItem('nao_daily_progress');
    if (storedData) {
      const { date, startCount } = JSON.parse(storedData);
      if (date === todayStr) {
        setTodayWords(Math.max(0, currentTotalWords - startCount));
      } else {
        localStorage.setItem('nao_daily_progress', JSON.stringify({ date: todayStr, startCount: currentTotalWords }));
        setTodayWords(0);
        setIsCheckedIn(false);
      }
    } else {
      localStorage.setItem('nao_daily_progress', JSON.stringify({ date: todayStr, startCount: currentTotalWords }));
    }
  }, [currentTotalWords]);

  const handleCheckIn = () => {
    const todayStr = new Date().toISOString().split('T')[0];
    localStorage.setItem('nao_last_checkin', todayStr);
    setIsCheckedIn(true);
    awardPoints(50, 'Daily Goal Completion'); // Award points!
    alert('恭喜达成今日目标！获得 50 积分奖励。');
  };

  return { dailyTarget, setDailyTarget, todayWords, isCheckedIn, handleCheckIn };
};

// --- Sub-Components ---

const PointsCard: React.FC = () => {
  const { user, performCheckIn, exchangePointsForCode, systemConfig, redemptionCodes } = useAuth();
  const { confirm } = useConfirm();
  const [exchangedCode, setExchangedCode] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [showShop, setShowShop] = useState(false);
  const [showInventory, setShowInventory] = useState(false);

  // Filter codes owned by the current user
  const myCodes = useMemo(() => {
    if (!user) return [];
    return redemptionCodes.filter(c => c.tags.includes(`owner:${user.id}`) || c.note?.includes(`[${user.username}]`));
  }, [redemptionCodes, user]);

  const hasCheckedInToday = useMemo(() => {
      if (!user || !user.lastCheckIn) return false;
      return new Date(user.lastCheckIn).toDateString() === new Date().toDateString();
  }, [user?.lastCheckIn]);

  const handleDoCheckIn = () => {
      setLoading(true);
      const result = performCheckIn();
      setLoading(false);
      if (result.success) {
          alert(`签到成功！获得 ${result.pointsAwarded} 积分。\n当前连签：${result.streak} 天`);
      } else {
          alert(result.message);
      }
  };

  const handleExchange = async (optionId: string, cost: number) => {
      if ((user?.points || 0) < cost) {
          alert("积分不足");
          return;
      }
       const ok = await confirm({
         title: `确定消耗 ${cost} 积分进行兑换吗？`,
         description: '兑换后将立即扣除积分。',
         confirmText: '确认兑换',
         cancelText: '取消',
         tone: 'default',
       });
       if (!ok) return;
      
      setLoading(true);
      
      // Simulate network delay for better UX
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const result = exchangePointsForCode(optionId);
      setLoading(false);
      
      if (result.success && result.code) {
          setExchangedCode(result.code);
          setShowShop(false);
      } else {
          alert(result.message || '兑换失败');
      }
  };

  return (
    <div className="bg-gradient-to-br from-indigo-600 to-purple-700 text-white rounded-[2.5rem] p-6 shadow-lg h-full flex flex-col justify-between relative overflow-hidden group">
       <div className="absolute top-0 right-0 p-3 opacity-20"><Trophy className="w-16 h-16 rotate-12" /></div>
       
       <div className="relative z-10">
          <div className="flex justify-between items-center mb-4">
             <h4 className="text-[10px] font-black uppercase tracking-widest flex items-center gap-2 text-white/80">
                <Coins className="w-4 h-4" /> 积分中心
             </h4>
             <span className="text-[10px] font-bold bg-white/20 px-2 py-1 rounded-full">{user?.points || 0} pts</span>
          </div>
          
          <div className="space-y-3">
             <div className="flex items-center justify-between bg-white/10 p-3 rounded-2xl border border-white/10">
                <div>
                   <p className="text-xs font-bold">每日签到</p>
                   <p className="text-[9px] opacity-70">连签 {user?.checkInStreak || 0} 天</p>
                </div>
                <button 
                  onClick={handleDoCheckIn}
                  disabled={hasCheckedInToday || loading}
                  className={`px-3 py-1.5 rounded-xl text-[10px] font-black uppercase tracking-wide transition-all ${hasCheckedInToday ? 'bg-white/20 text-white cursor-not-allowed' : 'bg-white text-indigo-700 hover:bg-indigo-50 shadow-md'}`}
                >
                  {hasCheckedInToday ? '已签到' : '签到 +积分'}
                </button>
             </div>

             <div className="flex gap-2">
                <button 
                  onClick={() => setShowShop(true)}
                  disabled={!systemConfig.enablePointsExchange || loading}
                  className="flex-1 flex items-center justify-center gap-2 px-3 py-3 bg-amber-400 text-amber-900 rounded-2xl text-[10px] font-black uppercase tracking-wide hover:bg-amber-300 transition-all shadow-md disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <Gift className="w-4 h-4" /> 兑换商品
                </button>
                <button 
                  onClick={() => setShowInventory(true)}
                  className="px-3 py-3 bg-white/10 text-white hover:bg-white/20 rounded-2xl text-[10px] font-black uppercase tracking-wide transition-all border border-white/10"
                  title="我的卡包"
                >
                   <Wallet className="w-4 h-4" />
                </button>
             </div>
          </div>
       </div>

       {/* Loading Overlay */}
       {loading && (
           <div className="absolute inset-0 bg-indigo-900/80 z-40 flex flex-col items-center justify-center animate-in fade-in">
               <Loader2 className="w-8 h-8 text-white animate-spin mb-2" />
               <p className="text-xs font-bold text-white">正在处理交易...</p>
           </div>
       )}

       {showShop && (
           <div className="absolute inset-0 bg-indigo-900/95 z-20 flex flex-col p-6 animate-in zoom-in-95 overflow-hidden">
               <div className="flex justify-between items-center mb-4">
                   <h3 className="text-sm font-black text-white flex items-center gap-2"><Gift className="w-4 h-4"/> 积分兑换</h3>
                   <button onClick={() => setShowShop(false)}><X className="w-5 h-5 text-white/50 hover:text-white"/></button>
               </div>
               <div className="flex-1 overflow-y-auto space-y-2 custom-scrollbar pr-1">
                   {systemConfig.exchangeOptions.map(opt => (
                       <div key={opt.id} className="bg-white/10 p-3 rounded-xl border border-white/10 flex justify-between items-center">
                           <div className="flex-1 min-w-0 mr-2">
                               <p className="text-xs font-bold truncate">{opt.name}</p>
                               <p className="text-[9px] opacity-60 truncate">{opt.description}</p>
                           </div>
                           <button 
                               onClick={() => handleExchange(opt.id, opt.cost)}
                               className="px-2 py-1 bg-amber-400 text-amber-900 text-[10px] font-black rounded-lg whitespace-nowrap active:scale-95 transition-transform"
                           >
                               {opt.cost} 积分
                           </button>
                       </div>
                   ))}
               </div>
           </div>
       )}

       {showInventory && (
           <div className="absolute inset-0 bg-indigo-900/95 z-30 flex flex-col p-6 animate-in zoom-in-95 overflow-hidden">
               <div className="flex justify-between items-center mb-4">
                   <h3 className="text-sm font-black text-white flex items-center gap-2"><Wallet className="w-4 h-4"/> 我的卡包</h3>
                   <button onClick={() => setShowInventory(false)}><X className="w-5 h-5 text-white/50 hover:text-white"/></button>
               </div>
               <div className="flex-1 overflow-y-auto space-y-2 custom-scrollbar pr-1">
                   {myCodes.length === 0 ? (
                       <p className="text-center text-white/50 text-xs mt-10">暂无兑换记录</p>
                   ) : (
                       myCodes.map(code => (
                           <div key={code.code} className="bg-white/10 p-3 rounded-xl border border-white/10 space-y-2">
                               <div className="flex justify-between items-center">
                                   <span className="text-[9px] bg-white/20 px-1.5 py-0.5 rounded text-white/80">{new Date(code.createdAt).toLocaleDateString()}</span>
                                   <span className={`text-[9px] px-1.5 py-0.5 rounded font-bold ${code.status === 'active' ? 'bg-emerald-500/20 text-emerald-300' : 'bg-red-500/20 text-red-300'}`}>
                                       {code.status === 'active' ? '未使用' : '已失效'}
                                   </span>
                               </div>
                               <div className="flex items-center gap-2 bg-black/20 p-2 rounded-lg">
                                   <Ticket className="w-3 h-3 text-amber-400" />
                                   <span className="font-mono text-xs font-bold flex-1 truncate text-amber-100">{code.code}</span>
                                   <button onClick={() => navigator.clipboard.writeText(code.code)}><Copy className="w-3 h-3 text-white/50 hover:text-white"/></button>
                               </div>
                               <p className="text-[9px] opacity-60 truncate">{code.tags.filter(t => !t.startsWith('owner:')).join(', ')}</p>
                           </div>
                       ))
                   )}
               </div>
           </div>
       )}

       {exchangedCode && (
           <div className="absolute inset-0 bg-indigo-900/95 z-30 flex flex-col items-center justify-center p-6 text-center animate-in zoom-in-95">
               <Gift className="w-10 h-10 text-amber-400 mb-2" />
               <h3 className="text-lg font-black text-white mb-2">兑换成功！</h3>
               <div className="bg-white/10 p-3 rounded-xl border border-white/20 flex items-center gap-2 mb-4 w-full justify-between">
                   <span className="font-mono font-bold text-sm tracking-widest text-amber-100">{exchangedCode}</span>
                   <button onClick={() => navigator.clipboard.writeText(exchangedCode)}><Copy className="w-4 h-4 text-white/70 hover:text-white" /></button>
               </div>
               <p className="text-[10px] text-white/60 mb-4">请保管好您的兑换码。您也可以在“我的卡包”中找到它。</p>
               <div className="flex gap-2 w-full">
                   <button onClick={() => { setExchangedCode(null); setShowInventory(true); }} className="flex-1 py-2 bg-white/10 rounded-xl text-xs font-bold hover:bg-white/20">查看卡包</button>
                   <button onClick={() => setExchangedCode(null)} className="flex-1 py-2 bg-white text-indigo-900 rounded-xl text-xs font-bold">关闭</button>
               </div>
           </div>
       )}
    </div>
  );
};

const InspirationCard: React.FC = () => {
  const INSPIRATION_CACHE_KEY = 'nao_daily_inspiration';
  const INSPIRATION_FETCH_KEY = 'nao_last_inspiration_fetch';
  const INSPIRATION_HOUR_KEY = 'nao_last_inspiration_hour';
  const INSPIRATION_DISABLE_KEY = 'nao_inspiration_disabled_until';
  const INSPIRATION_FAIL_COUNT_KEY = 'nao_inspiration_fail_count';

  const [inspiration, setInspiration] = useState<{quote: string, source: string, url?: string} | null>(() => {
    const saved = localStorage.getItem(INSPIRATION_CACHE_KEY);
    return saved ? JSON.parse(saved) : null;
  });
  const [loading, setLoading] = useState(false);
  const [disabledUntil, setDisabledUntil] = useState<number>(() => {
    const saved = localStorage.getItem(INSPIRATION_DISABLE_KEY);
    return saved ? parseInt(saved) : 0;
  });
  const initialFetchRef = useRef(false);

  const getBackoffMs = (failCount: number) => {
    const base = 5 * 60 * 1000;
    const max = 24 * 60 * 60 * 1000;
    const backoff = base * Math.pow(2, Math.max(0, failCount - 1));
    return Math.min(max, backoff);
  };

  const fetchNewInspiration = useCallback(async (force = false) => {
    const now = Date.now();
    if (!force && disabledUntil && disabledUntil > now) return;
    setLoading(true);
    try {
      const result = await fetchInternetInspiration(DEFAULT_AI_SETTINGS);
      setInspiration(result);
      localStorage.setItem(INSPIRATION_CACHE_KEY, JSON.stringify(result));
      localStorage.setItem(INSPIRATION_FETCH_KEY, Date.now().toString());
      localStorage.setItem(INSPIRATION_HOUR_KEY, Math.floor(Date.now() / (60 * 60 * 1000)).toString());
      localStorage.removeItem(INSPIRATION_DISABLE_KEY);
      localStorage.setItem(INSPIRATION_FAIL_COUNT_KEY, '0');
      setDisabledUntil(0);
    } catch (e) {
      const currentFail = parseInt(localStorage.getItem(INSPIRATION_FAIL_COUNT_KEY) || '0') + 1;
      const nextRetryAt = Date.now() + getBackoffMs(currentFail);
      localStorage.setItem(INSPIRATION_FAIL_COUNT_KEY, currentFail.toString());
      localStorage.setItem(INSPIRATION_DISABLE_KEY, nextRetryAt.toString());
      setDisabledUntil(nextRetryAt);
      console.warn("Failed to fetch inspiration", e);
    } finally {
      setLoading(false);
    }
  }, [disabledUntil]);

  useEffect(() => {
    if (initialFetchRef.current) return;
    initialFetchRef.current = true;
    const lastFetch = localStorage.getItem(INSPIRATION_FETCH_KEY);
    const lastHour = localStorage.getItem(INSPIRATION_HOUR_KEY);
    const oneHour = 60 * 60 * 1000;
    const now = Date.now();
    if (disabledUntil && disabledUntil > now) return;
    const currentHour = Math.floor(now / (60 * 60 * 1000));
    const needsRotation = !lastHour || parseInt(lastHour) !== currentHour;
    if (!inspiration || !lastFetch || (now - parseInt(lastFetch) > oneHour) || needsRotation) {
      fetchNewInspiration();
    }
  }, [fetchNewInspiration, inspiration, disabledUntil]);

  return (
    <div className="bg-white dark:bg-zinc-900 border border-paper-200 dark:border-zinc-800 p-6 rounded-[2.5rem] shadow-sm flex flex-col h-full relative group transition-all hover:shadow-lg">
        <div className="flex items-center justify-between mb-2 shrink-0">
           <h4 className="text-[10px] font-black uppercase tracking-widest text-ink-900 dark:text-zinc-100 flex items-center gap-2">
             <Quote className="w-4 h-4 text-indigo-500" /> 每日灵感
           </h4>
           <button 
            onClick={() => fetchNewInspiration(true)} 
            disabled={loading}
            className={`p-1.5 text-ink-300 hover:text-brand-500 transition-all ${loading ? 'animate-spin' : ''}`}
            title="手动刷新"
          >
            <RefreshCw className="w-3.5 h-3.5" />
          </button>
        </div>
        {disabledUntil > Date.now() && !loading && (
          <div className="text-[9px] text-amber-500 font-bold uppercase tracking-widest mb-3">网络不可用，已暂停自动获取，可手动刷新</div>
        )}
       <div className="flex-1 flex flex-col justify-center">
          {loading ? (
             <div className="flex flex-col items-center gap-3 opacity-40">
                <Loader2 className="w-6 h-6 animate-spin" />
                <span className="text-[10px] font-bold">漫步云端寻觅文字...</span>
             </div>
          ) : inspiration ? (
             <div className="space-y-4 animate-in fade-in duration-500">
                <p className="text-sm font-serif italic text-ink-800 dark:text-zinc-200 leading-relaxed text-center px-2">
                  “{inspiration.quote}”
                </p>
                <div className="text-right flex flex-col items-end">
                   <span className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase">—— {inspiration.source}</span>
                   {inspiration.url && (
                     <a href={inspiration.url} target="_blank" rel="noopener noreferrer" className="text-[8px] text-brand-500 hover:underline flex items-center gap-0.5 mt-1">
                       查看出处 <ExternalLink className="w-2 h-2" />
                     </a>
                   )}
                </div>
             </div>
          ) : (
            <p className="text-[10px] text-center text-ink-300 italic">暂无灵感数据</p>
          )}
       </div>
    </div>
  );
};

// ... other existing sub-components (QuickShortcuts, Achievements, QuestBoard, DailyGoalRing, RealHeatmap) kept as is ...
const QuickShortcuts: React.FC<{ onAction: (mode: ViewMode) => void, hasProject: boolean }> = ({ onAction, hasProject }) => (
  <div className="bg-white dark:bg-zinc-900 border border-paper-200 dark:border-zinc-800 p-6 rounded-[2rem] shadow-sm flex flex-col justify-between h-full group">
    <div className="flex justify-between items-center mb-4">
      <h4 className="text-[10px] font-black uppercase tracking-widest text-ink-900 dark:text-zinc-100 flex items-center gap-2">
        <Sparkles className="w-4 h-4 text-brand-500" /> 创作速访
      </h4>
    </div>
    <div className="grid grid-cols-1 gap-2">
      <button 
        disabled={!hasProject}
        onClick={() => onAction(ViewMode.PLANBOARD)}
        className="flex items-center gap-3 p-3 rounded-xl bg-paper-50 dark:bg-zinc-950/50 hover:bg-brand-50 dark:hover:bg-brand-900/20 text-ink-700 dark:text-zinc-300 transition-all border border-transparent hover:border-brand-100 disabled:opacity-30"
      >
        <Layout className="w-4 h-4 text-brand-500" />
        <span className="text-xs font-bold">我的大纲</span>
      </button>
      <button 
        disabled={!hasProject}
        onClick={() => onAction(ViewMode.WORLD)}
        className="flex items-center gap-3 p-3 rounded-xl bg-paper-50 dark:bg-zinc-950/50 hover:bg-emerald-50 dark:hover:bg-emerald-900/20 text-ink-700 dark:text-zinc-300 transition-all border border-transparent hover:border-emerald-100 disabled:opacity-30"
      >
        <Users className="w-4 h-4 text-emerald-500" />
        <span className="text-xs font-bold">人物设定</span>
      </button>
    </div>
  </div>
);

const Achievements: React.FC<{ totalWords: number }> = ({ totalWords }) => {
  const badges = [
    { limit: 0, label: '初出茅庐', color: 'bg-zinc-100 dark:bg-zinc-800 text-zinc-500' },
    { limit: 10000, label: '万字达成', color: 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600' },
    { limit: 50000, label: '笔耕不辍', color: 'bg-blue-100 dark:bg-blue-900/30 text-blue-600' },
    { limit: 100000, label: '著作等身', color: 'bg-amber-100 dark:bg-amber-900/30 text-amber-600' },
  ];

  return (
    <div className="bg-white dark:bg-zinc-900 border border-paper-200 dark:border-zinc-800 p-6 rounded-[2rem] shadow-sm h-full flex flex-col justify-between">
      <div className="flex justify-between items-center mb-4">
        <h4 className="text-[10px] font-black uppercase tracking-widest text-ink-900 dark:text-zinc-100 flex items-center gap-2">
          <Award className="w-4 h-4 text-amber-500" /> 成就勋章
        </h4>
      </div>
      <div className="grid grid-cols-2 gap-2">
        {badges.map((b, i) => (
          <div key={i} className={`flex items-center gap-2 p-2 rounded-xl transition-all ${totalWords >= b.limit ? 'opacity-100' : 'opacity-20 grayscale'}`}>
            <div className={`p-1.5 rounded-lg ${b.color}`}><Star className="w-3 h-3" /></div>
            <span className="text-[10px] font-black text-ink-800 dark:text-zinc-200 truncate">{b.label}</span>
          </div>
        ))}
      </div>
    </div>
  );
};

const QuestBoard: React.FC = () => {
  const [todos, setTodos] = useState<{id: string, text: string, done: boolean}[]>(() => {
    const saved = localStorage.getItem('nao_quest_board');
    return saved ? JSON.parse(saved) : [
      { id: '1', text: '构思第一卷高潮', done: false },
      { id: '2', text: '完善主角人设卡', done: false }
    ];
  });
  const [newTodo, setNewTodo] = useState('');

  useEffect(() => {
    localStorage.setItem('nao_quest_board', JSON.stringify(todos));
  }, [todos]);

  const addTodo = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && newTodo.trim()) {
      setTodos([...todos, { id: Date.now().toString(), text: newTodo.trim(), done: false }]);
      setNewTodo('');
    }
  };

  return (
    <div className="bg-white dark:bg-zinc-900 border border-paper-200 dark:border-zinc-800 p-6 rounded-[2rem] shadow-sm flex flex-col h-full relative overflow-hidden">
      <div className="flex justify-between items-center mb-4">
        <h4 className="text-[10px] font-black uppercase tracking-widest text-ink-900 dark:text-zinc-100 flex items-center gap-2">
          <ListChecks className="w-4 h-4 text-emerald-500" /> 创作待办
        </h4>
        <span className="text-[9px] font-bold text-ink-300 dark:text-zinc-600">{todos.filter(t => t.done).length}/{todos.length}</span>
      </div>
      <div className="flex-1 overflow-y-auto custom-scrollbar space-y-2 mb-4 max-h-[110px]">
        {todos.map(todo => (
          <div key={todo.id} className="flex items-center gap-2 group">
            <button onClick={() => setTodos(todos.map(t => t.id === todo.id ? {...t, done: !t.done} : t))} className={`shrink-0 ${todo.done ? 'text-emerald-500' : 'text-paper-300 dark:text-zinc-700'}`}>
              {todo.done ? <CheckCircle2 className="w-3.5 h-3.5" /> : <Circle className="w-3.5 h-3.5" />}
            </button>
            <span className={`text-[11px] font-bold truncate flex-1 ${todo.done ? 'text-ink-300 dark:text-zinc-600 line-through' : 'text-ink-700 dark:text-zinc-300'}`}>{todo.text}</span>
            <button onClick={() => setTodos(todos.filter(t => t.id !== todo.id))} className="opacity-0 group-hover:opacity-100 text-ink-200 hover:text-rose-500"><X className="w-3 h-3" /></button>
          </div>
        ))}
      </div>
      <input value={newTodo} onChange={e => setNewTodo(e.target.value)} onKeyDown={addTodo} placeholder="+ 添加任务" className="w-full bg-paper-50 dark:bg-zinc-950 border-none rounded-xl py-2 px-3 text-xs font-bold text-ink-900 dark:text-zinc-100 outline-none focus:ring-1 focus:ring-emerald-100 placeholder:text-ink-300" />
    </div>
  );
};

const DailyGoalRing: React.FC<{ 
  current: number, 
  target: number, 
  setTarget: (n: number) => void, 
  isCheckedIn: boolean,
  onCheckIn: () => void 
}> = ({ current, target, setTarget, isCheckedIn, onCheckIn }) => {
  const progress = Math.min(current / target, 1);
  const strokeDashoffset = (normalizedRadius: number) => (normalizedRadius * 2 * Math.PI) - progress * (normalizedRadius * 2 * Math.PI);
  const [isEditing, setIsEditing] = useState(false);
  const [editVal, setEditVal] = useState(String(target));

  return (
    <div className="bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 p-6 rounded-[2rem] flex items-center justify-between shadow-xl h-full relative overflow-hidden group border border-ink-800 dark:border-zinc-200">
       <div className="z-10 flex-1">
          <div className="flex items-center gap-2 mb-2">
             <Target className="w-4 h-4 text-brand-400 dark:text-brand-600" />
             <span className="text-[10px] font-black uppercase tracking-widest opacity-80">今日目标</span>
          </div>
          {isEditing ? (
             <input autoFocus className="w-20 bg-white/20 dark:bg-black/10 rounded text-xl font-black text-center outline-none" value={editVal} onChange={e => setEditVal(e.target.value)} onBlur={() => { const v = parseInt(editVal); if (v > 0) setTarget(v); setIsEditing(false); }} />
          ) : (
             <h3 className="text-3xl font-black font-serif tabular-nums leading-none cursor-pointer" onClick={() => setIsEditing(true)}>
                {current} <span className="text-sm opacity-50 font-sans">/ {target}</span>
             </h3>
          )}
          <button 
            onClick={onCheckIn}
            disabled={isCheckedIn}
            className={`mt-4 px-4 py-1.5 rounded-full text-[9px] font-black uppercase tracking-widest flex items-center gap-2 transition-all ${isCheckedIn ? 'bg-emerald-500 text-white shadow-lg' : 'bg-white/10 dark:bg-black/5 hover:bg-white/20 dark:hover:bg-black/10'}`}
          >
            {isCheckedIn ? <CheckCircle className="w-3 h-3" /> : <ListChecks className="w-3 h-3" />}
            {isCheckedIn ? '今日已达成打卡' : '目标打卡'}
          </button>
       </div>
       <div className="relative w-20 h-20 shrink-0">
          <svg viewBox="0 0 100 100" className="rotate-[-90deg]">
             <circle stroke="currentColor" fill="transparent" strokeWidth="8" strokeOpacity="0.1" r="40" cx="50" cy="50" />
             <circle stroke="currentColor" fill="transparent" strokeWidth="8" strokeDasharray={`${40 * 2 * Math.PI} ${40 * 2 * Math.PI}`} style={{ strokeDashoffset: strokeDashoffset(40), transition: 'stroke-dashoffset 0.8s ease' }} strokeLinecap="round" r="40" cx="50" cy="50" className={current >= target ? 'text-emerald-400 dark:text-emerald-600' : 'text-brand-400 dark:text-brand-600'} />
          </svg>
          <div className="absolute inset-0 flex items-center justify-center font-black text-[10px]">{Math.round(progress * 100)}%</div>
       </div>
    </div>
  );
};

const RealHeatmap: React.FC<{ projects: Project[] }> = ({ projects }) => {
  // ... existing implementation
  const days = useMemo(() => {
    const d = [];
    const now = new Date();
    for (let i = 0; i < 154; i++) { 
      const date = new Date(now);
      date.setDate(date.getDate() - i);
      d.unshift(date);
    }
    return d;
  }, []);

  const activityMap = useMemo(() => {
    const map: Record<string, number> = {};
    projects.forEach(p => p.documents.forEach(doc => {
      const ts = parseInt(doc.id.substring(1).split('_')[0]);
      if (!isNaN(ts)) {
        const dateStr = new Date(ts).toISOString().split('T')[0];
        map[dateStr] = (map[dateStr] || 0) + 1;
      }
    }));
    return map;
  }, [projects]);

  return (
    <div className="bg-white dark:bg-zinc-900 border border-paper-200 dark:border-zinc-800 p-5 rounded-[2rem] shadow-sm flex flex-col h-full">
        <div className="flex items-center justify-between mb-4 shrink-0">
            <h4 className="text-[10px] font-black uppercase tracking-widest text-ink-900 dark:text-zinc-100 flex items-center gap-2">
                <TrendingUp className="w-3.5 h-3.5 text-brand-500" /> 创作热力 (半年)
            </h4>
        </div>
        <div className="flex-1 grid grid-flow-col grid-rows-7 gap-1 overflow-hidden px-1">
            {days.map((date, i) => {
                const count = activityMap[date.toISOString().split('T')[0]] || 0;
                return (
                    <div key={i} className={`w-2 h-2 rounded-sm transition-all ${count === 0 ? 'bg-paper-100 dark:bg-zinc-800' : count === 1 ? 'bg-brand-200 dark:bg-brand-900/40' : count <= 3 ? 'bg-brand-400 dark:bg-brand-700/60' : 'bg-brand-600 dark:bg-brand-50'}`} title={`${date.toLocaleDateString()}: ${count}`} />
                );
            })}
        </div>
    </div>
  );
};

export const ProjectDashboard: React.FC<ProjectDashboardProps> = ({ onCreateNew }) => {
  const { projects, selectProject, deleteProject, createProject, setActiveDocumentId, setViewMode, activeProjectId } = useProject();
  const { user, hasAIAccess } = useAuth();
  const { confirm } = useConfirm();
  const { timeStr, dateStr, period } = useChineseTime();
  const [showManualCreate, setShowManualCreate] = useState(false);
  const [quickNote, setQuickNote] = useState(() => localStorage.getItem('nao_quick_note') || '');
  const [newTitle, setNewTitle] = useState('');
  const [newDesc, setNewDesc] = useState('');

  const totalWords = useMemo(() => projects.reduce((acc, p) => acc + p.documents.reduce((dAcc, d) => dAcc + (d.content?.length || 0), 0), 0), [projects]);
  const { dailyTarget, setDailyTarget, todayWords, isCheckedIn, handleCheckIn } = useDailyProgress(totalWords);

  useEffect(() => {
    localStorage.setItem('nao_quick_note', quickNote);
  }, [quickNote]);

  const lastActiveDoc = useMemo(() => {
    let latest = null;
    let maxTs = 0;
    projects.forEach(p => p.documents.forEach(d => {
      const ts = parseInt(d.id.substring(1).split('_')[0]);
      if (ts > maxTs) { maxTs = ts; latest = { doc: d, project: p }; }
    }));
    return latest;
  }, [projects]);

  const handleManualCreateSubmit = () => {
    if (!newTitle.trim()) return;
    const pid = `p${Date.now()}`;
    const vid = `v${Date.now()}`;
    createProject({
      id: pid, title: newTitle, coreConflict: newDesc || '核心冲突待定', characterArc: '', ultimateValue: '', volumes: [{ id: vid, title: '第一卷', order: 0, theme: '', coreGoal: '', boundaries: '', }],
      documents: [{ id: `d${Date.now()}`, volumeId: vid, title: '第一章', content: '', status: '草稿', order: 0, linkedIds: [], bookmarks: [] }],
      entities: [], templates: [], 
      aiSettings: { ...DEFAULT_AI_SETTINGS }
    });
    setShowManualCreate(false);
  };

  return (
    <div className="h-full w-full bg-paper-50 dark:bg-zinc-950 bg-dot-pattern overflow-y-auto custom-scrollbar transition-colors relative">
      
      {showManualCreate && (
        <div className="fixed inset-0 z-[110] bg-ink-900/40 dark:bg-black/60 backdrop-blur-md flex items-center justify-center p-6 animate-in fade-in duration-200">
          <div className="bg-white dark:bg-zinc-900 rounded-[2.5rem] shadow-2xl w-full max-w-lg p-8 border border-white/10">
             <div className="flex justify-between items-center mb-6">
                <div className="flex items-center gap-3"><PenTool className="w-5 h-5 text-ink-900 dark:text-zinc-400" /><h3 className="text-xl font-black text-ink-900 dark:text-zinc-100 font-serif">手动立项</h3></div>
                <button onClick={() => setShowManualCreate(false)}><X className="w-6 h-6 text-ink-300" /></button>
             </div>
             <div className="space-y-4 mb-8">
               <input autoFocus value={newTitle} onChange={e => setNewTitle(e.target.value)} className="w-full bg-paper-50 dark:bg-zinc-950 border-none rounded-xl p-4 text-lg font-bold text-ink-900 dark:text-zinc-100" placeholder="小说标题..." />
               <textarea value={newDesc} onChange={e => setNewDesc(e.target.value)} className="w-full h-32 bg-paper-50 dark:bg-zinc-950 border-none rounded-xl p-4 text-sm resize-none text-ink-900 dark:text-zinc-100" placeholder="简介..." />
             </div>
             <button onClick={handleManualCreateSubmit} className="w-full py-4 bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-2xl font-black text-xs uppercase tracking-widest shadow-xl">保存项目</button>
          </div>
        </div>
      )}

      <div className="max-w-[1400px] mx-auto p-8 md:p-12 space-y-8 pt-20">
        <div className="flex flex-col md:flex-row justify-between items-end gap-6 border-b border-paper-200 dark:border-zinc-800 pb-8">
          <div className="flex-1">
             <h1 className="text-4xl font-black text-ink-900 dark:text-zinc-100 font-serif mb-2">{period}好，创造者</h1>
             <p className="text-ink-400 font-medium font-serif italic">{dateStr}</p>
          </div>
          <div className="text-right">
            <div className="text-5xl font-black text-ink-900 dark:text-zinc-100 tabular-nums leading-none tracking-tighter mb-1">{timeStr}</div>
            <div className="text-[10px] font-black text-brand-500 uppercase tracking-[0.3em]">CST · 中国标准时间</div>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <div onClick={() => lastActiveDoc && (selectProject(lastActiveDoc.project.id), setActiveDocumentId(lastActiveDoc.doc.id), setViewMode(ViewMode.WRITER))} className={`col-span-1 md:col-span-2 relative overflow-hidden rounded-[2.5rem] p-8 flex flex-col justify-between group min-h-[220px] transition-all cursor-pointer ${lastActiveDoc ? 'bg-white dark:bg-zinc-900 border border-paper-200 dark:border-zinc-800 hover:shadow-xl' : 'bg-paper-100 dark:bg-zinc-800 cursor-default opacity-50'}`}>
             <div className="flex justify-between items-start relative z-10"><Zap className="w-6 h-6 text-brand-600" />{lastActiveDoc && <span className="px-3 py-1 bg-brand-100 dark:bg-brand-900/30 text-brand-700 dark:text-brand-300 rounded-full text-[10px] font-black uppercase">最近编辑</span>}</div>
             <div className="relative z-10 mt-6"><h3 className="text-2xl font-black font-serif mb-2 line-clamp-1 text-ink-900 dark:text-zinc-100">{lastActiveDoc ? lastActiveDoc.doc.title : '暂无最近记录'}</h3><p className="text-sm font-medium text-ink-500 line-clamp-1 uppercase">{lastActiveDoc ? lastActiveDoc.project.title : '开始你的第一步'}</p></div>
          </div>
          <div className="col-span-1"><DailyGoalRing current={todayWords} target={dailyTarget} setTarget={setDailyTarget} isCheckedIn={isCheckedIn} onCheckIn={handleCheckIn} /></div>
          
          {/* New Points Card */}
          <div className="col-span-1"><PointsCard /></div>
          
          <div className="col-span-1 h-56"><Achievements totalWords={totalWords} /></div>
          <div className="col-span-1 h-56"><QuickShortcuts hasProject={!!projects.length} onAction={(m) => { if(projects.length > 0) { if(!activeProjectId) selectProject(projects[0].id); setViewMode(m); } }} /></div>
          <div className="col-span-1 h-56"><InspirationCard /></div>
          <div className="col-span-1 h-56 bg-white dark:bg-zinc-900 border border-paper-200 dark:border-zinc-800 p-6 rounded-[2.5rem] shadow-sm flex flex-col relative group">
             <div className="flex items-center gap-2 mb-3"><Lightbulb className="w-4 h-4 text-amber-500" /><h4 className="text-[10px] font-black uppercase tracking-widest text-ink-900 dark:text-zinc-100">灵感闪念</h4></div>
             <textarea value={quickNote} onChange={e => setQuickNote(e.target.value)} placeholder="随手记下灵感..." className="flex-1 bg-paper-50 dark:bg-zinc-950/50 rounded-2xl p-4 text-xs font-medium text-ink-800 dark:text-zinc-300 resize-none border-none outline-none focus:ring-1 focus:ring-amber-200 transition-all custom-scrollbar" />
          </div>
          <div className="col-span-1 h-56"><QuestBoard /></div>
          <div className="col-span-1 h-56"><RealHeatmap projects={projects} /></div>
        </div>

        {/* Existing Project List ... */}
        <div className="space-y-6 pt-6">
           <div className="flex items-center justify-between border-b border-paper-200 dark:border-zinc-800 pb-4">
              <h2 className="text-xl font-black text-ink-900 dark:text-zinc-100 flex items-center gap-3"><Book className="w-5 h-5" /> 我的书架 <span className="text-xs bg-brand-500 text-white px-2 py-0.5 rounded-full">{projects.length}</span></h2>
              <button 
                onClick={hasAIAccess ? onCreateNew : () => alert('AI 订阅已过期，无法使用智能孵化功能。请使用下方“新建空白项目”进行手动创作。')} 
                className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-black uppercase transition-all shadow-lg active:scale-95 ${hasAIAccess ? 'bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 hover:bg-black dark:hover:bg-white' : 'bg-gray-200 dark:bg-zinc-800 text-gray-400 cursor-not-allowed'}`}
              >
                {hasAIAccess ? <Sparkles className="w-3.5 h-3.5" /> : <Lock className="w-3.5 h-3.5" />} 
                AI 孵化新宇宙
              </button>
           </div>
           <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 pb-20">
              <button onClick={() => setShowManualCreate(true)} className="group flex flex-col items-center justify-center p-8 border-2 border-dashed border-paper-200 dark:border-zinc-800 rounded-[2.5rem] hover:border-brand-500 hover:bg-brand-50/10 min-h-[280px] transition-all"><div className="p-5 bg-paper-50 dark:bg-zinc-900 rounded-full text-ink-300 group-hover:bg-brand-500 group-hover:text-white transition-all mb-4 border border-paper-100 dark:border-zinc-800 shadow-inner"><Plus className="w-8 h-8" /></div><span className="text-sm font-black text-ink-400 uppercase tracking-widest">新建空白项目</span></button>
               {projects.map(project => (
                 <div key={project.id} onClick={() => selectProject(project.id)} className="group relative bg-white dark:bg-zinc-900 p-8 rounded-[2.5rem] border border-paper-200 dark:border-zinc-800 shadow-sm hover:shadow-2xl transition-all duration-500 flex flex-col justify-between min-h-[280px] overflow-hidden cursor-pointer">
                    <div className="absolute -right-10 -top-10 w-40 h-40 bg-gradient-to-br from-brand-50 to-transparent dark:from-brand-900/10 rounded-full blur-3xl group-hover:scale-150 transition-transform pointer-events-none" />
                    <div className="relative z-10"><div className="flex justify-between items-start mb-6"><div className="p-3 bg-brand-50 dark:bg-zinc-800 rounded-2xl text-brand-600 dark:text-zinc-400 group-hover:bg-brand-600 group-hover:text-white transition-all shadow-sm"><Book className="w-6 h-6" /></div><ConfirmButton onClick={(e) => e.stopPropagation()} onConfirm={() => deleteProject(project.id)} title="确定永久删除此项目吗？" description="删除后将无法恢复。" confirmText="删除" cancelText="取消" tone="danger" className="p-2 text-ink-200 hover:text-red-500 transition-colors"><Trash2 className="w-4 h-4" /></ConfirmButton></div><h3 className="text-2xl font-black text-ink-900 dark:text-zinc-100 font-serif line-clamp-2 mb-2 group-hover:text-brand-700 transition-colors">{project.title}</h3><p className="text-xs text-ink-400 dark:text-zinc-500 line-clamp-3 leading-relaxed font-medium">{project.coreConflict || "未定义核心冲突"}</p></div>
                    <div className="relative z-10 pt-6 mt-4 border-t border-paper-100 dark:border-zinc-800 flex items-center justify-between"><div className="text-[10px] font-black uppercase text-ink-400 dark:text-zinc-600 tracking-widest">{project.documents.length} 章</div><div className="w-8 h-8 rounded-full bg-paper-50 dark:bg-zinc-800 flex items-center justify-center text-ink-400 dark:text-zinc-500 group-hover:bg-brand-600 group-hover:text-white transition-all"><ChevronRight className="w-4 h-4" /></div></div>
                 </div>
               ))}
           </div>
        </div>
      </div>
    </div>
  );
};
