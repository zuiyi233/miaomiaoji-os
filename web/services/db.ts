import { openDB } from 'idb';
import { Project } from '../types';

const DB_NAME = 'novel_agent_os_db';
const DB_VERSION = 2;

const STORES = {
  models: 'models_cache',
  projects: 'projects'
} as const;

export type ModelCacheRecord = {
  provider: string;
  data: any;
  timestamp: number;
};

export type ProjectRecord = {
  key: string;
  userId: string;
  project: Project;
  updatedAt: number;
};

export async function getDB() {
  return openDB(DB_NAME, DB_VERSION, {
    upgrade(db) {
      if (!db.objectStoreNames.contains(STORES.models)) {
        db.createObjectStore(STORES.models, { keyPath: 'provider' });
      }
      if (!db.objectStoreNames.contains(STORES.projects)) {
        db.createObjectStore(STORES.projects, { keyPath: 'key' });
      }
    }
  });
}

export async function getModelCache(provider: string): Promise<ModelCacheRecord | null> {
  const db = await getDB();
  return (await db.get(STORES.models, provider)) || null;
}

export async function setModelCache(record: ModelCacheRecord): Promise<void> {
  const db = await getDB();
  await db.put(STORES.models, record);
}

export async function clearModelCache(provider?: string): Promise<void> {
  const db = await getDB();
  if (provider) {
    await db.delete(STORES.models, provider);
    return;
  }
  await db.clear(STORES.models);
}

function projectKey(userId: string, projectId: string) {
  return `${userId}:${projectId}`;
}

export async function getProjectsByUser(userId: string): Promise<Project[]> {
  const db = await getDB();
  const records = await db.getAll(STORES.projects);
  return records
    .filter((item: ProjectRecord) => item.userId === userId)
    .map((item: ProjectRecord) => item.project);
}

export async function upsertProject(userId: string, project: Project): Promise<void> {
  const db = await getDB();
  const record: ProjectRecord = {
    key: projectKey(userId, project.id),
    userId,
    project,
    updatedAt: Date.now()
  };
  await db.put(STORES.projects, record);
}

export async function deleteProjectRecord(userId: string, projectId: string): Promise<void> {
  const db = await getDB();
  await db.delete(STORES.projects, projectKey(userId, projectId));
}
