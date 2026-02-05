
import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useProject } from '../contexts/ProjectContext';
import { Type } from '@google/genai';
import { buildWorkflowPayload, DEFAULT_AI_SETTINGS } from '../services/aiService';
import { apiRequest } from '../services/apiClient';
import { upsertProjectSnapshotApi } from '../services/projectApi';
import { createSSEClient, ConnectionState, ProgressUpdatedData, StepAppendedData, WorkflowDoneData } from '../services/sseClient';
import { runChapterGenerateApi, runWizardCharactersWorkflowApi, runWizardOutlineWorkflowApi, runWizardWorldWorkflowApi } from '../services/workflowApi';
import { Project, Volume, Document, StoryEntity, AIPromptTemplate, ViewMode } from '../types';
import { Sparkles, ArrowRight, Loader2, ArrowLeft, Check, AlertCircle, RefreshCw } from 'lucide-react';

interface NovelWizardProps {
  onCancel: () => void;
}

export const NovelWizard: React.FC<NovelWizardProps> = ({ onCancel }) => {
  const { createProject, selectSession, setViewMode, project: activeProject, projects, theme } = useProject();
  
  // Use requested default proxy settings as the seed for new projects
  const wizardAISettings = { ...DEFAULT_AI_SETTINGS };

  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [sparkInput, setSparkInput] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [blueprint, setBlueprint] = useState<any>(null);

  const [backendProjectId, setBackendProjectId] = useState<number | null>(null);
  const [wizardSessionId, setWizardSessionId] = useState<number | null>(null);
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected');
  const [progress, setProgress] = useState<number | null>(null);
  const [doneInfo, setDoneInfo] = useState<{ mode: string; documentId?: number } | null>(null);
  const [wizardLog, setWizardLog] = useState<Array<{ title: string; content: string }>>([]);
  const [generatedBackendDocumentId, setGeneratedBackendDocumentId] = useState<number | null>(null);
  const sseRef = useRef<ReturnType<typeof createSSEClient> | null>(null);

  const wizardExternalId = useMemo(() => `w${Date.now()}`, []);

  const ensureBackendProject = async (title: string): Promise<number | null> => {
    if (backendProjectId && backendProjectId > 0) return backendProjectId;
    try {
      // Wizard 阶段也需要后端 project_id 承载 session/step。
      // 使用 snapshot upsert 创建/更新后端 project 记录。
      const dto = await upsertProjectSnapshotApi({
        external_id: wizardExternalId,
        title: title || '未命名项目',
        ai_settings: wizardAISettings,
        snapshot: {
          id: wizardExternalId,
          title: title || '未命名项目',
          coreConflict: '',
          characterArc: '',
          ultimateValue: '',
          worldRules: '',
          characterCore: '',
          symbolSettings: '',
          aiSettings: wizardAISettings,
          volumes: [],
          documents: [],
          entities: [],
          templates: [],
        },
      });
      if (dto?.id) {
        setBackendProjectId(dto.id);
        return dto.id;
      }
      return null;
    } catch {
      return null;
    }
  };

  const parseJSONSafe = (text: string): any | null => {
    const raw = (text || '').trim();
    if (!raw) return null;
    try {
      return JSON.parse(raw);
    } catch {
      const match = raw.match(/\{[\s\S]*\}/);
      if (!match) return null;
      try {
        return JSON.parse(match[0]);
      } catch {
        return null;
      }
    }
  };

  useEffect(() => {
    if (!wizardSessionId || wizardSessionId <= 0) return;
    const client = createSSEClient({
      sessionId: wizardSessionId,
      onStateChange: (state) => setConnectionState(state),
      onStepAppended: (data: StepAppendedData) => {
        setWizardLog((prev) => {
          const next = [...prev];
          next.push({ title: data.title || '步骤', content: typeof data.content === 'string' ? data.content : '' });
          return next.slice(-8);
        });
      },
      onProgressUpdated: (data: ProgressUpdatedData) => {
        if (typeof data.progress === 'number') {
          setProgress(Math.max(0, Math.min(100, Math.round(data.progress))));
        }
      },
      onWorkflowDone: (data: WorkflowDoneData) => {
        setDoneInfo({ mode: data.mode || 'wizard', documentId: data.document_id });
        setProgress(100);
      },
    });
    sseRef.current = client;
    client.connect();
    return () => {
      client.disconnect();
      sseRef.current = null;
    };
  }, [wizardSessionId]);

  const handleGenerate = async () => {
    if (!sparkInput.trim()) return;
    setIsGenerating(true);
    setWizardSessionId(null);
    setBlueprint(null);
    setWizardLog([]);
    setDoneInfo(null);
    setProgress(null);
    setGeneratedBackendDocumentId(null);
    try {
      const projectId = await ensureBackendProject('新建项目');
      if (!projectId) {
        throw new Error('项目未同步到后端');
      }

      // Step 1: 世界观
      const worldSchema = {
        type: Type.OBJECT,
        properties: {
          title: { type: Type.STRING },
          coreConflict: { type: Type.STRING },
          characterArc: { type: Type.STRING },
          ultimateValue: { type: Type.STRING },
          worldRules: { type: Type.STRING },
          characterCore: { type: Type.STRING },
          symbolSettings: { type: Type.STRING },
        },
        required: ['title', 'coreConflict', 'worldRules'],
      };
      const worldSystem = '你是一位世界级的小说架构师。请输出结构化 JSON，不要输出多余文字。';
      const worldPrompt = `用户创意核心：${sparkInput}\n\n请生成小说核心设定蓝图。`;
      const worldPayload = buildWorkflowPayload(worldPrompt, worldSystem, wizardAISettings, worldSchema);
      const worldResult = await runWizardWorldWorkflowApi({
        project_id: projectId,
        session_id: wizardSessionId || undefined,
        title: '向导 · 小说蓝图',
        step_title: '世界观蓝图',
        provider: worldPayload.provider,
        path: worldPayload.path,
        body: worldPayload.body,
      });
      if (worldResult?.session?.id && !wizardSessionId) {
        setWizardSessionId(worldResult.session.id);
		selectSession(String(worldResult.session.id));
      }
      const sessionId = worldResult?.session?.id || wizardSessionId || undefined;
      const worldJSON = parseJSONSafe(worldResult?.content || '');
      if (!worldJSON) {
        throw new Error('世界观输出解析失败');
      }

      // Step 2: 角色
      const characterSchema = {
        type: Type.OBJECT,
        properties: {
          protagonistName: { type: Type.STRING },
          protagonistDesc: { type: Type.STRING },
        },
        required: ['protagonistName', 'protagonistDesc'],
      };
      const characterSystem = '你是一位角色塑造专家。请输出结构化 JSON，不要输出多余文字。';
      const characterPrompt = `书名：${worldJSON.title}\n核心冲突：${worldJSON.coreConflict}\n世界规则：${worldJSON.worldRules}\n\n请为主角设定姓名与简要介绍。`;
      const characterPayload = buildWorkflowPayload(characterPrompt, characterSystem, wizardAISettings, characterSchema);
      const characterResult = await runWizardCharactersWorkflowApi({
        project_id: projectId,
        session_id: sessionId,
        title: '向导 · 小说蓝图',
        step_title: '主角设定',
        provider: characterPayload.provider,
        path: characterPayload.path,
        body: characterPayload.body,
      });
      const characterJSON = parseJSONSafe(characterResult?.content || '');
      if (!characterJSON) {
        throw new Error('角色输出解析失败');
      }

      // Step 3: 大纲（第一卷/第一章标题）
      const outlineSchema = {
        type: Type.OBJECT,
        properties: {
          firstVolumeTitle: { type: Type.STRING },
          firstVolumeGoal: { type: Type.STRING },
          firstChapterTitle: { type: Type.STRING },
        },
        required: ['firstVolumeTitle', 'firstChapterTitle'],
      };
      const outlineSystem = '你是一位精通故事结构的小说策划。请输出结构化 JSON，不要输出多余文字。';
      const outlinePrompt = `书名：${worldJSON.title}\n主角：${characterJSON.protagonistName}\n核心冲突：${worldJSON.coreConflict}\n\n请生成第一卷标题/目标，以及第一章标题。`;
      const outlinePayload = buildWorkflowPayload(outlinePrompt, outlineSystem, wizardAISettings, outlineSchema);
      const outlineResult = await runWizardOutlineWorkflowApi({
        project_id: projectId,
        session_id: sessionId,
        title: '向导 · 小说蓝图',
        step_title: '第一卷与第一章',
        provider: outlinePayload.provider,
        path: outlinePayload.path,
        body: outlinePayload.body,
      });
      const outlineJSON = parseJSONSafe(outlineResult?.content || '');
      if (!outlineJSON) {
        throw new Error('大纲输出解析失败');
      }

      const merged = {
        ...worldJSON,
        ...characterJSON,
        ...outlineJSON,
      };

		// Step 4: 创建后端卷与文档，并写回第一章正文
		type VolumeDTO = { id: number; title: string };
		type DocumentDTO = { id: number; title: string };
		const volume = await apiRequest<VolumeDTO>(`/api/v1/projects/${projectId}/volumes`, {
			method: 'POST',
			body: JSON.stringify({
				title: merged.firstVolumeTitle || '第一卷',
				order_index: 1,
			}),
		});
		const backendDoc = await apiRequest<DocumentDTO>(`/api/v1/projects/${projectId}/documents`, {
			method: 'POST',
			body: JSON.stringify({
				title: merged.firstChapterTitle || '第一章',
				content: '',
				status: '草稿',
				order_index: 1,
				volume_id: volume?.id,
			}),
		});

		const chapterSystem = '你是一位职业小说作者。根据给定的设定与章节标题，写出完整第一章正文。只输出正文。';
		const chapterPrompt = `书名：${merged.title}\n核心冲突：${merged.coreConflict}\n世界规则：${merged.worldRules}\n主角：${merged.protagonistName}\n主角简介：${merged.protagonistDesc}\n\n第一卷：${merged.firstVolumeTitle}\n第一卷目标：${merged.firstVolumeGoal}\n\n第一章标题：${merged.firstChapterTitle}\n\n请写出第一章正文。`;
		const chapterPayload = buildWorkflowPayload(chapterPrompt, chapterSystem, wizardAISettings);
		const chapterResult = await runChapterGenerateApi({
			project_id: projectId,
			session_id: sessionId,
			document_id: backendDoc?.id,
			volume_id: volume?.id,
			title: merged.firstChapterTitle,
			order_index: 1,
			provider: chapterPayload.provider,
			path: chapterPayload.path,
			body: chapterPayload.body,
			write_back: { set_status: '草稿' },
		});
		const chapterText = (chapterResult?.content || '').trim();
		if (chapterText) {
			(merged as any).firstChapterContent = chapterText;
		}
		if (backendDoc?.id) {
			setGeneratedBackendDocumentId(backendDoc.id);
		}

      setBlueprint(merged);
      setStep(2);
    } catch (e) {
      alert("生成失败，请确认后端供应商配置是否已正确设置。");
    }
    setIsGenerating(false);
  };

  const handleFinalize = () => {
    if (!blueprint) return;

    const projectId = wizardExternalId;
    const volId = `v${Date.now()}`;
    const docId = `d${Date.now()}`;
    const entityId = `e${Date.now()}`;

    const newProject: Project = {
      id: projectId,
      title: blueprint.title,
      coreConflict: blueprint.coreConflict,
      characterArc: blueprint.characterArc || '',
      ultimateValue: blueprint.ultimateValue || '',
      worldRules: blueprint.worldRules || '',
      characterCore: blueprint.characterCore || '',
      symbolSettings: blueprint.symbolSettings || '',
      aiSettings: wizardAISettings,
      volumes: [{
        id: volId,
        title: blueprint.firstVolumeTitle || '第一卷',
        order: 0,
        theme: '',
        coreGoal: blueprint.firstVolumeGoal || '',
        boundaries: '',
        chapterLinkageLogic: '',
        volumeSpecificSettings: '',
        plotRoadmap: ''
      }],
      documents: [{
        id: docId,
		backendId: generatedBackendDocumentId || undefined,
        volumeId: volId,
        title: blueprint.firstChapterTitle || '第一章',
		content: blueprint.firstChapterContent || '',
        status: '草稿',
        order: 0,
        linkedIds: [{ targetId: entityId, type: 'character', relationName: '主角' }],
        bookmarks: [],
        chapterGoal: '',
        corePlot: '',
        hook: '',
        causeEffect: '',
        foreshadowingDetails: ''
      }],
      entities: [{
        id: entityId,
        type: 'character',
        title: blueprint.protagonistName || '主角',
        subtitle: '核心主角',
        content: blueprint.protagonistDesc || '',
        tags: ['主角'],
        linkedIds: [],
        importance: 'main'
      }],
      templates: [
         { id: `t${Date.now()}1`, name: '逻辑自查', description: '检查当前章节是否违背已建立的世界观与人设', category: 'logic', template: '检查当前章节是否违背了实体的基本设定：\n{{content}}' },
         { id: `t${Date.now()}2`, name: '核心设定完善', description: 'AI 辅助完善小说的核心设定', category: 'content', template: '根据用户提供的核心冲突、人物弧光、终极价值等信息，AI 辅助完善小说的核心设定，包括：世界基础规则、主角核心设定、以及贯穿全书的核心符号或道具设定。' },
         { id: `t${Date.now()}3`, name: '卷纲生成', description: '为当前选中的卷生成更详细的设定', category: 'content', template: '为当前选中的卷“{{volumeTitle}}”生成更详细的设定，包括：卷主题、卷核心目标、情节边界、章节联动逻辑、卷级设定补充、以及情节脉络图。' },
         { id: `t${Date.now()}4`, name: '新章构思', description: '构思并生成一个新章节', category: 'content', template: '在当前卷（{{volumeTitle}}）下，根据用户想法构思并生成一个新章节，包括标题、目标、核心情节、结尾钩子、时间节点、因果链和细节伏笔。' }
      ]
    };

    createProject(newProject);
  };

  return (
    <div className="min-h-screen bg-paper-50 dark:bg-zinc-950 text-ink-900 dark:text-zinc-100 flex flex-col items-center justify-center p-6 relative overflow-hidden transition-colors duration-300">
      {/* Background Ambience */}
      <div className="absolute top-0 left-0 w-full h-full opacity-10 dark:opacity-20 pointer-events-none">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-purple-400 dark:bg-purple-600 rounded-full blur-[128px]" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-indigo-400 dark:bg-indigo-600 rounded-full blur-[128px]" />
      </div>

      <div className="relative z-10 w-full max-w-2xl">
        <button onClick={onCancel} className="absolute -top-16 left-0 flex items-center gap-2 text-ink-400 dark:text-zinc-500 hover:text-ink-900 dark:hover:text-white transition-colors font-bold text-sm">
            <ArrowLeft className="w-4 h-4" /> 返回控制台
        </button>

        {step === 1 && (
          <div className="space-y-8 animate-in fade-in slide-in-from-bottom-8 duration-700">
            <div className="text-center space-y-4">
               <div className="inline-flex items-center justify-center p-4 bg-brand-100 dark:bg-indigo-500/20 rounded-full mb-4 ring-1 ring-brand-200 dark:ring-indigo-500/50">
                 <Sparkles className="w-8 h-8 text-brand-600 dark:text-indigo-300" />
               </div>
               <h2 className="text-4xl font-black font-serif">灵感起源</h2>
               <p className="text-ink-500 dark:text-zinc-400 max-w-lg mx-auto">告诉缪斯你想写什么。只需一个模糊的想法、一个画面，或者一个核心冲突。</p>
            </div>

            <div className="relative">
              <textarea 
                value={sparkInput}
                onChange={e => setSparkInput(e.target.value)}
                placeholder="例如：一个在古代长安城破解灵异案件的落魄道士，但他其实是外星人..."
                className="w-full h-48 bg-white dark:bg-white/5 border border-paper-200 dark:border-white/10 rounded-3xl p-6 text-lg focus:ring-2 focus:ring-brand-500 dark:focus:ring-indigo-500 focus:bg-white dark:focus:bg-white/10 outline-none resize-none transition-all placeholder:text-ink-200 dark:placeholder:text-gray-600 shadow-sm"
              />
              <div className="absolute bottom-4 right-4 text-xs text-ink-300 dark:text-gray-500 font-bold uppercase tracking-widest">
                 AI Powered Hub
              </div>
            </div>

            <button 
              onClick={handleGenerate}
              disabled={isGenerating || !sparkInput.trim()}
              className="w-full py-5 bg-gradient-to-r from-brand-600 to-indigo-600 dark:from-indigo-600 dark:to-purple-600 hover:opacity-90 rounded-2xl font-black text-lg shadow-2xl shadow-brand-100 dark:shadow-indigo-900/50 flex items-center justify-center gap-3 transition-all transform active:scale-[0.98] text-white"
            >
              {isGenerating ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" /> 正在构建世界观蓝图...
                </>
              ) : (
                <>
                  启动创世引擎 <ArrowRight className="w-5 h-5" />
                </>
              )}
            </button>

			{(isGenerating || wizardSessionId) && (
				<div className="bg-white dark:bg-white/5 border border-paper-200 dark:border-white/10 rounded-3xl p-6 space-y-4 shadow-sm">
					<div className="flex items-center justify-between">
						<div className="text-xs font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500">生成进度</div>
						<div className="text-xs font-bold text-ink-600 dark:text-zinc-300">
							{wizardSessionId ? `Session #${wizardSessionId}` : '等待会话创建...'}
						</div>
					</div>
					<div className="flex items-center gap-3">
						<div className="flex-1 h-2 bg-paper-100 dark:bg-zinc-800 rounded-full overflow-hidden border border-paper-200 dark:border-zinc-700">
							<div
								className="h-full bg-brand-600 transition-all duration-300"
								style={{ width: `${progress ?? 0}%` }}
							/>
						</div>
						<div className="text-[10px] font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500 tabular-nums">{progress ?? 0}%</div>
					</div>
					<div className="flex items-center justify-between">
						<div className="text-[10px] font-black uppercase tracking-widest text-ink-300 dark:text-zinc-600">
							连接状态：{connectionState}
						</div>
						{wizardSessionId && (
							<button
								onClick={() => {
									selectSession(String(wizardSessionId));
									setViewMode(ViewMode.WORKFLOW_DETAIL);
								}}
								className="text-[10px] font-black uppercase tracking-widest text-brand-700 dark:text-indigo-300 hover:opacity-80"
							>
								查看会话详情
							</button>
						)}
					</div>

					{wizardLog.length > 0 && (
						<div className="space-y-2">
							<div className="text-[10px] font-black uppercase tracking-widest text-ink-300 dark:text-zinc-600">实时输出</div>
							<div className="space-y-2 max-h-48 overflow-y-auto custom-scrollbar">
								{wizardLog.map((item, idx) => (
									<div key={idx} className="p-3 rounded-2xl bg-paper-50 dark:bg-zinc-900 border border-paper-100 dark:border-zinc-800">
										<div className="text-xs font-bold text-ink-800 dark:text-zinc-100">{item.title}</div>
										<pre className="mt-1 whitespace-pre-wrap text-xs text-ink-500 dark:text-zinc-400">{item.content}</pre>
									</div>
								))}
							</div>
						</div>
					)}
				</div>
			)}
          </div>
        )}

        {step === 2 && blueprint && (
          <div className="space-y-6 animate-in fade-in slide-in-from-right-8 duration-500">
             <div className="flex items-center justify-between">
                <h2 className="text-2xl font-black font-serif">蓝图确认</h2>
                <button onClick={() => setStep(1)} className="text-sm text-ink-400 dark:text-zinc-400 hover:text-ink-900 dark:hover:text-white font-bold">重新生成</button>
             </div>

             <div className="bg-white dark:bg-white/5 border border-paper-200 dark:border-white/10 rounded-3xl p-8 space-y-6 max-h-[60vh] overflow-y-auto shadow-xl">
                <div className="grid grid-cols-1 gap-6">
                   <div className="space-y-1">
                      <label className="text-[10px] font-black text-brand-600 dark:text-indigo-400 uppercase tracking-widest">书名</label>
                      <input 
                        value={blueprint.title} 
                        onChange={e => setBlueprint({...blueprint, title: e.target.value})}
                        className="w-full bg-transparent border-b border-paper-100 dark:border-gray-700 focus:border-brand-500 dark:focus:border-indigo-500 outline-none py-2 text-xl font-bold font-serif text-ink-900 dark:text-white"
                      />
                   </div>
                   <div className="space-y-1">
                      <label className="text-[10px] font-black text-brand-600 dark:text-indigo-400 uppercase tracking-widest">核心冲突</label>
                      <textarea 
                        value={blueprint.coreConflict} 
                        onChange={e => setBlueprint({...blueprint, coreConflict: e.target.value})}
                        className="w-full bg-transparent border border-paper-100 dark:border-gray-700 rounded-lg p-3 text-sm focus:border-brand-500 dark:focus:border-indigo-500 outline-none h-20 text-ink-700 dark:text-zinc-300"
                      />
                   </div>
                   <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-1">
                        <label className="text-[10px] font-black text-brand-600 dark:text-indigo-400 uppercase tracking-widest">主角姓名</label>
                        <input 
                            value={blueprint.protagonistName} 
                            onChange={e => setBlueprint({...blueprint, protagonistName: e.target.value})}
                            className="w-full bg-transparent border-b border-paper-100 dark:border-gray-700 focus:border-brand-500 dark:focus:border-indigo-500 outline-none py-1 text-sm font-bold text-ink-800 dark:text-white"
                        />
                      </div>
                      <div className="space-y-1">
                        <label className="text-[10px] font-black text-brand-600 dark:text-indigo-400 uppercase tracking-widest">世界规则摘要</label>
                        <input 
                            value={blueprint.worldRules} 
                            onChange={e => setBlueprint({...blueprint, worldRules: e.target.value})}
                            className="w-full bg-transparent border-b border-paper-100 dark:border-gray-700 focus:border-brand-500 dark:focus:border-indigo-500 outline-none py-1 text-sm font-bold truncate text-ink-800 dark:text-white"
                        />
                      </div>
                   </div>
                    <div className="p-4 bg-brand-50 dark:bg-indigo-500/10 rounded-xl border border-brand-100 dark:border-indigo-500/20">
                      <h4 className="text-xs font-bold text-brand-700 dark:text-indigo-300 mb-2">第一卷预设：{blueprint.firstVolumeTitle}</h4>
                      <p className="text-xs text-ink-500 dark:text-gray-400 leading-relaxed">{blueprint.firstVolumeGoal}</p>
                    </div>

					<div className="p-4 bg-white/70 dark:bg-zinc-900/60 rounded-xl border border-paper-200 dark:border-zinc-800">
						<div className="flex items-center justify-between gap-3">
							<h4 className="text-xs font-black uppercase tracking-widest text-ink-400 dark:text-zinc-500">第一章正文预览</h4>
							<div className="text-[10px] font-black uppercase tracking-widest text-ink-300 dark:text-zinc-600">
								{blueprint.firstChapterContent ? '已生成' : '未生成'}
							</div>
						</div>
						{blueprint.firstChapterContent ? (
							<pre className="mt-3 whitespace-pre-wrap text-sm leading-relaxed text-ink-700 dark:text-zinc-200 max-h-72 overflow-y-auto custom-scrollbar">
								{blueprint.firstChapterContent}
							</pre>
						) : (
							<div className="mt-3 text-sm text-ink-400 dark:text-zinc-400">
								第一章正文还在生成或解析失败。你仍然可以先创建项目，然后在编辑器里一键生成。
							</div>
						)}
					</div>
                 </div>
              </div>

             <div className="flex gap-4">
                <button 
                  onClick={handleFinalize}
                  className="flex-1 py-4 bg-ink-900 dark:bg-white text-white dark:text-gray-900 hover:opacity-90 rounded-2xl font-black text-sm shadow-xl flex items-center justify-center gap-2 transition-all"
                >
                  <Check className="w-4 h-4" /> 确认并创建
                </button>
             </div>
          </div>
        )}
      </div>
    </div>
  );
};
