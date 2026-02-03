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
import { useAuth } from './AuthContext';
import { fetchProjectsApi } from '../services/projectApi';
import { mapProjectFromApi } from '../services/mappers';

export type Theme = 'light' | 'dark';

interface ProjectContextType {
  projects: Project[];
  activeProjectId: string | null;
  activeSessionId: string | null;
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
  const { user } = useAuth();
  const [projects, setProjects] = useState<Project[]>([]);
  const [activeProjectId, setActiveProjectId] = useState<string | null>(null);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
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

  // Phase4：改为从后端拉取项目列表，避免 localStorage mock
  useEffect(() => {
    const load = async () => {
      if (!user) {
        setProjects([]);
        setActiveProjectId(null);
        return;
      }

      try {
        const data = await fetchProjectsApi(1, 50);
        const mapped = (data.list || []).map(mapProjectFromApi);
        setProjects(mapped);
      } catch {
        setProjects([]);
      }
    };

    load();
  }, [user?.id]);

  useEffect(() => {
    localStorage.setItem(THEME_KEY, theme);
  }, [theme]);

  useEffect(() => {
    localStorage.setItem(DEFAULT_AI_KEY, JSON.stringify(defaultAISettings));
  }, [defaultAISettings]);

  const activeProject = projects.find((p) => p.id === activeProjectId) || null;

  const setViewMode = useCallback(
    (mode: ViewMode) => {
      if (mode === ViewMode.SETTINGS && viewMode !== ViewMode.SETTINGS) {
        setPreviousViewMode(activeProjectId ? viewMode : 'DASHBOARD');
      } else if (mode !== ViewMode.SETTINGS) {
        setPreviousViewMode(null);
      }
      setViewModeState(mode);
    },
    [activeProjectId, viewMode]
  );

  const navigateBack = () => {
    if (previousViewMode === 'DASHBOARD') {
      exitProject();
    } else if (previousViewMode) {
      setViewModeState(previousViewMode as ViewMode);
    } else {
      if (activeProjectId) setViewModeState(ViewMode.WRITER);
      else exitProject();
    }
    setPreviousViewMode(null);
  };

  // Phase4：仅用于本地临时创建（不落后端）。联调闭环不依赖创建。
  const createProject = (newProject: Project) => {
    setProjects((prev) => [...prev, newProject]);
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
    setProjects((prev) => prev.filter((p) => p.id !== projectId));
    if (activeProjectId === projectId) setActiveProjectId(null);
  };

  const exitProject = () => {
    setActiveProjectId(null);
    setViewModeState(ViewMode.WRITER);
    setPreviousViewMode(null);
  };

  const selectSession = useCallback((sessionId: string | null) => {
    setActiveSessionId((prev) => (prev === sessionId ? prev : sessionId));
  }, []);

  const updateActiveProject = (updater: (p: Project) => Project) => {
    if (!activeProjectId) return;
    setProjects((prev) =>
      prev.map((p) => {
        if (p.id === activeProjectId) {
          return updater(p);
        }
        return p;
      })
    );
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
    updateActiveProject((p) => ({ ...p, aiSettings: { ...p.aiSettings, ...settings } }));
  };

  const updateDefaultAISettings = (settings: Partial<AISettings>) => {
    setDefaultAISettings((prev) => ({ ...prev, ...settings }));
  };

  const updateNovelDetails = (details: Partial<Project>) => {
    updateActiveProject((p) => ({ ...p, ...details }));
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
    updateActiveProject((p) => ({ ...p, volumes: [...p.volumes, newVol] }));
    setActiveVolumeId(newVol.id);
  };

  const updateVolume = (volumeId: string, updates: Partial<Volume>) => {
    updateActiveProject((p) => ({
      ...p,
      volumes: p.volumes.map((v) => (v.id === volumeId ? { ...v, ...updates } : v)),
    }));
  };

  const deleteVolume = (volumeId: string) => {
    updateActiveProject((p) => ({
      ...p,
      volumes: p.volumes.filter((v) => v.id !== volumeId),
      documents: p.documents.filter((d) => d.volumeId !== volumeId),
    }));
  };

  const updateDocument = (docId: string, updates: Partial<Document>) => {
    updateActiveProject((p) => ({
      ...p,
      documents: p.documents.map((d) => (d.id === docId ? { ...d, ...updates } : d)),
    }));
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
    updateActiveProject((p) => ({ ...p, documents: [...p.documents, newDoc] }));
    setActiveDocumentId(newDoc.id);
  };

  const deleteDocument = (id: string) => {
    updateActiveProject((p) => ({ ...p, documents: p.documents.filter((d) => d.id !== id) }));
  };

  const addBookmark = (docId: string, name: string, position: number) => {
    const newBookmark: Bookmark = { id: `bm${Date.now()}`, name, position, timestamp: Date.now() };
    updateActiveProject((p) => ({
      ...p,
      documents: p.documents.map((d) =>
        d.id === docId ? { ...d, bookmarks: [...d.bookmarks, newBookmark] } : d
      ),
    }));
  };

  const deleteBookmark = (docId: string, bookmarkId: string) => {
    updateActiveProject((p) => ({
      ...p,
      documents: p.documents.map((d) =>
        d.id === docId ? { ...d, bookmarks: d.bookmarks.filter((b) => b.id !== bookmarkId) } : d
      ),
    }));
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
    updateActiveProject((p) => ({ ...p, entities: [...p.entities, newEntity] }));
  };

  const updateEntity = (id: string, updates: Partial<StoryEntity>) => {
    updateActiveProject((p) => ({
      ...p,
      entities: p.entities.map((e) => (e.id === id ? { ...e, ...updates } : e)),
    }));
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
      return {
        ...p,
        documents: cleanedDocuments,
        entities: cleanedEntities,
      };
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
      return {
        ...p,
        documents: cleanedDocuments,
        entities: cleanedEntities,
      };
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
      return newP;
    });
  };

  const addTemplate = (name: string, template: string, description: string, category: 'logic' | 'style' | 'content' | 'character') => {
    updateActiveProject((p) => ({
      ...p,
      templates: [...(p.templates || []), { id: `t${Date.now()}`, name, template, description, category }],
    }));
  };

  const deleteTemplate = (id: string) => {
    updateActiveProject((p) => ({
      ...p,
      templates: (p.templates || []).filter((t) => t.id !== id),
    }));
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
