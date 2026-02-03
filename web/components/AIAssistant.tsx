
import React, { useState, useEffect, useRef } from 'react';
import { useProject } from '../contexts/ProjectContext';
import { chatWithMuseStream } from '../services/geminiService';
import { 
  X, Send, Sparkles, User, Bot, Loader2, Info, ChevronDown, Plus, Trash2, 
  History, Wand2, MessageSquare, BookOpen, Layers, Check, Copy, ExternalLink,
  Search, Bookmark, Zap, Command, Menu, MoreVertical, Edit2, ShoppingBag, PlusCircle, Save, FileText, Code, ChevronUp, PenTool, Megaphone, UserPlus, Flame, Target, Share2, Type as TypeIcon
} from 'lucide-react';
import { ChatMessage, AIPromptTemplate, ChatSession } from '../types';

const STORAGE_SESSIONS_KEY = 'novel_agent_chat_sessions_v1';

const MarkdownLite: React.FC<{ content: string }> = ({ content }) => {
  const parts = content.split(/(```[\s\S]*?```)/g);
  const renderTextWithLinks = (text: string) => {
    const urlRegex = /(?<!\[)https?:\/\/[^\s<]+(?![^<]*>)/g;
    return text.split(urlRegex).map((part, i) => {
      if (part.match(/https?:\/\/[^\s<]+/)) {
        return (
          <a key={i} href={part} target="_blank" rel="noopener noreferrer" className="text-brand-500 dark:text-brand-400 hover:underline inline-flex items-center gap-0.5">
            {part} <ExternalLink className="w-2.5 h-2.5" />
          </a>
        );
      }
      return part;
    });
  };

  return (
    <div className="space-y-3 font-sans text-sm leading-relaxed">
      {parts.map((part, i) => {
        if (part.startsWith('```')) {
          const code = part.replace(/```(\w+)?\n?/, '').replace(/```$/, '');
          const lang = part.match(/```(\w+)?/)?.[1] || 'code';
          return <CodeBlock key={i} code={code} lang={lang} />;
        }
        return (
          <div key={i} className="whitespace-pre-wrap">
            {part.split('\n').map((line, li) => {
              if (line.startsWith('- ') || line.startsWith('* ')) return <li key={li} className="ml-4 list-disc">{renderTextWithLinks(line.substring(2))}</li>;
              if (line.match(/^\d+\. /)) return <li key={li} className="ml-4 list-decimal">{renderTextWithLinks(line.replace(/^\d+\. /, ''))}</li>;
              return <p key={li} className={line.trim() === '' ? 'h-2' : ''}>{renderTextWithLinks(line)}</p>;
            })}
          </div>
        );
      })}
    </div>
  );
};

const CodeBlock: React.FC<{ code: string, lang: string }> = ({ code, lang }) => {
    // Fix: Corrected useState destructuring to prevent incorrect type inference as literal 'false'
    const [copied, setCopied] = useState(false);
    const handleCopy = () => {
        navigator.clipboard.writeText(code);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };
    return (
        <div className="bg-zinc-900 text-zinc-100 rounded-xl overflow-hidden my-3 border border-zinc-800 shadow-lg">
            <div className="flex items-center justify-between px-4 py-2 bg-zinc-800/50 border-b border-zinc-800">
                <span className="text-[10px] font-black text-zinc-500 uppercase tracking-widest">{lang}</span>
                <button onClick={handleCopy} className="flex items-center gap-1.5 px-2 py-1 rounded bg-zinc-700 hover:bg-zinc-600 text-[10px] text-zinc-300 transition-all">
                    {copied ? <Check className="w-3 h-3 text-green-400" /> : <Copy className="w-3 h-3" />}
                    <span>{copied ? '已复制' : '复制'}</span>
                </button>
            </div>
            <pre className="p-4 overflow-x-auto text-xs font-mono scrollbar-hide"><code>{code}</code></pre>
        </div>
    );
};

type TabType = 'chat' | 'history' | 'skills';

const PRESET_SKILLS: AIPromptTemplate[] = [
  { id: 'sk1', name: '逻辑自查', description: '扫描当前章节的设定冲突与逻辑漏洞', category: 'logic', template: '作为文学评论家，分析以下章节内容：\n{{content}}\n是否存在逻辑漏洞，特别是是否违背了设定集中的规则？' },
  { id: 'sk2', name: '环境渲染', description: '使用五感法增强当前场景细节', category: 'style', template: '请为以下内容增加更多的环境细节、感官体验（视觉、听觉、嗅觉等）：\n{{content}}' },
  { id: 'sk3', name: '台词校准', description: '使角色对话更符合人设语调', category: 'character', template: '润色以下段落中的对话。确保角色的语气符合其语调风格：\n{{content}}' },
  { id: 'sk4', name: '情节发散', description: '推演后续三个高吸引力走向', category: 'content', template: '根据当前情节：\n{{content}}\n请提出三个极具创意的后续发展方向。' },
];

const WRITING_TOOLS: AIPromptTemplate[] = [
  { id: 'wt1', name: '起名废救星', description: '根据小说风格生成角色/势力名', category: 'character', template: '根据小说类型（{{genre}}）和风格，生成10个独特、有内涵且符合人设的名字，并附带简短的起名含义说明。' },
  { id: 'wt2', name: '神级转折点', description: '为当前剧情设计出人意料的转折', category: 'content', template: '当前章节标题：{{title}}\n内容摘要：{{content}}\n请以此为基础，设计三个出人意料的转折点，要求能瞬间打破现状且符合逻辑。' },
  { id: 'wt3', name: '冲突压力测试', description: '加剧角色间的矛盾与两难处境', category: 'logic', template: '分析以下片段：\n{{content}}\n提出一种方法来加剧其中的冲突张力，让角色陷入更艰难的道德或现实困境。' },
  { id: 'wt4', name: '修辞实验室', description: '将平淡的叙述转化为华丽或深刻的文字', category: 'style', template: '请重写以下段落，运用高级修辞、隐喻和诗意的语言，提升其文学质感：\n{{content}}' },
  { id: 'wt5', name: '黄金钩子', description: '优化章节结尾的悬念设计', category: 'content', template: '分析本章结尾：\n{{content}}\n请优化这个结尾，使其成为一个不可抗拒的“黄金钩子”，强制吸引读者点击下一章。' },
];

const PROMO_TOOLS: AIPromptTemplate[] = [
  { id: 'pt1', name: '爆款标题党', description: '生成极具点击欲的标题', category: 'content', template: '根据小说核心冲突（{{coreConflict}}），设计5个符合当下流行趋势（如小红书、知乎风格）的极具传播力的吸引性标题。' },
  { id: 'pt2', name: '悬念简介', description: '撰写抓人的书籍背书文案', category: 'style', template: '根据小说的大纲，撰写一段200字以内的悬念简介，要求节奏紧凑，突出矛盾，让读者一眼入坑。' },
  { id: 'pt3', name: '社媒预告', description: '适合微博/小红书的分享短文', category: 'style', template: '从以下内容中摘取精彩片段：\n{{content}}\n并将其改编为一段适合在社交平台发布的精彩预告，要求文字具有网感，能引发广泛讨论。' },
  { id: 'pt4', name: '角色高光推介', description: '为主角打造的高能宣传词', category: 'character', template: '为角色（{{protagonistName}}）撰写一段“高光时刻”的推介文案，展示其独特魅力、金句以及最震撼的成长瞬间。' },
  { id: 'pt5', name: '分卷推广词', description: '总结本卷的情节亮点', category: 'content', template: '总结本卷（{{volumeTitle}}）的情节亮点，生成一段富有史诗感或吸引力、适合作为分卷开头的推广词。' },
];

export const AIAssistant: React.FC = () => {
  const { isAISidebarOpen, toggleAISidebar, project, activeDocumentId, addTemplate, deleteTemplate } = useProject();
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const [input, setInput] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const [activeTab, setActiveTab] = useState<TabType>('chat');
  const [copiedMsgId, setCopiedMsgId] = useState<string | null>(null);
  
  // States for dynamic tool menus
  const [openMenu, setOpenMenu] = useState<'writing' | 'promo' | null>(null);
  const [isQuickPickerOpen, setIsQuickPickerOpen] = useState(false);
  const [quickSearch, setQuickSearch] = useState('');
  
  // Custom Skill Management States
  const [isCreatingSkill, setIsCreatingSkill] = useState(false);
  const [newSkill, setNewSkill] = useState({ name: '', description: '', template: '', category: 'content' as any });

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const pickerRef = useRef<HTMLDivElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const saved = localStorage.getItem(STORAGE_SESSIONS_KEY);
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        setSessions(parsed);
        if (parsed.length > 0) setActiveSessionId(parsed[0].id);
      } catch (e) {
        createNewSession();
      }
    } else {
      createNewSession();
    }
  }, []);

  useEffect(() => {
    if (sessions.length > 0) {
      localStorage.setItem(STORAGE_SESSIONS_KEY, JSON.stringify(sessions));
    }
  }, [sessions]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [sessions, isTyping, activeSessionId, activeTab]);

  // Click outside handler for all floating menus
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (pickerRef.current && !pickerRef.current.contains(event.target as Node)) {
        setIsQuickPickerOpen(false);
      }
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setOpenMenu(null);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const createNewSession = () => {
    const newId = Date.now().toString();
    const newSession: ChatSession = {
      id: newId,
      title: '新对话',
      messages: [{ id: 'init', role: 'model', text: "你好！我是你的写作缪斯。今天想在你的故事宇宙中探索些什么？", timestamp: Date.now() }],
      lastUpdated: Date.now()
    };
    setSessions(prev => [newSession, ...prev]);
    setActiveSessionId(newId);
    setActiveTab('chat');
  };

  const deleteSession = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    const newSessions = sessions.filter(s => s.id !== id);
    setSessions(newSessions);
    if (activeSessionId === id) {
      setActiveSessionId(newSessions.length > 0 ? newSessions[0].id : null);
    }
  };

  const currentSession = sessions.find(s => s.id === activeSessionId);

  const handleSend = async (customText?: string) => {
    const textToSend = customText || input;
    if (!textToSend.trim() || isTyping || !project || !activeSessionId) return;

    const userMsg: ChatMessage = { id: Date.now().toString(), role: 'user', text: textToSend, timestamp: Date.now() };
    
    setSessions(prev => prev.map(s => {
      if (s.id === activeSessionId) {
        const newTitle = s.messages.length === 1 ? textToSend.substring(0, 15) + (textToSend.length > 15 ? '...' : '') : s.title;
        return { ...s, title: newTitle, messages: [...s.messages, userMsg], lastUpdated: Date.now() };
      }
      return s;
    }));

    setInput('');
    setIsTyping(true);

    const history = (currentSession?.messages || []).slice(-10).map(m => ({ role: m.role, parts: [{ text: m.text }] }));
    const botId = (Date.now() + 1).toString();
    const botMsg: ChatMessage = { id: botId, role: 'model', text: '', timestamp: Date.now() };
    
    setSessions(prev => prev.map(s => s.id === activeSessionId ? { ...s, messages: [...s.messages, botMsg] } : s));

    try {
      const stream = chatWithMuseStream(history, textToSend, project.aiSettings);
      let fullText = '';
      for await (const chunk of stream) {
        fullText += chunk;
        setSessions(prev => prev.map(s => {
          if (s.id === activeSessionId) {
            return { ...s, messages: s.messages.map(m => m.id === botId ? { ...m, text: fullText } : m) };
          }
          return s;
        }));
      }
    } catch (err) {
      console.error("AI Assistant Stream Error:", err);
    } finally {
      setIsTyping(false);
    }
  };

  const loadSkillToInput = (tpl: AIPromptTemplate) => {
    if (!project) return;
    const activeDoc = project.documents.find(d => d.id === activeDocumentId);
    let processed = tpl.template;
    if (activeDoc) {
      processed = processed.replace(/{{content}}/g, activeDoc.content);
      processed = processed.replace(/{{title}}/g, activeDoc.title);
      processed = processed.replace(/{{volumeTitle}}/g, project.volumes.find(v => v.id === activeDoc.volumeId)?.title || '');
    }
    processed = processed.replace(/{{genre}}/g, project.genre || '精致文学');
    processed = processed.replace(/{{coreConflict}}/g, project.coreConflict || '');
    processed = processed.replace(/{{protagonistName}}/g, project.entities.find(e => e.importance === 'main')?.title || '主角');

    setInput(processed);
    setIsQuickPickerOpen(false);
    setOpenMenu(null);
    setQuickSearch('');
    setActiveTab('chat');
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const val = e.target.value;
    setInput(val);
    if (val.endsWith('/')) {
      setIsQuickPickerOpen(true);
    }
  };

  // Safe concatenation of skills, guarding against project being null
  const filteredSkills = [
    ...PRESET_SKILLS, 
    ...(project?.templates || []), 
    ...WRITING_TOOLS, 
    ...PROMO_TOOLS
  ].filter(s => 
    s.name.toLowerCase().includes(quickSearch.toLowerCase()) || 
    s.description.toLowerCase().includes(quickSearch.toLowerCase())
  );

  const handleCreateCustomSkill = () => {
    if (!newSkill.name || !newSkill.template) return;
    addTemplate(newSkill.name, newSkill.template, newSkill.description, newSkill.category);
    setNewSkill({ name: '', description: '', template: '', category: 'content' as any });
    setIsCreatingSkill(false);
  };

  if (!isAISidebarOpen) return null;

  return (
    <div className="w-96 bg-white dark:bg-zinc-900 border-l border-gray-200 dark:border-zinc-800 flex flex-col h-full shadow-2xl relative z-30 transition-all animate-in slide-in-from-right duration-500 pt-14">
      {/* 头部控制栏 - 使用吸顶且高度适配 */}
      <div className="h-14 border-b border-gray-100 dark:border-zinc-800 flex items-center justify-between px-4 bg-white dark:bg-zinc-900 absolute top-0 left-0 right-0 z-40">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-1.5 text-brand-600 dark:text-brand-400 font-black tracking-tight shrink-0">
            <Sparkles className="w-4 h-4" /> 缪斯
          </div>
          
          {/* Quick Tool Access */}
          <div className="flex items-center gap-1 border-l border-gray-100 dark:border-zinc-800 pl-3">
             <div className="relative" ref={openMenu === 'writing' ? menuRef : null}>
                <button 
                    onClick={() => setOpenMenu(openMenu === 'writing' ? null : 'writing')}
                    className={`p-1.5 rounded-lg transition-all ${openMenu === 'writing' ? 'bg-brand-500 text-white' : 'text-gray-400 hover:bg-gray-100 dark:hover:bg-zinc-800'}`}
                    title="写作实验室"
                >
                    <PenTool className="w-4 h-4" />
                </button>
                {openMenu === 'writing' && (
                  <div className="absolute top-10 left-0 w-48 bg-white dark:bg-zinc-900 rounded-2xl shadow-2xl border border-gray-100 dark:border-zinc-800 p-2 animate-in zoom-in-95 duration-200">
                     <div className="px-3 py-1 text-[9px] font-black text-gray-400 uppercase tracking-widest mb-1">写作实验室</div>
                     {WRITING_TOOLS.map(tool => (
                        <button key={tool.id} onClick={() => loadSkillToInput(tool)} className="w-full text-left px-3 py-2 rounded-xl hover:bg-brand-50 dark:hover:bg-brand-900/20 flex items-center gap-2 group">
                           <Zap className="w-3 h-3 text-brand-500" />
                           <span className="text-[11px] font-bold text-gray-700 dark:text-zinc-300 group-hover:text-brand-600">{tool.name}</span>
                        </button>
                     ))}
                  </div>
                )}
             </div>

             <div className="relative" ref={openMenu === 'promo' ? menuRef : null}>
                <button 
                    onClick={() => setOpenMenu(openMenu === 'promo' ? null : 'promo')}
                    className={`p-1.5 rounded-lg transition-all ${openMenu === 'promo' ? 'bg-indigo-500 text-white' : 'text-gray-400 hover:bg-gray-100 dark:hover:bg-zinc-800'}`}
                    title="宣发中心"
                >
                    <Megaphone className="w-4 h-4" />
                </button>
                {openMenu === 'promo' && (
                  <div className="absolute top-10 left-0 w-48 bg-white dark:bg-zinc-900 rounded-2xl shadow-2xl border border-gray-100 dark:border-zinc-800 p-2 animate-in zoom-in-95 duration-200">
                     <div className="px-3 py-1 text-[9px] font-black text-gray-400 uppercase tracking-widest mb-1">宣发中心</div>
                     {PROMO_TOOLS.map(tool => (
                        <button key={tool.id} onClick={() => loadSkillToInput(tool)} className="w-full text-left px-3 py-2 rounded-xl hover:bg-indigo-50 dark:hover:bg-indigo-900/20 flex items-center gap-2 group">
                           <Share2 className="w-3 h-3 text-indigo-500" />
                           <span className="text-[11px] font-bold text-gray-700 dark:text-zinc-300 group-hover:text-indigo-600">{tool.name}</span>
                        </button>
                     ))}
                  </div>
                )}
             </div>
          </div>
        </div>

        <div className="flex items-center gap-1">
          <button onClick={createNewSession} className="p-2 text-gray-400 hover:text-brand-600 dark:hover:text-brand-400 transition-colors" title="新建对话"><Plus className="w-4 h-4" /></button>
          <button onClick={toggleAISidebar} className="p-2 hover:bg-gray-100 dark:hover:bg-zinc-800 rounded-lg text-gray-400 dark:text-zinc-500"><X className="w-4 h-4" /></button>
        </div>
      </div>

      {/* Tabs 导航 */}
      <div className="flex border-b border-gray-100 dark:border-zinc-800 bg-gray-50/50 dark:bg-zinc-900/50 relative z-30">
        {[
          { id: 'chat', label: '对话', icon: <MessageSquare className="w-3.5 h-3.5" /> },
          { id: 'history', label: '版本', icon: <History className="w-3.5 h-3.5" /> },
          { id: 'skills', label: '仓库', icon: <ShoppingBag className="w-3.5 h-3.5" /> }
        ].map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id as TabType)}
            className={`flex-1 flex items-center justify-center gap-2 py-3 text-[11px] font-black uppercase tracking-widest transition-all ${
              activeTab === tab.id 
                ? 'text-brand-600 dark:text-brand-400 bg-white dark:bg-zinc-900 border-b-2 border-brand-600 dark:border-brand-400' 
                : 'text-gray-400 hover:text-gray-600 dark:hover:text-zinc-300'
            }`}
          >
            {tab.icon} {tab.label}
          </button>
        ))}
      </div>

      {/* 主内容区 */}
      <div className="flex-1 flex flex-col min-h-0 relative">
        {activeTab === 'chat' && (
          <>
            <div className="flex-1 overflow-y-auto p-4 space-y-6 bg-gray-50/50 dark:bg-zinc-950/30 scroll-smooth">
              {currentSession ? (
                currentSession.messages.map((msg) => (
                  <div key={msg.id} className={`flex flex-col group ${msg.role === 'user' ? 'items-end' : 'items-start'}`}>
                    <span className="text-[9px] font-black text-gray-300 dark:text-zinc-700 uppercase tracking-widest mb-1 px-1">
                        {msg.role === 'user' ? '作者' : '缪斯'} · {new Date(msg.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                    </span>
                    <div className={`relative max-w-[90%] rounded-2xl px-4 py-3 shadow-sm transition-colors ${
                        msg.role === 'user' 
                        ? 'bg-zinc-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-tr-none' 
                        : 'bg-white dark:bg-zinc-900 border border-gray-100 dark:border-zinc-800 text-gray-800 dark:text-zinc-200 rounded-tl-none'
                    }`}>
                      <MarkdownLite content={msg.text} />
                      <button 
                        onClick={() => { navigator.clipboard.writeText(msg.text); setCopiedMsgId(msg.id); setTimeout(()=>setCopiedMsgId(null), 2000); }}
                        className={`absolute ${msg.role === 'user' ? '-left-8' : '-right-8'} top-1/2 -translate-y-1/2 p-1.5 text-gray-300 hover:text-brand-500 opacity-0 group-hover:opacity-100 transition-all`}
                      >
                        {copiedMsgId === msg.id ? <Check className="w-3 h-3 text-green-500" /> : <Copy className="w-3 h-3" />}
                      </button>
                    </div>
                  </div>
                ))
              ) : (
                <div className="h-full flex flex-col items-center justify-center text-center space-y-4 opacity-40">
                    <MessageSquare className="w-12 h-12" />
                    <p className="text-sm font-bold">选择一个会话开始探讨</p>
                </div>
              )}
              {isTyping && (
                <div className="flex gap-2 items-center p-2">
                  <Loader2 className="w-4 h-4 animate-spin text-brand-500" />
                  <span className="text-[10px] font-black text-brand-500 uppercase tracking-widest">灵感闪现中...</span>
                </div>
              )}
              <div ref={messagesEndRef} />
            </div>

            {/* 智能指令弹窗 (Quick Command Palette) */}
            {isQuickPickerOpen && (
              <div ref={pickerRef} className="absolute bottom-[90px] left-4 right-4 bg-white/90 dark:bg-zinc-900/95 backdrop-blur-xl border border-gray-100 dark:border-zinc-800 rounded-[2rem] shadow-[0_20px_50px_rgba(0,0,0,0.2)] z-50 flex flex-col max-h-[400px] overflow-hidden animate-in slide-in-from-bottom-4 zoom-in-95 duration-200">
                  <div className="p-4 border-b border-gray-50 dark:border-zinc-800 flex items-center gap-3 bg-white/50 dark:bg-zinc-900/50 backdrop-blur-md sticky top-0 z-10">
                      <Search className="w-4 h-4 text-gray-400" />
                      <input 
                        autoFocus
                        value={quickSearch}
                        onChange={e => setQuickSearch(e.target.value)}
                        placeholder="搜索技能、工具或提示词..."
                        className="flex-1 bg-transparent text-sm font-bold outline-none text-ink-900 dark:text-zinc-100"
                      />
                      <button onClick={() => setIsQuickPickerOpen(false)} className="p-1 hover:bg-gray-100 dark:hover:bg-zinc-800 rounded-lg text-gray-400"><X className="w-4 h-4" /></button>
                  </div>
                  <div className="flex-1 overflow-y-auto p-2 custom-scrollbar space-y-6">
                      {/* Presets Group */}
                      <div>
                          <h5 className="px-3 py-1 text-[9px] font-black text-gray-400 uppercase tracking-widest flex items-center gap-1.5"><Sparkles className="w-3 h-3"/> 核心创作工具</h5>
                          <div className="grid grid-cols-1 gap-1">
                              {filteredSkills.filter(s => s.id.startsWith('sk') || s.id.startsWith('wt')).map(skill => (
                                  <button 
                                      key={skill.id}
                                      onClick={() => loadSkillToInput(skill)}
                                      className="flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-brand-50 dark:hover:bg-brand-900/20 text-left transition-colors group/item"
                                  >
                                      <div className={`p-2 rounded-lg ${skill.id.startsWith('wt') ? 'bg-amber-50 dark:bg-amber-900/30 text-amber-500' : 'bg-brand-50 dark:bg-brand-900/30 text-brand-600'} group-hover/item:scale-110 transition-transform`}>
                                          {skill.id.startsWith('wt') ? <PenTool className="w-3.5 h-3.5"/> : <Zap className="w-3.5 h-3.5"/>}
                                      </div>
                                      <div className="flex flex-col truncate">
                                          <span className="text-xs font-bold text-ink-900 dark:text-zinc-100">{skill.name}</span>
                                          <span className="text-[10px] text-gray-400 truncate">{skill.description}</span>
                                      </div>
                                  </button>
                              ))}
                          </div>
                      </div>

                      {/* Promo Group */}
                      <div>
                          <h5 className="px-3 py-1 text-[9px] font-black text-indigo-400 uppercase tracking-widest flex items-center gap-1.5"><Megaphone className="w-3 h-3"/> 宣发推广组件</h5>
                          <div className="grid grid-cols-1 gap-1">
                              {filteredSkills.filter(s => s.id.startsWith('pt')).map(skill => (
                                  <button 
                                      key={skill.id}
                                      onClick={() => loadSkillToInput(skill)}
                                      className="flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-indigo-50 dark:hover:bg-indigo-900/20 text-left transition-colors group/item"
                                  >
                                      <div className="p-2 bg-indigo-50 dark:bg-indigo-900/30 rounded-lg text-indigo-600 dark:text-indigo-400 group-hover/item:scale-110 transition-transform"><Share2 className="w-3.5 h-3.5"/></div>
                                      <div className="flex flex-col truncate">
                                          <span className="text-xs font-bold text-ink-900 dark:text-zinc-100">{skill.name}</span>
                                          <span className="text-[10px] text-gray-400 truncate">{skill.description}</span>
                                      </div>
                                  </button>
                              ))}
                          </div>
                      </div>

                      {/* Custom Group */}
                      {project && (project?.templates || []).length > 0 && (
                        <div>
                            <h5 className="px-3 py-1 text-[9px] font-black text-emerald-400 uppercase tracking-widest flex items-center gap-1.5"><Bookmark className="w-3 h-3"/> 我的工坊</h5>
                            <div className="grid grid-cols-1 gap-1">
                                {filteredSkills.filter(s => s.id.startsWith('t')).map(skill => (
                                    <button 
                                        key={skill.id}
                                        onClick={() => loadSkillToInput(skill)}
                                        className="flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-emerald-50 dark:hover:bg-emerald-900/20 text-left transition-colors group/item"
                                    >
                                        <div className="p-2 bg-emerald-50 dark:bg-emerald-900/30 rounded-lg text-emerald-600 dark:text-emerald-400 group-hover/item:scale-110 transition-transform"><Bookmark className="w-3.5 h-3.5"/></div>
                                        <div className="flex flex-col truncate">
                                            <span className="text-xs font-bold text-ink-900 dark:text-zinc-100">{skill.name}</span>
                                            <span className="text-[10px] text-gray-400 truncate">{skill.description}</span>
                                        </div>
                                    </button>
                                ))}
                            </div>
                        </div>
                      )}

                      {filteredSkills.length === 0 && (
                          <div className="py-8 text-center text-gray-400">
                              <p className="text-xs font-bold">未找到匹配技能</p>
                              <button onClick={() => { setActiveTab('skills'); setIsQuickPickerOpen(false); }} className="text-[10px] font-black text-brand-600 uppercase mt-2 hover:underline">去仓库查看全部</button>
                          </div>
                      )}
                  </div>
              </div>
            )}

            {/* 输入栏 */}
            <div className="p-4 bg-white dark:bg-zinc-900 border-t border-gray-100 dark:border-zinc-800">
                <div className="flex items-end gap-2 bg-gray-50 dark:bg-zinc-950 p-2 rounded-2xl border border-gray-200 dark:border-zinc-800 focus-within:ring-2 focus-within:ring-brand-100 dark:focus-within:ring-brand-900/30 transition-all">
                <button 
                    onClick={() => setIsQuickPickerOpen(!isQuickPickerOpen)}
                    className={`p-3 rounded-xl transition-all ${isQuickPickerOpen ? 'bg-brand-600 text-white' : 'bg-white dark:bg-zinc-800 text-brand-600 shadow-sm'}`}
                    title="呼出指令集"
                >
                    <Zap className="w-4 h-4" />
                </button>
                <textarea
                    value={input}
                    onChange={handleInputChange}
                    onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend(); } }}
                    placeholder="键入指令，或输入 / 唤起技能..."
                    className="flex-1 bg-transparent resize-none p-2 text-sm outline-none min-h-[40px] max-h-[200px] text-ink-900 dark:text-zinc-100"
                    rows={1}
                />
                <button onClick={() => handleSend()} disabled={!input.trim() || isTyping || !activeSessionId} className="p-3 bg-zinc-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-xl hover:bg-black dark:hover:bg-zinc-100 disabled:opacity-30 transition-all shadow-lg">
                    <Send className="w-4 h-4" />
                </button>
                </div>
            </div>
          </>
        )}

        {activeTab === 'history' && (
          <div className="flex-1 overflow-y-auto p-4 space-y-2 bg-gray-50/30 dark:bg-zinc-950/20">
             <div className="flex justify-between items-center mb-4 px-2">
                <h4 className="text-[10px] font-black text-gray-400 uppercase tracking-widest">对话版本历史 ({sessions.length})</h4>
                <button onClick={createNewSession} className="text-[10px] font-black text-brand-600 dark:text-brand-400 flex items-center gap-1">
                    <Plus className="w-3 h-3" /> 新对话
                </button>
             </div>
             {sessions.map(s => (
               <div key={s.id} onClick={() => { setActiveSessionId(s.id); setActiveTab('chat'); }} className={`group flex items-center justify-between p-4 rounded-2xl border transition-all cursor-pointer ${activeSessionId === s.id ? 'bg-white dark:bg-zinc-900 border-brand-500/30 shadow-md ring-1 ring-brand-500/20' : 'bg-white dark:bg-zinc-900/50 border-gray-100 dark:border-zinc-800 hover:border-gray-200 dark:hover:border-zinc-700'}`}>
                 <div className="flex items-center gap-3 overflow-hidden">
                    <div className={`w-2 h-2 rounded-full shrink-0 ${activeSessionId === s.id ? 'bg-brand-500' : 'bg-gray-200 dark:bg-zinc-700'}`} />
                    <div className="flex flex-col truncate">
                        <span className={`text-xs font-bold truncate ${activeSessionId === s.id ? 'text-brand-600 dark:text-brand-400' : 'text-gray-700 dark:text-zinc-300'}`}>{s.title}</span>
                        <span className="text-[9px] text-gray-400 uppercase font-black">{new Date(s.lastUpdated).toLocaleDateString()} · {s.messages.length} 条消息</span>
                    </div>
                 </div>
                 <button onClick={(e) => deleteSession(s.id, e)} className="opacity-0 group-hover:opacity-100 p-2 text-gray-300 hover:text-red-500 transition-all"><Trash2 className="w-3.5 h-3.5" /></button>
               </div>
             ))}
          </div>
        )}

        {activeTab === 'skills' && (
          <div className="flex-1 overflow-y-auto p-6 space-y-8 bg-paper-50 dark:bg-zinc-950 transition-colors custom-scrollbar">
            {!isCreatingSkill ? (
              <>
                <header className="flex justify-between items-end">
                    <div className="space-y-1">
                        <h3 className="text-xl font-black text-gray-900 dark:text-zinc-100 font-serif italic flex items-center gap-2">
                           <ShoppingBag className="w-5 h-5 text-brand-500" /> 技能总库
                        </h3>
                        <p className="text-[10px] text-gray-400 font-black uppercase tracking-widest">预设馆 & 个人工作室</p>
                    </div>
                    <button 
                        onClick={() => setIsCreatingSkill(true)}
                        className="p-2 bg-brand-600 text-white rounded-xl shadow-lg shadow-brand-100 hover:bg-brand-700 transition-all"
                    >
                        <PlusCircle className="w-5 h-5" />
                    </button>
                </header>

                <section className="space-y-4">
                    <h4 className="text-[10px] font-black text-gray-400 uppercase tracking-[0.2em] px-2 flex items-center gap-2">
                        <Sparkles className="w-3 h-3" /> 缪斯推荐 (Presets)
                    </h4>
                    <div className="grid grid-cols-1 gap-3">
                        {PRESET_SKILLS.map(skill => (
                            <button
                                key={skill.id}
                                onClick={() => loadSkillToInput(skill)}
                                className="flex flex-col items-start text-left p-5 rounded-[1.5rem] bg-white dark:bg-zinc-900 border border-gray-100 dark:border-zinc-800 hover:border-brand-500/30 hover:shadow-xl transition-all group"
                            >
                                <div className="flex items-center justify-between w-full mb-2">
                                    <div className="flex items-center gap-2">
                                        <div className="p-2 bg-brand-50 dark:bg-brand-900/30 rounded-xl text-brand-600 dark:text-brand-400">
                                            <Zap className="w-4 h-4" />
                                        </div>
                                        <span className="text-xs font-black text-gray-900 dark:text-white">{skill.name}</span>
                                    </div>
                                    <span className="text-[8px] font-black text-gray-300 dark:text-zinc-600 uppercase tracking-widest">{skill.category}</span>
                                </div>
                                <p className="text-[10px] text-gray-500 dark:text-zinc-400 leading-relaxed line-clamp-2">{skill.description}</p>
                            </button>
                        ))}
                    </div>
                </section>

                <section className="space-y-4">
                    <h4 className="text-[10px] font-black text-amber-500 uppercase tracking-[0.2em] px-2 flex items-center gap-2">
                        <PenTool className="w-3 h-3" /> 创作实验室
                    </h4>
                    <div className="grid grid-cols-1 gap-3">
                        {WRITING_TOOLS.map(tool => (
                            <button
                                key={tool.id}
                                onClick={() => loadSkillToInput(tool)}
                                className="flex flex-col items-start text-left p-5 rounded-[1.5rem] bg-amber-50/10 dark:bg-amber-900/5 border border-amber-100 dark:border-amber-900/20 hover:border-amber-400 transition-all group"
                            >
                                <div className="flex items-center justify-between w-full mb-2">
                                    <div className="flex items-center gap-2">
                                        <div className="p-2 bg-amber-100 dark:bg-amber-900/30 rounded-xl text-amber-600 dark:text-amber-400">
                                            <Flame className="w-4 h-4" />
                                        </div>
                                        <span className="text-xs font-black text-gray-900 dark:text-white">{tool.name}</span>
                                    </div>
                                </div>
                                <p className="text-[10px] text-gray-500 dark:text-zinc-400 leading-relaxed line-clamp-2">{tool.description}</p>
                            </button>
                        ))}
                    </div>
                </section>

                <section className="space-y-4">
                    <h4 className="text-[10px] font-black text-indigo-500 uppercase tracking-[0.2em] px-2 flex items-center gap-2">
                        <Megaphone className="w-3 h-3" /> 宣发工具
                    </h4>
                    <div className="grid grid-cols-1 gap-3">
                        {PROMO_TOOLS.map(tool => (
                            <button
                                key={tool.id}
                                onClick={() => loadSkillToInput(tool)}
                                className="flex flex-col items-start text-left p-5 rounded-[1.5rem] bg-indigo-50/10 dark:bg-indigo-900/5 border border-indigo-100 dark:border-indigo-900/20 hover:border-indigo-400 transition-all group"
                            >
                                <div className="flex items-center justify-between w-full mb-2">
                                    <div className="flex items-center gap-2">
                                        <div className="p-2 bg-indigo-100 dark:bg-indigo-900/30 rounded-xl text-indigo-600 dark:text-indigo-400">
                                            <Target className="w-4 h-4" />
                                        </div>
                                        <span className="text-xs font-black text-gray-900 dark:text-white">{tool.name}</span>
                                    </div>
                                </div>
                                <p className="text-[10px] text-gray-500 dark:text-zinc-400 leading-relaxed line-clamp-2">{tool.description}</p>
                            </button>
                        ))}
                    </div>
                </section>

                {project && (project?.templates || []).length > 0 && (
                    <section className="space-y-4">
                        <h4 className="text-[10px] font-black text-emerald-400 uppercase tracking-[0.2em] px-2 flex items-center gap-2">
                            <Bookmark className="w-3 h-3" /> 我的创作组件 (Custom)
                        </h4>
                        <div className="grid grid-cols-1 gap-3">
                            {(project?.templates || []).map(skill => (
                                <div key={skill.id} className="group relative">
                                    <button
                                        onClick={() => loadSkillToInput(skill)}
                                        className="w-full flex flex-col items-start text-left p-5 rounded-[1.5rem] bg-emerald-50/10 dark:bg-emerald-900/5 border border-emerald-100 dark:border-emerald-900/20 hover:border-emerald-400 transition-all"
                                    >
                                        <div className="flex items-center gap-2 mb-2">
                                            <div className="p-2 bg-emerald-100 dark:bg-emerald-900/40 rounded-xl text-emerald-600 dark:text-emerald-400">
                                                <Bookmark className="w-4 h-4" />
                                            </div>
                                            <span className="text-xs font-black text-gray-900 dark:text-white">{skill.name}</span>
                                        </div>
                                        <p className="text-[10px] text-gray-500 dark:text-zinc-400 leading-relaxed line-clamp-2">{skill.description}</p>
                                    </button>
                                    <button 
                                        onClick={(e) => { e.stopPropagation(); if(confirm('删除此自定义技能？')) deleteTemplate(skill.id); }}
                                        className="absolute top-4 right-4 p-2 text-gray-300 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-all"
                                    >
                                        <Trash2 className="w-3.5 h-3.5" />
                                    </button>
                                </div>
                            ))}
                        </div>
                    </section>
                )}
              </>
            ) : (
              <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-300">
                <header className="flex items-center gap-3">
                   <button onClick={() => setIsCreatingSkill(false)} className="p-2 hover:bg-gray-100 dark:hover:bg-zinc-800 rounded-xl transition-all"><X className="w-5 h-5 text-gray-400" /></button>
                   <div>
                       <h3 className="text-lg font-black text-gray-900 dark:text-zinc-100">新建写作技能</h3>
                       <p className="text-[9px] text-gray-400 font-bold uppercase tracking-widest">定义你的专属缪斯指令</p>
                   </div>
                </header>

                <div className="space-y-4">
                    <div className="space-y-1">
                        <label className="text-[10px] font-black text-gray-400 uppercase tracking-widest px-1">技能名称</label>
                        <input 
                            value={newSkill.name} 
                            onChange={e => setNewSkill({...newSkill, name: e.target.value})}
                            className="w-full p-4 bg-white dark:bg-zinc-900 rounded-2xl border border-gray-100 dark:border-zinc-800 text-sm font-bold focus:ring-2 focus:ring-brand-100 dark:focus:ring-brand-900 outline-none"
                            placeholder="如：大师级对白优化"
                        />
                    </div>
                    <div className="space-y-1">
                        <label className="text-[10px] font-black text-gray-400 uppercase tracking-widest px-1">简短描述</label>
                        <input 
                            value={newSkill.description} 
                            onChange={e => setNewSkill({...newSkill, description: e.target.value})}
                            className="w-full p-4 bg-white dark:bg-zinc-900 rounded-2xl border border-gray-100 dark:border-zinc-800 text-sm focus:ring-2 focus:ring-brand-100 dark:focus:ring-brand-900 outline-none"
                            placeholder="描述该技能的作用..."
                        />
                    </div>
                    <div className="space-y-1">
                        <label className="text-[10px] font-black text-gray-400 uppercase tracking-widest px-1">提示词模板 (Template)</label>
                        <textarea 
                            value={newSkill.template} 
                            onChange={e => setNewSkill({...newSkill, template: e.target.value})}
                            className="w-full h-48 p-4 bg-white dark:bg-zinc-900 rounded-2xl border border-gray-100 dark:border-zinc-800 text-xs font-mono leading-relaxed focus:ring-2 focus:ring-brand-100 dark:focus:ring-brand-900 outline-none resize-none"
                            placeholder="请输入指令内容。可用变量：{{content}} (正文), {{title}} (标题)"
                        />
                        <div className="flex gap-2 mt-2">
                             {['{{content}}', '{{title}}', '{{volumeTitle}}'].map(tag => (
                                 <button key={tag} onClick={() => setNewSkill({...newSkill, template: newSkill.template + tag})} className="px-2 py-1 bg-paper-200 dark:bg-zinc-800 text-[9px] font-black rounded-lg text-ink-400 hover:text-brand-600 transition-colors border border-paper-300 dark:border-zinc-700">插入 {tag}</button>
                             ))}
                        </div>
                    </div>
                </div>

                <button 
                    onClick={handleCreateCustomSkill}
                    disabled={!newSkill.name || !newSkill.template}
                    className="w-full py-4 bg-brand-600 text-white rounded-2xl font-black text-xs uppercase tracking-widest shadow-xl shadow-brand-100 hover:bg-brand-700 disabled:opacity-50 transition-all flex items-center justify-center gap-2"
                >
                    <Save className="w-4 h-4" /> 保存到工作室
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
