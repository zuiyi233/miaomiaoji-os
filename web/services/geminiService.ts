
import { AISettings, EntityType, Project, Document } from "../types";
import { generateText, generateJSON, generateTextStream } from "./aiService";
import { Type } from "@google/genai";

// Helper to stringify entities for context
const formatEntities = (entities: any[]) => {
    if (!entities || entities.length === 0) return "无相关实体";
    return entities.map(e => `
- [${e.type}] ${e.title}: ${e.content}
  ${e.voiceStyle ? `(语调/风格: ${e.voiceStyle})` : ''}
`).join('\n');
};

export const generateStoryContent = async (
  prompt: string,
  project: Project,
  activeDoc: Document,
  prevDoc: Document | null,
  nextDoc: Document | null,
  wordCountRequest: string = "normal" // 'expand' | 'normal' | 'finish'
): Promise<string> => {
  try {
    const novelMeta = `
【小说元数据】
书名：${project.title}
类型：${project.genre || '未定义'}
标签：${(project.tags || []).join(', ')}
核心冲突：${project.coreConflict}
全书基调/世界观：${project.worldRules}
`;

    const worldContext = formatEntities(project.entities);

    const chapterContext = `
【当前章节信息】
标题：${activeDoc.title}
本章目标 (Goal)：${activeDoc.chapterGoal}
核心情节 (Core Plot)：${activeDoc.corePlot}
目标字数：${activeDoc.targetWordCount || 3000} 字
`;

    let continuity = "";
    if (prevDoc) {
        continuity += `\n【前文提要 (${prevDoc.title})】：...${prevDoc.content.slice(-1000)}`;
    } else {
        continuity += "\n【前文】：无（这是第一章）";
    }
    
    if (nextDoc) {
        continuity += `\n【后续预设 (${nextDoc.title})】：${nextDoc.summary || nextDoc.corePlot || '未设定详细大纲'}`;
    }

    let pacingInstruction = "";
    if (wordCountRequest === 'expand') {
        pacingInstruction = "【极其重要】：当前需要大幅度扩写。请放慢叙事节奏，增加大量的环境描写、心理描写、感官细节和肢体动作。不要急于推进剧情，专注于当前场景的沉浸感。严禁在此次生成中完结本章，除非情节已自然终结。";
    } else if (wordCountRequest === 'finish') {
        pacingInstruction = "【指令】：请根据核心情节，收束本章剧情，留下悬念（Hook），为下一章做铺垫。";
    } else {
        pacingInstruction = "【指令】：正常推进剧情。保持文学性，" + (activeDoc.content.length < (activeDoc.targetWordCount || 3000) / 2 ? "当前进度尚早，请从容铺垫，不要急于推向高潮。" : "推进冲突发展。");
    }

    const systemInstruction = `你是一位顶尖的${project.genre || '通俗'}小说作家。
    你的任务是撰写正文。请严格利用【世界观档案】中的设定，确保人物行为符合其性格（Voice Style）。
    
    必须遵守：
    1. Show, Don't Tell 原则。
    2. 严格遵循【当前章节信息】中的目标和情节安排，**严禁跑题**。
    3. 字数控制：请生成一段高质量正文，字数约 500-1000 字（基于单次输出限制），但要为长篇幅写作留出余地，不要像写大纲一样匆忙概括。
    4. 风格：${project.genre}风格，用词精准，氛围感强。
    `;

    const fullPrompt = `
${novelMeta}
${worldContext}
${chapterContext}
${continuity}
【即时上文】：
...${activeDoc.content.slice(-2000)}
${pacingInstruction}
用户额外指令：${prompt}
    `;

    return await generateText(fullPrompt, systemInstruction, project.aiSettings);
  } catch (error) {
    console.error("Story generation error:", error);
    return "灵感连接中断，请稍后重试。";
  }
};

export const generateChapterOutline = async (
  brief: string,
  worldContext: string,
  settings: AISettings
): Promise<{ title: string; content: string; timeNode: string } | null> => {
  try {
    const systemInstruction = `你是一位资深小说编辑和大纲策划专家。
    你的任务是根据用户提供的章节梗概，扩充为一个包含【起、承、转、合】的详细故事线。
    输出要求：
    - 文学性强的章节标题。
    - 内容需涵盖：环境氛围描写建议、关键对话切入点、冲突爆发点以及本章悬念结尾。
    - timeNode: 为本章节建议一个文学化的时间坐标。`;

    const prompt = `层级上下文与世界观：${worldContext}\n\n本章初步梗概：${brief}`;

    // Fix: Updated schema to use Type enum per Google GenAI guidelines
    const schema = {
      type: Type.OBJECT,
      properties: {
        title: { type: Type.STRING },
        content: { type: Type.STRING },
        timeNode: { type: Type.STRING }
      },
      required: ['title', 'content', 'timeNode']
    };

    return await generateJSON(prompt, systemInstruction, schema, settings);
  } catch (error) {
    console.error("Outline generation error:", error);
    return null;
  }
};

export const generateTimeSuggestions = async (
  chapterTitles: string[],
  worldContext: string,
  settings: AISettings
): Promise<{ suggestions: { title: string, timeNode: string }[] }> => {
  try {
    const prompt = `根据以下章节标题和世界观背景，为每个章节设计一个具有氛围感的时间节点。
    章节：${chapterTitles.join(', ')}
    世界观：${worldContext}
    请确保时间节点具有连续性或逻辑性。`;

    // Fix: Updated schema to use Type enum per Google GenAI guidelines
    const schema = {
      type: Type.OBJECT,
      properties: {
        suggestions: {
          type: Type.ARRAY,
          items: {
            type: Type.OBJECT,
            properties: {
              title: { type: Type.STRING },
              timeNode: { type: Type.STRING }
            },
            required: ['title', 'timeNode']
          }
        }
      },
      required: ['suggestions']
    };

    const result = await generateJSON(prompt, "你是一位擅长营造文学氛围的小说助手。", schema, settings);
    return result || { suggestions: [] };
  } catch (error) {
    console.error("Time suggestion error:", error);
    return { suggestions: [] };
  }
};

export async function* chatWithMuseStream(
  history: { role: string; parts: { text: string }[] }[],
  message: string,
  settings: AISettings
): AsyncGenerator<string> {
  const systemInstruction = "你是“灵感缪斯”，一个深刻、敏锐且充满创意的写作伙伴。你不仅能提供创意，还能从逻辑上分析故事的合理性。你会经常引用已建立的角色、地点和组织设定来辅助回答。输出支持标准 Markdown 格式。";
  
  try {
    const stream = generateTextStream(message, systemInstruction, settings, history);
    for await (const chunk of stream) {
      yield chunk;
    }
  } catch (error) {
    console.error("Chat error:", error);
    yield "灵感流转受阻...";
  }
}
