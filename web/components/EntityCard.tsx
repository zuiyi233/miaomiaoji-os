import React, { useState, useEffect } from 'react';
import { StoryEntity, Project } from '../types';
import { EntityCardView } from './EntityCardView';
import { EntityCardEditor } from './EntityCardEditor';

interface EntityCardProps {
  entity: StoryEntity;
  project: Project;
  categories: any[];
  onUpdate: (id: string, updates: Partial<StoryEntity>) => void;
  onDelete: (id: string) => void;
  onLink: (id: string) => void;
  autoEdit?: boolean;
}

export const EntityCard: React.FC<EntityCardProps> = ({ entity, project, onUpdate, onDelete, onLink, categories, autoEdit }) => {
  const [isEditing, setIsEditing] = useState(false);

  useEffect(() => {
    if (autoEdit) {
      setIsEditing(true);
    }
  }, [autoEdit]);

  const handleSave = (updates: Partial<StoryEntity>) => {
    onUpdate(entity.id, updates);
    setIsEditing(false);
  };

  return (
    <>
      {/* 正常视图状态的卡片 */}
      <div className={`group bg-white dark:bg-zinc-900 border-2 border-transparent dark:border-zinc-800 rounded-[2.5rem] p-8 transition-all relative overflow-hidden flex flex-col min-h-[420px] shadow-sm hover:shadow-2xl dark:hover:shadow-black/20 hover:-translate-y-2`}>
        <EntityCardView 
          entity={entity} 
          project={project} 
          categories={categories} 
          onEdit={() => setIsEditing(true)} 
          onDelete={onDelete} 
        />
      </div>

      {/* 沉浸式全屏聚焦层 */}
      {isEditing && (
        <div className="fixed inset-0 z-[1000] flex items-center justify-center p-4 md:p-8 animate-in fade-in duration-300">
          {/* 高级磨砂背景遮罩 */}
          <div 
            className="absolute inset-0 bg-gray-900/60 dark:bg-black/80 backdrop-blur-xl transition-all cursor-zoom-out" 
            onClick={() => setIsEditing(false)} 
          />
          
          {/* 扩张后的编辑器核心容器 */}
          <div className="relative w-full max-w-[96vw] h-[92vh] bg-white dark:bg-zinc-900 rounded-[4rem] shadow-[0_50px_100px_-20px_rgba(0,0,0,0.5)] overflow-hidden animate-in zoom-in-95 duration-500 flex flex-col border border-white/30 dark:border-white/5">
            <div className="flex-1 overflow-hidden p-10 md:p-20 lg:p-24">
              <EntityCardEditor 
                entity={entity} 
                project={project} 
                categories={categories} 
                onSave={handleSave} 
                onCancel={() => setIsEditing(false)} 
              />
            </div>
          </div>
        </div>
      )}
    </>
  );
};