
import React, { useState, useEffect } from 'react';
import { useProject } from '../contexts/ProjectContext';
import { X, Sparkles, Cpu, Globe, Info, RefreshCw, Check, Loader2, Trash2, Edit2, Keyboard, List, BrainCircuit, Hash, HardDrive, Laptop } from 'lucide-react';
import { AIProvider } from '../types';

interface AISettingsModalProps {
  onClose: () => void;
}

export const AISettingsModal: React.FC<AISettingsModalProps> = ({ onClose }) => {
  const { project, updateAISettings, availableModels, refreshModels, clearCache } = useProject();
  const [localSettings, setLocalSettings] = useState(project.aiSettings);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isManualInput, setIsManualInput] = useState(false);

  const providers: { id: AIProvider; name: string; icon: React.ReactNode; color: string; activeBg: string }[] = [
    { id: 'gemini', name: 'Google Gemini (原生)', icon: <Sparkles className="w-5 h-5" />, color: 'bg-blue-500', activeBg: 'bg-blue-50 border-blue-200' },
    { id: 'local', name: 'Local AI (本地模型)', icon: <HardDrive className="w-5 h-5" />, color: 'bg-amber-500', activeBg: 'bg-amber-50 border-amber-200' },
    { id: 'openai', name: 'OpenAI (官方)', icon: <Cpu className="w-5 h-5" />, color: 'bg-emerald-500', activeBg: 'bg-emerald-50 border-emerald-200' },
    { id: 'proxy', name: '第三方代理 (NewAPI)', icon: <Globe className="w-5 h-5" />, color: 'bg-indigo-500', activeBg: 'bg-indigo-50 border-indigo-200' },
  ];

  const currentProviderConfig = providers.find(p => p.id === localSettings.provider) || providers[0];

  useEffect(() => {
    if (localSettings.provider !== 'gemini' && (!availableModels || availableModels.length === 0)) {
        // Don't auto-refresh for local to prevent connection errors on load if local server isn't running
        if (localSettings.provider !== 'local') {
            handleRefresh();
        }
    }
  }, [localSettings.provider]);

  useEffect(() => {
    setLocalSettings(project.aiSettings);
  }, [project.aiSettings]);

  const handleRefresh = async () => {
    setIsRefreshing(true);
    updateAISettings({ 
      provider: localSettings.provider, 
      proxyEndpoint: localSettings.proxyEndpoint 
    });
    await refreshModels(); 
    setIsRefreshing(false);
  };

  const handleModelChange = (model: string) => {
    const updates = { ...localSettings, model };
    setLocalSettings(updates);
    updateAISettings({ model });
  };

  const handleProviderSelect = (id: AIProvider) => {
      let defaultEndpoint = localSettings.proxyEndpoint;
      if (id === 'local' && (!defaultEndpoint || !defaultEndpoint.includes('localhost'))) {
          defaultEndpoint = 'http://localhost:11434/v1';
      }
      
      const updates = { ...localSettings, provider: id, proxyEndpoint: defaultEndpoint };
      setLocalSettings(updates);
      if (id === 'gemini') setIsManualInput(false);
  };

  const handleSave = () => {
    updateAISettings(localSettings);
    onClose();
  };

  const isGeminiModel = localSettings.provider === 'gemini' || localSettings.model.includes('gemini');
  const supportsThinking = isGeminiModel && (localSettings.model.includes('3') || localSettings.model.includes('2.5'));

  return (
    <div className="fixed inset-0 z-[200] bg-gray-900/40 backdrop-blur-md flex items-center justify-center p-6 animate-in fade-in duration-300">
      <div className="bg-white dark:bg-zinc-900 rounded-[2.5rem] shadow-2xl w-full max-w-2xl overflow-hidden flex flex-col border border-white/20">
        <div className="p-8 border-b border-gray-100 dark:border-zinc-800 flex items-center justify-between bg-gray-50/50 dark:bg-zinc-950">
           <div className="flex items-center gap-4">
              <div className={`p-3 rounded-2xl shadow-lg text-white ${currentProviderConfig.color} transition-all duration-500`}>
                {currentProviderConfig.icon}
              </div>
              <div>
                <h2 className="text-xl font-black text-gray-900 dark:text-zinc-100 tracking-tight">AI 创作引擎设置</h2>
                <p className="text-[10px] text-indigo-600 font-black uppercase tracking-widest mt-0.5">多供应商接入 & 模型管理</p>
              </div>
           </div>
           <button onClick={onClose} className="p-2 hover:bg-white rounded-xl text-gray-400"><X className="w-6 h-6" /></button>
        </div>

        <div className="p-8 space-y-8 flex-1 overflow-y-auto max-h-[70vh] custom-scrollbar">
          <section>
            <label className="text-[10px] font-black text-gray-400 uppercase tracking-widest mb-4 block">选择 AI 供应商</label>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
              {providers.map((p) => (
                <button
                  key={p.id}
                  onClick={() => handleProviderSelect(p.id)}
                  className={`flex flex-col items-center justify-center p-4 rounded-2xl border-2 transition-all gap-2 ${
                    localSettings.provider === p.id 
                      ? `${p.activeBg} border-brand-500 shadow-sm` 
                      : 'border-gray-100 dark:border-zinc-800 hover:border-gray-200 bg-white dark:bg-zinc-950'
                  }`}
                >
                  <div className={`p-2 rounded-xl text-white ${p.color}`}>{p.icon}</div>
                  <span className={`text-[10px] font-bold text-center ${localSettings.provider === p.id ? 'text-brand-900 dark:text-brand-400' : 'text-gray-600'}`}>{p.name}</span>
                </button>
              ))}
            </div>
          </section>

          <section className="space-y-4">
            {(localSettings.provider === 'proxy' || localSettings.provider === 'openai' || localSettings.provider === 'local') && (
              <div className="animate-in slide-in-from-top-2">
                <label className="text-[10px] font-black text-gray-400 uppercase tracking-widest mb-2 block">
                    {localSettings.provider === 'local' ? '本地服务地址 (Localhost)' : '接口端点 (Endpoint)'}
                </label>
                <div className="relative group">
                  {localSettings.provider === 'local' ? (
                      <Laptop className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 group-focus-within:text-amber-500 transition-colors" />
                  ) : (
                      <Globe className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 group-focus-within:text-indigo-500 transition-colors" />
                  )}
                  <input 
                    className="w-full pl-12 pr-4 py-4 bg-gray-50 dark:bg-zinc-950 rounded-2xl border-none focus:ring-2 focus:ring-indigo-100 text-sm font-mono text-ink-900 dark:text-zinc-100"
                    value={localSettings.proxyEndpoint || ''}
                    onChange={(e) => setLocalSettings({ ...localSettings, proxyEndpoint: e.target.value })}
                    placeholder={localSettings.provider === 'local' ? "http://localhost:11434/v1" : "https://api.example.com/v1"}
                  />
                </div>
                {localSettings.provider === 'local' && (
                    <p className="text-[9px] text-gray-400 mt-2 px-2">
                        支持 Ollama, LM Studio, vLLM 等兼容 OpenAI 接口的本地服务。请确保本地服务允许 CORS 跨域请求。
                    </p>
                )}
              </div>
            )}

            <div>
              <div className="flex justify-between items-center mb-2">
                 <label className="text-[10px] font-black text-gray-400 uppercase tracking-widest block">模型选择</label>
                 <div className="flex items-center gap-3">
                   {localSettings.provider !== 'gemini' && (
                       <button 
                         onClick={() => setIsManualInput(!isManualInput)}
                         className={`text-[10px] font-bold flex items-center gap-1 transition-colors ${isManualInput ? 'text-indigo-600' : 'text-gray-400 hover:text-indigo-400'}`}
                       >
                         {isManualInput ? <List className="w-3 h-3" /> : <Keyboard className="w-3 h-3" />} 
                         {isManualInput ? '列表模式' : '手动模式'}
                       </button>
                   )}
                   <button 
                     onClick={handleRefresh} 
                     className="text-[10px] font-bold text-indigo-500 hover:text-indigo-700 flex items-center gap-1 group"
                     disabled={isRefreshing}
                   >
                     <RefreshCw className={`w-3 h-3 ${isRefreshing ? 'animate-spin' : 'group-hover:rotate-180 transition-transform duration-500'}`} /> 
                     {localSettings.provider === 'local' ? '检测连接' : '刷新'}
                   </button>
                 </div>
              </div>
              
              {isManualInput || (availableModels.length === 0 && localSettings.provider !== 'gemini') ? (
                  <div className="relative animate-in fade-in zoom-in-95">
                      <Edit2 className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                      <input 
                          className="w-full pl-12 pr-4 py-4 bg-gray-50 dark:bg-zinc-950 rounded-2xl border-none focus:ring-2 focus:ring-indigo-100 text-sm font-bold text-ink-900 dark:text-zinc-100"
                          value={localSettings.model}
                          onChange={(e) => handleModelChange(e.target.value)}
                          placeholder="输入模型 ID (如: llama3, qwen2.5)"
                      />
                  </div>
              ) : (
                  <div className="relative">
                    <select 
                      className={`w-full p-4 bg-gray-50 dark:bg-zinc-950 rounded-2xl border-none focus:ring-2 focus:ring-indigo-100 text-sm font-bold appearance-none cursor-pointer transition-opacity text-ink-900 dark:text-zinc-100 ${isRefreshing ? 'opacity-50' : 'opacity-100'}`}
                      value={localSettings.model}
                      onChange={(e) => handleModelChange(e.target.value)}
                      disabled={isRefreshing}
                    >
                      <option value="" disabled>-- 请选择模型 --</option>
                      {availableModels.map(m => (
                        <option key={m.id} value={m.id}>{m.name || m.id}</option>
                      ))}
                    </select>
                    <div className="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-gray-400">▼</div>
                  </div>
              )}
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <div className="flex justify-between items-center mb-2">
                  <label className="text-[10px] font-black text-gray-400 uppercase tracking-widest block">创造力 (Temp): {localSettings.temperature}</label>
                </div>
                <input 
                  type="range" min="0" max="1" step="0.1"
                  className="w-full h-1.5 bg-gray-200 dark:bg-zinc-800 rounded-lg appearance-none cursor-pointer accent-indigo-600"
                  value={localSettings.temperature}
                  onChange={(e) => setLocalSettings({ ...localSettings, temperature: parseFloat(e.target.value) })}
                />
              </div>

              <div>
                <div className="flex justify-between items-center mb-2">
                  <label className="text-[10px] font-black text-gray-400 uppercase tracking-widest block">最大输出 Token</label>
                </div>
                <div className="relative">
                    <Hash className="absolute left-3 top-1/2 -translate-y-1/2 w-3 h-3 text-gray-400" />
                    <input 
                        type="number"
                        className="w-full pl-10 pr-4 py-2 bg-gray-50 dark:bg-zinc-950 rounded-xl border-none focus:ring-2 focus:ring-indigo-100 text-xs font-bold text-ink-900 dark:text-zinc-100"
                        value={localSettings.maxOutputTokens || ''}
                        onChange={(e) => setLocalSettings({ ...localSettings, maxOutputTokens: parseInt(e.target.value) || undefined })}
                        placeholder="不限"
                    />
                </div>
              </div>
            </div>

            {supportsThinking && (
              <div className="p-6 bg-blue-50/50 dark:bg-blue-900/10 rounded-[2rem] border-2 border-blue-100/50 dark:border-blue-900/20 space-y-4 animate-in slide-in-from-top-2">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                        <BrainCircuit className="w-5 h-5 text-blue-600" />
                        <div>
                            <h4 className="text-sm font-black text-blue-900 dark:text-blue-300">深度思考预算 (Thinking Budget)</h4>
                            <p className="text-[9px] text-blue-600/60 uppercase font-black">限 Gemini 3 & 2.5 系列模型</p>
                        </div>
                    </div>
                    <div className="text-xs font-black text-blue-600">{localSettings.thinkingBudget || 0} Tokens</div>
                </div>
                <input 
                  type="range" min="0" max="32768" step="1024"
                  className="w-full h-1.5 bg-blue-200 dark:bg-blue-800 rounded-lg appearance-none cursor-pointer accent-blue-600"
                  value={localSettings.thinkingBudget || 0}
                  onChange={(e) => setLocalSettings({ ...localSettings, thinkingBudget: parseInt(e.target.value) })}
                />
                <p className="text-[9px] text-blue-500/80 leading-relaxed font-medium">
                  启用深度思考可让模型在回答前进行更缜密的逻辑推演。较高的预算适合复杂的剧情分析与逻辑校验。
                </p>
              </div>
            )}
          </section>

          <div className="flex items-start gap-3 p-4 bg-amber-50 dark:bg-amber-900/10 rounded-2xl border border-amber-100 dark:border-amber-900/20">
            <Info className="w-5 h-5 text-amber-500 shrink-0" />
            <div className="space-y-1">
               <p className="text-[10px] text-amber-700 dark:text-amber-400 font-bold leading-relaxed">
                 API 密钥由环境变量 process.env.API_KEY 提供 (Local 模式下可能不需要)。
               </p>
               <p className="text-[9px] text-amber-600/70 dark:text-amber-500/50 font-medium leading-relaxed">
                 设置将实时保存到当前项目。
               </p>
            </div>
          </div>
        </div>

        <div className="p-8 pt-0 flex gap-4">
           <button onClick={onClose} className="flex-1 py-4 text-xs font-black text-gray-400 uppercase hover:text-gray-900 transition-colors">取消</button>
           <button onClick={handleSave} className="flex-[2] py-4 bg-gray-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-2xl text-xs font-black shadow-xl hover:bg-black dark:hover:bg-white transition-all transform active:scale-95">保存配置</button>
        </div>
      </div>
    </div>
  );
};