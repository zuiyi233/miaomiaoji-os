
import React, { useState, useEffect } from 'react';
import { Sidebar } from './Sidebar';
import { Editor } from './Editor';
import { KanbanBoard } from './KanbanBoard';
import { WorldBible } from './WorldBible';
import { AIAssistant } from './AIAssistant';
import { ProjectDashboard } from './ProjectDashboard';
import { NovelWizard } from './NovelWizard';
import { PluginManager } from './PluginManager';
import { WorkflowSessions } from './WorkflowSessions';
import { WorkflowDetail } from './WorkflowDetail';
import { ConnectionState } from '../services/sseClient';
import { FileCenter } from './FileCenter';
import { CorpusCenter } from './CorpusCenter';
import { SettlementCenter } from './SettlementCenter';
import { AuthScreen } from './AuthScreen';
import { UserSettings } from './UserSettings';
import { useProject } from '../contexts/ProjectContext';
import { useAuth } from '../contexts/AuthContext';
import { ViewMode } from '../types';
import { Menu, Loader2, Shield, Settings, User as UserIcon, LayoutDashboard, ChevronDown } from 'lucide-react';
import { Routes, Route, useNavigate, useLocation, Navigate } from 'react-router-dom';

const LayoutContent: React.FC = () => {
  const { viewMode, activeProjectId, theme, setViewMode, project, isAISidebarOpen, exitProject, activeSessionId, selectSession, projectLoadError, clearProjectLoadError } = useProject();
  const { user, isLoading } = useAuth();
  const [isCreating, setIsCreating] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const isSettingsView = viewMode === ViewMode.SETTINGS;
  const showSidebar = !!activeProjectId && !isSettingsView;
  const isEditorPage = !!activeProjectId && (location.pathname === '/' || location.pathname === '/writer');

  // Sync theme with DOM
  useEffect(() => {
    if (theme === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [theme]);

  const resolveViewMode = (path: string): ViewMode | null => {
    if (path.startsWith('/workflows/')) return ViewMode.WORKFLOW_DETAIL;
    if (path === '/workflows') return ViewMode.WORKFLOWS;
    if (path === '/files') return ViewMode.FILES;
    if (path === '/corpus') return ViewMode.CORPUS;
    if (path === '/settlements') return ViewMode.SETTLEMENTS;
    if (path === '/plugins') return ViewMode.PLUGINS;
    if (path === '/world') return ViewMode.WORLD;
    if (path === '/planboard') return ViewMode.PLANBOARD;
    if (path === '/writer') return ViewMode.WRITER;
    if (path === '/settings') return ViewMode.SETTINGS;
    if (path === '/') return ViewMode.WRITER;
    return null;
  };

  useEffect(() => {
    const path = location.pathname;
    if (!activeProjectId && path !== '/' && path !== '/settings') {
      navigate('/', { replace: true });
      return;
    }
    if (path === '/' && activeProjectId) {
      return;
    }
    const nextView = resolveViewMode(path);
    if (path.startsWith('/workflows/')) {
      const sessionId = path.replace('/workflows/', '').trim();
      if (sessionId && sessionId !== activeSessionId) {
        selectSession(sessionId);
      }
    }

    if (nextView) {
      setViewMode(nextView);
    }
  }, [location.pathname, activeSessionId, selectSession, setViewMode, activeProjectId, navigate]);

  useEffect(() => {
    const target = (() => {
      if (viewMode === ViewMode.WORKFLOW_DETAIL) {
        return activeSessionId ? `/workflows/${activeSessionId}` : '/workflows';
      }
      if (viewMode === ViewMode.WRITER) {
        return activeProjectId ? '/writer' : '/';
      }
      const map: Record<ViewMode, string> = {
        [ViewMode.PLANBOARD]: '/planboard',
        [ViewMode.WORLD]: '/world',
        [ViewMode.PLUGINS]: '/plugins',
        [ViewMode.FILES]: '/files',
        [ViewMode.CORPUS]: '/corpus',
        [ViewMode.SETTLEMENTS]: '/settlements',
        [ViewMode.WORKFLOWS]: '/workflows',
        [ViewMode.ADMIN]: '/settings',
        [ViewMode.SETTINGS]: '/settings',
        [ViewMode.WRITER]: '/writer',
        [ViewMode.WORKFLOW_DETAIL]: '/workflows',
      };
      return map[viewMode];
    })();

    if (!target) return;
    if (location.pathname === target) return;
    const currentView = resolveViewMode(location.pathname);
    if (currentView === viewMode && location.pathname !== '/') return;
    navigate(target, { replace: true });
  }, [viewMode, activeProjectId, activeSessionId, navigate, location.pathname]);

  if (isLoading) {
    return (
      <div className="h-screen w-full flex items-center justify-center bg-paper-50 dark:bg-zinc-950">
        <Loader2 className="w-8 h-8 animate-spin text-brand-500" />
      </div>
    );
  }

  if (!user) {
    return <AuthScreen />;
  }

  if (isCreating) {
    return <NovelWizard onCancel={() => setIsCreating(false)} />;
  }

  return (
    <div className="flex h-screen w-full bg-paper-50 dark:bg-zinc-950 text-ink-900 dark:text-zinc-100 overflow-hidden font-sans transition-colors duration-300">
      {showSidebar && (
        <div className="hidden md:block h-full shrink-0">
          <Sidebar />
        </div>
      )}

      {showSidebar && mobileMenuOpen && (
        <div className="fixed inset-0 z-[150] md:hidden flex">
           <div className="fixed inset-0 bg-black/20 dark:bg-black/60 backdrop-blur-sm" onClick={() => setMobileMenuOpen(false)}></div>
           <div className="relative w-64 h-full bg-white dark:bg-zinc-900 shadow-xl animate-in slide-in-from-left duration-300">
              <Sidebar />
           </div>
        </div>
      )}

      <main className="flex-1 flex flex-col min-w-0 relative">
        {/* 顶部系统托盘 - 不遮挡主内容 */}
        <div
          className={`absolute z-[200] pointer-events-none flex px-4 ${
            isEditorPage
              ? 'top-[76px] right-3 left-auto translate-x-0 w-auto max-w-none justify-end px-0'
              : 'top-4 right-6 left-auto translate-x-0 w-auto max-w-none justify-end px-0'
          }`}
        >
          <div className={`flex items-center gap-1 p-1.5 bg-white/70 dark:bg-zinc-900/70 backdrop-blur-2xl border border-paper-200 dark:border-white/10 rounded-2xl shadow-[0_8px_32px_rgba(0,0,0,0.08)] pointer-events-auto transition-all duration-500 ${viewMode === ViewMode.WRITER ? 'opacity-40 hover:opacity-100' : 'opacity-100'}`}>
             
             {/* 返回仪表盘按钮 (仅在进入项目后显示) */}
              {activeProjectId && (
                <button 
                  onClick={exitProject}
                  className="flex items-center gap-2 px-3 py-1.5 hover:bg-paper-100 dark:hover:bg-zinc-800 rounded-xl transition-all group text-ink-500 dark:text-zinc-400"
                  title="返回仪表盘"
                >
                  <LayoutDashboard className="w-4 h-4 group-hover:scale-110 transition-transform" />
                  <span className="text-[10px] font-black uppercase tracking-widest hidden sm:block">Dashboard</span>
               </button>
             )}

              <div className="w-px h-4 bg-paper-200 dark:bg-zinc-800 mx-1"></div>

             {/* 设置中心入口 */}
              <button 
                onClick={() => setViewMode(ViewMode.SETTINGS)} 
                className={`flex items-center gap-2.5 pl-3 pr-2 py-1.5 rounded-xl transition-all group ${viewMode === ViewMode.SETTINGS ? 'bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 shadow-md' : 'hover:bg-paper-50 dark:hover:bg-zinc-800'}`}
              >
                 <div className="flex items-center gap-2">
                   <div className={`w-6 h-6 rounded-lg flex items-center justify-center text-[9px] font-black transition-colors ${viewMode === ViewMode.SETTINGS ? 'bg-white/20 dark:bg-black/20' : 'bg-ink-900 dark:bg-zinc-100 text-white dark:text-ink-900'}`}>
                     {user?.username.charAt(0).toUpperCase()}
                   </div>
                   <div className="flex flex-col items-start leading-none mr-2">
                     <span className="text-[10px] font-black uppercase tracking-widest">{user?.username}</span>
                     <span className="text-[8px] font-medium opacity-50 uppercase">{user?.role}</span>
                   </div>
                 </div>
                 <Settings className={`w-3.5 h-3.5 opacity-40 group-hover:opacity-100 group-hover:rotate-90 transition-all ${viewMode === ViewMode.SETTINGS ? 'opacity-100' : ''}`} />
              </button>
           </div>
        </div>

        {projectLoadError && (
          <div className="absolute top-20 right-6 z-[210] w-[360px] max-w-[90vw] pointer-events-auto">
            <div className="bg-white/90 dark:bg-zinc-900/90 backdrop-blur-xl border border-amber-200/60 dark:border-amber-800/50 text-amber-700 dark:text-amber-300 rounded-2xl shadow-[0_12px_40px_rgba(0,0,0,0.12)] px-4 py-3 flex items-start gap-3">
              <div className="mt-0.5">
                <Shield className="w-4 h-4" />
              </div>
              <div className="flex-1">
                <div className="text-[11px] font-black uppercase tracking-widest">本地项目读取异常</div>
                <div className="text-[11px] font-medium mt-1 leading-relaxed">{projectLoadError}</div>
              </div>
              <button
                onClick={clearProjectLoadError}
                className="text-[10px] font-black uppercase tracking-widest text-amber-600/70 hover:text-amber-700 dark:text-amber-300/70 dark:hover:text-amber-200"
              >
                关闭
              </button>
            </div>
          </div>
        )}

        {showSidebar && (
          <div className="md:hidden h-12 border-b border-paper-200 dark:border-zinc-800 flex items-center px-4 bg-paper-50 dark:bg-zinc-900 justify-between shrink-0">
              <button onClick={() => setMobileMenuOpen(true)} className="p-2 -ml-2 text-ink-500 dark:text-zinc-400">
                <Menu className="w-5 h-5" />
              </button>
              <span className="text-xs font-bold text-ink-900 dark:text-zinc-100">{project?.title || 'Novel Agent OS'}</span>
              <div className="w-8"></div>
          </div>
        )}

        <div className="flex-1 flex overflow-hidden relative">
          <div className="flex-1 min-w-0 flex flex-col relative overflow-hidden">
            <Routes>
              <Route
                path="/"
                element={activeProjectId ? <Editor /> : <ProjectDashboard onCreateNew={() => setIsCreating(true)} />}
              />
              <Route path="/writer" element={<Editor />} />
              <Route path="/planboard" element={<KanbanBoard />} />
              <Route path="/world" element={<WorldBible />} />
              <Route path="/plugins" element={<PluginManager />} />
              <Route path="/files" element={<FileCenter />} />
              <Route path="/corpus" element={<CorpusCenter />} />
              <Route path="/settlements" element={<SettlementCenter />} />
              <Route
                path="/workflows"
                element={
                  <WorkflowSessions
                    sessions={[
                      { id: '1', title: '世界观生成', mode: 'Normal', status: 'running', updatedAt: '刚刚' },
                      { id: '2', title: '章节润色', mode: 'Batch', status: 'success', updatedAt: '2 小时前' },
                    ]}
                    onSelect={(id) => {
                      selectSession(id);
                      setViewMode(ViewMode.WORKFLOW_DETAIL);
                    }}
                  />
                }
              />
              <Route
                path="/workflows/:sessionId"
                element={
                  <WorkflowDetail
                    sessionId={activeSessionId || '未知'}
                    connectionState={'disconnected' as ConnectionState}
                    onBack={() => {
                      setViewMode(ViewMode.WORKFLOWS);
                    }}
                  />
                }
              />
              <Route path="/settings" element={<UserSettings />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </div>
          {activeProjectId && !isSettingsView && <AIAssistant />}
        </div>
      </main>
    </div>
  );
};

export const Layout: React.FC = () => {
  return <LayoutContent />;
};
