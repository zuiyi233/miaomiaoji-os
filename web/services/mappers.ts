import { Project } from '../types';
import { ProjectListItemDTO } from './projectApi';

// 将后端 Project DTO 尽量映射到模板 Project 结构（最小可用：用于项目列表与进入项目）
export function mapProjectFromApi(dto: ProjectListItemDTO): Project {
  const pid = `p${dto.id}`;

  return {
    id: pid,
    title: dto.title || '未命名项目',
    genre: dto.genre,
    tags: dto.tags || [],
    coreConflict: dto.core_conflict || '',
    characterArc: dto.character_arc || '',
    ultimateValue: dto.ultimate_value || '',
    worldRules: dto.world_rules || '',
    volumes: [],
    documents: [],
    entities: [],
    templates: [],
    // 复用模板默认 AI 设置（避免引入新的配置体系）
    aiSettings: {
      provider: 'gemini',
      model: 'gemini-3-flash-preview',
      proxyEndpoint: '',
      temperature: 0.9,
    },
    plugins: [],
  };
}
