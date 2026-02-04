
import React, { useState, useEffect } from 'react';
import { useProject } from '../contexts/ProjectContext';
import { useConfirm } from '../contexts/ConfirmContext';
import { Plugin, PluginCapability } from '../types';
import {
  fetchPluginsApi,
  mapPluginFromApi,
  enablePluginApi,
  disablePluginApi,
  pingPluginApi,
  createPluginApi,
  deletePluginApi,
} from '../services/pluginService';
import { 
  Puzzle, Plus, Power, Trash2, Globe, Server, 
  Terminal, ShieldCheck, RefreshCw, ExternalLink, 
  Zap, Settings2, Info, Check, Copy, BookOpen, 
  Code, Cpu, Activity, AlertCircle, Play, X, Save, PlusCircle
} from 'lucide-react';

interface ConfigModalProps {
  plugin: Plugin;
  onSave: (config: Record<string, any>) => void;
  onClose: () => void;
}

const ConfigModal: React.FC<ConfigModalProps> = ({ plugin, onSave, onClose }) => {
  const [config, setConfig] = useState<Record<string, any>>(plugin.config || {});
  const [newKey, setNewKey] = useState('');
  const [newValue, setNewValue] = useState('');

  const handleAdd = () => {
    if (!newKey.trim()) return;
    setConfig({ ...config, [newKey.trim()]: newValue });
    setNewKey('');
    setNewValue('');
  };

  const handleRemove = (key: string) => {
    const next = { ...config };
    delete next[key];
    setConfig(next);
  };

  return (
    <div className="fixed inset-0 z-[200] bg-ink-900/60 backdrop-blur-md flex items-center justify-center p-6 animate-in fade-in duration-300">
      <div className="bg-white dark:bg-zinc-900 rounded-[2.5rem] shadow-2xl w-full max-w-xl overflow-hidden flex flex-col border border-white/10">
        <div className="p-8 border-b border-paper-100 dark:border-zinc-800 flex items-center justify-between bg-paper-50 dark:bg-zinc-950">
          <div className="flex items-center gap-4">
            <div className="p-3 bg-brand-50 dark:bg-brand-900/20 text-brand-600 rounded-2xl">
              <Settings2 className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-xl font-black text-ink-900 dark:text-zinc-100 font-serif">{plugin.name} 配置</h2>
              <p className="text-[10px] text-ink-400 font-black uppercase tracking-widest mt-0.5">环境变量与私有设定</p>
            </div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-paper-100 dark:hover:bg-zinc-800 rounded-xl text-ink-400"><X className="w-6 h-6" /></button>
        </div>

        <div className="p-8 space-y-6 max-h-[60vh] overflow-y-auto custom-scrollbar">
          <div className="space-y-4">
            {Object.entries(config).map(([key, value]) => (
              <div key={key} className="flex gap-3 items-center bg-paper-50 dark:bg-zinc-950 p-4 rounded-2xl border border-paper-100 dark:border-zinc-800 group animate-in slide-in-from-left-2">
                <div className="flex-1 space-y-1">
                  <p className="text-[10px] font-black text-brand-600 uppercase tracking-tighter">{key}</p>
                  <input 
                    type="password"
                    value={value} 
                    onChange={(e) => setConfig({...config, [key]: e.target.value})}
                    className="w-full bg-transparent text-sm font-bold text-ink-900 dark:text-zinc-100 outline-none"
                    autoComplete="off"
                  />
                </div>
                <button onClick={() => handleRemove(key)} className="p-2 text-ink-200 hover:text-rose-500 opacity-0 group-hover:opacity-100 transition-all">
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            ))}
          </div>

          <div className="pt-6 border-t border-paper-100 dark:border-zinc-800 space-y-4">
            <p className="text-[10px] font-black text-ink-300 uppercase tracking-widest">添加新配置项 (如 API_KEY)</p>
            <div className="flex gap-3">
              <input 
                placeholder="键名" 
                value={newKey} 
                onChange={e => setNewKey(e.target.value)}
                className="w-1/3 p-3 bg-paper-50 dark:bg-zinc-950 rounded-xl border-none text-xs font-bold focus:ring-2 focus:ring-brand-200 outline-none"
              />
              <input 
                placeholder="值" 
                value={newValue} 
                onChange={e => setNewValue(e.target.value)}
                className="flex-1 p-3 bg-paper-50 dark:bg-zinc-950 rounded-xl border-none text-xs font-bold focus:ring-2 focus:ring-brand-200 outline-none"
              />
              <button onClick={handleAdd} className="p-3 bg-brand-500 text-white rounded-xl hover:bg-brand-600 transition-all">
                <PlusCircle className="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>

        <div className="p-8 flex gap-4">
          <button onClick={onClose} className="flex-1 py-4 text-xs font-black text-ink-400 uppercase tracking-widest">取消</button>
          <button onClick={() => { onSave(config); onClose(); }} className="flex-[2] py-4 bg-ink-900 dark:bg-brand-500 text-white rounded-2xl text-xs font-black uppercase tracking-widest shadow-xl flex items-center justify-center gap-2">
            <Save className="w-4 h-4" /> 保存并应用
          </button>
        </div>
      </div>
    </div>
  );
};

export const PluginManager: React.FC = () => {
  const { project, setProject } = useProject();
  const { confirm } = useConfirm();
  const [newEndpoint, setNewEndpoint] = useState('');
  const [isConnecting, setIsConnecting] = useState(false);
  const [activeTab, setActiveTab] = useState<'manage' | 'docs'>('manage');
  const [docTab, setDocTab] = useState<'python' | 'nodejs' | 'specs'>('python');
  const [logs, setLogs] = useState<{msg: string, type: 'info' | 'error' | 'success', time: string}[]>([]);
  const [configuringPlugin, setConfiguringPlugin] = useState<Plugin | null>(null);

  if (!project) return null;

  const plugins = project.plugins || [];

  useEffect(() => {
    const loadPlugins = async () => {
      try {
        const data = await fetchPluginsApi(1, 50);
        const mapped = (data.plugins || []).map(mapPluginFromApi);
        setProject({
          ...project,
          plugins: mapped,
        });
      } catch (e) {
        addLog('插件列表加载失败，请稍后重试', 'error');
      }
    };

    loadPlugins();
  }, []);

  const addLog = (msg: string, type: 'info' | 'error' | 'success' = 'info') => {
    setLogs(prev => [{msg, type, time: new Date().toLocaleTimeString()}, ...prev].slice(0, 50));
  };

  const handlePing = async (id: string) => {
    const plugin = plugins.find(p => p.id === id);
    if (!plugin) return;
    
    addLog(`Pinging service ${plugin.name}...`, 'info');
    try {
      await pingPluginApi(id);
      addLog(`${plugin.name} ping 已发送`, 'success');
    } catch {
      addLog(`${plugin.name} ping 失败`, 'error');
    }
  };

  const handleConnectPlugin = async () => {
    if (!newEndpoint.trim()) return;
    setIsConnecting(true);
    addLog(`Initiating handshake with ${newEndpoint}...`, 'info');

    try {
      const payload = {
        name: `plugin-${Date.now()}`,
        endpoint: newEndpoint,
      };
      const created = await createPluginApi(payload);
      const mapped = mapPluginFromApi(created);
      setProject({
        ...project,
        plugins: [...plugins, mapped],
      });
      addLog(`已创建插件记录：${mapped.name}`, 'success');
      setNewEndpoint('');
    } catch {
      addLog(`创建插件失败，请检查后端连接`, 'error');
    }
    setIsConnecting(false);
  };

  const togglePlugin = async (id: string) => {
    const plugin = plugins.find(p => p.id === id);
    if (!plugin) return;

    try {
      if (plugin.isEnabled) {
        await disablePluginApi(id);
        addLog(`已禁用插件 ${plugin.name}`, 'info');
      } else {
        await enablePluginApi(id);
        addLog(`已启用插件 ${plugin.name}`, 'success');
      }
      setProject({
        ...project,
        plugins: plugins.map(p => p.id === id ? { ...p, isEnabled: !p.isEnabled } : p)
      });
    } catch {
      addLog('插件状态更新失败', 'error');
    }
  };

  const removePlugin = async (id: string) => {
    const plugin = plugins.find(p => p.id === id);
    if (!plugin) return;
    const ok = await confirm({
      title: '确定删除此插件吗？',
      description: '删除后将无法恢复。',
      confirmText: '删除',
      cancelText: '取消',
      tone: 'danger',
    });
    if (ok) {
      try {
        await deletePluginApi(id);
        setProject({
          ...project,
          plugins: plugins.filter(p => p.id !== id)
        });
        addLog(`已删除插件 ${plugin.name}`, 'success');
      } catch {
        addLog('删除插件失败', 'error');
      }
    }
  };

  const savePluginConfig = (pluginId: string, config: Record<string, any>) => {
    setProject({
      ...project,
      plugins: plugins.map(p => p.id === pluginId ? { ...p, config } : p)
    });
    addLog(`Configuration updated for plugin [${pluginId}]`, 'success');
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    addLog('Snippet copied to clipboard', 'success');
  };

  const CodeSnippets = {
    python: `from fastapi import FastAPI, Request
from pydantic import BaseModel
from typing import List

app = FastAPI()

@app.get("/manifest")
def get_manifest():
    return {
        "id": "py-logic-analyzer",
        "name": "Python Logic Analyzer",
        "version": "1.0.0",
        "capabilities": [
            {"id": "check_consistency", "name": "Check Consistency", "type": "logic_checker"}
        ]
    }

@app.post("/action")
async def handle_action(request: Request):
    data = await request.json()
    # Access user-provided config via: data['pluginConfig']
    # Access context via: data['context']
    
    return [
        {
            "type": "show_message",
            "payload": {"text": "Analysis complete: No plot holes found!", "type": "success"}
        }
    ]`,
    nodejs: `const express = require('express');
const app = express();
app.use(express.json());

app.get('/manifest', (req, res) => {
  res.json({
    id: "js-style-polisher",
    name: "JS Style Polisher",
    capabilities: [{ id: "polish", name: "Polish Style", type: "text_processor" }]
  });
});

app.post('/action', (req, res) => {
  const { actionId, pluginConfig, context } = req.body;
  // Use pluginConfig.API_KEY if needed
  res.json([{
    type: "update_document",
    payload: { content: context.activeDocument.content + "\\n\\n-- Polished by JS --" }
  }]);
});

app.listen(8080);`,
    specs: `// Manifest Response (GET /manifest)
{
  "id": string,
  "name": string,
  "capabilities": Array<{id, name, type, description}>
}

// Action Payload (POST /action)
{
  "actionId": string,
  "pluginConfig": Record<string, any>, // User defined key-values
  "context": {
    "project": { id, title, genre, entities, worldRules },
    "activeDocument": { id, title, content }
  }
}

// Action Instructions (Expected Response Array)
Array<{
  "type": "update_document" | "update_entity" | "show_message" | "add_log",
  "payload": object
}>`
  };

  return (
    <div className="flex-1 h-full bg-paper-50 dark:bg-zinc-950 transition-colors duration-300 overflow-hidden flex flex-col">
      {configuringPlugin && (
        <ConfigModal 
          plugin={configuringPlugin} 
          onSave={(config) => savePluginConfig(configuringPlugin.id, config)}
          onClose={() => setConfiguringPlugin(null)} 
        />
      )}

      {/* Header */}
      <div className="h-20 border-b border-paper-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 px-8 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-4">
          <div className="p-3 bg-brand-600 dark:bg-brand-500 rounded-2xl text-white shadow-lg">
            <Puzzle className="w-6 h-6" />
          </div>
          <div>
            <h1 className="text-xl font-black text-ink-900 dark:text-zinc-100 font-serif">Plugin & Extension Hub</h1>
            <div className="flex items-center gap-3 mt-1">
              <span className="flex items-center gap-1 text-[10px] font-black uppercase text-green-500">
                <ShieldCheck className="w-3 h-3" /> API Core Secure
              </span>
              <span className="flex items-center gap-1 text-[10px] font-black uppercase text-ink-300 dark:text-zinc-600">
                <Activity className="w-3 h-3" /> Runtime: Active
              </span>
            </div>
          </div>
        </div>

        <div className="flex bg-paper-100 dark:bg-zinc-800 p-1 rounded-xl">
           <button 
             onClick={() => setActiveTab('manage')}
             className={`px-6 py-2 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all ${activeTab === 'manage' ? 'bg-white dark:bg-zinc-700 text-ink-900 dark:text-zinc-100 shadow-sm' : 'text-ink-400 hover:text-ink-600'}`}
           >
             Manage Services
           </button>
           <button 
             onClick={() => setActiveTab('docs')}
             className={`px-6 py-2 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all ${activeTab === 'docs' ? 'bg-white dark:bg-zinc-700 text-ink-900 dark:text-zinc-100 shadow-sm' : 'text-ink-400 hover:text-ink-600'}`}
           >
             Developer Docs
           </button>
        </div>
      </div>

      <div className="flex-1 flex overflow-hidden">
        {activeTab === 'manage' ? (
          <div className="flex-1 flex flex-col lg:flex-row overflow-hidden">
            {/* Main Content */}
            <div className="flex-1 overflow-y-auto p-12 custom-scrollbar">
               <div className="max-w-4xl mx-auto space-y-12">
                 {/* Installer Card */}
                 <div className="bg-white dark:bg-zinc-900 rounded-[2.5rem] p-10 border border-paper-200 dark:border-zinc-800 shadow-sm">
                   <h3 className="text-lg font-black text-ink-900 dark:text-zinc-100 mb-6 flex items-center gap-3">
                     <Plus className="w-5 h-5 text-brand-500" /> Install Remote Microservice
                   </h3>
                   <div className="flex gap-4">
                      <div className="flex-1 relative">
                        <Globe className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-ink-300 dark:text-zinc-600" />
                        <input 
                          value={newEndpoint}
                          onChange={e => setNewEndpoint(e.target.value)}
                          placeholder="Endpoint URL (e.g., http://localhost:8080)"
                          className="w-full pl-12 pr-4 py-4 bg-paper-50 dark:bg-zinc-950 rounded-2xl border-none focus:ring-2 focus:ring-brand-200 dark:focus:ring-brand-500 text-sm font-bold text-ink-900 dark:text-zinc-100 outline-none transition-all shadow-inner"
                        />
                      </div>
                      <button 
                        onClick={handleConnectPlugin}
                        disabled={isConnecting || !newEndpoint.trim()}
                        className="px-8 bg-ink-900 dark:bg-brand-500 text-white rounded-2xl font-black text-xs uppercase tracking-widest hover:bg-black dark:hover:bg-brand-600 transition-all flex items-center gap-3 disabled:opacity-50 shadow-xl"
                      >
                        {isConnecting ? <RefreshCw className="w-4 h-4 animate-spin" /> : <Cpu className="w-4 h-4" />}
                        Connect Service
                      </button>
                   </div>
                   <p className="mt-4 text-[10px] text-ink-400 dark:text-zinc-500 flex items-center gap-2">
                     <Info className="w-3.5 h-3.5" /> 
                     System will request <code className="bg-paper-100 dark:bg-zinc-800 px-1.5 py-0.5 rounded">/manifest</code> to fetch capabilities.
                   </p>
                 </div>

                 {/* Plugin List */}
                 <div className="space-y-4">
                    <h4 className="px-4 text-[10px] font-black text-ink-300 dark:text-zinc-600 uppercase tracking-widest flex items-center gap-2">
                       <Server className="w-3.5 h-3.5" /> Registered Nodes ({plugins.length})
                    </h4>
                    {plugins.length === 0 ? (
                      <div className="bg-white/50 dark:bg-zinc-900/50 border-2 border-dashed border-paper-200 dark:border-zinc-800 rounded-[2.5rem] p-20 text-center text-ink-300 dark:text-zinc-600">
                        <Puzzle className="w-12 h-12 mx-auto mb-4 opacity-20" />
                        <p className="text-sm font-bold italic font-serif">No plugins connected. Use the panel above to expand NAO.</p>
                      </div>
                    ) : (
                      <div className="grid grid-cols-1 gap-4">
                        {plugins.map(p => (
                          <div key={p.id} className="bg-white dark:bg-zinc-900 rounded-[1.5rem] border border-paper-200 dark:border-zinc-800 p-8 flex items-center justify-between group transition-all hover:shadow-xl hover:-translate-y-1">
                            <div className="flex items-center gap-6">
                               <button 
                                 onClick={() => handlePing(p.id)}
                                 className={`p-5 rounded-2xl relative transition-all active:scale-95 ${p.isEnabled ? 'bg-brand-100 dark:bg-brand-900/30 text-brand-600' : 'bg-paper-100 dark:bg-zinc-800 text-ink-300'}`}
                                 title="Click to Refresh Status"
                               >
                                  <Zap className={`w-6 h-6 ${p.status === 'online' ? 'animate-pulse' : ''}`} />
                                  <div className={`absolute -top-1 -right-1 w-3 h-3 rounded-full border-2 border-white dark:border-zinc-900 ${p.status === 'online' ? 'bg-green-500' : 'bg-rose-500'}`} />
                               </button>
                               <div>
                                  <div className="flex items-center gap-2">
                                     <h5 className="text-lg font-black text-ink-900 dark:text-zinc-100 font-serif">{p.name}</h5>
                                     <span className="text-[9px] bg-paper-100 dark:bg-zinc-800 text-ink-500 dark:text-zinc-400 px-2 py-0.5 rounded-full font-bold">v{p.version}</span>
                                  </div>
                                  <p className="text-xs text-ink-500 dark:text-zinc-500 mt-1">{p.description || 'No description provided.'}</p>
                                  <div className="flex items-center gap-4 mt-3">
                                     <span className="text-[9px] font-black text-ink-300 dark:text-zinc-600 flex items-center gap-1.5 uppercase tracking-tighter">
                                        <Globe className="w-3 h-3" /> {p.endpoint}
                                     </span>
                                     <span className="text-[9px] font-black text-ink-300 dark:text-zinc-600 flex items-center gap-1.5 uppercase tracking-tighter">
                                        <Activity className="w-3 h-3" /> Latency: {p.latency || 0}ms
                                     </span>
                                  </div>
                               </div>
                            </div>
                            <div className="flex items-center gap-3">
                               <button 
                                 onClick={() => setConfiguringPlugin(p)}
                                 className="p-4 rounded-xl transition-all text-ink-300 hover:text-brand-600 hover:bg-brand-50 dark:hover:bg-brand-900/20"
                                 title="Configure"
                               >
                                 <Settings2 className="w-5 h-5" />
                               </button>
                               <button 
                                 onClick={() => togglePlugin(p.id)}
                                 className={`p-4 rounded-xl transition-all ${p.isEnabled ? 'bg-green-50 dark:bg-green-900/20 text-green-600' : 'bg-paper-50 dark:bg-zinc-800 text-ink-300'}`}
                                 title={p.isEnabled ? 'Disable Service' : 'Enable Service'}
                               >
                                 <Power className="w-5 h-5" />
                               </button>
                               <button 
                                 onClick={() => removePlugin(p.id)}
                                 className="p-4 text-ink-300 dark:text-zinc-600 hover:text-rose-500 hover:bg-rose-50 dark:hover:bg-rose-900/20 rounded-xl transition-all"
                                 title="Uninstall"
                               >
                                 <Trash2 className="w-5 h-5" />
                               </button>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                 </div>
               </div>
            </div>

            {/* Side Console */}
            <div className="w-full lg:w-[400px] border-l border-paper-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 flex flex-col overflow-hidden shrink-0">
               <div className="p-6 border-b border-paper-100 dark:border-zinc-800 flex items-center justify-between bg-paper-50 dark:bg-zinc-950">
                  <h4 className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest flex items-center gap-2">
                     <Terminal className="w-3.5 h-3.5" /> Runtime Live Console
                  </h4>
                  <button onClick={() => setLogs([])} className="text-[9px] font-black text-brand-600 dark:text-brand-400 uppercase hover:underline">Clear</button>
               </div>
               <div className="flex-1 overflow-y-auto p-4 font-mono text-[10px] space-y-3 custom-scrollbar bg-ink-950 text-emerald-400/90 selection:bg-emerald-500 selection:text-ink-950">
                  {logs.length === 0 && <p className="text-zinc-700 italic opacity-50">Waiting for interaction logs...</p>}
                  {logs.map((log, i) => (
                    <div key={i} className={`pb-2 border-b border-white/5 last:border-0 ${log.type === 'error' ? 'text-rose-400' : log.type === 'success' ? 'text-emerald-400' : 'text-zinc-400'}`}>
                       <div className="flex justify-between text-[8px] opacity-30 mb-1 font-sans">
                          <span>{log.time}</span>
                          <span className="uppercase tracking-widest">{log.type}</span>
                       </div>
                       <p className="leading-relaxed">{log.msg}</p>
                    </div>
                  ))}
               </div>
               <div className="p-4 bg-ink-950 border-t border-white/5 text-zinc-600 text-[9px] font-black uppercase tracking-widest text-center">
                  NAO Runtime Core v1.0.4 - Secure Bridge
               </div>
            </div>
          </div>
        ) : (
          <div className="flex-1 overflow-y-auto p-12 custom-scrollbar bg-paper-50 dark:bg-zinc-950">
             <div className="max-w-5xl mx-auto space-y-12">
                <header className="space-y-4">
                   <h2 className="text-3xl font-black text-ink-900 dark:text-zinc-100 font-serif">Integration Guide</h2>
                   <p className="text-sm text-ink-500 dark:text-zinc-400 max-w-3xl leading-relaxed">
                     NAO Plugins are independent microservices. You build the logic in any language you prefer, and NAO communicates via a simple 
                     standard JSON-RPC over HTTP protocol.
                   </p>
                </header>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                   <div className="bg-white dark:bg-zinc-900 p-8 rounded-[2rem] border border-paper-200 dark:border-zinc-800 space-y-4">
                      <div className="w-10 h-10 bg-brand-50 dark:bg-brand-900/30 rounded-xl flex items-center justify-center text-brand-600">
                        <Code className="w-5 h-5" />
                      </div>
                      <h4 className="font-black text-ink-900 dark:text-zinc-100">Language Agnostic</h4>
                      <p className="text-xs text-ink-500 leading-relaxed">FastAPI, Express, Go, or Rust. As long as it can speak JSON over HTTP, it can be a NAO Plugin.</p>
                   </div>
                   <div className="bg-white dark:bg-zinc-900 p-8 rounded-[2rem] border border-paper-200 dark:border-zinc-800 space-y-4">
                      <div className="w-10 h-10 bg-emerald-50 dark:bg-emerald-900/30 rounded-xl flex items-center justify-center text-emerald-600">
                        <Zap className="w-5 h-5" />
                      </div>
                      <h4 className="font-black text-ink-900 dark:text-zinc-100">Live Context</h4>
                      <p className="text-xs text-ink-500 leading-relaxed">Every action receives the full project state: characters, world rules, and current chapter content.</p>
                   </div>
                   <div className="bg-white dark:bg-zinc-900 p-8 rounded-[2rem] border border-paper-200 dark:border-zinc-800 space-y-4">
                      <div className="w-10 h-10 bg-amber-50 dark:bg-amber-900/30 rounded-xl flex items-center justify-center text-amber-600">
                        <Play className="w-5 h-5" />
                      </div>
                      <h4 className="font-black text-ink-900 dark:text-zinc-100">Command Control</h4>
                      <p className="text-xs text-ink-500 leading-relaxed">Plugins don't just return text; they return "Instructions" to modify the user's project state safely.</p>
                   </div>
                </div>

                {/* Snippets Section */}
                <div className="bg-ink-900 dark:bg-zinc-900 rounded-[2.5rem] overflow-hidden shadow-2xl">
                   <div className="p-6 border-b border-white/5 flex items-center justify-between bg-black/20">
                      <div className="flex gap-4">
                         {['python', 'nodejs', 'specs'].map(t => (
                           <button 
                             key={t}
                             onClick={() => setDocTab(t as any)}
                             className={`px-4 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all ${docTab === t ? 'bg-white text-ink-900' : 'text-zinc-500 hover:text-zinc-300'}`}
                           >
                             {t === 'specs' ? 'API Spec' : t.toUpperCase()}
                           </button>
                         ))}
                      </div>
                      <button 
                        onClick={() => copyToClipboard(CodeSnippets[docTab])}
                        className="p-2 text-zinc-500 hover:text-white transition-colors"
                      >
                        <Copy className="w-4 h-4" />
                      </button>
                   </div>
                   <pre className="p-8 text-[11px] font-mono text-emerald-400 overflow-x-auto custom-scrollbar h-[450px]">
                      <code>{CodeSnippets[docTab]}</code>
                   </pre>
                </div>

                <div className="p-10 bg-brand-50 dark:bg-brand-900/10 rounded-[2.5rem] border border-brand-100 dark:border-brand-900/20 flex flex-col md:flex-row items-center gap-8">
                   <div className="p-6 bg-brand-600 rounded-3xl text-white shadow-xl shrink-0"><BookOpen className="w-8 h-8" /></div>
                   <div className="space-y-2">
                      <h4 className="text-lg font-black text-brand-900 dark:text-brand-300">Read the full Developer Handbook</h4>
                      <p className="text-sm text-brand-700/60 dark:text-brand-400/60">Explore advanced interaction patterns, multi-step agent flows, and UI extension methods.</p>
                      <button className="flex items-center gap-2 text-brand-600 dark:text-brand-400 font-black text-xs uppercase tracking-widest mt-4 group">
                        Open Documentation <ExternalLink className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
                      </button>
                   </div>
                </div>
             </div>
          </div>
        )}
      </div>
    </div>
  );
};
