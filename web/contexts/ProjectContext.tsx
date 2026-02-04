import React, { createContext, useContext, useState, ReactNode, useEffect, useCallback } from 'react';
import {
  Project,
  Document,
  Volume,
  ViewMode,
  StoryEntity,
  EntityLink,
  EntityType,
  AISettings,
  Bookmark,
  ModelInfo,
  AIPromptTemplate,
} from '../types';
import { fetchAvailableModels, clearModelCache, DEFAULT_AI_SETTINGS } from '../services/aiService';
import { deleteProjectRecord, getProjectsByUser, upsertProject } from '../services/db';
import { backupProjectSnapshotApi, downloadBackupFile, getLatestBackupApi, listProjectsApi, projectFromSnapshot, toSnapshotPayload, upsertProjectSnapshotApi } from '../services/projectApi';
import { useAuth } from './AuthContext';
// 单机版本不从后端加载项目

export type Theme = 'light' | 'dark';
const CLOUD_SYNC_KEY = 'nao_cloud_sync_enabled';

interface ProjectContextType {
  projects: Project[];
  activeProjectId: string | null;
  activeSessionId: string | null;
  projectLoadError: string | null;
  clearProjectLoadError: () => void;
  backupProject: () => void;
  restoreLatestBackup: () => Promise<void>;
  cloudSyncEnabled: boolean;
  setCloudSyncEnabled: (enabled: boolean) => void;
  createProject: (project: Project) => void;
  selectProject: (projectId: string) => void;
  deleteProject: (projectId: string) => void;
  exitProject: () => void;
  selectSession: (sessionId: string | null) => void;
  project: Project | null;
  activeDocumentId: string | null;
  activeVolumeId: string | null;
  viewMode: ViewMode;
  previousViewMode: ViewMode | 'DASHBOARD' | null;
  theme: Theme;
  setTheme: (theme: Theme) => void;
  isAISidebarOpen: boolean;
  availableModels: ModelInfo[];
  defaultAISettings: AISettings;
  setProject: (project: Project) => void;
  setActiveDocumentId: (id: string | null) => void;
  setActiveVolumeId: (id: string | null) => void;
  setViewMode: (mode: ViewMode) => void;
  navigateBack: () => void;
  toggleAISidebar: () => void;
  updateAISettings: (settings: Partial<AISettings>) => void;
  updateDefaultAISettings: (settings: Partial<AISettings>) => void;
  refreshModels: (settings?: AISettings) => Promise<void>;
  clearCache: () => void;
  updateNovelDetails: (details: Partial<Project>) => void;
  addVolume: (initialData?: Partial<Volume>) => void;
  updateVolume: (volumeId: string, updates: Partial<Volume>) => void;
  deleteVolume: (volumeId: string) => void;
  updateDocument: (docId: string, updates: Partial<Document>) => void;
  addDocument: (volumeId: string, initialData?: Partial<Document>) => void;
  deleteDocument: (id: string) => void;
  addBookmark: (docId: string, name: string, position: number) => void;
  deleteBookmark: (docId: string, bookmarkId: string) => void;
  addEntity: (type: EntityType, initialData?: Partial<StoryEntity>) => void;
  updateEntity: (id: string, updates: Partial<StoryEntity>) => void;
  deleteEntity: (id: string) => void;
  batchDeleteEntities: (ids: string[]) => void;
  linkEntities: (sourceId: string, targetId: string, type: EntityType, relation: string) => void;
  unlinkEntities: (sourceId: string, targetId: string) => void;
  batchLinkEntities: (sourceIds: string[], targetId: string, targetType: EntityType, relation: string) => void;
  addTemplate: (name: string, template: string, description: string, category: 'logic' | 'style' | 'content' | 'character') => void;
  deleteTemplate: (id: string) => void;
}

const ProjectContext = createContext<ProjectContextType | undefined>(undefined);
const THEME_KEY = 'novel_agent_theme';
const DEFAULT_AI_KEY = 'nao_default_ai_v1';

export const ProjectProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const { user, deviceId } = useAuth();
  const [projects, setProjects] = useState<Project[]>([]);
  const [activeProjectId, setActiveProjectId] = useState<string | null>(null);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const [projectLoadError, setProjectLoadError] = useState<string | null>(null);
  const syncTimerRef = React.useRef<number | null>(null);
  const [activeDocumentId, setActiveDocumentId] = useState<string | null>(null);
  const [activeVolumeId, setActiveVolumeId] = useState<string | null>(null);
  const [viewMode, setViewModeState] = useState<ViewMode>(ViewMode.WRITER);
  const [previousViewMode, setPreviousViewMode] = useState<ViewMode | 'DASHBOARD' | null>(null);
  const [theme, setTheme] = useState<Theme>(() => (localStorage.getItem(THEME_KEY) as Theme) || 'light');
  const [isAISidebarOpen, setIsAISidebarOpen] = useState<boolean>(true);
  const [availableModels, setAvailableModels] = useState<ModelInfo[]>([]);
  const [defaultAISettings, setDefaultAISettings] = useState<AISettings>(() => {
    const saved = localStorage.getItem(DEFAULT_AI_KEY);
    return saved ? JSON.parse(saved) : DEFAULT_AI_SETTINGS;
  });
  const [cloudSyncEnabled, setCloudSyncEnabledState] = useState<boolean>(() => {
    const stored = localStorage.getItem(CLOUD_SYNC_KEY);
    if (stored === 'true' || stored === 'false') return stored === 'true';
    return (import.meta as any).env?.VITE_ENABLE_CLOUD_SYNC === 'true';
  });

  // 本地单机：从本地数据库加载项目
  useEffect(() => {
    const load = async () => {
      const ownerId = user?.id ? String(user.id) : deviceId;
      if (!ownerId) return;

      try {
        const localProjects = await getProjectsByUser(ownerId);
        setProjects(localProjects);
        setActiveProjectId((prev) => (localProjects.find((p) => p.id === prev) ? prev : null));
        setProjectLoadError(null);
      } catch {
        setProjectLoadError('本地项目读取失败，请检查浏览器存储权限或稍后重试。');
      }
    };

    load();
  }, [user?.id, deviceId]);

  useEffect(() => {
    const loadRemote = async () => {
      if (!user || !cloudSyncEnabled) return;
      try {
        const data = await listProjectsApi(1, 200);
        const ownerId = String(user.id);
        const merged: Project[] = [];

        for (const item of data.list || []) {
          const snapshot = projectFromSnapshot(item);
          if (!snapshot) continue;
          merged.push(snapshot);
          await upsertProject(ownerId, snapshot);
        }

        if (merged.length > 0) {
          setProjects((prev) => {
            const map = new Map(prev.map((p) => [p.id, p]));
            merged.forEach((p) => map.set(p.id, p));
            return Array.from(map.values());
          });
        }
      } catch {
        // 远端同步失败不影响本地
      }
    };

    loadRemote();
  }, [user?.id, cloudSyncEnabled]);

  useEffect(() => {
    if (!user) return;
    if (!cloudSyncEnabled) {
      setCloudSyncEnabled(true);
    }
  }, [user]);

  useEffect(() => {
    localStorage.setItem(THEME_KEY, theme);
  }, [theme]);

  useEffect(() => {
    localStorage.setItem(DEFAULT_AI_KEY, JSON.stringify(defaultAISettings));
  }, [defaultAISettings]);

  const activeProject = projects.find((p) => p.id === activeProjectId) || null;

  const setViewMode = useCallback(
    (mode: ViewMode) => {
      setViewModeState((prev) => {
        if (prev === mode) return prev;
        if (mode === ViewMode.SETTINGS && prev !== ViewMode.SETTINGS) {
          setPreviousViewMode(activeProjectId ? prev : 'DASHBOARD');
        } else if (mode !== ViewMode.SETTINGS) {
          setPreviousViewMode(null);
        }
        return mode;
      });
    },
    [activeProjectId]
  );

  const navigateBack = () => {
    if (previousViewMode === 'DASHBOARD') {
      exitProject();
    } else if (previousViewMode) {
      setViewModeState(previousViewMode as ViewMode);
    } else {
      if (activeProjectId) setViewModeState(ViewMode.WRITER);
      else {
        setViewModeState(ViewMode.WRITER);
      }
    }
    setPreviousViewMode(null);
  };

  // Phase4：仅用于本地临时创建（不落后端）。联调闭环不依赖创建。
  const createProject = (newProject: Project) => {
    setProjects((prev) => {
      const next = [...prev, newProject];
      const ownerId = user?.id ? String(user.id) : deviceId;
      if (ownerId) {
        upsertProject(ownerId, newProject).catch(() => {});
      }
      return next;
    });
    scheduleRemoteSync(newProject);
    setActiveProjectId(newProject.id);
    setPreviousViewMode(null);
  };

  const selectProject = (projectId: string) => {
    setActiveProjectId(projectId);
    const proj = projects.find((p) => p.id === projectId);
    if (proj && proj.documents.length > 0) {
      setActiveDocumentId(proj.documents[0].id);
      setActiveVolumeId(proj.documents[0].volumeId);
    }
    setViewModeState(ViewMode.WRITER);
    setPreviousViewMode(null);
  };

  const deleteProject = (projectId: string) => {
    setProjects((prev) => {
      const next = prev.filter((p) => p.id !== projectId);
      const ownerId = user?.id ? String(user.id) : deviceId;
      if (ownerId) {
        deleteProjectRecord(ownerId, projectId).catch(() => {});
      }
      return next;
    });
    if (activeProjectId === projectId) setActiveProjectId(null);
  };

  const exitProject = () => {
    setActiveProjectId(null);
    setViewModeState(ViewMode.WRITER);
    setPreviousViewMode('DASHBOARD');
  };

  const selectSession = useCallback((sessionId: string | null) => {
    setActiveSessionId((prev) => (prev === sessionId ? prev : sessionId));
  }, []);

  const updateActiveProject = (updater: (p: Project) => Project) => {
    if (!activeProjectId) return;
    setProjects((prev) => {
      const next = prev.map((p) => {
        if (p.id === activeProjectId) {
          const updated = updater(p);
          const ownerId = user?.id ? String(user.id) : deviceId;
          if (ownerId) {
            upsertProject(ownerId, updated).catch(() => {});
          }
          return updated;
        }
        return p;
      });
      return next;
    });
  };

  const scheduleRemoteSync = useCallback(
    (project: Project) => {
      if (!user || !cloudSyncEnabled) return;
      if (syncTimerRef.current) {
        window.clearTimeout(syncTimerRef.current);
      }
      syncTimerRef.current = window.setTimeout(() => {
        upsertProjectSnapshotApi(toSnapshotPayload(project, project.id)).catch(() => {});
        backupProjectSnapshotApi(toSnapshotPayload(project, project.id)).catch(() => {});
      }, 800);
    },
    [user, cloudSyncEnabled]
  );

  const setCloudSyncEnabled = (enabled: boolean) => {
    setCloudSyncEnabledState(enabled);
    localStorage.setItem(CLOUD_SYNC_KEY, String(enabled));
  };

  const clearProjectLoadError = () => setProjectLoadError(null);

  const restoreLatestBackup = async () => {
    if (!user || !cloudSyncEnabled || !activeProject) return;
    const latest = await getLatestBackupApi(Number(activeProject.id));
    const snapshot = await downloadBackupFile(latest.id);
    if (!snapshot || !snapshot.id) return;
    const restored = snapshot as Project;
    const ownerId = String(user.id);
    await upsertProject(ownerId, restored);
    setProjects((prev) => {
      const map = new Map(prev.map((p) => [p.id, p]));
      map.set(restored.id, restored);
      return Array.from(map.values());
    });
    setActiveProjectId(restored.id);
  };

  const refreshModels = async (settings?: AISettings) => {
    const targetSettings = settings || activeProject?.aiSettings || defaultAISettings;
    const models = await fetchAvailableModels(targetSettings);
    setAvailableModels(models);
  };

  const clearCache = () => {
    clearModelCache();
  };

  const updateAISettings = (settings: Partial<AISettings>) => {
    updateActiveProject((p) => {
      const updated = { ...p, aiSettings: { ...p.aiSettings, ...settings } };
      scheduleRemoteSync(updated);
      return updated;
    });
    refreshModels({ ...activeProject?.aiSettings, ...settings } as AISettings).catch(() => {});
  };

  const updateDefaultAISettings = (settings: Partial<AISettings>) => {
    setDefaultAISettings((prev) => ({ ...prev, ...settings }));
    refreshModels({ ...defaultAISettings, ...settings } as AISettings).catch(() => {});
  };

  const updateNovelDetails = (details: Partial<Project>) => {
    updateActiveProject((p) => {
      const updated = { ...p, ...details };
      scheduleRemoteSync(updated);
      return updated;
    });
  };

  const addVolume = (initialData: Partial<Volume> = {}) => {
    if (!activeProject) return;
    const newVol: Volume = {
      id: `v${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      title: initialData.title || `新卷 ${activeProject.volumes.length + 1}`,
      order: activeProject.volumes.length,
      theme: initialData.theme || '',
      coreGoal: initialData.coreGoal || '',
      boundaries: initialData.boundaries || '',
    };
    updateActiveProject((p) => {
      const updated = { ...p, volumes: [...p.volumes, newVol] };
      scheduleRemoteSync(updated);
      return updated;
    });
    setActiveVolumeId(newVol.id);
  };

  const updateVolume = (volumeId: string, updates: Partial<Volume>) => {
    updateActiveProject((p) => {
      const updated = {
        ...p,
        volumes: p.volumes.map((v) => (v.id === volumeId ? { ...v, ...updates } : v)),
      };
      scheduleRemoteSync(updated);
      return updated;
    });
  };

  const deleteVolume = (volumeId: string) => {
    updateActiveProject((p) => {
      const updated = {
        ...p,
        volumes: p.volumes.filter((v) => v.id !== volumeId),
        documents: p.documents.filter((d) => d.volumeId !== volumeId),
      };
      scheduleRemoteSync(updated);
      return updated;
    });
  };

  const updateDocument = (docId: string, updates: Partial<Document>) => {
    updateActiveProject((p) => {
      const updated = {
        ...p,
        documents: p.documents.map((d) => (d.id === docId ? { ...d, ...updates } : d)),
      };
      scheduleRemoteSync(updated);
      return updated;
    });
  };

  const addDocument = (volumeId: string, initialData: Partial<Document> = {}) => {
    if (!activeProject) return;
    const volDocs = activeProject.documents.filter((d) => d.volumeId === volumeId);
    const newDoc: Document = {
      id: `d${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      volumeId,
      title: initialData.title || `新章节 ${volDocs.length + 1}`,
      content: initialData.content || '',
      status: '草稿',
      order: volDocs.length,
      linkedIds: [],
      bookmarks: [],
      targetWordCount: 3000,
      ...initialData,
    };
    updateActiveProject((p) => {
      const updated = { ...p, documents: [...p.documents, newDoc] };
      scheduleRemoteSync(updated);
      return updated;
    });
    setActiveDocumentId(newDoc.id);
  };

  const deleteDocument = (id: string) => {
    updateActiveProject((p) => {
      const updated = { ...p, documents: p.documents.filter((d) => d.id !== id) };
      scheduleRemoteSync(updated);
      return updated;
    });
  };

  const addBookmark = (docId: string, name: string, position: number) => {
    const newBookmark: Bookmark = { id: `bm${Date.now()}`, name, position, timestamp: Date.now() };
    updateActiveProject((p) => {
      const updated = {
        ...p,
        documents: p.documents.map((d) =>
          d.id === docId ? { ...d, bookmarks: [...d.bookmarks, newBookmark] } : d
        ),
      };
      scheduleRemoteSync(updated);
      return updated;
    });
  };

  const deleteBookmark = (docId: string, bookmarkId: string) => {
    updateActiveProject((p) => {
      const updated = {
        ...p,
        documents: p.documents.map((d) =>
          d.id === docId ? { ...d, bookmarks: d.bookmarks.filter((b) => b.id !== bookmarkId) } : d
        ),
      };
      scheduleRemoteSync(updated);
      return updated;
    });
  };

  const addEntity = (type: EntityType, initialData: Partial<StoryEntity> = {}) => {
    const newEntity: StoryEntity = {
      id: `e${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      type,
      title: '未命名实体',
      subtitle: '',
      content: '',
      tags: [],
      linkedIds: [],
      importance: 'secondary',
      customFields: [],
      referenceCount: 0,
      ...initialData,
    };
    updateActiveProject((p) => {
      const updated = { ...p, entities: [...p.entities, newEntity] };
      scheduleRemoteSync(updated);
      return updated;
    });
  };

  const updateEntity = (id: string, updates: Partial<StoryEntity>) => {
    updateActiveProject((p) => {
      const updated = {
        ...p,
        entities: p.entities.map((e) => (e.id === id ? { ...e, ...updates } : e)),
      };
      scheduleRemoteSync(updated);
      return updated;
    });
  };

  const deleteEntity = (id: string) => {
    updateActiveProject((p) => {
      const cleanedDocuments = p.documents.map((d) => ({
        ...d,
        linkedIds: d.linkedIds.filter((l) => l.targetId !== id),
      }));
      const cleanedEntities = p.entities
        .filter((e) => e.id !== id)
        .map((e) => ({
          ...e,
          linkedIds: e.linkedIds.filter((l) => l.targetId !== id),
        }));
      const updated = {
        ...p,
        documents: cleanedDocuments,
        entities: cleanedEntities,
      };
      scheduleRemoteSync(updated);
      return updated;
    });
  };

  const batchDeleteEntities = (ids: string[]) => {
    updateActiveProject((p) => {
      const cleanedDocuments = p.documents.map((d) => ({
        ...d,
        linkedIds: d.linkedIds.filter((l) => !ids.includes(l.targetId)),
      }));
      const cleanedEntities = p.entities
        .filter((e) => !ids.includes(e.id))
        .map((e) => ({
          ...e,
          linkedIds: e.linkedIds.filter((l) => !ids.includes(l.targetId)),
        }));
      const updated = {
        ...p,
        documents: cleanedDocuments,
        entities: cleanedEntities,
      };
      scheduleRemoteSync(updated);
      return updated;
    });
  };

  const linkEntities = (sourceId: string, targetId: string, type: EntityType, relation: string) => {
    const link: EntityLink = { targetId, type, relationName: relation };
    updateActiveProject((p) => {
      const newP = { ...p };
      if (sourceId.startsWith('d')) {
        newP.documents = newP.documents.map((d) => (d.id === sourceId ? { ...d, linkedIds: [...d.linkedIds, link] } : d));
      } else {
        newP.entities = newP.entities.map((e) => (e.id === sourceId ? { ...e, linkedIds: [...e.linkedIds, link] } : e));
      }
      scheduleRemoteSync(newP);
      return newP;
    });
  };

  const unlinkEntities = (sourceId: string, targetId: string) => {
    updateActiveProject((p) => {
      const newP = { ...p };
      if (sourceId.startsWith('d')) {
        newP.documents = newP.documents.map((d) => (d.id === sourceId ? { ...d, linkedIds: d.linkedIds.filter((l) => l.targetId !== targetId) } : d));
      } else {
        newP.entities = newP.entities.map((e) => (e.id === sourceId ? { ...e, linkedIds: e.linkedIds.filter((l) => l.targetId !== targetId) } : e));
      }
      scheduleRemoteSync(newP);
      return newP;
    });
  };

  const batchLinkEntities = (sourceIds: string[], targetId: string, targetType: EntityType, relation: string) => {
    const link: EntityLink = { targetId, type: targetType, relationName: relation };
    updateActiveProject((p) => {
      const newP = { ...p };
      sourceIds.forEach((id) => {
        if (id.startsWith('d')) {
          newP.documents = newP.documents.map((d) => (d.id === id ? { ...d, linkedIds: [...d.linkedIds, link] } : d));
        } else {
          newP.entities = newP.entities.map((e) => (e.id === id ? { ...e, linkedIds: [...e.linkedIds, link] } : e));
        }
      });
      scheduleRemoteSync(newP);
      return newP;
    });
  };

  const addTemplate = (name: string, template: string, description: string, category: 'logic' | 'style' | 'content' | 'character') => {
    updateActiveProject((p) => {
      const updated = {
        ...p,
        templates: [...(p.templates || []), { id: `t${Date.now()}`, name, template, description, category }],
      };
      scheduleRemoteSync(updated);
      return updated;
    });
  };

  const deleteTemplate = (id: string) => {
    updateActiveProject((p) => {
      const updated = {
        ...p,
        templates: (p.templates || []).filter((t) => t.id !== id),
      };
      scheduleRemoteSync(updated);
      return updated;
    });
  };

  const toggleAISidebar = useCallback(() => {
    setIsAISidebarOpen((prev) => !prev);
  }, []);

  return (
    <ProjectContext.Provider
      value={{
        projects,
        activeProjectId,
        activeSessionId,
        projectLoadError,
        clearProjectLoadError,
        backupProject: () => {
          if (!user || !activeProject || !cloudSyncEnabled) return;
          backupProjectSnapshotApi(toSnapshotPayload(activeProject, activeProject.id)).catch(() => {});
        },
        restoreLatestBackup,
        cloudSyncEnabled,
        setCloudSyncEnabled,
        createProject,
        selectProject,
        deleteProject,
        exitProject,
        selectSession,
        project: activeProject,
        activeDocumentId,
        activeVolumeId,
        viewMode,
        previousViewMode,
        theme,
        setTheme,
        isAISidebarOpen,
        availableModels,
        defaultAISettings,
        setProject: (p) => updateActiveProject(() => p),
        setActiveDocumentId,
        setActiveVolumeId,
        setViewMode,
        navigateBack,
        toggleAISidebar,
        updateAISettings,
        updateDefaultAISettings,
        refreshModels,
        clearCache,
        updateNovelDetails,
        addVolume,
        updateVolume,
        deleteVolume,
        updateDocument,
        addDocument,
        deleteDocument,
        addBookmark,
        deleteBookmark,
        addEntity,
        updateEntity,
        deleteEntity,
        batchDeleteEntities,
        linkEntities,
        unlinkEntities,
        batchLinkEntities,
        addTemplate,
        deleteTemplate,
      }}
    >
      {children}
    </ProjectContext.Provider>
  );
};

export const useProject = () => {
  const context = useContext(ProjectContext);
  if (!context) throw new Error('useProject must be used within a ProjectProvider');
  return context;
};
