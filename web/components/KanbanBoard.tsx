
import React, { useState } from 'react';
import { useProject } from '../contexts/ProjectContext';
import { useAuth } from '../contexts/AuthContext';
import { GripVertical, MoreHorizontal, Plus, BookOpen, Sparkles, X, Loader2, Calendar, LayoutGrid, Clock, ChevronRight, Target, Flag, Layers, Book, Edit3, Trash2, Anchor, Key, Route, Bookmark, Wand2, FileText, Save, Tag, Hash, FileType } from 'lucide-react';
import { ViewMode, Document, Volume } from '../types';
import { generateJSON, generateVolumeOutline } from '../services/aiService';

const PLOT_STRUCTURES = [
  { id: 'none', label: '自由构思', prompt: '自由发挥，不做结构限制。' },
  { id: 'in_medias_res', label: '开篇突发', prompt: '采用“中间开端”法，直接从冲突或危机的高潮切入，省略铺垫。' },
  { id: 'three_act_setup', label: '三幕式：铺垫', prompt: '展示常态世界，引入激励事件，打破平衡。' },
  { id: 'three_act_confrontation', label: '三幕式：对抗', prompt: '主角遭遇阻碍，尝试失败，冲突升级。' },
  { id: 'cliffhanger', label: '悬念收尾', prompt: '重点在于构建结尾的巨大悬念，在最高潮处戛而止。' },
  { id: 'emotional_deep_dive', label: '情感深潜', prompt: '弱化外部动作，着重描写角色的内心冲突、回忆与情感转变。' },
  { id: 'hero_threshold', label: '跨越门槛', prompt: '英雄之旅节点：主角正式离开舒适区，进入未知的冒险世界。' }
];

export const KanbanBoard: React.FC = () => {
  const { project, setActiveDocumentId, setViewMode, addDocument, updateDocument, updateNovelDetails, addVolume, updateVolume, deleteVolume, deleteDocument } = useProject();
  const { hasAIAccess } = useAuth();
  const [isAiModalOpen, setIsAiModalOpen] = useState(false);
  const [isOutlineModalOpen, setIsOutlineModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [editingDoc, setEditingDoc] = useState<Partial<Document> | null>(null);

  const [aiPrompt, setAiPrompt] = useState('');
  const [selectedStructure, setSelectedStructure] = useState<string>('none');
  const [isGenerating, setIsGenerating] = useState(false);
  
  if (!project) return null;

  const [activeVolumeTab, setActiveVolumeTab] = useState<string>(project.volumes[0]?.id || '');
  const currentVolume = project.volumes.find(v => v.id === activeVolumeTab);

  const openEditModal = (doc?: Document) => {
    if (doc) {
      setEditingDoc({ ...doc });
    } else {
      setEditingDoc({
        title: '',
        chapterGoal: '',
        corePlot: '',
        hook: '',
        causeEffect: '',
        foreshadowingDetails: '',
        duration: '',
        content: '',
        status: '草稿'
      });
    }
    setIsEditModalOpen(true);
  };

  const handleSaveManualChapter = () => {
    if (!editingDoc || !activeVolumeTab) return;
    
    if (!editingDoc.title) {
        alert("请输入章节标题");
        return;
    }

    if ((editingDoc as any).id) {
        updateDocument((editingDoc as any).id, editingDoc);
    } else {
        addDocument(activeVolumeTab, editingDoc);
    }
    setIsEditModalOpen(false);
    setEditingDoc(null);
  };

  const handleAiGenerateChapter = async () => {
     if (!aiPrompt.trim() || !activeVolumeTab) return;
     setIsGenerating(true);
     
     const structurePrompt = PLOT_STRUCTURES.find(s => s.id === selectedStructure)?.prompt || '';

     const context = `
     小说主题：${project.title} | 核心冲突：${project.coreConflict}
     类型：${project.genre} | 标签：${(project.tags || []).join(',')}
     人物弧光：${project.characterArc} | 世界规则：${project.worldRules}
     当前卷：${currentVolume?.title} | 卷目标：${currentVolume?.coreGoal}
     卷逻辑：${currentVolume?.chapterLinkageLogic || '未设定'}
     `;

     const schema = {
       type: "object",
       properties: {
         title: { type: "string" },
         chapterGoal: { type: "string" },
         corePlot: { type: "string" },
         hook: { type: "string" },
         timeNode: { type: "string" },
         causeEffect: { type: "string" },
         foreshadowingDetails: { type: "string" },
         duration: { type: "string", description: "预计叙事时长" }
       },
       required: ['title', 'chapterGoal', 'corePlot', 'hook']
     };

     const finalPrompt = `根据指令构思一个新章节：${aiPrompt}\n\n【结构要求】：${structurePrompt}`;

     const result = await generateJSON(
       finalPrompt,
       `你是一位专业的架构师，确保章节服务于卷和小说整体。层级上下文：${context}`,
       schema,
       project.aiSettings
     );
     
     if (result) {
        addDocument(activeVolumeTab, result);
        setIsAiModalOpen(false);
        setAiPrompt('');
     }
     setIsGenerating(false);
  };

  const handleGenerateOutline = async () => {
    if (!currentVolume) return;
    setIsGenerating(true);
    
    const volumeContext = `标题: ${currentVolume.title}\n主题: ${currentVolume.theme}\n目标: ${currentVolume.coreGoal}\n脉络: ${currentVolume.plotRoadmap}`;
    const structurePrompt = PLOT_STRUCTURES.find(s => s.id === selectedStructure)?.label || '自由构思';

    try {
      const chapters = await generateVolumeOutline(volumeContext, structurePrompt, project.aiSettings);
      if (chapters && chapters.length > 0) {
        chapters.forEach((chapterData: any) => {
          addDocument(currentVolume.id, {
            title: chapterData.title,
            chapterGoal: chapterData.chapterGoal,
            corePlot: chapterData.corePlot,
            hook: chapterData.hook,
            content: ''
          });
        });
        setIsOutlineModalOpen(false);
      }
    } catch (e) {
      console.error(e);
      alert("生成失败，请重试");
    } finally {
      setIsGenerating(false);
    }
  };

  const NovelHeader = () => (
    <div className="bg-white dark:bg-zinc-900 rounded-[2rem] p-8 border border-paper-200 dark:border-zinc-800 shadow-sm mb-10 space-y-8 transition-colors">
      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4 max-w-[40%]">
          <div className="p-4 bg-brand-600 dark:bg-brand-500 rounded-2xl shadow-lg shadow-brand-100 dark:shadow-black/20 shrink-0"><Book className="w-6 h-6 text-white" /></div>
          <div className="min-w-0">
            <h2 className="text-2xl font-black text-ink-900 dark:text-zinc-100 font-serif truncate">{project.title} · 总纲</h2>
            <p className="text-xs text-brand-600 dark:text-brand-400 font-bold uppercase tracking-widest">三级联动核心锚点</p>
          </div>
        </div>
        <div className="max-w-[40%] flex justify-end">
            <button onClick={() => setViewMode(ViewMode.SETTINGS)} className="flex items-center gap-2 px-4 py-2 bg-paper-100 dark:bg-zinc-800 text-ink-600 dark:text-zinc-400 rounded-xl text-xs font-bold hover:bg-paper-200 dark:hover:bg-zinc-700 transition-all border border-paper-200 dark:border-zinc-700 shrink truncate">
               <Sparkles className="w-3 h-3" /> {project.aiSettings.model}
            </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 pb-6 border-b border-paper-50 dark:border-zinc-800">
        <div className="space-y-1">
            <label className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest flex items-center gap-1">
                <FileType className="w-3 h-3" /> 类型 / 流派
            </label>
            <input 
                className="w-full bg-paper-50 dark:bg-zinc-950 rounded-xl px-3 py-2 text-sm font-bold text-ink-800 dark:text-zinc-200 border-none outline-none focus:ring-2 focus:ring-brand-200 dark:focus:ring-brand-900"
                value={project.genre || ''}
                onChange={e => updateNovelDetails({ genre: e.target.value })}
                placeholder="例如：东方玄幻..."
            />
        </div>
        <div className="space-y-1">
            <label className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest flex items-center gap-1">
                <Tag className="w-3 h-3" /> 标签
            </label>
            <input 
                className="w-full bg-paper-50 dark:bg-zinc-950 rounded-xl px-3 py-2 text-sm font-bold text-ink-800 dark:text-zinc-200 border-none outline-none focus:ring-2 focus:ring-brand-200 dark:focus:ring-brand-900"
                value={(project.tags || []).join(' ')}
                onChange={e => updateNovelDetails({ tags: e.target.value.split(/[\s,，]+/) })}
                placeholder="空格分隔..."
            />
        </div>
        <div className="space-y-1">
            <label className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest flex items-center gap-1">
                <Hash className="w-3 h-3" /> 预计总字数
            </label>
            <input 
                type="number"
                className="w-full bg-paper-50 dark:bg-zinc-950 rounded-xl px-3 py-2 text-sm font-bold text-ink-800 dark:text-zinc-200 border-none outline-none focus:ring-2 focus:ring-brand-200 dark:focus:ring-brand-900"
                value={project.totalWordCount || ''}
                onChange={e => updateNovelDetails({ totalWordCount: parseInt(e.target.value) || 0 })}
                placeholder="2000000"
            />
        </div>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="space-y-2 p-4 bg-red-50/50 dark:bg-red-900/10 rounded-2xl group border border-transparent hover:border-red-200 dark:hover:border-red-900/40 transition-all">
          <label className="text-[10px] font-black text-red-400 dark:text-red-500/60 uppercase tracking-widest flex items-center gap-2">
             <Target className="w-3 h-3" /> 灵魂锚点
          </label>
          <div className="space-y-3 mt-2">
            <input 
              className="w-full bg-transparent border-b border-red-100 dark:border-red-900/20 focus:border-red-300 dark:focus:border-red-700 outline-none text-sm font-bold text-ink-700 dark:text-zinc-300 pb-1 placeholder:text-red-200 dark:placeholder:text-red-900/40"
              value={project.coreConflict}
              onChange={e => updateNovelDetails({ coreConflict: e.target.value })}
              placeholder="核心冲突..."
            />
            <input 
              className="w-full bg-transparent border-b border-red-100 dark:border-red-900/20 focus:border-red-300 dark:focus:border-red-700 outline-none text-sm font-bold text-ink-700 dark:text-zinc-300 pb-1 placeholder:text-red-200 dark:placeholder:text-red-900/40"
              value={project.characterArc}
              onChange={e => updateNovelDetails({ characterArc: e.target.value })}
              placeholder="人物成长弧光..."
            />
          </div>
        </div>

        <div className="space-y-2 p-4 bg-blue-50/50 dark:bg-blue-900/10 rounded-2xl group border border-transparent hover:border-blue-200 dark:hover:border-blue-900/40 transition-all">
          <label className="text-[10px] font-black text-blue-400 dark:text-blue-500/60 uppercase tracking-widest flex items-center gap-2">
             <Anchor className="w-3 h-3" /> 设定锚点
          </label>
           <div className="space-y-3 mt-2">
            <input 
              className="w-full bg-transparent border-b border-blue-100 dark:border-blue-900/20 focus:border-blue-300 dark:focus:border-blue-700 outline-none text-sm font-bold text-ink-700 dark:text-zinc-300 pb-1 placeholder:text-blue-200 dark:placeholder:text-blue-900/40"
              value={project.worldRules || ''}
              onChange={e => updateNovelDetails({ worldRules: e.target.value })}
              placeholder="世界基础规则..."
            />
          </div>
        </div>

        <div className="space-y-2 p-4 bg-amber-50/50 dark:bg-amber-900/10 rounded-2xl group border border-transparent hover:border-amber-200 dark:hover:border-amber-900/40 transition-all">
          <label className="text-[10px] font-black text-amber-400 dark:text-amber-500/60 uppercase tracking-widest flex items-center gap-2">
             <Key className="w-3 h-3" /> 符号锚点
          </label>
           <div className="space-y-3 mt-2">
            <textarea 
              className="w-full bg-transparent border-none outline-none text-sm font-bold text-ink-700 dark:text-zinc-300 resize-none h-16 placeholder:text-amber-200 dark:placeholder:text-amber-900/40"
              value={project.symbolSettings || ''}
              onChange={e => updateNovelDetails({ symbolSettings: e.target.value })}
              placeholder="贯穿全书的核心符号..."
            />
          </div>
        </div>
      </div>
    </div>
  );

  const VolumeTabs = () => (
    <div className="flex items-center gap-2 mb-8 overflow-x-auto pb-2 scrollbar-hide">
      {project.volumes.map(vol => (
        <button
          key={vol.id}
          onClick={() => setActiveVolumeTab(vol.id)}
          className={`flex-shrink-0 px-6 py-3 rounded-2xl text-xs font-black transition-all border ${
            activeVolumeTab === vol.id 
              ? 'bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 border-ink-900 dark:border-zinc-100 shadow-xl' 
              : 'bg-white dark:bg-zinc-900 text-ink-400 dark:text-zinc-500 border-paper-200 dark:border-zinc-800 hover:border-paper-300 dark:hover:border-zinc-700'
          }`}
        >
          {vol.title}
        </button>
      ))}
      <button 
        onClick={() => addVolume()}
        className="flex-shrink-0 w-10 h-10 bg-white dark:bg-zinc-900 border border-paper-200 dark:border-zinc-800 rounded-2xl flex items-center justify-center text-ink-400 dark:text-zinc-500 hover:text-brand-600 dark:hover:text-brand-400 hover:border-brand-500 dark:hover:border-brand-400 transition-all"
      >
        <Plus className="w-4 h-4" />
      </button>
    </div>
  );

  const VolumeMetadata = () => {
    if (!currentVolume) return null;
    return (
      <div className="bg-white dark:bg-zinc-900 rounded-[2rem] p-8 border border-paper-200 dark:border-zinc-800 mb-10 shadow-sm transition-colors">
        <div className="flex justify-between items-start mb-6">
          <div className="flex items-center gap-3">
             <div className="p-3 bg-brand-600 dark:bg-brand-500 rounded-xl text-white shadow-md"><Layers className="w-5 h-5" /></div>
             <input 
               value={currentVolume.title}
               onChange={e => updateVolume(currentVolume.id, { title: e.target.value })}
               className="text-xl font-black text-ink-900 dark:text-zinc-100 bg-transparent border-none outline-none font-serif w-64"
             />
             <span className="text-[10px] bg-brand-100 dark:bg-brand-900/30 text-brand-700 dark:text-brand-400 px-2 py-1 rounded-md font-bold uppercase">分卷骨架</span>
          </div>
          <div className="flex items-center gap-2">
            {hasAIAccess && (
              <button 
                onClick={() => setIsOutlineModalOpen(true)}
                className="px-3 py-1.5 bg-paper-50 dark:bg-zinc-800 text-brand-600 dark:text-brand-400 rounded-lg text-[10px] font-black uppercase hover:bg-brand-50 dark:hover:bg-zinc-700 transition-colors flex items-center gap-1 shadow-sm border border-brand-100 dark:border-zinc-700"
              >
                <Wand2 className="w-3 h-3" /> AI 自动规划卷纲
              </button>
            )}
            <button onClick={() => { if(confirm(`确定删除卷 "${currentVolume.title}" 吗？`)) deleteVolume(currentVolume.id); }} className="p-2 text-ink-300 dark:text-zinc-600 hover:text-red-500 transition-colors"><Trash2 className="w-4 h-4" /></button>
          </div>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
           <div className="space-y-1">
              <label className="text-[10px] font-black text-brand-400 dark:text-zinc-600 uppercase tracking-[0.2em] flex items-center gap-1"><Target className="w-3 h-3"/> 核心目标</label>
              <textarea 
                className="w-full bg-transparent border-none outline-none text-xs font-bold text-ink-600 dark:text-zinc-400 resize-none h-16"
                value={currentVolume.coreGoal}
                onChange={e => updateVolume(currentVolume.id, { coreGoal: e.target.value })}
                placeholder="本卷需解决的关键冲突..."
              />
           </div>
           <div className="space-y-1">
              <label className="text-[10px] font-black text-brand-400 dark:text-zinc-600 uppercase tracking-[0.2em] flex items-center gap-1"><Route className="w-3 h-3"/> 情节脉络</label>
              <textarea 
                className="w-full bg-transparent border-none outline-none text-xs font-bold text-ink-600 dark:text-zinc-400 resize-none h-16"
                value={currentVolume.plotRoadmap || ''}
                onChange={e => updateVolume(currentVolume.id, { plotRoadmap: e.target.value })}
                placeholder="A事件 -> B事件 -> C高潮..."
              />
           </div>
           <div className="space-y-1">
              <label className="text-[10px] font-black text-brand-400 dark:text-zinc-600 uppercase tracking-[0.2em] flex items-center gap-1"><BookOpen className="w-3 h-3"/> 卷级设定</label>
              <textarea 
                className="w-full bg-transparent border-none outline-none text-xs font-bold text-ink-600 dark:text-zinc-400 resize-none h-16"
                value={currentVolume.volumeSpecificSettings || ''}
                onChange={e => updateVolume(currentVolume.id, { volumeSpecificSettings: e.target.value })}
                placeholder="本卷特有的环境设定..."
              />
           </div>
        </div>
      </div>
    );
  };

  const KanbanView = () => {
    const chapters = project.documents.filter(d => d.volumeId === activeVolumeTab);
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-8 pb-20">
        {chapters.map((doc, index) => (
          <div 
            key={doc.id}
            className="group bg-white dark:bg-zinc-900 rounded-3xl shadow-sm border border-paper-200 dark:border-zinc-800 p-6 hover:shadow-2xl transition-all cursor-pointer flex flex-col min-h-[280px]"
            onClick={() => {
              setActiveDocumentId(doc.id);
              setViewMode(ViewMode.WRITER);
            }}
          >
            <div className="flex justify-between items-start mb-4">
              <span className="text-[10px] font-black text-ink-300 dark:text-zinc-700 uppercase tracking-widest">第 {index + 1} 章</span>
              <div className="flex items-center gap-1">
                 <button 
                    onClick={(e) => { e.stopPropagation(); openEditModal(doc); }}
                    className="p-1.5 text-ink-300 dark:text-zinc-700 hover:text-brand-600 rounded-lg opacity-0 group-hover:opacity-100"
                 >
                    <Edit3 className="w-3.5 h-3.5" />
                 </button>
                 <div className={`w-2 h-2 rounded-full ml-1 ${doc.status === '完成' ? 'bg-green-400' : doc.status === '修改中' ? 'bg-amber-400' : 'bg-paper-200'}`} />
              </div>
            </div>
            
            <h3 className="text-lg font-black text-ink-800 dark:text-zinc-200 mb-2 line-clamp-2 font-serif">{doc.title}</h3>
            <p className="text-xs text-ink-500 dark:text-zinc-500 mb-6 line-clamp-3 leading-relaxed flex-1">{doc.summary || doc.content.substring(0, 80) + '...'}</p>
            
            <div className="space-y-2 pt-4 border-t border-paper-50 dark:border-zinc-800">
              {doc.chapterGoal && (
                <div className="flex items-center gap-2 text-[10px] font-bold text-brand-500 bg-brand-50 dark:bg-brand-900/20 p-2 rounded-lg">
                  <Target className="w-3 h-3 flex-shrink-0" />
                  <span className="truncate">{doc.chapterGoal}</span>
                </div>
              )}
            </div>
          </div>
        ))}
        
        <div className="rounded-3xl border-2 border-dashed border-paper-200 dark:border-zinc-800 flex flex-col items-center justify-center p-8 text-ink-400 dark:text-zinc-600 hover:border-brand-500 transition-all min-h-[280px] gap-6 group">
          <div className="p-4 bg-paper-50 dark:bg-zinc-800 rounded-full group-hover:scale-110 transition-transform"><Plus className="w-6 h-6" /></div>
          {hasAIAccess ? (
            <button 
               onClick={() => setIsAiModalOpen(true)}
               className="w-full py-2.5 bg-white dark:bg-zinc-900 border border-paper-200 dark:border-zinc-800 rounded-xl text-xs font-bold flex items-center justify-center gap-2 shadow-sm"
            >
               <Sparkles className="w-3.5 h-3.5" /> AI 灵感生成
            </button>
          ) : (
            <button 
               onClick={() => openEditModal()}
               className="w-full py-2.5 bg-white dark:bg-zinc-900 border border-paper-200 dark:border-zinc-800 rounded-xl text-xs font-bold flex items-center justify-center gap-2 shadow-sm"
            >
               <Edit3 className="w-3.5 h-3.5" /> 手动创建章节
            </button>
          )}
        </div>
      </div>
    );
  };

  return (
    <div className="flex-1 h-full bg-paper-50 dark:bg-zinc-950 bg-dot-pattern overflow-y-auto relative transition-colors duration-300">
      <div className="max-w-[1600px] mx-auto p-12">
        <NovelHeader />
        
        {project.volumes.length > 0 ? (
          <>
            <VolumeTabs />
            <VolumeMetadata />
            <KanbanView />
          </>
        ) : (
          <div className="flex flex-col items-center justify-center py-20 text-ink-400 dark:text-zinc-600">
            <BookOpen className="w-16 h-16 mb-4 text-paper-200 dark:text-zinc-800" />
            <p className="text-sm font-bold">暂无分卷，请点击上方“+”创建第一卷</p>
          </div>
        )}
      </div>

      {/* AI 生成弹窗 */}
      {isAiModalOpen && hasAIAccess && (
        <div className="fixed inset-0 z-[150] bg-ink-900/40 dark:bg-black/60 backdrop-blur-sm flex items-center justify-center p-6">
          <div className="bg-white dark:bg-zinc-900 rounded-[2rem] shadow-2xl w-full max-w-lg p-8 animate-in zoom-in-95 border border-white/10">
            <div className="flex justify-between items-center mb-6">
              <h3 className="text-xl font-black text-ink-900 dark:text-zinc-100 font-serif">AI 章节构造</h3>
              <button onClick={() => setIsAiModalOpen(false)}><X className="w-6 h-6 text-ink-300" /></button>
            </div>
            
            <div className="space-y-4 mb-6">
              <div>
                <label className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest block mb-2">事件梗概</label>
                <textarea 
                  value={aiPrompt}
                  onChange={e => setAiPrompt(e.target.value)}
                  placeholder="描述本章发生的核心事件..."
                  className="w-full h-32 p-4 bg-paper-50 dark:bg-zinc-950 rounded-2xl border-none focus:ring-4 focus:ring-brand-50 text-sm resize-none text-ink-900 dark:text-zinc-100"
                />
              </div>
              
              <div>
                <label className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest block mb-2">情节结构模式</label>
                <select 
                  value={selectedStructure}
                  onChange={(e) => setSelectedStructure(e.target.value)}
                  className="w-full p-3 bg-paper-50 dark:bg-zinc-950 rounded-xl border-none text-sm font-bold text-ink-700 dark:text-zinc-300 cursor-pointer"
                >
                  {PLOT_STRUCTURES.map(s => <option key={s.id} value={s.id}>{s.label}</option>)}
                </select>
              </div>
            </div>

            <button 
              onClick={handleAiGenerateChapter}
              disabled={isGenerating || !aiPrompt.trim()}
              className="w-full py-4 bg-brand-600 text-white rounded-xl font-black text-sm shadow-lg flex items-center justify-center gap-2"
            >
              {isGenerating ? <Loader2 className="w-5 h-5 animate-spin" /> : <Sparkles className="w-5 h-5" />}
              生成章节骨架
            </button>
          </div>
        </div>
      )}

      {/* 手动编辑弹窗 */}
      {isEditModalOpen && editingDoc && (
        <div className="fixed inset-0 z-[150] bg-ink-900/40 dark:bg-black/60 backdrop-blur-sm flex items-center justify-center p-6">
            <div className="bg-white dark:bg-zinc-900 rounded-[2rem] shadow-2xl w-full max-w-3xl p-8 animate-in zoom-in-95 max-h-[90vh] overflow-y-auto border border-white/10">
            <div className="flex justify-between items-center mb-8 border-b border-paper-100 dark:border-zinc-800 pb-4">
                <div className="flex items-center gap-3">
                    <div className="p-3 bg-paper-100 rounded-xl"><Edit3 className="w-5 h-5 text-ink-600" /></div>
                    <h3 className="text-xl font-black text-ink-900 dark:text-zinc-100 font-serif">编辑章节</h3>
                </div>
                <button onClick={() => setIsEditModalOpen(false)}><X className="w-6 h-6 text-ink-300" /></button>
            </div>
            
            <div className="space-y-6">
                <div className="grid grid-cols-3 gap-6">
                    <div className="col-span-2 space-y-1">
                        <label className="text-[10px] font-black text-ink-400 uppercase tracking-widest block">章节标题</label>
                        <input 
                            value={editingDoc.title || ''}
                            onChange={e => setEditingDoc({...editingDoc, title: e.target.value})}
                            className="w-full p-3 bg-paper-50 dark:bg-zinc-950 rounded-xl border-none text-lg font-bold font-serif text-ink-900 dark:text-zinc-100"
                        />
                    </div>
                </div>
            </div>

            <div className="pt-8 flex gap-4">
                <button onClick={() => setIsEditModalOpen(false)} className="flex-1 py-4 text-xs font-black text-ink-400 uppercase">取消</button>
                <button 
                    onClick={handleSaveManualChapter}
                    className="flex-[2] py-4 bg-ink-900 text-white rounded-xl text-xs font-black uppercase shadow-lg flex items-center justify-center gap-2"
                >
                   <Save className="w-4 h-4" /> 保存章节
                </button>
            </div>
            </div>
        </div>
      )}
    </div>
  );
};
