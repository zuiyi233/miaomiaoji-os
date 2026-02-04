
import React, { useState, useEffect, useMemo } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { getExportCodesUrl } from '../services/redemptionApi';
import { fetchProviderConfigApi, testProviderConfigApi, updateProviderConfigApi } from '../services/aiConfigApi';
import { useProject } from '../contexts/ProjectContext';
import { getBackupDownloadUrl, listProjectBackupsApi } from '../services/projectApi';
import { useConfirm } from '../contexts/ConfirmContext';
import { 
  UserCircle, Shield, Calendar, Trash2, LogOut, Palette, Cpu, 
  ArrowLeft, Sparkles, Globe, BrainCircuit, RefreshCw, 
  Check, Info, Keyboard, List, Hash, ShieldCheck, Zap, AlertCircle,
  Loader2, Activity, Database, Laptop, Users, LayoutDashboard, ChevronLeft, Key, Plus,
  Search, Download, Copy, FileText, XCircle, PlayCircle, Sliders, AlertTriangle, Save, Filter, Coins, Gift, ChevronRight, Lock
} from 'lucide-react';
import { AIProvider, CodeStatus } from '../types';

type SettingsTab = 'profile' | 'ai-engine' | 'appearance' | 'data' | 'admin';

const StatusBadge: React.FC<{ status: CodeStatus, expiresAt: number }> = ({ status, expiresAt }) => {
  const isExpired = Date.now() > expiresAt;
  
  if (status === 'disabled') {
    return <span className="px-2 py-0.5 rounded-full bg-gray-100 dark:bg-zinc-800 text-gray-500 text-[9px] font-black uppercase tracking-wider border border-gray-200 dark:border-zinc-700">已禁用</span>;
  }
  if (isExpired) {
    return <span className="px-2 py-0.5 rounded-full bg-orange-50 dark:bg-orange-900/20 text-orange-600 dark:text-orange-400 text-[9px] font-black uppercase tracking-wider border border-orange-100 dark:border-orange-900/30">已过期</span>;
  }
  if (status === 'depleted') {
    return <span className="px-2 py-0.5 rounded-full bg-yellow-50 dark:bg-yellow-900/20 text-yellow-600 dark:text-yellow-400 text-[9px] font-black uppercase tracking-wider border border-yellow-100 dark:border-yellow-900/30">已耗尽</span>;
  }
  return <span className="px-2 py-0.5 rounded-full bg-emerald-50 dark:bg-emerald-900/20 text-emerald-600 dark:text-emerald-400 text-[9px] font-black uppercase tracking-wider border border-emerald-100 dark:border-emerald-900/30">正常</span>;
};

export const UserSettings: React.FC = () => {
  const { user, allUsers, logout, redemptionCodes, batchGenerateCodes, batchUpdateCodes, fetchCodes, deviceId, hasAIAccess, systemConfig, updateSystemConfig } = useAuth();
  const { project, theme, setTheme, defaultAISettings, updateDefaultAISettings, updateAISettings, availableModels, refreshModels, previousViewMode, navigateBack, projects, backupProject, restoreLatestBackup, cloudSyncEnabled, setCloudSyncEnabled } = useProject();
  const [backupList, setBackupList] = useState<Array<{ id: number; file_name: string; created_at: string; size_bytes: number }>>([]);
  const [backupTotal, setBackupTotal] = useState(0);
  const [backupLoading, setBackupLoading] = useState(false);
  const [showCloudSyncTip, setShowCloudSyncTip] = useState(false);
  const { confirm } = useConfirm();
  
  const [activeTab, setActiveTab] = useState<SettingsTab>('profile');
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isManualInput, setIsManualInput] = useState(false);
  const [providerConfig, setProviderConfig] = useState<{ provider: AIProvider; baseUrl: string; apiKey: string }>({
    provider: 'gemini',
    baseUrl: '',
    apiKey: ''
  });
   const [isSavingProvider, setIsSavingProvider] = useState(false);
   const [isTestingProvider, setIsTestingProvider] = useState(false);
  
  // --- Admin State ---
  const [genPrefix, setGenPrefix] = useState('');
  const [genCount, setGenCount] = useState(10);
  const [genValidity, setGenValidity] = useState(30); 
  const [genMaxUses, setGenMaxUses] = useState(1);
  const [genTags, setGenTags] = useState('');
  const [genNote, setGenNote] = useState('');
  const [genCharType, setGenCharType] = useState<'alphanum' | 'num' | 'alpha'>('alphanum');
  const [genLength, setGenLength] = useState(8);
  const [showPreview, setShowPreview] = useState(false);

  // Points Config State
  const [dailyPointsMax, setDailyPointsMax] = useState(systemConfig.checkInBasePointsMax);
  const [enableExchange, setEnableExchange] = useState(systemConfig.enablePointsExchange);
  const [newOption, setNewOption] = useState({ name: '', cost: 1000, days: 30 });

  // List Management
  const [selectedCodes, setSelectedCodes] = useState<Set<string>>(new Set());
  const [filterStatus, setFilterStatus] = useState<'all' | 'active' | 'disabled' | 'expired' | 'depleted' | 'rewards'>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [sortOrder, setSortOrder] = useState<'desc' | 'asc'>('desc');
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 20;

  // Stats
  const stats = useMemo(() => {
    const total = redemptionCodes.length;
    const active = redemptionCodes.filter(c => c.status === 'active' && Date.now() < c.expiresAt).length;
    const expiring = redemptionCodes.filter(c => c.status === 'active' && Date.now() < c.expiresAt && c.expiresAt - Date.now() < 3 * 24 * 60 * 60 * 1000).length;
    return { total, active, expiring };
  }, [redemptionCodes]);

   const providers: { id: AIProvider; name: string; icon: React.ReactNode; color: string; activeBg: string }[] = [
     { id: 'gemini', name: 'Google Gemini', icon: <Sparkles className="w-4 h-4" />, color: 'bg-blue-500', activeBg: 'bg-blue-50 border-blue-200 dark:bg-blue-900/20' },
     { id: 'openai', name: 'OpenAI (官方)', icon: <Cpu className="w-4 h-4" />, color: 'bg-emerald-500', activeBg: 'bg-emerald-50 border-emerald-200 dark:bg-emerald-900/20' },
     { id: 'openrouter', name: 'OpenRouter', icon: <Globe className="w-4 h-4" />, color: 'bg-sky-500', activeBg: 'bg-sky-50 border-sky-200 dark:bg-sky-900/20' },
     { id: 'anthropic', name: 'Anthropic', icon: <Globe className="w-4 h-4" />, color: 'bg-fuchsia-500', activeBg: 'bg-fuchsia-50 border-fuchsia-200 dark:bg-fuchsia-900/20' },
     { id: 'proxy', name: '第三方代理', icon: <Globe className="w-4 h-4" />, color: 'bg-indigo-500', activeBg: 'bg-indigo-50 border-indigo-200 dark:bg-indigo-900/20' },
     { id: 'local', name: '本地模型', icon: <Cpu className="w-4 h-4" />, color: 'bg-amber-500', activeBg: 'bg-amber-50 border-amber-200 dark:bg-amber-900/20' },
   ];

  useEffect(() => {
    if (availableModels.length === 0) handleRefreshModels();
  }, []);

  useEffect(() => {
    if (!cloudSyncEnabled) {
      setShowCloudSyncTip(true);
    }
  }, [cloudSyncEnabled]);

   useEffect(() => {
     if (user?.role === 'admin') {
       loadProviderConfig(providerConfig.provider);
     }
   }, [user?.role, providerConfig.provider]);

  const handleRefreshModels = async () => {
    setIsRefreshing(true);
    try {
      await refreshModels(project?.aiSettings || defaultAISettings);
    } catch (e) {
      console.error(e);
    } finally {
      setIsRefreshing(false);
    }
  };

   const loadProviderConfig = async (provider: AIProvider) => {
    if (!user || user.role !== 'admin') return;
    try {
      const data = await fetchProviderConfigApi(provider);
      setProviderConfig({
        provider: provider,
        baseUrl: data.base_url || '',
        apiKey: data.api_key || ''
      });
    } catch (e) {
      setProviderConfig({ provider, baseUrl: '', apiKey: '' });
    }
  };

  const saveProviderConfig = async () => {
    if (!user || user.role !== 'admin') return;
    setIsSavingProvider(true);
    try {
      await updateProviderConfigApi({
        provider: providerConfig.provider,
        base_url: providerConfig.baseUrl,
        api_key: providerConfig.apiKey
      });
      await loadProviderConfig(providerConfig.provider);
      alert('供应商配置已保存');
    } catch (e) {
      alert('保存失败，请检查配置');
    } finally {
      setIsSavingProvider(false);
    }
  };

  const testProviderConfig = async () => {
    if (!user || user.role !== 'admin') return;
    setIsTestingProvider(true);
    try {
      await testProviderConfigApi(providerConfig.provider);
      alert('连接测试成功');
    } catch (e) {
      alert('连接测试失败');
    } finally {
      setIsTestingProvider(false);
    }
  };

  const handleSavePointsConfig = () => {
      updateSystemConfig({
          checkInBasePointsMax: dailyPointsMax,
          enablePointsExchange: enableExchange
      });
  };

  const handleAddOption = () => {
      if (!newOption.name) return;
      const newOpt = {
          id: `opt_${Date.now()}`,
          name: newOption.name,
          cost: newOption.cost,
          durationDays: newOption.days,
          description: `${newOption.days}天有效期`
      };
      updateSystemConfig({
          exchangeOptions: [...systemConfig.exchangeOptions, newOpt]
      });
      setNewOption({ name: '', cost: 1000, days: 30 });
  };

  const handleRemoveOption = (id: string) => {
      updateSystemConfig({
          exchangeOptions: systemConfig.exchangeOptions.filter(o => o.id !== id)
      });
  };

  const handleGenerateConfirm = () => {
    batchGenerateCodes({
        prefix: genPrefix.toUpperCase(),
        length: genLength,
        charType: genCharType,
        count: genCount,
        validityDays: genValidity,
        maxUses: genMaxUses,
        tags: genTags.split(/[,，\s]+/).filter(Boolean),
        note: genNote
    });
    setShowPreview(false);
    setGenNote('');
    alert(`成功生成 ${genCount} 个兑换码。`);
  };

  const filteredCodes = useMemo(() => redemptionCodes, [redemptionCodes]);
  const paginatedCodes = filteredCodes;
  const totalPages = Math.max(1, Math.ceil(filteredCodes.length / itemsPerPage));

  useEffect(() => {
    if (user?.role !== 'admin') return;
    const status = filterStatus === 'rewards' ? 'all' : filterStatus;
    const search = filterStatus === 'rewards' ? 'points_exchange' : searchQuery;
    fetchCodes(currentPage, itemsPerPage, status, search, sortOrder).catch(() => {});
  }, [filterStatus, searchQuery, sortOrder, currentPage, user?.role]);

  const handleSelectAll = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.checked) {
      setSelectedCodes(new Set(filteredCodes.map(c => c.code)));
    } else {
      setSelectedCodes(new Set());
    }
  };

  const handleBatchAction = async (action: 'disable' | 'enable' | 'delete' | 'renew') => {
    if (selectedCodes.size === 0) return;
    const actionMap = { disable: '禁用', enable: '启用', delete: '删除', renew: '续期30天' };
    const ok = await confirm({
      title: `确定要批量${actionMap[action]}选中的 ${selectedCodes.size} 个兑换码吗？`,
      description: '该操作将应用到所有选中的兑换码。',
      confirmText: '确认执行',
      cancelText: '取消',
      tone: action === 'delete' ? 'danger' : 'default',
    });
    if (!ok) return;
    
    let val = undefined;
    if (action === 'renew') val = 30; 

    batchUpdateCodes(Array.from(selectedCodes), action, val);
    setSelectedCodes(new Set());
  };

  const handleExport = () => {
    const link = document.createElement("a");
    const status = filterStatus === 'rewards' ? 'all' : filterStatus;
    const search = filterStatus === 'rewards' ? 'points_exchange' : searchQuery;
    link.setAttribute("href", getExportCodesUrl(status, search));
    link.setAttribute("download", `nao_codes_export_${Date.now()}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const handleClearCache = async () => {
    const ok = await confirm({
      title: '确定清除所有本地缓存？',
      description: '包含模型列表与临时状态，但项目数据不会丢失。',
      confirmText: '清除缓存',
      cancelText: '取消',
      tone: 'danger',
    });
    if (ok) {
      localStorage.clear();
      alert("缓存已清除，页面将刷新。");
      window.location.reload();
    }
  };

  const handleExportAllData = () => {
    const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(projects));
    const downloadAnchorNode = document.createElement('a');
    downloadAnchorNode.setAttribute("href",     dataStr);
    downloadAnchorNode.setAttribute("download", "nao_backup_all.json");
    document.body.appendChild(downloadAnchorNode);
    downloadAnchorNode.click();
    downloadAnchorNode.remove();
  };

  const loadBackups = async () => {
    if (!project) return;
    setBackupLoading(true);
    try {
      const data = await listProjectBackupsApi(Number(project.id), 1, 20);
      setBackupList(data.files || []);
      setBackupTotal(data.total || 0);
    } catch {
      setBackupList([]);
      setBackupTotal(0);
    } finally {
      setBackupLoading(false);
    }
  };

  const navItem = (id: SettingsTab, label: string, icon: React.ReactNode) => (
    <button
      onClick={() => setActiveTab(id)}
      className={`w-full flex items-center justify-between px-5 py-3.5 rounded-2xl text-[11px] font-black uppercase tracking-widest transition-all ${
        activeTab === id 
          ? 'bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 shadow-xl translate-x-1' 
          : 'text-ink-400 dark:text-zinc-500 hover:bg-paper-100 dark:hover:bg-zinc-800 hover:text-ink-900 dark:hover:text-zinc-300'
      }`}
    >
      <div className="flex items-center gap-3">
        {icon}
        {label}
      </div>
      {activeTab === id && <ChevronRight className="w-3 h-3 opacity-50" />}
    </button>
  );

  const currentSettings = project ? project.aiSettings : defaultAISettings;
  const isGeminiModel = currentSettings.provider === 'gemini' || currentSettings.model.includes('gemini');
  const supportsThinking = isGeminiModel && (currentSettings.model.includes('3') || currentSettings.model.includes('2.5'));

  const renderAdminPanel = () => (
    <div className="space-y-8 animate-in fade-in duration-500">
       <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="bg-white dark:bg-zinc-900 p-6 rounded-[2rem] border border-paper-200 dark:border-zinc-800 shadow-sm flex items-center justify-between">
             <div>
                <p className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest">活跃兑换码</p>
                <p className="text-3xl font-black text-emerald-500 tabular-nums">{stats.active}</p>
             </div>
             <div className="p-3 bg-emerald-50 dark:bg-emerald-900/20 rounded-2xl"><Check className="w-6 h-6 text-emerald-500"/></div>
          </div>
          <div className="bg-white dark:bg-zinc-900 p-6 rounded-[2rem] border border-paper-200 dark:border-zinc-800 shadow-sm flex items-center justify-between">
             <div>
                <p className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest">累计生成</p>
                <p className="text-3xl font-black text-brand-600 dark:text-brand-400 tabular-nums">{stats.total}</p>
             </div>
             <div className="p-3 bg-brand-50 dark:bg-brand-900/20 rounded-2xl"><Database className="w-6 h-6 text-brand-500"/></div>
          </div>
          <div className="bg-white dark:bg-zinc-900 p-6 rounded-[2rem] border border-paper-200 dark:border-zinc-800 shadow-sm flex items-center justify-between cursor-pointer hover:border-indigo-500 transition-colors" onClick={() => setFilterStatus('rewards')}>
             <div>
                <p className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest">积分兑换单</p>
                <p className="text-3xl font-black text-indigo-600 dark:text-indigo-400 tabular-nums">{redemptionCodes.filter(c => c.source === 'points_exchange').length}</p>
             </div>
             <div className="p-3 bg-indigo-50 dark:bg-indigo-900/20 rounded-2xl"><Gift className="w-6 h-6 text-indigo-500"/></div>
          </div>
          <div className="bg-white dark:bg-zinc-900 p-6 rounded-[2rem] border border-paper-200 dark:border-zinc-800 shadow-sm flex items-center justify-between">
             <div>
                <p className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest">系统负载</p>
                <p className="text-sm font-black text-ink-900 dark:text-zinc-200 mt-1 uppercase tracking-wider">Optimal</p>
             </div>
             <div className="p-3 bg-paper-100 dark:bg-zinc-800 rounded-2xl"><Activity className="w-6 h-6 text-ink-500"/></div>
          </div>
       </div>

       <div className="grid grid-cols-12 gap-8">
          <div className="col-span-12 xl:col-span-4 space-y-8">
             {/* Points Config */}
             <div className="bg-indigo-600 text-white rounded-[2.5rem] p-8 shadow-xl shadow-indigo-100 dark:shadow-none space-y-6">
                <div className="flex items-center gap-3">
                    <div className="p-2.5 bg-white/20 rounded-xl"><Coins className="w-5 h-5" /></div>
                    <h3 className="text-lg font-black uppercase tracking-widest">积分商城管理</h3>
                </div>
                <div className="space-y-4">
                    <div className="flex items-center justify-between bg-white/10 p-4 rounded-2xl border border-white/10">
                        <span className="text-xs font-bold">商城访问权限</span>
                        <button 
                            onClick={() => { setEnableExchange(!enableExchange); handleSavePointsConfig(); }}
                            className={`w-12 h-6 rounded-full relative transition-all ${enableExchange ? 'bg-amber-400' : 'bg-white/20'}`}
                        >
                            <div className={`absolute top-1 left-1 w-4 h-4 bg-white rounded-full transition-all shadow-md ${enableExchange ? 'translate-x-6' : ''}`} />
                        </button>
                    </div>
                    <div className="space-y-2">
                        <label className="text-[10px] font-black uppercase tracking-widest opacity-60">每日签到基准分</label>
                        <input type="number" value={dailyPointsMax} onChange={e => setDailyPointsMax(parseInt(e.target.value))} onBlur={handleSavePointsConfig} className="w-full bg-white/10 border border-white/20 rounded-xl p-3 text-sm font-bold text-white outline-none focus:ring-2 focus:ring-white/30" />
                    </div>
                    <div className="pt-4 border-t border-white/10">
                        <p className="text-[10px] font-black uppercase tracking-widest opacity-60 mb-3">商品预设</p>
                        <div className="space-y-2 max-h-[200px] overflow-y-auto custom-scrollbar pr-2">
                            {systemConfig.exchangeOptions.map(opt => (
                                <div key={opt.id} className="flex items-center justify-between bg-white/10 p-3 rounded-xl border border-white/5 text-xs">
                                    <div className="flex flex-col"><span className="font-bold">{opt.name}</span><span className="text-[9px] opacity-60">{opt.cost} pts / {opt.durationDays} days</span></div>
                                    <button onClick={() => handleRemoveOption(opt.id)} className="p-1.5 hover:bg-white/10 rounded-lg text-white/50 hover:text-white transition-all"><Trash2 className="w-3.5 h-3.5"/></button>
                                </div>
                            ))}
                        </div>
                        <div className="grid grid-cols-1 gap-2 mt-4">
                            <input value={newOption.name} onChange={e => setNewOption({...newOption, name: e.target.value})} placeholder="商品名称" className="bg-white/10 border border-white/10 rounded-xl p-2.5 text-xs text-white placeholder:text-white/30 outline-none" />
                            <div className="flex gap-2">
                                <input type="number" value={newOption.cost} onChange={e => setNewOption({...newOption, cost: parseInt(e.target.value)})} placeholder="所需积分" className="flex-1 bg-white/10 border border-white/10 rounded-xl p-2.5 text-xs text-white outline-none" />
                                <input type="number" value={newOption.days} onChange={e => setNewOption({...newOption, days: parseInt(e.target.value)})} placeholder="天数" className="flex-1 bg-white/10 border border-white/10 rounded-xl p-2.5 text-xs text-white outline-none" />
                            </div>
                            <button onClick={handleAddOption} className="w-full mt-2 py-3 bg-white text-indigo-700 rounded-xl text-xs font-black uppercase tracking-widest hover:bg-indigo-50 transition-all flex items-center justify-center gap-2">
                                <Plus className="w-4 h-4" /> 添加商品
                            </button>
                        </div>
                    </div>
                </div>
             </div>

             {/* Generator Panel */}
             <div className="bg-white dark:bg-zinc-900 rounded-[2.5rem] border border-paper-200 dark:border-zinc-800 shadow-sm overflow-hidden flex flex-col">
                <div className="p-6 border-b border-paper-100 dark:border-zinc-800 bg-paper-50/50 dark:bg-zinc-950/50">
                   <h3 className="text-sm font-black text-ink-900 dark:text-zinc-100 uppercase tracking-widest flex items-center gap-2">
                      <Zap className="w-4 h-4 text-brand-500" /> 批量生产
                   </h3>
                </div>
                <div className="p-8 space-y-6">
                   <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                         <label className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest">前缀</label>
                         <input value={genPrefix} onChange={e => setGenPrefix(e.target.value.toUpperCase())} className="w-full bg-paper-50 dark:bg-zinc-950 border border-paper-200 dark:border-zinc-800 rounded-xl p-3 text-sm font-mono font-bold uppercase outline-none focus:ring-2 focus:ring-brand-500 transition-all" placeholder="VIP_" />
                      </div>
                      <div className="space-y-2">
                         <label className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest">长度</label>
                         <input type="number" min="4" max="16" value={genLength} onChange={e => setGenLength(parseInt(e.target.value))} className="w-full bg-paper-50 dark:bg-zinc-950 border border-paper-200 dark:border-zinc-800 rounded-xl p-3 text-sm font-bold outline-none focus:ring-2 focus:ring-brand-500 transition-all" />
                      </div>
                   </div>
                   <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                         <label className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest">生成数量</label>
                         <input type="number" min="1" max="1000" value={genCount} onChange={e => setGenCount(parseInt(e.target.value))} className="w-full bg-paper-50 dark:bg-zinc-950 border border-paper-200 dark:border-zinc-800 rounded-xl p-3 text-sm font-bold outline-none focus:ring-2 focus:ring-brand-500 transition-all" />
                      </div>
                      <div className="space-y-2">
                         <label className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest">有效期 (天)</label>
                         <input type="number" min="1" value={genValidity} onChange={e => setGenValidity(parseInt(e.target.value))} className="w-full bg-paper-50 dark:bg-zinc-950 border border-paper-200 dark:border-zinc-800 rounded-xl p-3 text-sm font-bold outline-none focus:ring-2 focus:ring-brand-500 transition-all" />
                      </div>
                   </div>
                   <div className="space-y-2">
                      <label className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest">备注信息</label>
                      <textarea value={genNote} onChange={e => setGenNote(e.target.value)} className="w-full bg-paper-50 dark:bg-zinc-950 border border-paper-200 dark:border-zinc-800 rounded-xl p-3 text-sm font-bold outline-none focus:ring-2 focus:ring-brand-500 transition-all h-24 resize-none" placeholder="记录本次生成原因..." />
                   </div>
                   <button onClick={() => setShowPreview(true)} className="w-full py-4 bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-[1.5rem] font-black text-xs uppercase tracking-widest shadow-2xl hover:bg-black dark:hover:bg-white transition-all">预览并执行生成</button>
                </div>
             </div>
          </div>

          {/* Code Table */}
          <div className="col-span-12 xl:col-span-8 bg-white dark:bg-zinc-900 rounded-[2.5rem] border border-paper-200 dark:border-zinc-800 shadow-sm flex flex-col overflow-hidden">
             <div className="p-6 border-b border-paper-100 dark:border-zinc-800 bg-paper-50/50 dark:bg-zinc-950/50 flex items-center justify-between flex-wrap gap-4">
                <div className="flex items-center gap-2 bg-white dark:bg-zinc-800 border border-paper-200 dark:border-zinc-700 rounded-2xl px-4 py-2.5 flex-1 min-w-[300px] shadow-inner">
                   <Search className="w-4 h-4 text-ink-300" />
                   <input value={searchQuery} onChange={e => setSearchQuery(e.target.value)} className="bg-transparent border-none outline-none text-sm font-bold text-ink-900 dark:text-zinc-100 w-full" placeholder="全局搜索..." />
                </div>
                <div className="flex items-center gap-3">
                   <select value={filterStatus} onChange={e => setFilterStatus(e.target.value as any)} className="bg-white dark:bg-zinc-800 border border-paper-200 dark:border-zinc-700 rounded-xl px-4 py-2.5 text-xs font-black uppercase text-ink-600 dark:text-zinc-400 outline-none cursor-pointer shadow-sm">
                     <option value="all">所有兑换码</option>
                     <option value="active">正常有效</option>
                     <option value="expired">过期失效</option>
                     <option value="rewards">积分兑换</option>
                   </select>
                   <button onClick={handleExport} className="p-3 bg-white dark:bg-zinc-800 border border-paper-200 dark:border-zinc-700 rounded-xl text-ink-500 hover:text-brand-600 transition-all shadow-sm"><Download className="w-4 h-4" /></button>
                </div>
             </div>

             {selectedCodes.size > 0 && (
                <div className="px-6 py-3 bg-brand-500 text-white flex items-center justify-between animate-in slide-in-from-top-4">
                   <span className="text-xs font-black uppercase tracking-widest">已选中 {selectedCodes.size} 项</span>
                   <div className="flex gap-2">
                      <button onClick={() => handleBatchAction('disable')} className="px-4 py-1.5 bg-white/20 hover:bg-white/30 rounded-lg text-[10px] font-black uppercase tracking-widest">禁用</button>
                      <button onClick={() => handleBatchAction('renew')} className="px-4 py-1.5 bg-white/20 hover:bg-white/30 rounded-lg text-[10px] font-black uppercase tracking-widest">续期</button>
                      <button onClick={() => handleBatchAction('delete')} className="px-4 py-1.5 bg-rose-600 hover:bg-rose-700 rounded-lg text-[10px] font-black uppercase tracking-widest">彻底删除</button>
                   </div>
                </div>
             )}

             <div className="flex-1 overflow-auto custom-scrollbar">
                <table className="w-full text-left">
                   <thead className="bg-paper-50/50 dark:bg-zinc-950/50 sticky top-0 z-10 backdrop-blur-md">
                      <tr className="border-b border-paper-100 dark:border-zinc-800">
                         <th className="p-6 w-12"><input type="checkbox" onChange={handleSelectAll} checked={selectedCodes.size > 0 && selectedCodes.size === filteredCodes.length} className="rounded accent-brand-600" /></th>
                         <th className="p-6 text-[10px] font-black text-ink-400 uppercase tracking-widest">兑换码 / 备注</th>
                         <th className="p-6 text-[10px] font-black text-ink-400 uppercase tracking-widest">来源</th>
                         <th className="p-6 text-[10px] font-black text-ink-400 uppercase tracking-widest">状态</th>
                         <th className="p-6 text-[10px] font-black text-ink-400 uppercase tracking-widest">使用率</th>
                         <th className="p-6 text-[10px] font-black text-ink-400 uppercase tracking-widest">到期时间</th>
                         <th className="p-6 text-right text-[10px] font-black text-ink-400 uppercase tracking-widest">操作</th>
                      </tr>
                   </thead>
                   <tbody className="divide-y divide-paper-100 dark:divide-zinc-800">
                      {paginatedCodes.map(code => (
                         <tr key={code.code} className="group hover:bg-paper-50 dark:hover:bg-zinc-800/50 transition-colors">
                            <td className="p-6"><input type="checkbox" checked={selectedCodes.has(code.code)} onChange={(e) => { const next = new Set(selectedCodes); if(e.target.checked) next.add(code.code); else next.delete(code.code); setSelectedCodes(next); }} className="rounded accent-brand-600" /></td>
                            <td className="p-6">
                               <div className="flex items-center gap-2 mb-1">
                                  <span className="font-mono text-xs font-black text-ink-900 dark:text-zinc-200">{code.code}</span>
                                  <button onClick={() => { navigator.clipboard.writeText(code.code); alert('Copied'); }} className="opacity-0 group-hover:opacity-100 text-ink-300 hover:text-brand-500 transition-all"><Copy className="w-3.5 h-3.5" /></button>
                               </div>
                               {code.note && <p className="text-[10px] text-ink-400 dark:text-zinc-500 italic truncate max-w-[200px]">{code.note}</p>}
                            </td>
                            <td className="p-6">
                               <div className="flex items-center gap-1.5 px-2.5 py-1 bg-paper-100 dark:bg-zinc-800 rounded-full w-fit">
                                  {code.source === 'points_exchange' ? <Coins className="w-3 h-3 text-indigo-500" /> : <Shield className="w-3 h-3 text-gray-400" />}
                                  <span className="text-[9px] font-black uppercase text-ink-500 dark:text-zinc-400">{code.source === 'points_exchange' ? '积分兑换' : '系统生成'}</span>
                               </div>
                            </td>
                            <td className="p-6"><StatusBadge status={code.status} expiresAt={code.expiresAt} /></td>
                            <td className="p-6">
                               <div className="flex items-center gap-2">
                                  <div className="w-16 h-1.5 bg-paper-100 dark:bg-zinc-800 rounded-full overflow-hidden shrink-0"><div className="h-full bg-brand-500 transition-all" style={{ width: `${(code.usedCount / code.maxUses) * 100}%` }} /></div>
                                  <span className="text-[9px] font-black text-ink-400">{code.usedCount}/{code.maxUses}</span>
                               </div>
                            </td>
                            <td className="p-6 text-[10px] font-mono font-bold text-ink-500 dark:text-zinc-400">{new Date(code.expiresAt).toLocaleDateString()}</td>
                            <td className="p-6 text-right">
                               <div className="flex justify-end gap-1 opacity-0 group-hover:opacity-100 transition-all">
                                  <button onClick={() => batchUpdateCodes([code.code], code.status === 'active' ? 'disable' : 'enable')} className="p-2 hover:bg-paper-100 dark:hover:bg-zinc-800 rounded-xl text-ink-400" title={code.status === 'active' ? '禁用' : '启用'}>
                                     {code.status === 'active' ? <XCircle className="w-4 h-4" /> : <PlayCircle className="w-4 h-4" />}
                                  </button>
                                  <button onClick={() => batchUpdateCodes([code.code], 'delete')} className="p-2 hover:bg-rose-50 dark:hover:bg-rose-900/20 rounded-xl text-rose-400" title="彻底删除"><Trash2 className="w-4 h-4" /></button>
                               </div>
                            </td>
                         </tr>
                      ))}
                   </tbody>
                </table>
                {paginatedCodes.length === 0 && <div className="py-20 text-center"><p className="text-sm font-black text-ink-200 uppercase tracking-[0.2em]">No Records Found</p></div>}
             </div>

             <div className="p-6 border-t border-paper-100 dark:border-zinc-800 flex justify-between items-center bg-paper-50/30 dark:bg-zinc-950/30">
                <span className="text-[10px] font-black text-ink-300 uppercase tracking-widest">Showing {paginatedCodes.length} of {filteredCodes.length} entries</span>
                <div className="flex gap-2">
                   <button disabled={currentPage === 1} onClick={() => setCurrentPage(p => p - 1)} className="px-4 py-2 bg-white dark:bg-zinc-800 border border-paper-200 dark:border-zinc-700 rounded-xl text-[10px] font-black uppercase disabled:opacity-30">Previous</button>
                   <button disabled={currentPage >= totalPages} onClick={() => setCurrentPage(p => p + 1)} className="px-4 py-2 bg-white dark:bg-zinc-800 border border-paper-200 dark:border-zinc-700 rounded-xl text-[10px] font-black uppercase disabled:opacity-30">Next</button>
                </div>
             </div>
          </div>
       </div>

       {showPreview && (
          <div className="fixed inset-0 z-[1000] bg-black/60 backdrop-blur-md flex items-center justify-center p-6 animate-in fade-in duration-300">
             <div className="bg-white dark:bg-zinc-900 rounded-[3rem] shadow-2xl w-full max-w-md p-10 animate-in zoom-in-95 duration-500 border border-white/20">
                <h3 className="text-2xl font-black text-ink-900 dark:text-zinc-100 mb-6 font-serif italic tracking-tight">确认生成规格</h3>
                <div className="space-y-4 mb-10 bg-paper-50 dark:bg-zinc-950 p-6 rounded-[2rem] border border-paper-200 dark:border-zinc-800">
                   <div className="flex justify-between items-center"><span className="text-[10px] font-black text-ink-400 uppercase tracking-widest">单次生产量</span> <span className="text-sm font-black text-brand-600">{genCount} 个</span></div>
                   <div className="flex justify-between items-center"><span className="text-[10px] font-black text-ink-400 uppercase tracking-widest">规格前缀</span> <span className="text-sm font-bold">{genPrefix || '(无)'}</span></div>
                   <div className="flex justify-between items-center"><span className="text-[10px] font-black text-ink-400 uppercase tracking-widest">安全有效期</span> <span className="text-sm font-black">{genValidity} 天</span></div>
                   <div className="flex justify-between items-center"><span className="text-[10px] font-black text-ink-400 uppercase tracking-widest">最大利用率</span> <span className="text-sm font-black">{genMaxUses} 次/码</span></div>
                </div>
                <div className="flex gap-4">
                   <button onClick={() => setShowPreview(false)} className="flex-1 py-4 text-xs font-black text-ink-400 uppercase tracking-widest hover:text-ink-900 transition-colors">放弃</button>
                   <button onClick={handleGenerateConfirm} className="flex-[2] py-4 bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-2xl font-black text-xs uppercase tracking-widest shadow-xl hover:scale-95 transition-all">确认并执行生产</button>
                </div>
             </div>
          </div>
       )}
    </div>
  );

  // --- Main Render ---

  if (!user) return null;

  return (
    <div className="flex-1 h-full bg-paper-50 dark:bg-zinc-950 transition-colors flex overflow-hidden">
      {/* Dynamic Sidebar */}
      <div className="w-80 border-r border-paper-200 dark:border-zinc-800 bg-white/60 dark:bg-zinc-900/60 backdrop-blur-3xl p-8 flex flex-col pt-24 shrink-0 shadow-2xl z-10 transition-colors">
        
        <div className="mb-12 group">
          <button 
            onClick={navigateBack}
            className="w-full flex items-center gap-4 px-6 py-4 bg-white dark:bg-zinc-100 text-ink-900 dark:text-zinc-900 rounded-3xl shadow-[0_15px_35px_rgba(0,0,0,0.08)] dark:shadow-none transition-all hover:translate-x-1 active:scale-95 border border-paper-100 dark:border-zinc-200 group"
          >
            <div className="p-2 bg-ink-900 dark:bg-zinc-900 rounded-xl text-white dark:text-zinc-100 group-hover:-translate-x-1 transition-transform">
                <ChevronLeft className="w-4 h-4" />
            </div>
            <div className="flex flex-col items-start leading-tight">
               <span className="text-[12px] font-black uppercase tracking-widest mb-0.5">
                 {previousViewMode === 'DASHBOARD' ? '仪表盘' : '创作空间'}
               </span>
               <span className="text-[9px] font-bold opacity-40 uppercase truncate max-w-[140px]">
                 {previousViewMode === 'DASHBOARD' ? 'Dashboard Home' : (project?.title || 'Back to Manuscript')}
               </span>
            </div>
          </button>
        </div>

        <div className="space-y-1">
           <h3 className="text-[10px] font-black text-ink-300 dark:text-zinc-600 uppercase tracking-[0.4em] mb-6 px-5">System Preferences</h3>
           <nav className="space-y-1.5">
             {navItem('profile', '账户概览', <UserCircle className="w-4 h-4" />)}
             {navItem('ai-engine', 'AI 创作引擎', <Cpu className="w-4 h-4" />)}
             {navItem('appearance', '感官与外观', <Palette className="w-4 h-4" />)}
             {navItem('data', '数据与实验室', <Database className="w-4 h-4" />)}
             {user.role === 'admin' && (
                 <div className="pt-4 mt-4 border-t border-paper-100 dark:border-zinc-800">
                    <h3 className="text-[10px] font-black text-rose-500 uppercase tracking-[0.4em] mb-4 px-5">Root Access</h3>
                    {navItem('admin', '超级控制台', <Shield className="w-4 h-4" />)}
                 </div>
             )}
           </nav>
        </div>
        
        <div className="mt-auto pt-8 border-t border-paper-100 dark:border-zinc-800">
           <button onClick={logout} className="w-full flex items-center justify-between px-6 py-4 text-rose-500 hover:bg-rose-50 dark:hover:bg-rose-900/20 rounded-2xl text-[11px] font-black uppercase tracking-widest transition-all group">
             <div className="flex items-center gap-3">
                <LogOut className="w-4 h-4 group-hover:rotate-180 transition-transform duration-500" /> 退出系统
             </div>
             <ChevronRight className="w-3 h-3 opacity-30" />
           </button>
        </div>
      </div>

      {/* Optimized Content Area */}
      <div className="flex-1 overflow-y-auto pt-24 pb-20 px-10 md:px-16 lg:px-24 custom-scrollbar bg-dot-pattern transition-colors">
        <div className="max-w-[1400px] mx-auto min-h-full">
          
          {activeTab === 'profile' && (
            <div className="grid grid-cols-12 gap-8 animate-in fade-in slide-in-from-bottom-6 duration-700">
               <div className="col-span-12 lg:col-span-8 space-y-8">
                  <section className="bg-white dark:bg-zinc-900 p-10 rounded-[3.5rem] border border-paper-200 dark:border-zinc-800 shadow-xl shadow-paper-100 dark:shadow-none flex items-center gap-10 group overflow-hidden relative">
                    <div className="absolute -right-20 -bottom-20 w-80 h-80 bg-paper-50 dark:bg-zinc-800 rounded-full blur-3xl opacity-50 group-hover:scale-110 transition-transform duration-1000"></div>
                    <div className="relative z-10 w-32 h-32 bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-[2.5rem] flex items-center justify-center text-5xl font-black shadow-2xl transform group-hover:rotate-3 transition-transform">
                      {user.username.charAt(0).toUpperCase()}
                    </div>
                    <div className="relative z-10 space-y-3">
                      <div className="flex items-center gap-3">
                         <h2 className="text-4xl font-black text-ink-900 dark:text-white font-serif tracking-tight">{user.username}</h2>
                         <span className="px-3 py-1 bg-brand-500 text-white rounded-full text-[9px] font-black uppercase tracking-widest shadow-md">LV.1 创造者</span>
                      </div>
                      <div className="flex items-center gap-5">
                        <span className="px-3 py-1 bg-paper-100 dark:bg-zinc-800 text-ink-600 dark:text-zinc-400 rounded-xl text-[10px] font-black uppercase tracking-widest border border-paper-200 dark:border-zinc-700">{user.role} Account</span>
                        <span className="text-[11px] text-ink-400 font-bold flex items-center gap-2"><Calendar className="w-4 h-4 text-brand-500" /> 启程于 {new Date(user.createdAt).toLocaleDateString()}</span>
                      </div>
                    </div>
                  </section>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                     <div className="bg-white dark:bg-zinc-900 p-10 rounded-[3rem] border border-paper-200 dark:border-zinc-800 shadow-sm flex flex-col justify-between group transition-all hover:shadow-2xl">
                        <h4 className="text-[11px] font-black text-ink-300 dark:text-zinc-600 uppercase mb-8 tracking-[0.3em] flex items-center gap-3">
                          <ShieldCheck className="w-4 h-4 text-emerald-500" /> Security Status
                        </h4>
                        {hasAIAccess ? (
                          <div className="flex items-center gap-4 text-emerald-600 bg-emerald-50 dark:bg-emerald-900/20 p-5 rounded-[1.5rem] border border-emerald-100 dark:border-emerald-800/40">
                            <div className="w-10 h-10 bg-white dark:bg-zinc-900 rounded-full flex items-center justify-center shadow-sm"><Check className="w-5 h-5" /></div>
                            <div className="flex flex-col"><span className="font-black text-sm uppercase">AI Access Active</span><span className="text-[10px] font-bold opacity-60">灵感引擎已就绪，全功能开放</span></div>
                          </div>
                        ) : (
                          <div className="flex items-center gap-4 text-rose-600 bg-rose-50 dark:bg-rose-900/20 p-5 rounded-[1.5rem] border border-rose-100 dark:border-rose-800/40">
                            <div className="w-10 h-10 bg-white dark:bg-zinc-900 rounded-full flex items-center justify-center shadow-sm"><AlertCircle className="w-5 h-5" /></div>
                            <div className="flex flex-col"><span className="font-black text-sm uppercase">Access Limited</span><span className="text-[10px] font-bold opacity-60">订阅已失效，核心功能受限</span></div>
                          </div>
                        )}
                     </div>
                     <div className="bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 p-10 rounded-[3rem] shadow-2xl flex flex-col justify-between group overflow-hidden relative">
                        <div className="absolute -right-10 -bottom-10 w-40 h-40 bg-white/5 dark:bg-black/5 rounded-full blur-3xl group-hover:scale-150 transition-transform duration-700"></div>
                        <h4 className="text-[11px] font-black uppercase mb-8 tracking-[0.3em] flex items-center gap-3 opacity-50 relative z-10">
                           <Coins className="w-4 h-4 text-amber-400" /> Credit Balance
                        </h4>
                        <div className="flex items-end gap-3 relative z-10">
                            <span className="text-6xl font-black font-serif tabular-nums tracking-tighter">{user.points || 0}</span>
                            <span className="text-xs font-black uppercase tracking-widest mb-2.5 opacity-40">Points</span>
                        </div>
                     </div>
                  </div>
               </div>

               <div className="col-span-12 lg:col-span-4 space-y-8">
                  <div className="bg-white dark:bg-zinc-900 p-8 rounded-[3rem] border border-paper-200 dark:border-zinc-800 shadow-sm flex flex-col h-full">
                     <h4 className="text-[11px] font-black text-ink-300 dark:text-zinc-600 uppercase mb-8 tracking-[0.3em] flex items-center gap-3"><Activity className="w-4 h-4 text-brand-500" /> Device Meta</h4>
                     <div className="space-y-6">
                        <div className="flex flex-col gap-1.5"><span className="text-[10px] font-black text-ink-400 uppercase tracking-widest">Unique Device Identifier</span><p className="font-mono text-xs font-bold bg-paper-50 dark:bg-zinc-950 p-4 rounded-2xl break-all border border-paper-100 dark:border-zinc-800 shadow-inner">{deviceId}</p></div>
                        <div className="flex flex-col gap-1.5"><span className="text-[10px] font-black text-ink-400 uppercase tracking-widest">Auth Certificate</span><div className="flex items-center gap-2 text-xs font-bold text-emerald-500"><ShieldCheck className="w-3.5 h-3.5" /> Hardware-Linked Signature</div></div>
                     </div>
                  </div>
               </div>
            </div>
          )}

          {activeTab === 'ai-engine' && (
            <div className="grid grid-cols-12 gap-8 animate-in fade-in slide-in-from-bottom-6 duration-700">
               <div className="col-span-12 lg:col-span-4 space-y-8">
                  <div className="bg-white dark:bg-zinc-900 p-8 rounded-[3rem] border border-paper-200 dark:border-zinc-800 shadow-sm space-y-6">
                     <h3 className="text-xl font-black text-ink-900 dark:text-white font-serif italic">供应商选择</h3>
                     <div className="grid grid-cols-1 gap-3">
                         {providers.map((p) => (
                            <button
                               key={p.id}
                               onClick={() => {
                                 project ? updateAISettings({ provider: p.id }) : updateDefaultAISettings({ provider: p.id });
                                 if (user?.role === 'admin') {
                                   loadProviderConfig(p.id);
                                 }
                               }}
                               className={`flex items-center gap-4 p-5 rounded-2xl border-2 transition-all ${
                                  currentSettings.provider === p.id 
                                  ? `${p.activeBg} border-brand-500 shadow-lg` 
                                  : 'border-paper-50 dark:border-zinc-800 hover:border-paper-100 dark:hover:border-zinc-700'
                               }`}
                            >
                               <div className={`p-3 rounded-xl text-white ${p.color} shadow-md`}>{p.icon}</div>
                               <div className="text-left flex flex-col leading-tight"><span className="text-[11px] font-black uppercase tracking-widest">{p.name}</span><span className="text-[9px] font-bold text-ink-300 uppercase mt-0.5">Runtime Node</span></div>
                            </button>
                         ))}
                     </div>
                  </div>

                  <div className="bg-amber-50 dark:bg-amber-900/10 p-8 rounded-[2.5rem] border border-amber-100 dark:border-amber-900/20 space-y-3">
                     <div className="flex items-center gap-3 text-amber-600"><Info className="w-5 h-5" /><span className="text-xs font-black uppercase tracking-widest">Configuration Note</span></div>
                     <p className="text-[11px] font-medium text-amber-700 dark:text-amber-400 leading-relaxed">系统优先读取“项目级设置”。如果您正在写作中，这里的改动会立即影响当前手稿的生成效果。</p>
                  </div>
               </div>

               <div className="col-span-12 lg:col-span-8 bg-white dark:bg-zinc-900 p-12 rounded-[3.5rem] border border-paper-200 dark:border-zinc-800 shadow-xl space-y-12">
                  <div className="flex justify-between items-center border-b border-paper-50 dark:border-zinc-800 pb-8">
                     <div className="space-y-1">
                        <h3 className="text-2xl font-black text-ink-900 dark:text-white font-serif italic flex items-center gap-3"><BrainCircuit className="w-7 h-7 text-brand-600" /> 引擎参数微调</h3>
                        <p className="text-[10px] text-ink-400 dark:text-zinc-600 font-black uppercase tracking-widest">{project ? `Current Project: ${project.title}` : 'System Default Mode'}</p>
                     </div>
                     <button onClick={handleRefreshModels} disabled={isRefreshing} className="flex items-center gap-2 px-5 py-2.5 bg-brand-50 dark:bg-brand-900/30 text-brand-600 dark:text-brand-400 rounded-2xl text-[10px] font-black uppercase tracking-widest hover:bg-brand-500 hover:text-white transition-all">
                        <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} /> 同步模型库
                     </button>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-10">
                     <div className="space-y-8 md:col-span-2">
                         {(currentSettings.provider === 'proxy' || currentSettings.provider === 'openai' || currentSettings.provider === 'openrouter' || currentSettings.provider === 'anthropic' || currentSettings.provider === 'local') && (
                            <div className="space-y-2 animate-in slide-in-from-top-4">
                               <label className="text-[10px] font-black text-ink-400 uppercase tracking-widest ml-2">Endpoint Proxy Address</label>
                               <div className="relative group">
                                  <Globe className="absolute left-5 top-1/2 -translate-y-1/2 w-4 h-4 text-ink-300 group-focus-within:text-brand-500 transition-colors" />
                                  <input value={currentSettings.proxyEndpoint || ''} onChange={(e) => project ? updateAISettings({ proxyEndpoint: e.target.value }) : updateDefaultAISettings({ proxyEndpoint: e.target.value })} className="w-full pl-14 pr-6 py-5 bg-paper-50 dark:bg-zinc-950 rounded-[1.5rem] border-none text-sm font-mono font-bold text-ink-900 dark:text-zinc-100 shadow-inner focus:ring-2 focus:ring-brand-100 transition-all" placeholder="https://api.example.com/v1" />
                               </div>
                            </div>
                         )}
                         {user?.role === 'admin' && (
                           <div className="space-y-3 bg-paper-50 dark:bg-zinc-950/50 p-6 rounded-[1.75rem] border border-paper-100 dark:border-zinc-800">
                              <div className="flex items-center justify-between">
                                <div className="text-[11px] font-black uppercase tracking-widest text-ink-500">供应商配置（管理员）</div>
                                <div className="flex items-center gap-2">
                                  <button onClick={testProviderConfig} disabled={isTestingProvider} className="px-3 py-2 bg-paper-100 dark:bg-zinc-800 text-ink-700 dark:text-zinc-200 rounded-xl text-[10px] font-black uppercase tracking-widest disabled:opacity-50">
                                    {isTestingProvider ? '测试中' : '测试连接'}
                                  </button>
                                  <button onClick={saveProviderConfig} disabled={isSavingProvider} className="px-4 py-2 bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-xl text-[10px] font-black uppercase tracking-widest disabled:opacity-50">
                                    {isSavingProvider ? '保存中' : '保存'}
                                  </button>
                                </div>
                              </div>
                             <div className="grid grid-cols-1 gap-3">
                               <input
                                 value={providerConfig.baseUrl}
                                 onChange={(e) => setProviderConfig((prev) => ({ ...prev, baseUrl: e.target.value }))}
                                 className="w-full px-4 py-3 bg-white dark:bg-zinc-900 rounded-xl border border-paper-200 dark:border-zinc-800 text-xs font-mono"
                                 placeholder="Base URL (例如 https://api.openai.com/v1)"
                               />
                               <input
                                 value={providerConfig.apiKey}
                                 onChange={(e) => setProviderConfig((prev) => ({ ...prev, apiKey: e.target.value }))}
                                 className="w-full px-4 py-3 bg-white dark:bg-zinc-900 rounded-xl border border-paper-200 dark:border-zinc-800 text-xs font-mono"
                                 placeholder="API Key（保存后将脱敏显示）"
                               />
                             </div>
                           </div>
                         )}
                        <div className="space-y-4">
                           <div className="flex justify-between items-center ml-2 mr-2">
                              <label className="text-[10px] font-black text-ink-400 uppercase tracking-widest">Active Model Brain</label>
                              {currentSettings.provider !== 'gemini' && <button onClick={() => setIsManualInput(!isManualInput)} className="text-[9px] font-black text-ink-300 uppercase hover:text-brand-500 transition-colors">{isManualInput ? 'Toggle Selector' : 'Manual Entry'}</button>}
                           </div>
                           {isManualInput || (availableModels.length === 0 && currentSettings.provider !== 'gemini') ? (
                              <input value={currentSettings.model} onChange={(e) => project ? updateAISettings({ model: e.target.value }) : updateDefaultAISettings({ model: e.target.value })} className="w-full p-5 bg-paper-50 dark:bg-zinc-950 rounded-[1.5rem] border-none text-sm font-black text-ink-900 dark:text-zinc-100 shadow-inner focus:ring-2 focus:ring-brand-100 outline-none" placeholder="e.g. gpt-4o-mini" />
                           ) : (
                              <div className="relative">
                                 <select value={currentSettings.model} onChange={(e) => project ? updateAISettings({ model: e.target.value }) : updateDefaultAISettings({ model: e.target.value })} className="w-full p-5 bg-paper-50 dark:bg-zinc-950 rounded-[1.5rem] border-none text-sm font-black appearance-none text-ink-900 dark:text-zinc-100 shadow-inner cursor-pointer focus:ring-2 focus:ring-brand-100 outline-none transition-all">
                                    {availableModels.map(m => <option key={m.id} value={m.id}>{m.name}</option>)}
                                 </select>
                                 <div className="absolute right-6 top-1/2 -translate-y-1/2 pointer-events-none text-ink-200"><ChevronRight className="w-5 h-5 rotate-90" /></div>
                              </div>
                           )}
                        </div>
                     </div>

                     <div className="bg-paper-50 dark:bg-zinc-950/50 p-8 rounded-[2.5rem] border border-paper-100 dark:border-zinc-800 space-y-6">
                        <div className="flex justify-between items-center"><label className="text-[10px] font-black text-ink-400 uppercase tracking-widest">Temperature: {currentSettings.temperature}</label><Sliders className="w-4 h-4 text-ink-200" /></div>
                        <input type="range" min="0" max="1" step="0.1" className="w-full h-1.5 bg-paper-200 dark:bg-zinc-800 rounded-lg appearance-none cursor-pointer accent-brand-600" value={currentSettings.temperature} onChange={(e) => { const val = parseFloat(e.target.value); project ? updateAISettings({ temperature: val }) : updateDefaultAISettings({ temperature: val }); }} />
                        <p className="text-[9px] font-medium text-ink-300 leading-relaxed uppercase tracking-tight">Higher temperature increases random variety and creative leaps. Lower values yield stable, deterministic logic.</p>
                     </div>

                     {supportsThinking && (
                        <div className="bg-brand-50/50 dark:bg-brand-900/10 p-8 rounded-[2.5rem] border border-brand-100 dark:border-brand-900/20 space-y-6">
                           <div className="flex justify-between items-center"><label className="text-[10px] font-black text-brand-600 uppercase tracking-widest">Thinking Budget: {currentSettings.thinkingBudget || 0} Tokens</label><BrainCircuit className="w-4 h-4 text-brand-500" /></div>
                           <input type="range" min="0" max="32768" step="1024" className="w-full h-1.5 bg-brand-100 dark:bg-brand-800 rounded-lg appearance-none cursor-pointer accent-brand-500" value={currentSettings.thinkingBudget || 0} onChange={(e) => { const val = parseInt(e.target.value); project ? updateAISettings({ thinkingBudget: val }) : updateDefaultAISettings({ thinkingBudget: val }); }} />
                           <p className="text-[9px] font-medium text-brand-600/70 leading-relaxed uppercase tracking-tight">Allows Gemini to reason internally before output. Essential for complex narrative logic and plot-hole prevention.</p>
                        </div>
                     )}
                  </div>
               </div>
            </div>
          )}

          {activeTab === 'appearance' && (
            <div className="max-w-4xl space-y-8 animate-in fade-in slide-in-from-bottom-6 duration-700">
               <section className="bg-white dark:bg-zinc-900 p-12 rounded-[3.5rem] border border-paper-200 dark:border-zinc-800 shadow-xl space-y-10">
                  <div className="space-y-1">
                     <h3 className="text-2xl font-black text-ink-900 dark:text-white font-serif italic flex items-center gap-3"><Palette className="w-7 h-7 text-indigo-500" /> 系统感官体验</h3>
                     <p className="text-[10px] text-ink-400 uppercase tracking-[0.2em]">Visual & Sensory Aesthetics</p>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                     <div className="bg-paper-50 dark:bg-zinc-950 p-10 rounded-[3rem] border border-paper-100 dark:border-zinc-800 shadow-inner flex flex-col justify-between group h-64 transition-all hover:bg-white dark:hover:bg-zinc-900">
                        <div className="space-y-2">
                           <p className="text-sm font-black uppercase tracking-widest text-ink-900 dark:text-white">Day / Night Cycle</p>
                           <p className="text-[10px] text-ink-300 dark:text-zinc-600 font-bold uppercase tracking-widest">白昼与深夜模式切换</p>
                        </div>
                        <button onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')} className={`flex items-center justify-center gap-4 px-10 py-5 rounded-[1.5rem] text-[11px] font-black uppercase tracking-widest shadow-2xl transition-all active:scale-95 group overflow-hidden relative ${theme === 'light' ? 'bg-ink-900 text-white' : 'bg-white text-ink-900'}`}>
                           <div className="absolute inset-0 bg-white/10 opacity-0 group-hover:opacity-100 transition-opacity"></div>
                           {theme === 'light' ? <Activity className="w-4 h-4 text-indigo-400 group-hover:rotate-45 transition-transform" /> : <Laptop className="w-4 h-4 text-brand-600 group-hover:-rotate-12 transition-transform" />}
                           {theme === 'light' ? 'Switch to Midnight' : 'Switch to Daylight'}
                        </button>
                     </div>
                     <div className="bg-paper-50 dark:bg-zinc-950 p-10 rounded-[3rem] border border-paper-100 dark:border-zinc-800 shadow-inner flex flex-col justify-center items-center h-64 opacity-30 grayscale cursor-not-allowed">
                        <Lock className="w-10 h-10 mb-4 text-ink-200" />
                        <p className="text-[11px] font-black uppercase tracking-widest">Custom Themes coming soon</p>
                     </div>
                  </div>
               </section>
            </div>
          )}

          {activeTab === 'data' && (
            <div className="max-w-6xl space-y-8 animate-in fade-in slide-in-from-bottom-6 duration-700">
               <div className="bg-white dark:bg-zinc-900 p-12 rounded-[3.5rem] border border-paper-200 dark:border-zinc-800 shadow-xl space-y-12">
                  <div className="space-y-1">
                     <h3 className="text-2xl font-black text-ink-900 dark:text-white font-serif italic flex items-center gap-3"><Database className="w-7 h-7 text-emerald-500" /> 数据管护与备份</h3>
                     <p className="text-[10px] text-ink-400 uppercase tracking-[0.2em]">Safe Storage & Laboratory</p>
                  </div>
                  <div className="flex items-center justify-between bg-paper-50 dark:bg-zinc-950 p-6 rounded-[2.5rem] border border-paper-100 dark:border-zinc-800">
                    <div className="space-y-2">
                      <p className="text-sm font-black uppercase text-ink-900 dark:text-white">Cloud Sync</p>
                      <p className="text-[10px] text-ink-300 dark:text-zinc-600 font-bold uppercase tracking-widest">开启后允许同步与云端备份（默认仅本地）</p>
                    </div>
                    <button
                      onClick={() => setCloudSyncEnabled(!cloudSyncEnabled)}
                      className={`px-5 py-3 rounded-2xl text-[10px] font-black uppercase tracking-widest transition-all ${
                        cloudSyncEnabled
                          ? 'bg-emerald-500 text-white shadow-lg'
                          : 'bg-white dark:bg-zinc-900 border border-paper-200 dark:border-zinc-800 text-ink-500'
                      }`}
                    >
                      {cloudSyncEnabled ? '已开启' : '未开启'}
                    </button>
                  </div>
                  {showCloudSyncTip && (
                    <div className="flex items-start gap-3 p-4 bg-amber-50 dark:bg-amber-900/10 rounded-2xl border border-amber-100 dark:border-amber-900/20">
                      <Info className="w-4 h-4 text-amber-500 mt-0.5" />
                      <div className="text-[10px] text-amber-700 dark:text-amber-400 font-bold leading-relaxed">
                        当前为仅本地模式。开启云端同步后，项目快照会上传用于多端恢复。
                      </div>
                      <button onClick={() => setShowCloudSyncTip(false)} className="text-[9px] font-black uppercase tracking-widest text-amber-500">关闭</button>
                    </div>
                  )}
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                     <div className="p-8 bg-paper-50 dark:bg-zinc-950 rounded-[2.5rem] border border-paper-100 dark:border-zinc-800 shadow-inner space-y-6 flex flex-col justify-between">
                        <div className="space-y-2"><p className="text-sm font-black uppercase text-ink-900 dark:text-white">Full Backup</p><p className="text-[10px] text-ink-300 dark:text-zinc-600 font-bold uppercase tracking-widest">全量数据 JSON 导出</p></div>
                        <button onClick={handleExportAllData} className="w-full py-4 bg-white dark:bg-zinc-900 border border-paper-200 dark:border-zinc-800 rounded-2xl text-[10px] font-black uppercase tracking-widest shadow-lg flex items-center justify-center gap-3 hover:translate-y-[-2px] transition-all"><Download className="w-4 h-4" /> Export All Data</button>
                     </div>
                     <div className="p-8 bg-emerald-50/40 dark:bg-emerald-900/10 rounded-[2.5rem] border-2 border-emerald-100 dark:border-emerald-900/30 shadow-inner space-y-6 flex flex-col justify-between">
                        <div className="space-y-2"><p className="text-sm font-black uppercase text-emerald-700 dark:text-emerald-300">Local Backup</p><p className="text-[10px] text-emerald-500/70 font-bold uppercase tracking-widest">写入本地文件作为保底备份</p></div>
                        <button onClick={backupProject} className="w-full py-4 bg-white dark:bg-zinc-900 border-2 border-emerald-100 dark:border-emerald-900/40 rounded-2xl text-[10px] font-black uppercase tracking-widest shadow-lg text-emerald-600 hover:bg-emerald-500 hover:text-white transition-all" disabled={!cloudSyncEnabled}><Download className="w-4 h-4" /> Backup to Local Storage</button>
                        <button
                          onClick={async () => {
                            const ok = await confirm({
                              title: '确认恢复备份',
                              description: '恢复后将覆盖本地当前项目内容，是否继续？',
                              confirmText: '确认恢复',
                              cancelText: '取消',
                              tone: 'danger'
                            });
                            if (ok) await restoreLatestBackup();
                          }}
                          className="w-full py-3 bg-white/80 dark:bg-zinc-900 border border-emerald-100 dark:border-emerald-900/40 rounded-2xl text-[10px] font-black uppercase tracking-widest text-emerald-700 hover:bg-emerald-500 hover:text-white transition-all"
                          disabled={!cloudSyncEnabled}
                        >
                          恢复最新备份
                        </button>
                     </div>
                     <div className="p-8 bg-white dark:bg-zinc-900 rounded-[2.5rem] border border-paper-200 dark:border-zinc-800 shadow-inner space-y-6 flex flex-col">
                        <div className="flex items-center justify-between">
                           <div className="space-y-2">
                             <p className="text-sm font-black uppercase text-ink-900 dark:text-white">Backup History</p>
                             <p className="text-[10px] text-ink-300 dark:text-zinc-600 font-bold uppercase tracking-widest">最近备份记录</p>
                           </div>
                           <button onClick={loadBackups} className="px-4 py-2 bg-paper-100 dark:bg-zinc-800 rounded-xl text-[9px] font-black uppercase tracking-widest">刷新</button>
                         </div>
                         <div className="space-y-3 text-[11px]">
                           {backupLoading && <div className="text-ink-400">加载中...</div>}
                           {!backupLoading && backupList.length === 0 && <div className="text-ink-400">暂无备份</div>}
                           {!backupLoading && backupList.map((item) => (
                             <div key={item.id} className="flex items-center justify-between bg-paper-50 dark:bg-zinc-950 rounded-xl px-4 py-3">
                               <div className="flex flex-col">
                                 <span className="font-black text-ink-800 dark:text-zinc-100 truncate max-w-[220px]">{item.file_name}</span>
                                 <span className="text-[9px] text-ink-400 uppercase tracking-widest">{new Date(item.created_at).toLocaleString()}</span>
                               </div>
                               <a href={getBackupDownloadUrl(item.id)} className="text-[10px] font-black uppercase tracking-widest text-brand-600">下载</a>
                             </div>
                           ))}
                         </div>
                        <div className="text-[9px] text-ink-300 uppercase tracking-widest">总计 {backupTotal} 条</div>
                     </div>
                      <div className="p-8 bg-paper-50 dark:bg-zinc-950 rounded-[2.5rem] border border-paper-100 dark:border-zinc-800 shadow-inner space-y-6 flex flex-col justify-between">
                         <div className="space-y-2"><p className="text-sm font-black uppercase text-ink-900 dark:text-white">Project Clean</p><p className="text-[10px] text-ink-300 dark:text-zinc-600 font-bold uppercase tracking-widest">清理本地未使用的资源</p></div>
                         <button className="w-full py-4 bg-white dark:bg-zinc-900 border border-paper-200 dark:border-zinc-800 rounded-2xl text-[10px] font-black uppercase tracking-widest shadow-lg opacity-40 cursor-not-allowed">Run Cleanup Tool</button>
                      </div>
                      <div className="p-8 bg-rose-50/30 dark:bg-rose-950/20 rounded-[2.5rem] border-2 border-rose-100 dark:border-rose-900/30 shadow-inner space-y-6 flex flex-col justify-between">
                         <div className="space-y-2"><p className="text-sm font-black uppercase text-rose-700 dark:text-rose-400">Nuclear Option</p><p className="text-[10px] text-rose-400/50 font-bold uppercase tracking-widest">彻底清除所有本地缓存</p></div>
                         <button onClick={handleClearCache} className="w-full py-4 bg-white dark:bg-zinc-900 border-2 border-rose-100 dark:border-rose-900/40 rounded-2xl text-[10px] font-black uppercase tracking-widest shadow-lg text-rose-500 hover:bg-rose-500 hover:text-white transition-all"><RefreshCw className="w-4 h-4" /> Reset Application</button>
                      </div>
                  </div>
               </div>
            </div>
          )}

          {activeTab === 'admin' && user.role === 'admin' && renderAdminPanel()}

        </div>
      </div>
    </div>
  );
};
