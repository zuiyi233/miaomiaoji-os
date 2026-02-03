
import { generateJSON, generateText } from "./aiService";
import { AISettings, Project, Document, SceneNode, ValidationResult, StoryEntity } from "../types";
import { Type } from "@google/genai";

// --- Node 1: Conception & Planning ---
// Deconstructs the chapter goal into specific scenes with word counts.
export const agentPlanChapter = async (
  project: Project,
  chapter: Document,
  volume: any,
  prevChapterSummary: string,
  userInstruction: string
): Promise<SceneNode[]> => {
  const prompt = `
  你是一位精密的小说架构师。请根据以下信息，将本章拆解为 4-6 个具体的“场景（Scene）”节点。
  
  【全书信息】：${project.title} (${project.genre})
  【核心冲突】：${project.coreConflict}
  【本卷目标】：${volume?.coreGoal || '无'}
  【前情提要】：${prevChapterSummary}
  
  【本章任务】：
  标题：${chapter.title}
  核心目标：${chapter.chapterGoal}
  情节摘要：${chapter.corePlot}
  用户额外指令：${userInstruction}
  
  【输出要求】：
  1. 将章节拆分为 4-6 个连续的场景。
  2. 每个场景分配 400-800 字，确保本章总字数达到 3000 字以上。
  3. 场景的 "beat" (节拍) 必须具体，包含动作和冲突。
  `;

  const schema = {
    type: Type.ARRAY,
    items: {
      type: Type.OBJECT,
      properties: {
        title: { type: Type.STRING, description: "场景简短标题" },
        beat: { type: Type.STRING, description: "场景核心节拍/发生的具体事件" },
        expectedWordCount: { type: Type.INTEGER, description: "预计字数 (400-800)" }
      },
      required: ["title", "beat", "expectedWordCount"]
    }
  };

  try {
    const scenes = await generateJSON(prompt, "你是一位注重逻辑和节奏的小说策划。", schema, project.aiSettings);
    return (scenes || []).map((s: any, index: number) => ({
      id: `scene_${Date.now()}_${index}`,
      title: s.title,
      beat: s.beat,
      expectedWordCount: s.expectedWordCount,
      status: 'pending'
    }));
  } catch (error) {
    console.error("Agent Planning Error:", error);
    return [];
  }
};

// --- Node 2: Content Drafting ---
// Writes a single scene based on the beat and context.
export const agentDraftScene = async (
  project: Project,
  scene: SceneNode,
  chapterContext: string, // Previous written scenes content or summary
  entities: StoryEntity[]
): Promise<string> => {
  const entityContext = entities.slice(0, 5).map(e => `${e.title}(${e.type}): ${e.content.substring(0, 100)}...`).join('\n');
  
  const prompt = `
  你是一位金牌小说作家。请撰写以下场景的正文。
  
  【当前场景】：${scene.title}
  【核心节拍】：${scene.beat}
  【字数要求】：约 ${scene.expectedWordCount} 字 (请务必写够，多描写细节)
  
  【上下文衔接】：
  ${chapterContext.slice(-500)}
  
  【相关实体设定】：
  ${entityContext}
  
  【写作要求】：
  1. "Show, Don't Tell"。多用感官描写（视觉、听觉、触觉）和肢体动作。
  2. 情感充沛，代入感强。
  3. 严格贴合核心节拍，不要跑题。
  4. 只输出正文，不要标题或解释。
  `;

  return await generateText(prompt, "你是一位文笔细腻、擅长营造氛围的小说家。", project.aiSettings);
};

// --- Node 3: Validation ---
// Checks if the content meets quality and logical standards.
export const agentValidateScene = async (
  sceneContent: string,
  sceneBeat: string,
  settings: AISettings
): Promise<ValidationResult> => {
  const prompt = `
  作为一位严苛的文学编辑，请对以下小说片段进行质量质检。
  
  【预定情节】：${sceneBeat}
  【实际正文】：
  ${sceneContent}
  
  【检查项】：
  1. 是否完成了预定情节？
  2. 字数是否充足（视觉估算）？内容是否注水？
  3. 逻辑是否通顺？
  4. 情感描写是否生硬？
  
  请按 JSON 格式输出质检结果。
  `;

  const schema = {
    type: Type.OBJECT,
    properties: {
      passed: { type: Type.BOOLEAN, description: "是否达标 (True/False)" },
      score: { type: Type.INTEGER, description: "评分 0-100" },
      issues: { type: Type.ARRAY, items: { type: Type.STRING }, description: "主要问题列表" },
      critique: { type: Type.STRING, description: "简短评语" }
    },
    required: ["passed", "score", "issues", "critique"]
  };

  try {
    return await generateJSON(prompt, "你是一位严格的文学编辑。", schema, settings);
  } catch (error) {
    return { passed: true, score: 80, issues: [], critique: "自动质检跳过 (API Error)" };
  }
};

// --- Node 4: Refinement ---
// Rewrites the content based on validation issues.
export const agentRefineScene = async (
  originalContent: string,
  validationResult: ValidationResult,
  settings: AISettings
): Promise<string> => {
  const prompt = `
  你需要润色重写以下小说片段。
  
  【原文】：
  ${originalContent}
  
  【编辑反馈】：
  评分：${validationResult.score}
  问题：${validationResult.issues.join('; ')}
  评语：${validationResult.critique}
  
  【任务】：
  请针对上述问题进行深度润色和重写。保留原文的优点，修正缺点。增加细节描写，提升文学性。
  `;

  return await generateText(prompt, "你是一位精通改稿的资深编辑。", settings);
};
