
import React, { useState, useMemo, useEffect } from 'react';
import { useProject } from '../contexts/ProjectContext';
import { useAuth } from '../contexts/AuthContext';
import { useConfirm } from '../contexts/ConfirmContext';
import { 
  Users, BookOpen, Plus, Tag, Trash2, Link as LinkIcon, X, Sparkles, 
  Activity, Check, Search, Loader2, Shield, Sword, Scroll, Zap, Globe, ChevronRight, Wand2,
  MoreVertical, Edit3, SortAsc, SortDesc, FileJson, FileText as FileTextIcon, Type, Mic2, RefreshCw, PlusCircle
} from 'lucide-react';
import { EntityType, StoryEntity, ViewMode } from '../types';
import { buildWorkflowPayload, generateJSON, refineNovelCore } from '../services/aiService';
import { runWorldWorkflowApi } from '../services/workflowApi';
import { toSnapshotPayload, upsertProjectSnapshotApi } from '../services/projectApi';
import { GraphVisualizer } from './GraphVisualizer';
import { EntityCard } from './EntityCard';

type SortOption = 'title' | 'id' | 'importance';

export const WorldBible: React.FC = () => {
  const { project, addEntity, updateEntity, deleteEntity, updateNovelDetails, setViewMode } = useProject();
  const { hasAIAccess } = useAuth();
  const { confirm } = useConfirm();
  
  const [activeCategory, setActiveCategory] = useState<EntityType | 'all'>('all');
  const [showGraph, setShowGraph] = useState(false);
  
  // Debounced search logic
  const [inputValue, setInputValue] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    const handler = setTimeout(() => {
      setSearchQuery(inputValue);
    }, 300); // 300ms debounce

    return () => {
      clearTimeout(handler);
    };
  }, [inputValue]);

  const [sortBy, setSortBy] = useState<SortOption>('id');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');

  const [isAICreating, setIsAICreating] = useState<EntityType | null>(null);
  const [aiBrief, setAiBrief] = useState('');
  const [isGeneratingDraft, setIsGeneratingDraft] = useState(false);
  const [isRefiningCore, setIsRefiningCore] = useState(false);

  const [lastCreatedId, setLastCreatedId] = useState<string | null>(null);

  if (!project) return null;

  const categories = [
    { id: 'all', label: '全部要素', icon: <Globe className="w-4 h-4" />, color: 'text-ink-400 dark:text-zinc-500', activeBg: 'bg-ink-900 dark:bg-zinc-100' },
    { id: 'character', label: '角色人物', icon: <Users className="w-4 h-4" />, color: 'text-rose-500 dark:text-rose-400', activeBg: 'bg-rose-600 dark:bg-rose-500' },
    { id: 'organization', label: '组织机构', icon: <Shield className="w-4 h-4" />, color: 'text-amber-500 dark:text-amber-400', activeBg: 'bg-amber-600 dark:bg-amber-500' },
    { id: 'item', label: '神兵法宝', icon: <Sword className="w-4 h-4" />, color: 'text-cyan-500 dark:text-cyan-400', activeBg: 'bg-cyan-600 dark:bg-cyan-500' },
    { id: 'setting', label: '地理风貌', icon: <BookOpen className="w-4 h-4" />, color: 'text-emerald-500 dark:text-emerald-400', activeBg: 'bg-emerald-600 dark:bg-emerald-500' },
    { id: 'magic', label: '力量体系', icon: <Zap className="w-4 h-4" />, color: 'text-indigo-500 dark:text-indigo-400', activeBg: 'bg-indigo-600 dark:bg-indigo-500' },
    { id: 'event', label: '大事年表', icon: <Scroll className="w-4 h-4" />, color: 'text-orange-500 dark:text-orange-400', activeBg: 'bg-orange-600 dark:bg-orange-500' },
  ];

  const filteredEntities = useMemo(() => {
    let result = project.entities.filter(e => {
      const matchesCategory = activeCategory === 'all' || e.type === activeCategory;
      const matchesSearch = e.title.toLowerCase().includes(searchQuery.toLowerCase()) || 
                            e.content.toLowerCase().includes(searchQuery.toLowerCase());
      return matchesCategory && matchesSearch;
    });

    result.sort((a, b) => {
      let valA: any = a[sortBy as keyof StoryEntity] || '';
      let valB: any = b[sortBy as keyof StoryEntity] || '';
      if (valA < valB) return sortOrder === 'asc' ? -1 : 1;
      if (valA > valB) return sortOrder === 'asc' ? 1 : -1;
      return 0;
    });

    return result;
  }, [project.entities, activeCategory, searchQuery, sortBy, sortOrder]);

  const handleManualCreate = () => {
    const type = activeCategory === 'all' ? 'character' : activeCategory;
    const newId = `e${Date.now()}_${Math.random().toString(36).substr(2, 5)}`;
    addEntity(type, { 
      id: newId, 
      title: '未命名实体', 
      content: '', 
      tags: [],
      importance: 'secondary'
    });
    setLastCreatedId(newId);
  };

  const ensureProjectId = async (): Promise<number | null> => {
    if (!project) return null;
    try {
      const dto = await upsertProjectSnapshotApi(toSnapshotPayload(project, project.id));
      return dto.id || null;
    } catch (e) {
      return null;
    }
  };

  const handleAICreate = async () => {
    if (!aiBrief.trim() || !isAICreating || !hasAIAccess) return;
    setIsGeneratingDraft(true);

    const schema = {
      type: "object",
      properties: {
        title: { type: "string" },
        subtitle: { type: "string" },
        content: { type: "string" },
        tags: { type: "array", items: { type: "string" } },
        voiceStyle: isAICreating === 'character' ? { type: "string" } : undefined
      },
      required: ['title', 'subtitle', 'content', 'tags']
    };

    const prompt = `根据用户描述构思一个${isAICreating}设定：${aiBrief}。如果是角色，请包含语调风格。`;
    const systemInstruction = `你是一位创意非凡的小说设定专家。书名：${project.title}`;

    let draft: any = null;
    try {
      const projectId = await ensureProjectId();
      if (!projectId) {
        throw new Error('项目未同步到后端');
      }
      const payload = buildWorkflowPayload(prompt, systemInstruction, project.aiSettings, schema);
      const result = await runWorldWorkflowApi({
        project_id: projectId,
        title: `世界观生成 · ${isAICreating}`,
        step_title: `孵化${isAICreating}`,
        provider: payload.provider,
        path: payload.path,
        body: payload.body,
      });
      try {
        draft = JSON.parse(result.content || '{}');
      } catch {
        draft = null;
      }
    } catch (e) {
      draft = await generateJSON(prompt, systemInstruction, schema, project.aiSettings);
    }

    if (draft) {
      const newId = `e${Date.now()}_${Math.random().toString(36).substr(2, 5)}`;
      addEntity(isAICreating, { ...draft, id: newId });
      setIsAICreating(null);
      setAiBrief('');
    }
    setIsGeneratingDraft(false);
  };

  const handleRefineCore = async () => {
    if (!hasAIAccess) return;
    const ok = await confirm({
      title: '确定执行 AI 重新生成吗？',
      description: '将根据现有信息重新生成/完善核心冲突等要素。',
      confirmText: '确认执行',
      cancelText: '取消',
      tone: 'default',
    });
    if (ok) {
      setIsRefiningCore(true);
      try {
        const settings = project.aiSettings;
        const systemInstruction = `你是一位世界顶尖的小说架构师和文学顾问。你的任务是分析现有的小说草案，并深度完善其核心世界观设定。`;
        const context = `书名：${project.title}\n类型：${project.genre || '未定义'}\n核心冲突：${project.coreConflict || '无'}\n世界规则：${project.worldRules || '无'}`;
        const prompt = `分析以上信息，完善以下核心要素：\n${context}`;
        const schema = {
          type: 'object',
          properties: {
            genre: { type: 'string' },
            tags: { type: 'array', items: { type: 'string' } },
            coreConflict: { type: 'string' },
            characterArc: { type: 'string' },
            ultimateValue: { type: 'string' },
            worldRules: { type: 'string' },
            characterCore: { type: 'string' },
            symbolSettings: { type: 'string' }
          },
          required: ['coreConflict', 'characterArc', 'worldRules', 'characterCore']
        };

        let refined: any = null;
        try {
          const projectId = await ensureProjectId();
          if (!projectId) {
            throw new Error('项目未同步到后端');
          }
          const payload = buildWorkflowPayload(prompt, systemInstruction, settings, schema);
          const result = await runWorldWorkflowApi({
            project_id: projectId,
            title: '世界观生成 · 核心精修',
            step_title: '核心精修',
            provider: payload.provider,
            path: payload.path,
            body: payload.body,
          });
          try {
            refined = JSON.parse(result.content || '{}');
          } catch {
            refined = null;
          }
        } catch (e) {
          refined = await refineNovelCore(project);
        }
        if (refined) {
          updateNovelDetails(refined);
          alert("核心设定已完善！");
        }
      } catch (e) {
        console.error("完善核心设定失败", e);
      }
      setIsRefiningCore(false);
    }
  };

  return (
    <div className="flex-1 h-full bg-paper-50 dark:bg-zinc-950 bg-dot-pattern flex overflow-hidden relative transition-colors duration-300">
      {showGraph && <GraphVisualizer project={project} onClose={() => setShowGraph(false)} />}
      
      {/* 侧边分类栏 */}
      <div className="w-64 border-r border-paper-200 dark:border-zinc-800 bg-white/80 dark:bg-zinc-900/80 backdrop-blur-md p-6 flex flex-col gap-8 transition-colors">
        <div>
          <h3 className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-[0.3em] mb-6 px-2">档案分类索引</h3>
          <nav className="space-y-1.5">
            {categories.map(cat => (
              <button
                key={cat.id}
                onClick={() => setActiveCategory(cat.id as any)}
                className={`w-full flex items-center justify-between px-4 py-3 rounded-2xl text-xs font-bold transition-all group ${
                  activeCategory === cat.id ? `${cat.activeBg} text-white dark:text-zinc-900 shadow-xl shadow-paper-300 dark:shadow-black/20` : 'text-ink-500 dark:text-zinc-400 hover:bg-paper-100 dark:hover:bg-zinc-800 hover:text-ink-900 dark:hover:text-zinc-100'
                }`}
              >
                <div className="flex items-center gap-3">
                  <span className={`${activeCategory === cat.id ? (activeCategory === 'all' ? 'text-white dark:text-zinc-900' : 'text-white') : cat.color}`}>{cat.icon}</span>
                  {cat.label}
                </div>
                <span className={`text-[10px] font-black ${activeCategory === cat.id ? 'text-white/60 dark:text-zinc-900/60' : 'text-ink-300 dark:text-zinc-600'}`}>
                  {cat.id === 'all' ? project.entities.length : project.entities.filter(e => e.type === cat.id).length}
                </span>
              </button>
            ))}
          </nav>
        </div>

        <div className="mt-auto space-y-3">
          <button onClick={handleManualCreate} className="w-full py-4 bg-white dark:bg-zinc-800 border-2 border-paper-200 rounded-2xl text-[11px] font-black uppercase tracking-widest shadow-sm hover:border-brand-500 transition-all flex items-center justify-center gap-2">
            <PlusCircle className="w-4 h-4" /> 手动建立
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-12 scroll-smooth">
        <header className="mb-12">
          <div className="flex justify-between items-start mb-10 overflow-hidden">
            <div className="space-y-1 max-w-[40%]">
              <h1 className="text-4xl font-black text-ink-900 dark:text-zinc-100 tracking-tight font-serif italic truncate">设定集 (Bible)</h1>
              <div className="flex items-center gap-2 mt-2">
                 <button onClick={() => setViewMode(ViewMode.SETTINGS)} className="px-2 py-1 bg-paper-100 dark:bg-zinc-800 rounded-lg text-[10px] font-black text-brand-600 dark:text-brand-400 hover:bg-brand-50 transition-colors flex items-center gap-1 border border-paper-200 shrink truncate">
                    {project.aiSettings.model}
                 </button>
              </div>
            </div>
            <div className="flex gap-3 max-w-[40%] justify-end shrink-0">
              {hasAIAccess && (
                <button onClick={handleRefineCore} disabled={isRefiningCore} className="hidden sm:flex items-center gap-2 px-5 py-3.5 bg-brand-50 dark:bg-zinc-800 text-brand-700 dark:text-brand-400 border border-brand-100 rounded-2xl font-black text-[11px] uppercase tracking-widest hover:bg-brand-500 hover:text-white transition-all">
                  {isRefiningCore ? <Loader2 className="w-4 h-4 animate-spin" /> : <RefreshCw className="w-4 h-4" />} 
                  精修核心
                </button>
              )}
              <button onClick={() => setShowGraph(true)} className="flex items-center gap-2 px-5 py-3.5 bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-2xl font-black text-[11px] uppercase tracking-widest shadow-xl hover:bg-black transition-all">
                <Activity className="w-4 h-4" /> 图谱
              </button>
            </div>
          </div>

          <div className="relative group max-w-xl">
            <Search className="absolute left-6 top-1/2 -translate-y-1/2 w-4 h-4 text-ink-400 dark:text-zinc-500" />
            <input 
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              placeholder="搜索档案..."
              className="w-full pl-14 pr-6 py-4.5 bg-white dark:bg-zinc-900 border border-paper-200 rounded-[2rem] shadow-sm outline-none transition-all font-bold text-sm"
            />
          </div>
        </header>

        {/* AI Creating Modal */}
        {isAICreating && hasAIAccess && (
          <div className="fixed inset-0 z-[150] bg-ink-900/40 dark:bg-black/60 backdrop-blur-md flex items-center justify-center p-6">
            <div className="bg-white dark:bg-zinc-900 rounded-[3.5rem] shadow-2xl w-full max-w-2xl p-10 animate-in zoom-in-95 duration-300 border border-white/10">
              <div className="flex justify-between items-center mb-8">
                <h4 className="text-2xl font-black text-ink-900 dark:text-white font-serif">AI 档案孵化</h4>
                <button onClick={() => setIsAICreating(null)}><X className="w-7 h-7 text-gray-400" /></button>
              </div>
              <textarea 
                value={aiBrief}
                onChange={e => setAiBrief(e.target.value)}
                placeholder="描述你的灵感雏形..."
                className="w-full h-40 p-6 bg-paper-50 dark:bg-zinc-800 rounded-[2rem] border-none text-sm font-bold leading-relaxed resize-none text-ink-900 dark:text-zinc-100"
              />
              <button 
                onClick={handleAICreate}
                disabled={isGeneratingDraft || !aiBrief.trim()}
                className="w-full mt-6 py-5 bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-[1.5rem] font-black text-xs uppercase tracking-widest flex items-center justify-center gap-3"
              >
                {isGeneratingDraft ? <Loader2 className="w-5 h-5 animate-spin" /> : <Wand2 className="w-5 h-5" />}
                确认孵化
              </button>
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 pb-20">
          {filteredEntities.map(entity => (
            <EntityCard
              key={entity.id}
              entity={entity}
              project={project}
              categories={categories}
              onUpdate={updateEntity}
              onDelete={deleteEntity}
              onLink={(id) => console.log('Link', id)}
              autoEdit={lastCreatedId === entity.id}
            />
          ))}
          <button onClick={handleManualCreate} className="group rounded-[2.5rem] border-2 border-dashed border-paper-200 flex flex-col items-center justify-center p-12 text-ink-400 hover:border-brand-500 transition-all min-h-[400px] gap-4">
            <Plus className="w-10 h-10" />
            <span className="text-xs font-black uppercase">新建档案</span>
          </button>
        </div>
      </div>
    </div>
  );
};
