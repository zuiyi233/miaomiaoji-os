import React from 'react';
import { StoryEntity, Project } from '../types';
import { Mic2, Trash2, Tag as TagIcon, Layout, Edit3, Link as LinkIcon } from 'lucide-react';
import { getTypeStyle, getEntityIcon } from './EntityVisuals';

interface EntityCardViewProps {
  entity: StoryEntity;
  project: Project;
  categories: any[];
  onEdit: () => void;
  onDelete: (id: string) => void;
}

export const EntityCardView: React.FC<EntityCardViewProps> = ({ entity, project, categories, onEdit, onDelete }) => {
  return (
    <>
      <div className={`absolute top-6 right-6 px-4 py-2 rounded-full text-[9px] font-black uppercase tracking-widest flex items-center gap-2 border ${getTypeStyle(entity.type)} shadow-sm z-10 dark:bg-zinc-800`}>
        {getEntityIcon(entity.type)} {categories.find(c => c.id === entity.type)?.label}
      </div>

      <div className="mb-6 pt-12 cursor-pointer group/title" onClick={onEdit}>
        <div className="flex items-center gap-2 mb-2">
            <h4 className="text-2xl font-black text-gray-900 dark:text-zinc-100 group-hover/title:text-brand-600 dark:group-hover/title:text-brand-400 transition-colors font-serif leading-tight">{entity.title}</h4>
            {entity.referenceCount !== undefined && entity.referenceCount > 0 && (
                <span className="bg-gray-100 dark:bg-zinc-800 text-gray-500 dark:text-zinc-400 px-2 py-0.5 rounded-full text-[9px] font-bold flex items-center gap-1" title="关联引用次数">
                    <LinkIcon className="w-2.5 h-2.5" /> {entity.referenceCount}
                </span>
            )}
        </div>
        <p className="text-[10px] font-black text-gray-400 dark:text-zinc-500 uppercase tracking-widest mt-2 flex items-center gap-2">
          {entity.subtitle || '点击编辑详情'}
          {entity.importance === 'main' && <span className="bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400 px-1.5 py-0.5 rounded text-[8px]">核心</span>}
        </p>
      </div>

      {entity.tags && entity.tags.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mb-6">
          {entity.tags.map((tag, idx) => (
            <span key={idx} className="flex items-center gap-1 px-2 py-0.5 bg-gray-100 dark:bg-zinc-800 text-[9px] font-bold text-gray-500 dark:text-zinc-400 rounded-md border border-gray-200 dark:border-zinc-700">
              <TagIcon className="w-2.5 h-2.5" /> {tag}
            </span>
          ))}
        </div>
      )}
      
      <div className="flex-1 space-y-5 mb-8">
          {entity.type === 'character' && entity.voiceStyle && (
            <div className="flex items-start gap-3 bg-rose-50/50 dark:bg-rose-900/10 p-4 rounded-2xl border border-rose-100 dark:border-rose-900/30">
                <Mic2 className="w-3.5 h-3.5 text-rose-400 dark:text-rose-500 mt-0.5 shrink-0" />
                <div className="space-y-1">
                  <span className="text-[8px] font-black text-rose-300 dark:text-rose-500/50 uppercase tracking-widest">语调风格 (Voice Style)</span>
                  <p className="text-[11px] font-bold text-rose-800 dark:text-rose-300 leading-relaxed italic">“{entity.voiceStyle}”</p>
                </div>
            </div>
          )}
          
          <div className="space-y-2">
            <span className="text-[9px] font-black text-gray-300 dark:text-zinc-700 uppercase tracking-widest block">档案内容</span>
            <p className="text-sm text-gray-600 dark:text-zinc-400 leading-[1.8] font-serif line-clamp-6">{entity.content || '暂无详细背景描写...'}</p>
          </div>

          {entity.customFields && entity.customFields.length > 0 && (
            <div className="pt-4 border-t border-gray-50 dark:border-zinc-800">
              <span className="text-[9px] font-black text-gray-300 dark:text-zinc-700 uppercase tracking-widest block mb-3">关键属性</span>
              <div className="grid grid-cols-2 gap-2">
                  {entity.customFields.map((f, i) => (
                    <div key={i} className="px-3 py-2 bg-white dark:bg-zinc-800/50 border border-gray-100 dark:border-zinc-800 rounded-xl text-[10px] shadow-sm flex flex-col gap-0.5">
                        <span className="font-black text-gray-400 dark:text-zinc-600 uppercase text-[8px] tracking-tighter">{f.key}</span>
                        <span className="font-bold text-gray-700 dark:text-zinc-300 truncate">{f.value}</span>
                    </div>
                  ))}
              </div>
            </div>
          )}
      </div>

      <div className="pt-6 border-t border-gray-50 dark:border-zinc-800 flex flex-wrap gap-2 justify-between items-center mt-auto">
          <div className="flex flex-wrap gap-2">
            {entity.linkedIds.length > 0 ? (
              entity.linkedIds.map(link => (
                <span key={link.targetId} className="flex items-center gap-2 px-3 py-1.5 bg-gray-50 dark:bg-zinc-800/50 border border-gray-200 dark:border-zinc-800 text-[10px] font-black rounded-xl text-gray-500 dark:text-zinc-400">
                  <Layout className="w-3 h-3 text-gray-300 dark:text-zinc-700" /> {link.relationName} · {project.entities.find(e => e.id === link.targetId)?.title}
                </span>
              ))
            ) : (
              <span className="text-[9px] font-bold text-gray-300 dark:text-zinc-700 italic">暂无外部关联</span>
            )}
          </div>
          <div className="flex items-center gap-1">
            <button onClick={(e) => { e.stopPropagation(); onEdit(); }} className="text-gray-400 dark:text-zinc-500 hover:text-indigo-600 dark:hover:text-indigo-400 p-2 transition-colors">
              <Edit3 className="w-4 h-4" />
            </button>
            <button 
                onClick={(e) => { 
                    e.stopPropagation(); 
                    if(confirm(`确定永久删除档案 "${entity.title}" 吗？\n此操作不可撤销，且会清除所有关联引用。`)) {
                        onDelete(entity.id); 
                    }
                }} 
                className="text-gray-300 dark:text-zinc-600 hover:text-red-400 dark:hover:text-red-500 p-2 transition-colors"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
      </div>
    </>
  );
};