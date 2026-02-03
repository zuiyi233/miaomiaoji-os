
import React, { useState, useEffect, useRef } from 'react';
import { useProject } from '../contexts/ProjectContext';
import { agentPlanChapter, agentDraftScene, agentValidateScene, agentRefineScene } from '../services/agentService';
import { SceneNode, AgentStage, Document } from '../types';
import { 
  Bot, Play, Pause, CheckCircle2, Circle, AlertCircle, RefreshCw, 
  ChevronRight, Terminal, BookOpen, PenTool, BrainCircuit, Activity,
  Maximize2, Minimize2, X
} from 'lucide-react';

interface AgentWriterProps {
  onClose: () => void;
  onAppendContent: (text: string) => void;
  activeDoc: Document;
}

export const AgentWriter: React.FC<AgentWriterProps> = ({ onClose, onAppendContent, activeDoc }) => {
  const { project } = useProject();
  const [stage, setStage] = useState<AgentStage>('idle');
  const [scenes, setScenes] = useState<SceneNode[]>([]);
  const [currentSceneIdx, setCurrentSceneIdx] = useState(0);
  const [logs, setLogs] = useState<{msg: string, type: 'info'|'success'|'error'}[]>([]);
  const [isPaused, setIsPaused] = useState(false);
  const [userInstruction, setUserInstruction] = useState('');
  
  const [isExpanded, setIsExpanded] = useState(false);
  const logsEndRef = useRef<HTMLDivElement>(null);

  if (!project) return null;

  const addLog = (msg: string, type: 'info'|'success'|'error' = 'info') => {
    setLogs(prev => [...prev, {msg, type}]);
    setTimeout(() => logsEndRef.current?.scrollIntoView({ behavior: 'smooth' }), 100);
  };

  const startAgent = async () => {
    if (stage !== 'idle' && stage !== 'completed' && stage !== 'error') return;
    
    setStage('planning');
    setScenes([]);
    setCurrentSceneIdx(0);
    setLogs([]);
    addLog("Agent initialized. Analyzing chapter context...", 'info');

    // 1. Plan
    try {
      addLog("Node: Planning... Designing scene structure.", 'info');
      const plannedScenes = await agentPlanChapter(
        project, 
        activeDoc, 
        project.volumes.find(v => v.id === activeDoc.volumeId), 
        "暂无前文摘要", // In real app, fetch prev doc summary
        userInstruction
      );
      
      if (plannedScenes.length === 0) {
        throw new Error("Failed to generate scene plan.");
      }
      
      setScenes(plannedScenes);
      addLog(`Plan generated: ${plannedScenes.length} scenes created.`, 'success');
      setStage('drafting');
      
      // Start Execution Loop
      executeLoop(plannedScenes, 0);
      
    } catch (e) {
      addLog(`Error during planning: ${e}`, 'error');
      setStage('error');
    }
  };

  const executeLoop = async (currentScenes: SceneNode[], idx: number) => {
    if (idx >= currentScenes.length) {
      setStage('completed');
      addLog("All scenes completed. Workflow finished.", 'success');
      return;
    }

    if (isPaused) {
      addLog("Workflow paused by user.", 'info');
      return;
    }

    setCurrentSceneIdx(idx);
    const scene = currentScenes[idx];
    
    // Update scene status
    updateSceneStatus(idx, 'drafting', currentScenes);

    try {
      // 2. Draft
      addLog(`[Scene ${idx+1}] Node: Drafting... "${scene.title}"`, 'info');
      const prevContext = currentScenes.slice(0, idx).map(s => s.content).join('\n').slice(-1000) || activeDoc.content.slice(-1000);
      let content = await agentDraftScene(project, scene, prevContext, project.entities);
      
      updateSceneStatus(idx, 'validating', currentScenes, content);
      
      // 3. Validate
      addLog(`[Scene ${idx+1}] Node: Validating... Checking logic & quality.`, 'info');
      const validation = await agentValidateScene(content, scene.beat, project.aiSettings);
      
      if (!validation.passed) {
        addLog(`[Scene ${idx+1}] Quality Check Failed (Score: ${validation.score}). Issues: ${validation.issues.join(', ')}`, 'error');
        
        // 4. Refine (Loop back)
        updateSceneStatus(idx, 'refining', currentScenes);
        addLog(`[Scene ${idx+1}] Node: Refining... Optimizing content based on feedback.`, 'info');
        content = await agentRefineScene(content, validation, project.aiSettings);
        addLog(`[Scene ${idx+1}] Refinement complete.`, 'success');
      } else {
        addLog(`[Scene ${idx+1}] Quality Check Passed (Score: ${validation.score}).`, 'success');
      }

      // Finalize Scene
      updateSceneStatus(idx, 'completed', currentScenes, content);
      
      // Append to Editor immediately
      onAppendContent(`\n\n### ${scene.title}\n\n${content}`);
      
      // Next Scene
      executeLoop(currentScenes, idx + 1);

    } catch (e) {
      addLog(`Error in Scene ${idx+1}: ${e}`, 'error');
      updateSceneStatus(idx, 'failed', currentScenes);
      setStage('error');
    }
  };

  const updateSceneStatus = (idx: number, status: SceneNode['status'], list: SceneNode[], content?: string) => {
    const updated = [...list];
    updated[idx] = { ...updated[idx], status, content: content || updated[idx].content };
    setScenes(updated);
  };

  return (
    <div className={`fixed bottom-4 right-4 z-50 bg-ink-900/95 dark:bg-zinc-900/95 backdrop-blur-xl border border-white/10 text-white rounded-3xl shadow-2xl transition-all duration-500 flex flex-col overflow-hidden ${isExpanded ? 'w-[600px] h-[80vh]' : 'w-[400px] h-[500px]'}`}>
      
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-white/10 bg-white/5">
        <div className="flex items-center gap-3">
          <div className={`p-2 rounded-xl ${stage === 'idle' ? 'bg-white/10' : stage === 'error' ? 'bg-red-500' : 'bg-brand-500 animate-pulse'}`}>
            <Bot className="w-5 h-5" />
          </div>
          <div>
            <h3 className="text-sm font-black uppercase tracking-widest">Deep Creation Agent</h3>
            <p className="text-[10px] opacity-60 font-mono">State: {stage.toUpperCase()}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => setIsExpanded(!isExpanded)} className="p-2 hover:bg-white/10 rounded-lg transition-all">
            {isExpanded ? <Minimize2 className="w-4 h-4"/> : <Maximize2 className="w-4 h-4"/>}
          </button>
          <button onClick={onClose} className="p-2 hover:bg-red-500/20 hover:text-red-400 rounded-lg transition-all">
            <X className="w-4 h-4"/>
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {scenes.length === 0 && stage === 'idle' ? (
          <div className="flex-1 flex flex-col items-center justify-center p-8 text-center space-y-6">
             <div className="w-16 h-16 bg-white/5 rounded-full flex items-center justify-center">
                <BrainCircuit className="w-8 h-8 opacity-50" />
             </div>
             <div className="space-y-2">
               <h4 className="text-lg font-bold">Ready to Architect</h4>
               <p className="text-xs opacity-60 max-w-xs mx-auto">Agent will decompose the chapter into scenes, draft each one, check logic, and refine automatically.</p>
             </div>
             <textarea 
               value={userInstruction}
               onChange={e => setUserInstruction(e.target.value)}
               placeholder="Additional instructions (e.g. 'Make it a cliffhanger', 'Focus on rain atmosphere')..."
               className="w-full bg-black/20 border border-white/10 rounded-xl p-3 text-xs focus:ring-1 focus:ring-brand-500 outline-none resize-none h-20"
             />
             <button onClick={startAgent} className="px-8 py-3 bg-brand-600 hover:bg-brand-500 rounded-2xl text-xs font-black uppercase tracking-widest transition-all shadow-lg flex items-center gap-2">
               <Play className="w-4 h-4" /> Start Agent Loop
             </button>
          </div>
        ) : (
          <div className="flex-1 flex flex-col">
             {/* Scene Graph Visualization */}
             <div className="flex-1 overflow-y-auto p-4 space-y-3 custom-scrollbar bg-black/20">
                {scenes.map((scene, i) => (
                  <div key={scene.id} className={`p-4 rounded-2xl border transition-all ${i === currentSceneIdx ? 'bg-white/10 border-brand-500/50' : 'bg-transparent border-white/5 opacity-60'}`}>
                     <div className="flex justify-between items-center mb-2">
                        <span className="text-[10px] font-black opacity-50 uppercase tracking-widest">Scene {i + 1}</span>
                        <div className="flex items-center gap-2">
                           {scene.status === 'completed' && <span className="text-[9px] bg-green-500/20 text-green-400 px-2 py-0.5 rounded-full font-bold">DONE</span>}
                           {scene.status === 'drafting' && <span className="text-[9px] bg-blue-500/20 text-blue-400 px-2 py-0.5 rounded-full font-bold flex items-center gap-1"><PenTool className="w-3 h-3"/> WRITING</span>}
                           {scene.status === 'validating' && <span className="text-[9px] bg-amber-500/20 text-amber-400 px-2 py-0.5 rounded-full font-bold flex items-center gap-1"><Activity className="w-3 h-3"/> CHECKING</span>}
                           {scene.status === 'refining' && <span className="text-[9px] bg-purple-500/20 text-purple-400 px-2 py-0.5 rounded-full font-bold flex items-center gap-1"><RefreshCw className="w-3 h-3"/> REFINING</span>}
                        </div>
                     </div>
                     <h4 className="text-sm font-bold mb-1">{scene.title}</h4>
                     <p className="text-[10px] opacity-60 line-clamp-2">{scene.beat}</p>
                  </div>
                ))}
             </div>

             {/* Live Terminal */}
             <div className="h-40 bg-black/40 border-t border-white/10 p-4 overflow-y-auto font-mono text-[10px] space-y-1.5 custom-scrollbar">
                <div className="flex items-center gap-2 opacity-50 mb-2 sticky top-0 bg-transparent">
                   <Terminal className="w-3 h-3" /> Agent Live Logs
                </div>
                {logs.map((log, i) => (
                  <div key={i} className={`${log.type === 'error' ? 'text-red-400' : log.type === 'success' ? 'text-green-400' : 'text-zinc-400'}`}>
                    <span className="opacity-30 mr-2">[{new Date().toLocaleTimeString()}]</span>
                    {log.msg}
                  </div>
                ))}
                <div ref={logsEndRef} />
             </div>
          </div>
        )}
      </div>
    </div>
  );
};
