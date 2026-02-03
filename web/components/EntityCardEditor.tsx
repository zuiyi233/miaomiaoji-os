import React, { useState } from 'react';
import { StoryEntity, Project } from '../types';
import { Mic2, Loader2, Wand2, Plus, Trash2, Save, Tag as TagIcon, X, PlusCircle, Sparkles, User, FileText, Settings2 } from 'lucide-react';
import { generateText } from '../services/aiService';
import { getTypeStyle, getEntityIcon } from './EntityVisuals';

interface EntityCardEditorProps {
  entity: StoryEntity;
  project: Project;
  categories: any[];
  onSave: (updates: Partial<StoryEntity>) => void;
  onCancel: () => void;
}

type TabType = 'identity' | 'content' | 'attributes';

export const EntityCardEditor: React.FC<EntityCardEditorProps> = ({ entity, project, categories, onSave, onCancel }) => {
  const [activeTab, setActiveTab] = useState<TabType>('identity');
  const [editTitle, setEditTitle] = useState(entity.title);
  const [editSubtitle, setEditSubtitle] = useState(entity.subtitle || '');
  const [editVoiceStyle, setEditVoiceStyle] = useState(entity.voiceStyle || '');
  const [editContent, setEditContent] = useState(entity.content);
  const [editImportance, setEditImportance] = useState<'main' | 'secondary' | 'minor'>(entity.importance || 'secondary');
  const [isExpandingContent, setIsExpandingContent] = useState(false);
  
  const [editTags, setEditTags] = useState(entity.tags || []);
  const [newTag, setNewTag] = useState('');

  const [customFields, setCustomFields] = useState<{key: string, value: string}[]>(entity.customFields || []);
  const [newFieldKey, setNewFieldKey] = useState('');
  const [newFieldValue, setNewFieldValue] = useState('');

  const handleExpandContent = async () => {
    if (!editTitle.trim()) {
      alert("请先输入实体名称，以便 AI 锚定创作对象");
      return;
    }
    setIsExpandingContent(true);
    try {
      const prompt = `为小说《${project.title}》中的${categories.find(c => c.id === entity.type)?.label}设定“${editTitle}”创作深度背景。
        【当前已知信息】：
        - 类型：${entity.type}
        - 现有副标题：${editSubtitle}
        - 核心标签：${editTags.join(', ')}
        ${entity.type === 'character' ? `- 语调风格：${editVoiceStyle}` : ''}
        
        【创作要求】：
        请以此为基础，撰写一段详细设定。文字风格需符合“${project.genre || '精致文学'}”的调性。`;

      const expanded = await generateText(
        prompt,
        `你是一位世界级的小说架构师。`,
        project.aiSettings
      );
      if (expanded) {
        setEditContent(expanded);
        setActiveTab('content'); 
      }
    } catch (e) {
      console.error("AI 扩写失败", e);
    }
    setIsExpandingContent(false);
  };

  const handleAddField = () => {
    if (newFieldKey.trim() && newFieldValue.trim()) {
      setCustomFields([...customFields, { key: newFieldKey.trim(), value: newFieldValue.trim() }]);
      setNewFieldKey('');
      setNewFieldValue('');
    }
  };

  const handleRemoveField = (index: number) => {
    setCustomFields(customFields.filter((_, i) => i !== index));
  };

  const handleAddTag = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && newTag.trim()) {
      if (!editTags.includes(newTag.trim())) {
        setEditTags([...editTags, newTag.trim()]);
      }
      setNewTag('');
    }
  };

  const handleRemoveTag = (tagToRemove: string) => {
    setEditTags(editTags.filter(t => t !== tagToRemove));
  };

  const handleSave = () => {
    onSave({
      title: editTitle,
      subtitle: editSubtitle,
      voiceStyle: editVoiceStyle,
      content: editContent,
      tags: editTags,
      importance: editImportance,
      customFields: customFields
    });
  };

  return (
    <div className="flex flex-col h-full bg-white dark:bg-zinc-900 relative transition-colors duration-300">
      <div className={`absolute -top-4 -right-4 px-6 py-3 rounded-bl-[2.5rem] text-[10px] font-black uppercase tracking-widest flex items-center gap-2 border-l border-b ${getTypeStyle(entity.type)} shadow-sm z-30 dark:bg-zinc-800`}>
        {getEntityIcon(entity.type)} {categories.find(c => c.id === entity.type)?.label}
      </div>

      <div className="flex border-b-2 border-gray-50 dark:border-zinc-800 mb-10 -mx-4">
        <button 
          onClick={() => setActiveTab('identity')}
          className={`flex-1 py-4 text-xs font-black uppercase tracking-[0.2em] flex items-center justify-center gap-3 transition-all ${activeTab === 'identity' ? 'text-brand-600 border-b-4 border-brand-600' : 'text-gray-300 dark:text-zinc-600 hover:text-gray-500 dark:hover:text-zinc-400'}`}
        >
          <User className="w-4 h-4" /> 身份档案
        </button>
        <button 
          onClick={() => setActiveTab('content')}
          className={`flex-1 py-4 text-xs font-black uppercase tracking-[0.2em] flex items-center justify-center gap-3 transition-all ${activeTab === 'content' ? 'text-brand-600 border-b-4 border-brand-600' : 'text-gray-300 dark:text-zinc-600 hover:text-gray-500 dark:hover:text-zinc-400'}`}
        >
          <FileText className="w-4 h-4" /> 核心设定
        </button>
        <button 
          onClick={() => setActiveTab('attributes')}
          className={`flex-1 py-4 text-xs font-black uppercase tracking-[0.2em] flex items-center justify-center gap-3 transition-all ${activeTab === 'attributes' ? 'text-brand-600 border-b-4 border-brand-600' : 'text-gray-300 dark:text-zinc-600 hover:text-gray-500 dark:hover:text-zinc-400'}`}
        >
          <Settings2 className="w-4 h-4" /> 数值/特质
        </button>
      </div>

      <div className="flex-1 overflow-y-auto custom-scrollbar px-2">
        {activeTab === 'identity' && (
          <div className="space-y-12 animate-in fade-in slide-in-from-bottom-4 duration-500">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-10 items-end">
              <div className="md:col-span-3 space-y-3">
                <label className="text-[11px] font-black text-gray-400 dark:text-zinc-600 uppercase tracking-[0.3em] block">实体名称</label>
                <div className="relative group">
                  <input 
                    className="w-full text-5xl font-black outline-none bg-transparent font-serif border-b-2 border-gray-100 dark:border-zinc-800 focus:border-brand-500 transition-all pb-4 pr-14 text-ink-900 dark:text-white" 
                    value={editTitle} 
                    onChange={e => setEditTitle(e.target.value)} 
                    placeholder="例如: 沈孤舟"
                  />
                  <button 
                    onClick={handleExpandContent}
                    disabled={isExpandingContent}
                    className="absolute right-0 bottom-6 p-2 text-brand-500 hover:text-brand-700 hover:bg-brand-50 dark:hover:bg-zinc-800 rounded-full transition-all disabled:opacity-30"
                    title="AI 智能构思设定"
                  >
                    {isExpandingContent ? <Loader2 className="w-6 h-6 animate-spin" /> : <Sparkles className="w-6 h-6" />}
                  </button>
                </div>
              </div>
              <div className="space-y-3">
                <label className="text-[11px] font-black text-gray-400 dark:text-zinc-600 uppercase tracking-[0.3em] block">重要度</label>
                <select 
                  value={editImportance}
                  onChange={e => setEditImportance(e.target.value as any)}
                  className="w-full bg-gray-50 dark:bg-zinc-800 border-2 border-gray-100 dark:border-zinc-700 rounded-2xl p-4 text-sm font-black text-gray-700 dark:text-zinc-300 outline-none focus:border-brand-500 transition-all"
                >
                  <option value="main">核心 (Main)</option>
                  <option value="secondary">重要 (Secondary)</option>
                  <option value="minor">次要 (Minor)</option>
                </select>
              </div>
            </div>

            <div className="space-y-3">
              <label className="text-[11px] font-black text-gray-400 dark:text-zinc-600 uppercase tracking-[0.3em] block">简短副标题</label>
              <input 
                className="w-full text-lg font-bold text-gray-500 dark:text-zinc-400 outline-none bg-transparent border-b-2 border-gray-100 dark:border-zinc-800 focus:border-brand-500 transition-all pb-2" 
                value={editSubtitle} 
                onChange={e => setEditSubtitle(e.target.value)} 
                placeholder="例如: 万剑山庄的叛徒 / 隐秘的遗迹入口"
              />
            </div>

            <div className="space-y-4 bg-gray-50/50 dark:bg-zinc-800/50 p-8 rounded-[2.5rem] border-2 border-gray-100 dark:border-zinc-800">
              <label className="text-[11px] font-black text-gray-500 dark:text-zinc-500 uppercase tracking-[0.3em] block flex items-center gap-3">
                <TagIcon className="w-4 h-4" /> 设定标签
              </label>
              <div className="flex flex-wrap gap-3 items-center">
                {editTags.map((tag, idx) => (
                  <span key={idx} className="flex items-center gap-2 px-4 py-2.5 bg-white dark:bg-zinc-900 border-2 border-gray-200 dark:border-zinc-700 rounded-2xl text-xs font-black text-gray-700 dark:text-zinc-300 shadow-sm animate-in zoom-in-95">
                    {tag}
                    <button onClick={() => handleRemoveTag(tag)} className="text-gray-300 dark:text-zinc-600 hover:text-red-500 transition-colors"><X className="w-4 h-4" /></button>
                  </span>
                ))}
                <input 
                  value={newTag}
                  onChange={e => setNewTag(e.target.value)}
                  onKeyDown={handleAddTag}
                  placeholder="+ 键入新标签并回车"
                  className="bg-transparent text-xs font-bold text-gray-400 dark:text-zinc-600 outline-none px-4 py-2 w-48"
                />
              </div>
            </div>
          </div>
        )}

        {activeTab === 'content' && (
          <div className="space-y-6 h-full flex flex-col animate-in fade-in slide-in-from-bottom-4 duration-500">
            <div className="flex justify-between items-center px-2">
              <label className="text-[11px] font-black text-gray-400 dark:text-zinc-600 uppercase tracking-[0.3em] block">详细档案设定</label>
              <button 
                onClick={handleExpandContent} 
                disabled={isExpandingContent} 
                className="text-[11px] font-black text-brand-600 dark:text-brand-400 uppercase flex items-center gap-2 hover:text-brand-800 dark:hover:text-brand-300 transition-all disabled:opacity-50"
              >
                {isExpandingContent ? <Loader2 className="w-4 h-4 animate-spin" /> : <Wand2 className="w-4 h-4" />} 
                {isExpandingContent ? '缪斯推演中...' : 'AI 智能扩写'}
              </button>
            </div>
            <textarea 
              className="flex-1 min-h-[400px] text-lg text-gray-700 dark:text-zinc-300 bg-gray-50/30 dark:bg-zinc-800/30 p-10 rounded-[3rem] outline-none font-serif leading-relaxed resize-none focus:ring-8 focus:ring-brand-50 dark:focus:ring-brand-900/10 border-2 border-transparent focus:border-brand-200 dark:focus:border-zinc-700 transition-all shadow-inner" 
              value={editContent} 
              onChange={e => setEditContent(e.target.value)} 
              placeholder="在这里深入描写该实体的历史渊源、外貌特征、核心矛盾、以及在故事中所扮演的关键角色..."
              spellCheck={false}
            />
          </div>
        )}

        {activeTab === 'attributes' && (
          <div className="space-y-10 animate-in fade-in slide-in-from-bottom-4 duration-500">
            {entity.type === 'character' && (
              <div className="space-y-4 bg-rose-50/40 dark:bg-rose-900/10 p-8 rounded-[2.5rem] border-2 border-rose-100 dark:border-rose-900/20">
                <label className="text-[11px] font-black text-rose-500 dark:text-rose-400 uppercase tracking-[0.3em] flex items-center gap-3">
                  <Mic2 className="w-5 h-5" /> 语调与语言风格
                </label>
                <textarea 
                  className="w-full bg-transparent border-none outline-none text-sm font-bold text-rose-900 dark:text-rose-200 placeholder:text-rose-200 dark:placeholder:text-rose-900/50 leading-loose resize-none h-24"
                  value={editVoiceStyle}
                  onChange={e => setEditVoiceStyle(e.target.value)}
                  placeholder="例如：冷冽寡言，多用短句。总是带有隐晦的嘲讽，拒绝使用敬语..."
                />
              </div>
            )}

            <div className="space-y-6 bg-cyan-50/20 dark:bg-cyan-900/10 p-8 rounded-[2.5rem] border-2 border-cyan-100/50 dark:border-cyan-900/20">
              <label className="text-[11px] font-black text-cyan-600 dark:text-cyan-400 uppercase tracking-[0.3em] block">动态数值/特质属性</label>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {customFields.map((field, idx) => (
                  <div key={idx} className="flex gap-3 items-center bg-white dark:bg-zinc-800 border-2 border-cyan-50 dark:border-zinc-700 p-3 rounded-[1.5rem] shadow-sm animate-in slide-in-from-left-2">
                    <span className="px-4 py-2 bg-cyan-50 dark:bg-cyan-900/30 rounded-xl text-[10px] font-black text-cyan-700 dark:text-cyan-400 uppercase tracking-tighter min-w-[80px] text-center">{field.key}</span>
                    <span className="flex-1 text-sm font-bold text-gray-700 dark:text-zinc-300 truncate">{field.value}</span>
                    <button onClick={() => handleRemoveField(idx)} className="p-2 text-gray-300 dark:text-zinc-600 hover:text-red-500 transition-colors"><Trash2 className="w-5 h-5" /></button>
                  </div>
                ))}
              </div>
              
              <div className="flex gap-4 mt-8 pt-8 border-t-2 border-cyan-100/30 dark:border-zinc-800">
                <input 
                  value={newFieldKey}
                  onChange={e => setNewFieldKey(e.target.value)}
                  placeholder="属性名 (如: 战力)"
                  className="w-1/3 px-5 py-4 bg-white dark:bg-zinc-800 border-2 border-gray-100 dark:border-zinc-700 rounded-2xl text-sm font-black outline-none focus:ring-4 focus:ring-cyan-100 dark:focus:ring-cyan-900/30 transition-all text-ink-900 dark:text-zinc-100"
                />
                <input 
                  value={newFieldValue}
                  onChange={e => setNewFieldValue(e.target.value)}
                  placeholder="值 (如: S级 / 剑道宗师)"
                  className="flex-1 px-5 py-4 bg-white dark:bg-zinc-800 border-2 border-gray-100 dark:border-zinc-700 rounded-2xl text-sm font-black outline-none focus:ring-4 focus:ring-cyan-100 dark:focus:ring-cyan-900/30 transition-all text-ink-900 dark:text-zinc-100"
                />
                <button 
                  onClick={handleAddField}
                  disabled={!newFieldKey.trim() || !newFieldValue.trim()}
                  className="p-4 bg-cyan-600 dark:bg-cyan-500 text-white rounded-2xl hover:bg-cyan-700 dark:hover:bg-cyan-600 disabled:opacity-30 transition-all shadow-lg active:scale-95"
                >
                  <PlusCircle className="w-6 h-6" />
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      <div className="flex gap-6 pt-10 mt-6 border-t-2 border-gray-50 dark:border-zinc-800 bg-white dark:bg-zinc-900 transition-colors">
        <button 
          onClick={onCancel} 
          className="flex-1 py-5 text-xs font-black text-gray-400 dark:text-zinc-600 uppercase tracking-widest hover:text-gray-900 dark:hover:text-zinc-300 transition-all bg-gray-50 dark:bg-zinc-800 rounded-2xl hover:bg-gray-100 dark:hover:bg-zinc-700"
        >
          丢弃修改
        </button>
        <button 
          onClick={handleSave} 
          className="flex-[2] py-5 bg-gray-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-2xl text-xs font-black uppercase tracking-widest flex items-center justify-center gap-3 hover:bg-black dark:hover:bg-white shadow-[0_20px_40px_-10px_rgba(0,0,0,0.3)] dark:shadow-black/20 transition-all active:scale-[0.98]"
        >
           <Save className="w-5 h-5" /> 确认保存档案
        </button>
      </div>
    </div>
  );
};