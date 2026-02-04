
import React, { useState } from 'react';
import { useProject } from '../contexts/ProjectContext';
import { useAuth } from '../contexts/AuthContext';
import { ViewMode } from '../types';
import { Book, Layout, Globe, FileText, Settings, Plus, Trash2, HelpCircle, Link, Check, Sparkles, ChevronDown, ChevronRight, Folder, Home, Moon, Sun, Puzzle, LogOut, Shield, UserCircle, Activity, Archive, Database, Coins } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { UserGuide } from './UserGuide';
import { useConfirm } from '../contexts/ConfirmContext';

export const Sidebar: React.FC = () => {
  const { project, activeDocumentId, setActiveDocumentId, viewMode, setViewMode, addDocument, deleteDocument, addVolume, deleteVolume, exitProject, theme, setTheme } = useProject();
  const navigate = useNavigate();
  const navigateTo = (path: string, mode: ViewMode) => {
    setViewMode(mode);
    navigate(path, { replace: true });
  };
  const { user, logout } = useAuth();
  const { confirm } = useConfirm();
  const [showGuide, setShowGuide] = useState(false);
  const [expandedVolumes, setExpandedVolumes] = useState<Record<string, boolean>>(() => {
    if (!project) return {};
    const initial: Record<string, boolean> = {};
    project.volumes.forEach(v => { initial[v.id] = true; });
    return initial;
  });

  const toggleVolume = (id: string) => {
    setExpandedVolumes(prev => ({ ...prev, [id]: !prev[id] }));
  };

  const navItemClass = (isActive: boolean) => 
    `flex items-center w-full px-3 py-2.5 mb-1 text-sm font-medium rounded-lg transition-all ${
      isActive ? 'bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 shadow-md' : 'text-ink-500 dark:text-zinc-400 hover:bg-paper-100 dark:hover:bg-zinc-800 hover:text-ink-900 dark:hover:text-zinc-100'
    }`;

  const docItemClass = (isActive: boolean) =>
    `group flex items-center justify-between w-full pl-8 pr-3 py-2 text-xs transition-colors border-l ${
      isActive ? 'border-ink-900 dark:border-zinc-100 bg-paper-100 dark:bg-zinc-800 text-ink-900 dark:text-zinc-100 font-bold' : 'border-transparent text-ink-400 dark:text-zinc-500 hover:bg-paper-50 dark:hover:bg-zinc-800/50 hover:text-ink-900 dark:hover:text-zinc-100'
    }`;

  return (
    <div className="w-full md:w-64 h-full bg-paper-50 dark:bg-zinc-900 border-r border-paper-200 dark:border-zinc-800 flex flex-col flex-shrink-0 relative transition-colors duration-300">
      {showGuide && <UserGuide onClose={() => setShowGuide(false)} />}
      
      {/* Project Title Section */}
      <div className="p-4 sm:p-5 border-b border-paper-200 dark:border-zinc-800">
        {project && (
          <>
            <button 
              onClick={exitProject}
              className="mb-4 flex items-center gap-2 text-[10px] font-bold text-ink-400 dark:text-zinc-500 hover:text-ink-900 dark:hover:text-zinc-100 transition-colors uppercase tracking-wider"
            >
              <Home className="w-3 h-3" /> 返回仪表盘
            </button>
            <h1 className="text-lg font-black text-ink-900 dark:text-zinc-100 tracking-tight flex items-center gap-2 font-serif">
              <Book className="w-5 h-5" />
              <span className="truncate">{project.title}</span>
            </h1>
          </>
        )}
      </div>

      <div className="p-3 flex-1 overflow-y-auto custom-scrollbar">
        <div className="mb-8">
          <h3 className="px-3 text-[10px] font-black text-ink-300 dark:text-zinc-600 uppercase tracking-widest mb-3">创作导航</h3>
          
          {project && (
            <>
               <button onClick={() => navigateTo('/writer', ViewMode.WRITER)} className={navItemClass(viewMode === ViewMode.WRITER)}>
                 <FileText className="w-4 h-4 mr-3" /> 写作手稿
               </button>
              <button onClick={() => navigateTo('/planboard', ViewMode.PLANBOARD)} className={navItemClass(viewMode === ViewMode.PLANBOARD)}>
                <Layout className="w-4 h-4 mr-3" /> 情节大纲
              </button>
              <button onClick={() => navigateTo('/world', ViewMode.WORLD)} className={navItemClass(viewMode === ViewMode.WORLD)}>
                <Globe className="w-4 h-4 mr-3" /> 世界观档案
              </button>
              <button onClick={() => navigateTo('/workflows', ViewMode.WORKFLOWS)} className={navItemClass(viewMode === ViewMode.WORKFLOWS)}>
                <Activity className="w-4 h-4 mr-3" /> 工作流会话
              </button>
              <button onClick={() => navigateTo('/files', ViewMode.FILES)} className={navItemClass(viewMode === ViewMode.FILES)}>
                <Archive className="w-4 h-4 mr-3" /> 文件中心
              </button>
              <button onClick={() => navigateTo('/corpus', ViewMode.CORPUS)} className={navItemClass(viewMode === ViewMode.CORPUS)}>
                <Database className="w-4 h-4 mr-3" /> 语料库
              </button>
              <button onClick={() => navigateTo('/settlements', ViewMode.SETTLEMENTS)} className={navItemClass(viewMode === ViewMode.SETTLEMENTS)}>
                <Coins className="w-4 h-4 mr-3" /> 积分结算
              </button>
              <button onClick={() => navigateTo('/plugins', ViewMode.PLUGINS)} className={navItemClass(viewMode === ViewMode.PLUGINS)}>
                <Puzzle className="w-4 h-4 mr-3" /> 扩展插件
              </button>
            </>
          )}
        </div>

        {project && (
          <div>
            <h3 className="px-3 text-[10px] font-black text-ink-300 dark:text-zinc-600 uppercase tracking-widest mb-3 flex justify-between items-center">
              <span>目录结构</span>
              <button onClick={() => addVolume()} className="text-ink-300 dark:text-zinc-600 hover:text-ink-900 dark:hover:text-zinc-100 transition-colors" title="添加新卷">
                <Plus className="w-3.5 h-3.5" />
              </button>
            </h3>
            <div className="space-y-1">
              {project.volumes.sort((a,b) => a.order - b.order).map(vol => (
                <div key={vol.id} className="mb-1">
                  <div className="group flex items-center justify-between px-3 py-2 hover:bg-paper-100 dark:hover:bg-zinc-800 rounded-lg cursor-pointer transition-colors">
                    <div 
                      className="flex items-center gap-2 flex-1 min-w-0" 
                      onClick={() => toggleVolume(vol.id)}
                    >
                      {expandedVolumes[vol.id] ? <ChevronDown className="w-3 h-3 text-ink-400" /> : <ChevronRight className="w-3 h-3 text-ink-400" />}
                      <Folder className="w-3.5 h-3.5 text-ink-500 dark:text-zinc-500 flex-shrink-0" />
                      <span className="text-xs font-bold text-ink-700 dark:text-zinc-300 truncate">{vol.title}</span>
                    </div>
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button 
                        onClick={() => addDocument(vol.id)} 
                        className="p-1 text-ink-400 dark:text-zinc-500 hover:text-ink-900 dark:hover:text-zinc-100 rounded"
                        title="添加章节"
                      >
                        <Plus className="w-3 h-3" />
                      </button>
                    </div>
                  </div>

                  {expandedVolumes[vol.id] && (
                    <div className="mt-1 space-y-0.5 relative">
                      <div className="absolute left-[19px] top-0 bottom-2 w-px bg-paper-200 dark:bg-zinc-800"></div>
                      {project.documents
                        .filter(d => d.volumeId === vol.id)
                        .sort((a,b) => a.order - b.order)
                        .map((doc) => (
                          <div key={doc.id} className="relative group">
                            <div className="relative group">
                              <button
                                onClick={() => {
                                  setActiveDocumentId(doc.id);
                                  setViewMode(ViewMode.WRITER);
                                }}
                                className={docItemClass(activeDocumentId === doc.id)}
                              >
                                <span className="truncate pr-4">{doc.title}</span>
                              </button>
                              <div className="absolute right-3 top-1/2 -translate-y-1/2 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                <button
                                  onClick={async (e) => {
                                    e.stopPropagation();
                                    const ok = await confirm({
                                      title: '确定删除此章节吗？',
                                      description: '删除后将无法恢复。',
                                      confirmText: '删除',
                                      cancelText: '取消',
                                      tone: 'danger',
                                    });
                                    if (ok) deleteDocument(doc.id);
                                  }}
                                  className="p-1 text-ink-300 dark:text-zinc-600 hover:text-red-500 rounded"
                                >
                                  <Trash2 className="w-3 h-3" />
                                </button>
                              </div>
                            </div>
                          </div>
                        ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

       <div className="p-4 border-t border-paper-200 dark:border-zinc-800 space-y-1 bg-paper-50 dark:bg-zinc-900">
        <button 
          onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}
          className="w-full flex items-center justify-center gap-2 py-2 text-[10px] font-black uppercase tracking-widest bg-paper-100 dark:bg-zinc-800 text-ink-600 dark:text-zinc-400 rounded-lg border border-paper-200 dark:border-zinc-700 hover:bg-paper-200 dark:hover:bg-zinc-700 transition-all mb-2"
        >
          {theme === 'light' ? <Moon className="w-3.5 h-3.5" /> : <Sun className="w-3.5 h-3.5" />}
          {theme === 'light' ? '切换深夜' : '切换白昼'}
        </button>
        <button 
          onClick={() => setShowGuide(true)}
          className="flex items-center w-full px-3 py-2 text-xs font-bold text-ink-400 dark:text-zinc-500 hover:text-ink-900 dark:hover:text-zinc-100 transition-colors"
        >
          <HelpCircle className="w-3.5 h-3.5 mr-2" /> 用户手册
        </button>
      </div>
    </div>
  );
};
