
import React, { useState, useEffect, useRef, useCallback } from 'react';
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
  const [toolbarPosition, setToolbarPosition] = useState<{ x: number; y: number } | null>(null);
  const [isDraggingToolbar, setIsDraggingToolbar] = useState(false);
  const [isToolbarCollapsed, setIsToolbarCollapsed] = useState(false);
  const toolbarRef = useRef<HTMLDivElement | null>(null);
  const dragOffsetRef = useRef<{ x: number; y: number } | null>(null);
  const TOOLBAR_STORAGE_KEY = 'layout.floatingToolbar.position';
  const TOOLBAR_COLLAPSE_KEY = 'layout.floatingToolbar.collapsed';
  const TOOLBAR_PADDING = 16;
  const navigate = useNavigate();
  const location = useLocation();
  const isSettingsView = viewMode === ViewMode.SETTINGS;
  const showSidebar = !!activeProjectId && !isSettingsView;
  const isEditorPage = !!activeProjectId && (location.pathname === '/' || location.pathname === '/writer');

  const clampToolbarPosition = useCallback((position: { x: number; y: number }) => {
    const rect = toolbarRef.current?.getBoundingClientRect();
    if (!rect) return position;
    const maxX = Math.max(TOOLBAR_PADDING, window.innerWidth - rect.width - TOOLBAR_PADDING);
    const maxY = Math.max(TOOLBAR_PADDING, window.innerHeight - rect.height - TOOLBAR_PADDING);
    return {
      x: Math.min(Math.max(position.x, TOOLBAR_PADDING), maxX),
      y: Math.min(Math.max(position.y, TOOLBAR_PADDING), maxY),
    };
  }, []);

  const getDefaultToolbarPosition = useCallback(() => {
    const rect = toolbarRef.current?.getBoundingClientRect();
    if (!rect) {
      return { x: TOOLBAR_PADDING, y: TOOLBAR_PADDING };
    }
    return clampToolbarPosition({
      x: window.innerWidth - rect.width - TOOLBAR_PADDING,
      y: TOOLBAR_PADDING,
    });
  }, [clampToolbarPosition]);

  // Sync theme with DOM
  useEffect(() => {
    if (theme === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [theme]);

  useEffect(() => {
    const raw = localStorage.getItem(TOOLBAR_STORAGE_KEY);
    if (!raw) return;
    try {
      const parsed = JSON.parse(raw) as { x: number; y: number } | null;
      if (!parsed || typeof parsed.x !== 'number' || typeof parsed.y !== 'number') return;
      setToolbarPosition(parsed);
    } catch {
      return;
    }
  }, []);

  useEffect(() => {
    const raw = localStorage.getItem(TOOLBAR_COLLAPSE_KEY);
    if (!raw) return;
    setIsToolbarCollapsed(raw === 'true');
  }, []);

  useEffect(() => {
    if (!toolbarRef.current) return;
    if (!toolbarPosition) {
      setToolbarPosition(getDefaultToolbarPosition());
      return;
    }
    const clamped = clampToolbarPosition(toolbarPosition);
    if (clamped.x !== toolbarPosition.x || clamped.y !== toolbarPosition.y) {
      setToolbarPosition(clamped);
    }
  }, [toolbarPosition, clampToolbarPosition, getDefaultToolbarPosition]);

  useEffect(() => {
    if (!toolbarPosition) return;
    localStorage.setItem(TOOLBAR_STORAGE_KEY, JSON.stringify(toolbarPosition));
  }, [toolbarPosition]);

  useEffect(() => {
    localStorage.setItem(TOOLBAR_COLLAPSE_KEY, String(isToolbarCollapsed));
  }, [isToolbarCollapsed]);

  useEffect(() => {
    if (!isDraggingToolbar) return;

    const handleMouseMove = (event: MouseEvent) => {
      if (!dragOffsetRef.current) return;
      const next = clampToolbarPosition({
        x: event.clientX - dragOffsetRef.current.x,
        y: event.clientY - dragOffsetRef.current.y,
      });
      setToolbarPosition(next);
    };

    const handleTouchMove = (event: TouchEvent) => {
      if (!dragOffsetRef.current) return;
      const touch = event.touches[0];
      if (!touch) return;
      const next = clampToolbarPosition({
        x: touch.clientX - dragOffsetRef.current.x,
        y: touch.clientY - dragOffsetRef.current.y,
      });
      setToolbarPosition(next);
      event.preventDefault();
    };

    const stopDragging = () => {
      setIsDraggingToolbar(false);
      dragOffsetRef.current = null;
    };

    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('mouseup', stopDragging);
    window.addEventListener('touchmove', handleTouchMove, { passive: false });
    window.addEventListener('touchend', stopDragging);

    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', stopDragging);
      window.removeEventListener('touchmove', handleTouchMove);
      window.removeEventListener('touchend', stopDragging);
    };
  }, [isDraggingToolbar, clampToolbarPosition]);

  useEffect(() => {
    const handleResize = () => {
      if (!toolbarPosition) return;
      const clamped = clampToolbarPosition(toolbarPosition);
      if (clamped.x !== toolbarPosition.x || clamped.y !== toolbarPosition.y) {
        setToolbarPosition(clamped);
      }
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, [toolbarPosition, clampToolbarPosition]);

  const shouldStartToolbarDrag = (target: EventTarget | null) => {
    if (!target || !(target instanceof HTMLElement)) return true;
    return !target.closest('button');
  };

  const startToolbarDrag = (clientX: number, clientY: number) => {
    const rect = toolbarRef.current?.getBoundingClientRect();
    if (!rect) return;
    dragOffsetRef.current = { x: clientX - rect.left, y: clientY - rect.top };
    setIsDraggingToolbar(true);
  };

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
        <div className="hidden lg:block h-full shrink-0">
          <Sidebar />
        </div>
      )}

       {showSidebar && mobileMenuOpen && (
         <div className="fixed inset-0 z-[150] lg:hidden flex">
            <div className="fixed inset-0 bg-black/20 dark:bg-black/60 backdrop-blur-sm" onClick={() => setMobileMenuOpen(false)}></div>
            <div className="relative w-[82vw] max-w-xs h-full bg-white dark:bg-zinc-900 shadow-xl animate-in slide-in-from-left duration-300">
               <Sidebar />
            </div>
         </div>
       )}

      <main className="flex-1 flex flex-col min-w-0 relative">
        {/* 顶部系统托盘 - 移至右上角避免遮挡 */}
        <div
          className={`fixed z-[100] pointer-events-none flex ${
            toolbarPosition ? '' : 'top-4 right-4'
          }`}
          style={
            toolbarPosition
              ? { left: toolbarPosition.x, top: toolbarPosition.y }
              : undefined
          }
        >
          <div
            ref={toolbarRef}
            onMouseDown={(event) => {
              if (event.button !== 0) return;
              if (!shouldStartToolbarDrag(event.target)) return;
              startToolbarDrag(event.clientX, event.clientY);
            }}
            onTouchStart={(event) => {
              if (!shouldStartToolbarDrag(event.target)) return;
              const touch = event.touches[0];
              if (!touch) return;
              startToolbarDrag(touch.clientX, touch.clientY);
            }}
            className={`flex items-center ${isToolbarCollapsed ? 'gap-1 p-1 rounded-full' : 'gap-1.5 p-1.5 rounded-2xl'} bg-white/70 dark:bg-zinc-900/70 backdrop-blur-2xl border border-paper-200 dark:border-white/10 shadow-[0_8px_32px_rgba(0,0,0,0.08)] pointer-events-auto transition-all duration-500 cursor-grab active:cursor-grabbing ${viewMode === ViewMode.WRITER ? 'opacity-40 hover:opacity-100' : 'opacity-100'}`}
          >
            
              {isToolbarCollapsed ? (
                <button
                  onClick={() => setViewMode(ViewMode.SETTINGS)}
                  className={`p-2 rounded-full transition-all ${viewMode === ViewMode.SETTINGS ? 'bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 shadow-md' : 'hover:bg-paper-50 dark:hover:bg-zinc-800 text-ink-500 dark:text-zinc-400'}`}
                  title="设置中心"
                >
                  <Settings className="w-4 h-4" />
                </button>
              ) : (
                <button
                  onClick={() => setViewMode(ViewMode.SETTINGS)}
                  className={`flex items-center gap-2 pl-2.5 pr-2 py-1 rounded-xl transition-all group ${viewMode === ViewMode.SETTINGS ? 'bg-ink-900 dark:bg-zinc-100 text-white dark:text-zinc-900 shadow-md' : 'hover:bg-paper-50 dark:hover:bg-zinc-800'}`}
                >
                  <div className="flex items-center gap-2">
                    <div className={`w-5 h-5 rounded-lg flex items-center justify-center text-[8px] font-black transition-colors ${viewMode === ViewMode.SETTINGS ? 'bg-white/20 dark:bg-black/20' : 'bg-ink-900 dark:bg-zinc-100 text-white dark:text-ink-900'}`}>
                      {user?.username.charAt(0).toUpperCase()}
                    </div>
                    <div className="flex flex-col items-start leading-none mr-1">
                      <span className="text-[9px] font-black uppercase tracking-widest">{user?.username}</span>
                      <span className="text-[8px] font-medium opacity-50 uppercase">{user?.role}</span>
                    </div>
                  </div>
                  <Settings className={`w-3.5 h-3.5 opacity-40 group-hover:opacity-100 group-hover:rotate-90 transition-all ${viewMode === ViewMode.SETTINGS ? 'opacity-100' : ''}`} />
                </button>
              )}

              <button
                onClick={() => setIsToolbarCollapsed((prev) => !prev)}
                className="p-2 rounded-full hover:bg-paper-100 dark:hover:bg-zinc-800 transition-all text-ink-500 dark:text-zinc-400"
                title={isToolbarCollapsed ? '展开浮窗' : '收起浮窗'}
              >
                <ChevronDown className={`w-3.5 h-3.5 transition-transform ${isToolbarCollapsed ? 'rotate-0' : 'rotate-180'}`} />
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
          <div className="lg:hidden h-12 border-b border-paper-200 dark:border-zinc-800 flex items-center px-4 bg-paper-50 dark:bg-zinc-900 justify-between shrink-0">
              <button onClick={() => setMobileMenuOpen(true)} className="p-2 -ml-2 text-ink-500 dark:text-zinc-400">
                <Menu className="w-5 h-5" />
              </button>
              <span className="text-xs font-bold text-ink-900 dark:text-zinc-100 truncate max-w-[60%]">{project?.title || 'Novel Agent OS'}</span>
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
