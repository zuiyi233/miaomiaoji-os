import React, { useState } from 'react';
import { X, Book, Feather, Globe, Layout, Sparkles, Link, Zap, HelpCircle, ArrowRight, Check } from 'lucide-react';

interface UserGuideProps {
  onClose: () => void;
}

export const UserGuide: React.FC<UserGuideProps> = ({ onClose }) => {
  const [view, setView] = useState<'guide' | 'onboarding'>('guide');
  const [step, setStep] = useState(1);

  const QuickStart = () => (
    <div className="flex-1 flex flex-col items-center justify-center p-12 space-y-8 text-center animate-in fade-in slide-in-from-right duration-500">
       <div className="w-16 h-16 bg-brand-100 text-brand-600 rounded-full flex items-center justify-center mb-4">
          {step === 1 && <Book className="w-8 h-8" />}
          {step === 2 && <Feather className="w-8 h-8" />}
          {step === 3 && <Globe className="w-8 h-8" />}
       </div>
       
       <div className="max-w-md space-y-4">
         <h3 className="text-2xl font-black text-gray-900">
            {step === 1 && "第一步：创世"}
            {step === 2 && "第二步：编织"}
            {step === 3 && "第三步：扩展"}
         </h3>
         <p className="text-gray-500 font-medium leading-relaxed">
            {step === 1 && "点击主控台的“孵化新宇宙”，输入你的灵感碎片（如“蒸汽朋克背景下的侦探故事”）。AI 将为你生成完整的小说蓝图，包括核心冲突、主角人设和第一卷大纲。"}
            {step === 2 && "进入“写作手稿”，在编辑器中尝试输入一段文字，然后点击右下角的 AI 助手，选择“细节扩写”。观察 AI 如何根据世界观自动填充环境描写。"}
            {step === 3 && "打开“世界观档案”，创建几个关键角色或地点。回到编辑器，你会发现 AI 在续写时会自动引用这些新设定，保持逻辑连贯。"}
         </p>
       </div>

       <div className="flex gap-4">
         {step > 1 && (
            <button onClick={() => setStep(step - 1)} className="px-6 py-3 rounded-xl text-gray-400 hover:text-gray-600 font-bold text-sm">上一步</button>
         )}
         {step < 3 ? (
            <button onClick={() => setStep(step + 1)} className="px-8 py-3 bg-gray-900 text-white rounded-xl font-bold text-sm hover:bg-black flex items-center gap-2">
                下一步 <ArrowRight className="w-4 h-4" />
            </button>
         ) : (
            <button onClick={onClose} className="px-8 py-3 bg-brand-600 text-white rounded-xl font-bold text-sm hover:bg-brand-700 flex items-center gap-2 shadow-lg shadow-brand-200">
                开始创作 <Check className="w-4 h-4" />
            </button>
         )}
       </div>
    </div>
  );

  return (
    <div className="fixed inset-0 z-[200] bg-gray-900/40 backdrop-blur-md flex items-center justify-center p-6 animate-in fade-in duration-300">
      <div className="bg-white rounded-[3rem] shadow-2xl w-full max-w-4xl h-[85vh] overflow-hidden flex flex-col">
        <div className="p-8 border-b border-gray-100 flex items-center justify-between bg-gray-50/50">
           <div className="flex items-center gap-4">
              <div className="bg-brand-500 p-3 rounded-2xl shadow-lg shadow-brand-100">
                <HelpCircle className="w-6 h-6 text-white" />
              </div>
              <div>
                <h2 className="text-2xl font-black text-gray-900 tracking-tight">操作手册 & 设计哲学</h2>
                <div className="flex gap-4 mt-1">
                   <button 
                     onClick={() => setView('guide')}
                     className={`text-xs font-black uppercase tracking-[0.2em] transition-colors ${view === 'guide' ? 'text-brand-600' : 'text-gray-400 hover:text-gray-600'}`}
                   >
                     核心理念
                   </button>
                   <button 
                     onClick={() => setView('onboarding')}
                     className={`text-xs font-black uppercase tracking-[0.2em] transition-colors ${view === 'onboarding' ? 'text-brand-600' : 'text-gray-400 hover:text-gray-600'}`}
                   >
                     快速开始
                   </button>
                </div>
              </div>
           </div>
           <button onClick={onClose} className="p-3 hover:bg-white rounded-2xl text-gray-400 transition-all border border-transparent hover:border-gray-100">
             <X className="w-6 h-6" />
           </button>
        </div>

        {view === 'onboarding' ? (
            <QuickStart />
        ) : (
            <>
                <div className="flex-1 overflow-y-auto p-12 space-y-12">
                <section className="space-y-6">
                    <h3 className="text-xl font-black text-gray-900 flex items-center gap-3">
                    <Zap className="w-5 h-5 text-amber-400" /> 核心理念
                    </h3>
                    <p className="text-gray-500 leading-relaxed font-medium">
                    Novel Agent OS 不仅仅是一个写作工具，它是一个“基于图谱的叙事引擎”。我们通过将传统的<b>线性文档</b>与<b>模块化世界观</b>相链接，利用 AI 确保创作过程中的逻辑一致性与灵感爆发。
                    </p>
                </section>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                    <div className="p-8 bg-brand-50 rounded-[2.5rem] border border-brand-100 space-y-4">
                    <Feather className="w-8 h-8 text-brand-600" />
                    <h4 className="font-black text-gray-900 text-lg">智能写作模式</h4>
                    <p className="text-sm text-gray-600 leading-relaxed">
                        在“写作手稿”中，你可以通过 <b>AI 深度续写</b> 来延续情节。AI 会自动抓取你关联的世界观实体（角色/地点），确保续写内容不偏离设定。
                    </p>
                    </div>
                    <div className="p-8 bg-indigo-50 rounded-[2.5rem] border border-indigo-100 space-y-4">
                    <Globe className="w-8 h-8 text-indigo-600" />
                    <h4 className="font-black text-gray-900 text-lg">世界观神经图谱</h4>
                    <p className="text-sm text-gray-600 leading-relaxed">
                        在“世界观/人设”中，你可以构建实体间的链接。使用 <b>AI 辅助生成</b> 快速填充设定，并利用 <b>深度连锁影响分析</b> 来检查设定变更对全局的影响。
                    </p>
                    </div>
                </div>

                <section className="space-y-6">
                    <h3 className="text-xl font-black text-gray-900 flex items-center gap-3">
                    <Link className="w-5 h-5 text-brand-500" /> 快速上手技巧
                    </h3>
                    <ul className="space-y-4">
                    {[
                        { title: '建立链接', desc: '在世界观页面，点击“插入”即可将两个实体关联。AI 会根据这些关联来理解你的故事宇宙。' },
                        { title: '批量管理', desc: '利用“批量管理”功能，一次性为多个角色添加同一个背景设定或删除过时的内容。' },
                        { title: '可视化探索', desc: '点击“可视化”按钮，在 2D 物理引擎驱动的图谱中观察你的故事脉络。' },
                        { title: '咒语模板', desc: '在 AI 助手侧边栏，你可以自定义提示词模板（如“逻辑自查”），一键处理当前章节。' }
                    ].map((item, i) => (
                        <li key={i} className="flex gap-4 p-4 hover:bg-gray-50 rounded-2xl transition-colors group">
                        <div className="w-8 h-8 bg-white border border-gray-100 rounded-xl flex items-center justify-center text-xs font-black text-brand-500 group-hover:scale-110 transition-transform">
                            {i + 1}
                        </div>
                        <div>
                            <h5 className="font-bold text-gray-900">{item.title}</h5>
                            <p className="text-sm text-gray-400 mt-1">{item.desc}</p>
                        </div>
                        </li>
                    ))}
                    </ul>
                </section>
                </div>

                <div className="p-8 border-t border-gray-100 bg-gray-50/30 flex justify-center">
                <button 
                    onClick={() => setView('onboarding')}
                    className="px-12 py-4 bg-gray-900 text-white rounded-2xl font-black text-sm hover:bg-gray-800 transition-all shadow-xl shadow-gray-200"
                >
                    我准备好了，开始创作
                </button>
                </div>
            </>
        )}
      </div>
    </div>
  );
};