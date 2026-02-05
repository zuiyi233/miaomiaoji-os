
import React, { useState, useEffect, useRef } from 'react';
import { useProject } from '../contexts/ProjectContext';
import { useAuth } from '../contexts/AuthContext';
import { generateStoryContent, generateTimeSuggestions } from '../services/geminiService';
import { buildWorkflowPayload, generateText } from '../services/aiService';
import { runChapterAnalyzeApi, runChapterGenerateApi, runChapterRewriteApi, runPolishWorkflowApi } from '../services/workflowApi';
import { toSnapshotPayload, upsertProjectSnapshotApi } from '../services/projectApi';
import { callPluginAction, executePluginActions } from '../services/pluginService';
import { createDocumentApi } from '../services/documentApi';
import { AgentWriter } from './AgentWriter';
import { 
  Loader2, Maximize2, Sparkles, Bookmark as BookmarkIcon, Plus, X, Trash2, 
  ArrowRight, Book, Layers, Target, Layout, Route, Anchor, GitMerge, 
  AlignLeft, Save, Edit3, FileText, Hash, Puzzle, Play, ShieldAlert, ChevronRight, AlertCircle, Settings, Bot
} from 'lucide-react';
import { ViewMode } from '../types';

export const Editor: React.FC = () => {
  const { activeDocumentId, project, updateDocument, updateEntity, toggleAISidebar, addBookmark, deleteBookmark, setViewMode, selectSession } = useProject();
  const { hasAIAccess } = useAuth();
  const [content, setContent] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [isPolishing, setIsPolishing] = useState(false);
  const [isChapterGenerating, setIsChapterGenerating] = useState(false);
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [isPluginRunning, setIsPluginRunning] = useState<string | null>(null);
  const [pluginError, setPluginError] = useState<string | null>(null);
  const [showBookmarks, setShowBookmarks] = useState(false);
  const [showHierarchy, setShowHierarchy] = useState(false);
  const [showAgentWriter, setShowAgentWriter] = useState(false);
  const [bookmarkName, setBookmarkName] = useState('');
  const [customPrompt, setCustomPrompt] = useState('');
  const [isGeneratingTime, setIsGeneratingTime] = useState(false);
  const [wordCountRequest, setWordCountRequest] = useState<'normal' | 'expand' | 'finish'>('normal');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const activeDoc = project?.documents.find(d => d.id === activeDocumentId);
  const activeVolume = project?.volumes.find(v => v.id === activeDoc?.volumeId);
  const enabledPlugins = (project?.plugins || []).filter(p => p.isEnabled) || [];

  const sortedDocs = (project?.documents || [])
    .filter(d => d.volumeId === activeDoc?.volumeId)
    .sort((a, b) => a.order - b.order);
  const currentIndex = sortedDocs.findIndex(d => d.id === activeDoc?.id);
  const prevDoc = currentIndex > 0 ? sortedDocs[currentIndex - 1] : null;
  const nextDoc = currentIndex !== -1 && currentIndex < sortedDocs.length - 1 ? sortedDocs[currentIndex + 1] : null;

  useEffect(() => {
    if (activeDoc) {
      setContent(activeDoc.content);
      if (!activeDoc.timeNode && activeDoc.title && !isGeneratingTime && hasAIAccess) {
        handleGenerateTimeNode();
      }
    }
  }, [activeDoc?.id]);

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = `${Math.max(textareaRef.current.scrollHeight, 600)}px`;
    }
  }, [content, showHierarchy]);

  const handleAIRequest = async () => {
    if (!activeDoc || activeDoc.status === '完成' || !hasAIAccess) return;
    setIsGenerating(true);

    const result = await generateStoryContent(
        customPrompt,
        project,
        { ...activeDoc, content },
        prevDoc || null,
        nextDoc || null,
        wordCountRequest
    );
    
    if (result) {
      const cleanedResult = result.trim().replace(/^[\n\r]+/, '');
      const newText = content + (content.endsWith('\n') ? '' : '\n') + cleanedResult;
      setContent(newText);
      updateDocument(activeDoc.id, { content: newText });
    }
    setIsGenerating(false);
    setCustomPrompt('');
  };

  const ensureProjectId = async (): Promise<number | null> => {
    if (!project) return null;
    try {
      const dto = await upsertProjectSnapshotApi(toSnapshotPayload(project, project.id));
      return dto.id || null;
    } catch {
      return null;
    }
  };

  const ensureBackendDocumentId = async (projectId: number): Promise<number | null> => {
    if (!activeDoc) return null;
    if (typeof activeDoc.backendId === 'number' && activeDoc.backendId > 0) return activeDoc.backendId;

    // 后端文档表是独立的：这里做一次懒创建，拿到 document_id 供章节工作流写回。
    try {
      const created = await createDocumentApi(projectId, {
        title: activeDoc.title || '未命名章节',
        content: content || '',
        summary: activeDoc.summary || '',
        status: activeDoc.status || '草稿',
        order_index: (typeof activeDoc.order === 'number' ? activeDoc.order : 0) + 1,
      });
      if (created?.id) {
        updateDocument(activeDoc.id, { backendId: created.id });
        return created.id;
      }
      return null;
    } catch {
      return null;
    }
  };

  const handlePolishChapter = async () => {
    if (!activeDoc || !project || !hasAIAccess) return;
    if (!content.trim()) return;
    setIsPolishing(true);
    try {
      const systemInstruction = '你是一位资深文学编辑，请在不改变核心情节的前提下润色文本，提升文采与节奏。只输出润色后的正文。';
      const prompt = `请润色以下章节内容：\n\n${content}`;
      let polished = '';

      try {
        const projectId = await ensureProjectId();
        if (!projectId) throw new Error('项目未同步到后端');

        const backendDocumentId = await ensureBackendDocumentId(projectId);
        if (!backendDocumentId) throw new Error('文档未同步到后端');

        const payload = buildWorkflowPayload(prompt, systemInstruction, project.aiSettings);
        const result = await runChapterRewriteApi({
          project_id: projectId,
          document_id: backendDocumentId,
          rewrite_mode: 'polish',
          provider: payload.provider,
          path: payload.path,
          body: payload.body,
          write_back: { set_status: '修改中' },
        });
        polished = result.content || '';
        if (result.session?.id) {
          selectSession(String(result.session.id));
          setViewMode(ViewMode.WORKFLOW_DETAIL);
        }
      } catch {
        // 兜底：先走旧的工作流接口（不依赖后端文档表）；再兜底直连
        try {
          const projectId = await ensureProjectId();
          if (!projectId) throw new Error('项目未同步到后端');
          const payload = buildWorkflowPayload(prompt, systemInstruction, project.aiSettings);
          const result = await runPolishWorkflowApi({
            project_id: projectId,
            title: `章节润色 · ${activeDoc.title}`,
            step_title: '润色输出',
            provider: payload.provider,
            path: payload.path,
            body: payload.body,
          });
          polished = result.content || '';
          if (result.session?.id) {
            selectSession(String(result.session.id));
            setViewMode(ViewMode.WORKFLOW_DETAIL);
          }
        } catch {
          polished = await generateText(prompt, systemInstruction, project.aiSettings);
        }
      }

      if (polished.trim()) {
        setContent(polished.trim());
        updateDocument(activeDoc.id, { content: polished.trim() });
      }
    } finally {
      setIsPolishing(false);
    }
  };

  const handleGenerateChapter = async () => {
    if (!activeDoc || !project || !hasAIAccess) return;
    setIsChapterGenerating(true);
    try {
      const projectId = await ensureProjectId();
      if (!projectId) return;
      const backendDocumentId = await ensureBackendDocumentId(projectId);
      if (!backendDocumentId) return;

      const systemInstruction = '你是一位职业小说作者。根据给定的章节目标、核心情节与钩子，写出完整章节正文。只输出正文。';
      const prompt = `章节标题：${activeDoc.title}\n章节目标：${activeDoc.chapterGoal || '无'}\n核心情节：${activeDoc.corePlot || '无'}\n因果链：${activeDoc.causeEffect || '无'}\n细节伏笔：${activeDoc.foreshadowingDetails || '无'}\n结尾钩子：${activeDoc.hook || '无'}\n\n请写出本章正文。`;
      const payload = buildWorkflowPayload(prompt, systemInstruction, project.aiSettings);

      const result = await runChapterGenerateApi({
        project_id: projectId,
        document_id: backendDocumentId,
        title: activeDoc.title,
        provider: payload.provider,
        path: payload.path,
        body: payload.body,
        write_back: { set_status: '草稿' },
      });

      const text = (result?.content || '').trim();
      if (text) {
        setContent(text);
        updateDocument(activeDoc.id, { content: text });
      }

      if (result.session?.id) {
        selectSession(String(result.session.id));
        setViewMode(ViewMode.WORKFLOW_DETAIL);
      }
    } finally {
      setIsChapterGenerating(false);
    }
  };

  const handleAnalyzeChapter = async () => {
    if (!activeDoc || !project || !hasAIAccess) return;
    if (!content.trim()) return;
    setIsAnalyzing(true);
    try {
      const systemInstruction = '你是一位文学编辑与剧情分析师。请输出一段结构化摘要（不超过300字），并列出3-8条关键剧情点。仅输出摘要与要点。';
      const prompt = `请分析并总结以下章节内容：\n\n${content}`;

      const projectId = await ensureProjectId();
      if (!projectId) return;
      const backendDocumentId = await ensureBackendDocumentId(projectId);
      if (!backendDocumentId) return;

      const payload = buildWorkflowPayload(prompt, systemInstruction, project.aiSettings);
      const result = await runChapterAnalyzeApi({
        project_id: projectId,
        document_id: backendDocumentId,
        provider: payload.provider,
        path: payload.path,
        body: payload.body,
        write_back: { set_summary: true },
      });

      const summary = (result?.document as any)?.summary || result?.content || '';
      if (typeof summary === 'string' && summary.trim()) {
        updateDocument(activeDoc.id, { summary: summary.trim() });
      }

      if (result.session?.id) {
        selectSession(String(result.session.id));
        setViewMode(ViewMode.WORKFLOW_DETAIL);
      }
    } finally {
      setIsAnalyzing(false);
    }
  };

  const handleAgentAppend = (text: string) => {
      const newText = content + text;
      setContent(newText);
      updateDocument(activeDoc.id, { content: newText });
  };

  const handleRunPluginAction = async (pluginId: string, actionId: string) => {
    if (!activeDoc) return;
    const plugin = project.plugins?.find(p => p.id === pluginId);
    if (!plugin) return;

    setIsPluginRunning(actionId);
    setPluginError(null);
    
    try {
      const result = await callPluginAction(plugin, actionId, project, { ...activeDoc, content });
      
      if (!result.success) {
        setPluginError(result.error || "Action failed.");
        console.error(`[Plugin Error]: ${result.error}`);
      } else {
        executePluginActions(result.actions, {
          updateDocument,
          updateEntity,
          activeDocId: activeDoc.id,
          onMessage: (msg, type) => {
            if (type === 'error') setPluginError(msg);
            else console.log(`[Plugin Notification]: ${msg} (${type})`);
          }
        });
      }
    } catch (e) {
      setPluginError("An unexpected system error occurred while calling the plugin.");
    } finally {
      setIsPluginRunning(null);
    }
  };

  const handleGenerateTimeNode = async () => {
    if (!activeDoc || !project || !hasAIAccess) return;
    setIsGeneratingTime(true);
    try {
        const context = `小说标题：${project.title}\n世界观：${project.worldRules}\n前序章节：${prevDoc?.title || '无'}`;
        const result = await generateTimeSuggestions([activeDoc.title], context, project.aiSettings);
        
        if (result && result.suggestions && result.suggestions.length > 0) {
             const suggestion = result.suggestions[0].timeNode;
             if (suggestion) {
                 updateDocument(activeDoc.id, { timeNode: suggestion });
             }
        }
    } catch (e) {
        console.error("生成时间节点失败", e);
    }
    setIsGeneratingTime(false);
  };

  const handleAddBookmark = () => {
    if (!activeDoc || !bookmarkName.trim()) return;
    const pos = textareaRef.current?.selectionStart || content.length;
    addBookmark(activeDoc.id, bookmarkName, pos);
    setBookmarkName('');
  };

  const jumpToBookmark = (pos: number) => {
    if (textareaRef.current) {
      textareaRef.current.focus();
      textareaRef.current.setSelectionRange(pos, pos);
      const scrollContainer = document.getElementById('editor-scroll-body');
      if (scrollContainer) {
        const lines = content.substring(0, pos).split('\n').length;
        scrollContainer.scrollTop = lines * 36 - 200;
      }
    }
  };

  const handlePaste = (e: React.ClipboardEvent<HTMLTextAreaElement>) => {
    const pastedText = e.clipboardData.getData('text');
    if (!pastedText) return;

    const codeKeywords = ['function ', 'const ', 'let ', 'var ', 'import ', 'class ', 'return ', 'if (', 'for (', '=>', 'public ', 'private '];
    const hasCodeKeywords = codeKeywords.some(keyword => pastedText.includes(keyword));
    const hasBraces = pastedText.includes('{') && pastedText.includes('}');
    const isMultiLine = pastedText.split('\n').length > 1;

    if ((hasCodeKeywords || (hasBraces && isMultiLine)) && !pastedText.trim().startsWith('```')) {
      e.preventDefault();
      const textArea = textareaRef.current;
      if (!textArea) return;

      const start = textArea.selectionStart;
      const end = textArea.selectionEnd;
      const formattedCode = `\n\`\`\`\n${pastedText.trim()}\n\`\`\`\n`;
      const newContent = content.substring(0, start) + formattedCode + content.substring(end);
      
      setContent(newContent);
      updateDocument(activeDoc.id, { content: newContent });
      
      setTimeout(() => {
        const newCursorPos = start + formattedCode.length;
        textArea.setSelectionRange(newCursorPos, newCursorPos);
      }, 0);
    }
  };

  if (!project) return null;
  if (!activeDoc) return <div className="flex-1 flex items-center justify-center text-ink-400 dark:text-zinc-500 bg-paper-50 dark:bg-zinc-950 bg-dot-pattern">点击左侧章节开始创作</div>;

  const currentWordCount = content.length;
  const targetWordCount = activeDoc.targetWordCount || 3000;
  const progressPercent = Math.min((currentWordCount / targetWordCount) * 100, 100);

  return (
    <div className="flex-1 flex flex-col h-full relative overflow-hidden bg-paper-50 dark:bg-zinc-950 transition-colors duration-300">
      
      {showAgentWriter && (
          <AgentWriter 
            activeDoc={activeDoc} 
            onClose={() => setShowAgentWriter(false)} 
            onAppendContent={handleAgentAppend} 
          />
      )}

      {/* 写作页头部 */}
      <div className="min-h-[56px] bg-paper-50/95 dark:bg-zinc-900/95 border-b border-paper-200 dark:border-zinc-800 sticky top-0 z-20 backdrop-blur-sm shadow-sm transition-colors flex items-center">
        <div className="w-full max-w-3xl mx-auto flex flex-col lg:flex-row lg:items-center gap-2 px-4 sm:px-6 md:px-8 text-[11px] font-medium tracking-wide">
          {/* 左侧面包屑 */}
          <div className="flex items-center gap-2 w-full min-w-0 lg:flex-1 truncate">
            <span className="text-ink-400 dark:text-zinc-500 flex items-center gap-2 shrink-0"><Book className="w-3.5 h-3.5" /> {project.title}</span>
            <span className="text-ink-300 dark:text-zinc-700 shrink-0">/</span>
            <span className="text-ink-500 dark:text-zinc-500 truncate">{activeVolume?.title}</span>
            <span className="text-ink-300 dark:text-zinc-700 shrink-0">/</span>
            <span className="text-ink-900 dark:text-zinc-100 font-bold truncate">{activeDoc.title}</span>
          </div>
          
          {/* 右侧工具组 */}
          <div className="ml-auto flex items-center gap-2 sm:gap-3 md:gap-4 w-full lg:w-auto justify-end flex-wrap">
            <button 
              onClick={() => setViewMode(ViewMode.SETTINGS)}
              className="hidden lg:flex items-center gap-2 px-3 py-1.5 bg-paper-100 dark:bg-zinc-800 hover:bg-paper-200 dark:hover:bg-zinc-700 rounded-lg text-[10px] font-bold text-ink-600 dark:text-zinc-300 transition-all border border-paper-200 dark:border-zinc-700 shrink truncate group"
              title="进入控制中心配置模型"
            >
              <Sparkles className="w-3 h-3 text-brand-500 group-hover:animate-pulse" />
              <span className="truncate max-w-[80px]">{project.aiSettings.model}</span>
              <Settings className="w-2.5 h-2.5 opacity-30 group-hover:opacity-100" />
            </button>

            {hasAIAccess && (
              <button
                onClick={handlePolishChapter}
                disabled={isPolishing}
                className="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-emerald-600 text-white rounded-lg text-[10px] font-black uppercase tracking-widest hover:bg-emerald-500 transition-all shadow-lg disabled:opacity-60"
                title="章节润色"
              >
                {isPolishing ? <Loader2 className="w-3 h-3 animate-spin" /> : <FileText className="w-3 h-3" />} 润色
              </button>
            )}

            {hasAIAccess && (
              <button
                onClick={handleGenerateChapter}
                disabled={isChapterGenerating}
                className="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-lg text-[10px] font-black uppercase tracking-widest hover:bg-black dark:hover:bg-white transition-all shadow-lg disabled:opacity-60"
                title="章节生成"
              >
                {isChapterGenerating ? <Loader2 className="w-3 h-3 animate-spin" /> : <Sparkles className="w-3 h-3" />} 生成
              </button>
            )}

            {hasAIAccess && (
              <button
                onClick={handleAnalyzeChapter}
                disabled={isAnalyzing}
                className="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-brand-600 text-white rounded-lg text-[10px] font-black uppercase tracking-widest hover:bg-brand-500 transition-all shadow-lg disabled:opacity-60"
                title="章节分析"
              >
                {isAnalyzing ? <Loader2 className="w-3 h-3 animate-spin" /> : <ShieldAlert className="w-3 h-3" />} 分析
              </button>
            )}
            <div className="hidden md:flex items-center gap-2 text-[10px] font-semibold text-ink-400 dark:text-zinc-400 bg-paper-100 dark:bg-zinc-800 px-3 py-1 rounded-full border border-paper-200 dark:border-zinc-700 shrink-0">
                <span className="tabular-nums">{currentWordCount} / {targetWordCount}</span>
                <div className="w-10 h-1 bg-paper-200 dark:bg-zinc-700 rounded-full overflow-hidden">
                    <div className="h-full bg-ink-800 dark:bg-zinc-100 transition-all duration-500" style={{ width: `${progressPercent}%` }}></div>
                </div>
            </div>
            
            <div className="flex items-center border-l border-paper-200 dark:border-zinc-800 pl-3 gap-1.5 shrink-0">
                <button 
                  onClick={() => setShowHierarchy(!showHierarchy)} 
                  className={`p-2 rounded-lg transition-all ${showHierarchy ? 'bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900' : 'text-ink-400 dark:text-zinc-500 hover:text-ink-900 dark:hover:text-zinc-100 hover:bg-paper-100 dark:hover:bg-zinc-800'}`}
                  title="章节逻辑结构"
                >
                  <Layers className="w-4 h-4" />
                </button>
                <button 
                  onClick={() => setShowBookmarks(!showBookmarks)} 
                  className={`p-2 rounded-lg transition-all ${showBookmarks ? 'bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900' : 'text-ink-400 dark:text-zinc-500 hover:text-ink-900 dark:hover:text-zinc-100 hover:bg-paper-100 dark:hover:bg-zinc-800'}`}
                  title="书签与批注"
                >
                  <BookmarkIcon className="w-4 h-4" />
                </button>
                {hasAIAccess && (
                  <button onClick={toggleAISidebar} className="p-2 text-ink-400 dark:text-zinc-500 hover:text-ink-900 dark:hover:text-zinc-100 hover:bg-paper-100 dark:hover:bg-zinc-800 rounded-lg transition-all">
                    <Maximize2 className="w-4 h-4" />
                  </button>
                )}
            </div>
          </div>
        </div>
      </div>

      <div className="flex-1 flex overflow-hidden relative">
        {/* Logic Panel */}
        {showHierarchy && (
          <>
            <div
              className="fixed inset-0 z-40 bg-black/30 dark:bg-black/60 backdrop-blur-sm lg:hidden"
              onClick={() => setShowHierarchy(false)}
            ></div>
            <div className="fixed inset-y-0 left-0 z-50 w-[min(90vw,20rem)] bg-paper-50 dark:bg-zinc-900 border-r border-paper-200 dark:border-zinc-800 overflow-y-auto p-6 space-y-8 animate-in slide-in-from-left duration-300 shrink-0 custom-scrollbar shadow-2xl transition-colors lg:static lg:z-10 lg:w-80">
             {enabledPlugins.length > 0 && (
               <div className="bg-brand-50 dark:bg-brand-900/10 p-4 rounded-2xl border border-brand-100 dark:border-brand-900/30">
                 <h4 className="text-[10px] font-black text-brand-600 dark:text-brand-400 uppercase tracking-widest mb-3 flex items-center gap-2">
                   <Puzzle className="w-3 h-3"/> 已挂载插件
                 </h4>
                 <div className="space-y-2">
                   {enabledPlugins.map(p => (
                     <div key={p.id} className="space-y-1">
                       <p className="text-[9px] font-black text-ink-300 dark:text-zinc-600 uppercase px-1">{p.name}</p>
                       {p.capabilities.map(cap => (
                         <button 
                           key={cap.id}
                           onClick={() => handleRunPluginAction(p.id, cap.id)}
                           disabled={!!isPluginRunning}
                           className="w-full flex items-center justify-between p-2.5 bg-white dark:bg-zinc-950 hover:bg-brand-500 hover:text-white dark:hover:bg-brand-600 transition-all rounded-xl border border-paper-100 dark:border-zinc-800 group shadow-sm"
                         >
                           <div className="flex items-center gap-2">
                             {isPluginRunning === cap.id ? <Loader2 className="w-3 h-3 animate-spin"/> : <Play className="w-3 h-3 text-brand-500 group-hover:text-white"/>}
                             <span className="text-[11px] font-bold">{cap.name}</span>
                           </div>
                           <ChevronRight className="w-3 h-3 opacity-30"/>
                         </button>
                       ))}
                     </div>
                   ))}
                 </div>
                 {pluginError && (
                   <div className="mt-4 p-3 bg-rose-50 dark:bg-rose-950/30 border border-rose-100 dark:border-rose-900/40 rounded-xl flex items-start gap-2 animate-in fade-in duration-300">
                     <AlertCircle className="w-3.5 h-3.5 text-rose-500 shrink-0 mt-0.5" />
                     <p className="text-[10px] font-bold text-rose-600 dark:text-rose-400 leading-tight">{pluginError}</p>
                   </div>
                 )}
               </div>
             )}

             <div>
               <h4 className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest mb-3 flex items-center gap-2">
                 <Target className="w-3 h-3"/> 核心目标 (Goal)
               </h4>
               <textarea 
                  value={activeDoc.chapterGoal || ''} 
                  onChange={e => updateDocument(activeDoc.id, { chapterGoal: e.target.value })}
                  className="w-full bg-white dark:bg-zinc-950 border border-paper-200 dark:border-zinc-800 rounded-xl p-3 text-xs text-ink-800 dark:text-zinc-300 outline-none focus:border-ink-400 dark:focus:border-zinc-500 transition-colors h-24 resize-none leading-relaxed shadow-inner"
                  placeholder="本章目标..."
               />
             </div>
             <div>
               <h4 className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest mb-3 flex items-center gap-2">
                 <Layout className="w-3 h-3"/> 核心情节 (Plot)
               </h4>
               <textarea 
                  value={activeDoc.corePlot || ''} 
                  onChange={e => updateDocument(activeDoc.id, { corePlot: e.target.value })}
                  className="w-full bg-white dark:bg-zinc-950 border border-paper-200 dark:border-zinc-800 rounded-xl p-3 text-xs text-ink-800 dark:text-zinc-300 outline-none focus:border-brand-300 dark:focus:border-zinc-500 transition-colors h-32 resize-none leading-relaxed shadow-inner"
                  placeholder="核心情节..."
               />
             </div>
             <div className="space-y-4">
                <div>
                   <h4 className="text-[10px] font-black text-amber-500 dark:text-amber-600 uppercase tracking-widest mb-2 flex items-center gap-2">
                     <Route className="w-3 h-3"/> 因果链 (Logic)
                   </h4>
                   <input 
                      value={activeDoc.causeEffect || ''} 
                      onChange={e => updateDocument(activeDoc.id, { causeEffect: e.target.value })}
                      className="w-full bg-paper-100 dark:bg-zinc-950 border border-paper-200 dark:border-zinc-800 rounded-lg p-2.5 text-xs text-ink-800 dark:text-zinc-300 outline-none focus:border-amber-300 dark:focus:border-amber-700 transition-colors"
                      placeholder="因果衔接..."
                   />
                </div>
                <div>
                   <h4 className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest mb-2 flex items-center gap-2">
                     <Anchor className="w-3 h-3"/> 细节伏笔 (Details)
                   </h4>
                   <textarea 
                      value={activeDoc.foreshadowingDetails || ''} 
                      onChange={e => updateDocument(activeDoc.id, { foreshadowingDetails: e.target.value })}
                      className="w-full bg-white dark:bg-zinc-950 border border-paper-200 dark:border-zinc-800 rounded-xl p-3 text-xs text-ink-800 dark:text-zinc-300 outline-none focus:border-ink-400 dark:focus:border-zinc-500 transition-colors h-20 resize-none leading-relaxed shadow-inner"
                      placeholder="伏笔细节..."
                   />
                </div>
             </div>
             <div className="pb-6">
               <h4 className="text-[10px] font-black text-rose-500 dark:text-rose-600 uppercase tracking-widest mb-3 flex items-center gap-2">
                 <GitMerge className="w-3 h-3"/> 结尾钩子 (Hook)
               </h4>
               <textarea 
                  value={activeDoc.hook || ''} 
                  onChange={e => updateDocument(activeDoc.id, { hook: e.target.value })}
                  className="w-full bg-paper-100 dark:bg-zinc-950 border border-paper-200 dark:border-zinc-800 rounded-xl p-3 text-xs text-ink-800 dark:text-zinc-300 outline-none focus:border-rose-300 dark:focus:border-rose-800 transition-colors h-20 resize-none leading-relaxed shadow-inner"
                  placeholder="悬念钩子..."
               />
             </div>
            </div>
          </>
        )}

        {/* Editor Scroller */}
        <div id="editor-scroll-body" className="flex-1 overflow-y-auto bg-dot-pattern scroll-smooth transition-colors">
             <div className="w-full mx-auto py-6 sm:py-8 md:py-10 pb-40 sm:pb-48 flex justify-center px-[5vw]">
                 <div className="w-full bg-white dark:bg-zinc-900 min-h-[75vh] px-4 sm:px-6 md:px-8 lg:px-10 py-6 sm:py-8 md:py-10 relative shadow-[0_30px_100px_rgba(0,0,0,0.05)] dark:shadow-[0_30px_100px_rgba(0,0,0,0.5)] border border-paper-200 dark:border-zinc-800/50 rounded-lg transition-colors">
                    
                    <div className="mb-6 text-center space-y-3">
                        <div className="flex items-center justify-center gap-2 opacity-50 hover:opacity-100 transition-opacity">
                            <input 
                                value={activeDoc.timeNode || ''} 
                                onChange={e => updateDocument(activeDoc.id, { timeNode: e.target.value })}
                                className="text-center text-[10px] font-serif text-ink-400 dark:text-zinc-400 bg-transparent outline-none border-b border-transparent hover:border-paper-200 dark:hover:border-zinc-700 focus:border-ink-500 dark:focus:border-zinc-500 transition-all w-48"
                                placeholder="氛围/时间坐标..."
                            />
                            {hasAIAccess && (
                              <button onClick={handleGenerateTimeNode} disabled={isGeneratingTime} className="text-ink-400 dark:text-zinc-500 hover:text-ink-900 dark:hover:text-zinc-200 transition-colors">
                                  {isGeneratingTime ? <Loader2 className="w-3 h-3 animate-spin"/> : <Sparkles className="w-3 h-3"/>}
                              </button>
                            )}
                        </div>
                        
                        <input 
                            value={activeDoc.title}
                            onChange={e => updateDocument(activeDoc.id, { title: e.target.value })}
                            className="text-lg sm:text-xl md:text-2xl font-black text-ink-900 dark:text-zinc-100 text-center w-full bg-transparent outline-none font-serif tracking-tight placeholder:text-paper-100 dark:placeholder:text-zinc-800"
                            placeholder="章节标题"
                        />
                        
                        <div className="flex justify-center gap-2 flex-wrap">
                            {activeDoc.linkedIds.map(link => (
                                <span key={link.targetId} className="px-3 py-1 bg-paper-100 dark:bg-zinc-800 text-[10px] font-bold text-ink-500 dark:text-zinc-500 rounded-full border border-paper-200 dark:border-zinc-700">
                                    {project.entities.find(e => e.id === link.targetId)?.title}
                                </span>
                            ))}
                        </div>
                    </div>

                    <textarea
                        ref={textareaRef}
                        value={content}
                        onChange={(e) => {
                            setContent(e.target.value);
                            updateDocument(activeDoc.id, { content: e.target.value });
                        }}
                        onPaste={handlePaste}
                        className="w-full bg-transparent resize-none outline-none font-serif text-sm sm:text-base md:text-lg leading-[1.7] text-ink-800 dark:text-zinc-300 placeholder:text-paper-100 dark:placeholder:text-zinc-800 selection:bg-paper-200 dark:selection:bg-zinc-700 selection:text-ink-900 dark:selection:text-white overflow-hidden subpixel-antialiased"
                        style={{
                             textRendering: 'optimizeLegibility',
                             WebkitFontSmoothing: 'antialiased',
                             fontFeatureSettings: '"kern", "liga", "clig", "calt"',
                        }}
                        placeholder="笔尖沙沙作响..."
                        spellCheck={false}
                    />
                </div>
            </div>

        {/* AI Control Dock */}
        {hasAIAccess && (
          <div className="absolute bottom-4 sm:bottom-8 left-4 right-4 z-40 max-w-lg mx-auto">
              <div className="bg-white/90 dark:bg-zinc-900/90 backdrop-blur-2xl shadow-2xl border border-paper-200 dark:border-white/5 rounded-2xl p-1.5 flex flex-col gap-2 ring-1 ring-black/5 transition-all">
                  <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between px-2 pt-1 gap-2">
                      <div className="flex gap-1 overflow-x-auto scrollbar-hide">
                          <button onClick={() => setWordCountRequest('normal')} className={`px-3 py-1 text-[10px] font-bold rounded-lg transition-all ${wordCountRequest === 'normal' ? 'bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900' : 'text-ink-400 dark:text-zinc-500 hover:text-ink-900 dark:hover:text-zinc-200'}`}>推进剧情</button>
                          <button onClick={() => setWordCountRequest('expand')} className={`px-3 py-1 text-[10px] font-bold rounded-lg transition-all ${wordCountRequest === 'expand' ? 'bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900' : 'text-ink-400 dark:text-zinc-500 hover:text-ink-900 dark:hover:text-zinc-200'}`}>细节扩写</button>
                          <button onClick={() => setWordCountRequest('finish')} className={`px-3 py-1 text-[10px] font-bold rounded-lg transition-all ${wordCountRequest === 'finish' ? 'bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900' : 'text-ink-400 dark:text-zinc-500 hover:text-ink-900 dark:hover:text-zinc-200'}`}>章节收尾</button>
                      </div>
                      <div className="flex items-center gap-2">
                        <button 
                          onClick={handlePolishChapter}
                          disabled={isPolishing}
                          className="flex items-center gap-1.5 px-3 py-1 bg-emerald-600 text-white rounded-lg text-[10px] font-black uppercase tracking-widest hover:bg-emerald-500 transition-all shadow-lg disabled:opacity-60"
                        >
                          {isPolishing ? <Loader2 className="w-3 h-3 animate-spin" /> : <FileText className="w-3 h-3" />} 润色
                        </button>

                        <button
                          onClick={handleGenerateChapter}
                          disabled={isChapterGenerating}
                          className="flex items-center gap-1.5 px-3 py-1 bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-lg text-[10px] font-black uppercase tracking-widest hover:bg-black dark:hover:bg-white transition-all shadow-lg disabled:opacity-60"
                        >
                          {isChapterGenerating ? <Loader2 className="w-3 h-3 animate-spin" /> : <Sparkles className="w-3 h-3" />} 生成
                        </button>

                        <button
                          onClick={handleAnalyzeChapter}
                          disabled={isAnalyzing}
                          className="flex items-center gap-1.5 px-3 py-1 bg-brand-600 text-white rounded-lg text-[10px] font-black uppercase tracking-widest hover:bg-brand-500 transition-all shadow-lg disabled:opacity-60"
                        >
                          {isAnalyzing ? <Loader2 className="w-3 h-3 animate-spin" /> : <ShieldAlert className="w-3 h-3" />} 分析
                        </button>
                        <button 
                          onClick={() => setShowAgentWriter(true)} 
                          className="flex items-center gap-1.5 px-3 py-1 bg-brand-600 text-white rounded-lg text-[10px] font-black uppercase tracking-widest hover:bg-brand-500 transition-all shadow-lg"
                        >
                          <Bot className="w-3 h-3" /> Agent Mode
                        </button>
                      </div>
                  </div>
                      <div className="relative flex gap-1">
                          <input 
                              value={customPrompt}
                              onChange={e => setCustomPrompt(e.target.value)}
                              onKeyDown={(e) => e.key === 'Enter' && handleAIRequest()}
                              placeholder="给缪斯的指令..."
                              className="flex-1 bg-paper-50 dark:bg-zinc-800/50 border border-paper-200 dark:border-zinc-700/50 rounded-xl px-4 py-3 text-sm text-ink-900 dark:text-zinc-100 font-medium focus:bg-white dark:focus:bg-zinc-800 outline-none transition-all"
                          />
                          <button onClick={handleAIRequest} disabled={isGenerating} className="px-4 bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-xl hover:bg-black dark:hover:bg-white transition-all disabled:opacity-50 w-12 shrink-0 flex items-center justify-center">
                              {isGenerating ? <Loader2 className="w-4 h-4 animate-spin" /> : <ArrowRight className="w-4 h-4" />}
                          </button>
                      </div>
                  </div>
              </div>
            )}
        </div>

        {/* Bookmarks Panel */}
        {showBookmarks && (
          <>
            <div
              className="fixed inset-0 z-40 bg-black/30 dark:bg-black/60 backdrop-blur-sm lg:hidden"
              onClick={() => setShowBookmarks(false)}
            ></div>
            <div className="fixed top-14 bottom-0 right-0 z-50 w-[min(90vw,18rem)] bg-paper-50 dark:bg-zinc-900 border-l border-paper-200 dark:border-zinc-800 overflow-y-auto p-6 space-y-6 animate-in slide-in-from-right duration-300 shrink-0 shadow-2xl custom-scrollbar transition-colors lg:static lg:top-auto lg:bottom-auto lg:right-auto lg:z-10 lg:w-72">
            <div>
               <h4 className="text-[10px] font-black text-ink-400 dark:text-zinc-500 uppercase tracking-widest mb-4 flex items-center gap-2">
                 <BookmarkIcon className="w-3 h-3"/> 快速书签
               </h4>
               <div className="flex gap-2 mb-4">
                 <input 
                   value={bookmarkName}
                   onChange={e => setBookmarkName(e.target.value)}
                   placeholder="书签名称..."
                   className="flex-1 bg-white dark:bg-zinc-950 border border-paper-200 dark:border-zinc-800 rounded-lg px-3 py-2 text-xs text-ink-900 dark:text-zinc-100 outline-none focus:border-ink-400 dark:focus:border-zinc-500"
                   onKeyDown={e => e.key === 'Enter' && handleAddBookmark()}
                 />
                 <button onClick={handleAddBookmark} disabled={!bookmarkName.trim()} className="p-2 bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-lg">
                   <Plus className="w-4 h-4" />
                 </button>
               </div>
               <div className="space-y-2">
                 {activeDoc.bookmarks.length === 0 && <p className="text-xs text-ink-400 dark:text-zinc-700 text-center py-4 italic">暂无书签</p>}
                 {activeDoc.bookmarks.map(bm => (
                   <div key={bm.id} className="group flex items-center justify-between p-3 bg-white dark:bg-zinc-800/50 border border-paper-200 dark:border-zinc-800 rounded-xl hover:bg-paper-100 dark:hover:bg-zinc-800 transition-all cursor-pointer shadow-sm" onClick={() => jumpToBookmark(bm.position)}>
                      <div className="flex items-center gap-3 truncate">
                        <div className="w-1 h-6 bg-brand-600 rounded-full shrink-0"></div>
                        <span className="text-xs font-bold text-ink-900 dark:text-zinc-300 truncate">{bm.name}</span>
                      </div>
                      <button onClick={(e) => { e.stopPropagation(); deleteBookmark(activeDoc.id, bm.id); }} className="opacity-0 group-hover:opacity-100 text-ink-300 dark:text-zinc-500 hover:text-rose-500 transition-opacity">
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                   </div>
                 ))}
               </div>
            </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
};
