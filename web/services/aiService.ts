
import { GoogleGenAI, GenerateContentResponse, Type } from "@google/genai";
import { AIProvider, AISettings, EntityType, ModelInfo, Project } from "../types";

// Base interface for the response
export interface AIResponse {
  text: string;
  raw?: any;
}

const CACHE_KEY_PREFIX = 'novel_agent_models_';
const CACHE_DURATION = 300 * 60 * 60 * 1000; // 300 hours as per request

// New Default Configuration: Removed specific hardcoded proxy, defaults to Gemini or standard OpenAI
export const DEFAULT_AI_SETTINGS: AISettings = {
  provider: 'gemini',
  model: 'gemini-3-flash-preview',
  proxyEndpoint: '', // Clean default
  temperature: 0.9
};

// Hardcoded recommended models for Gemini to ensure stability and guideline compliance
const GEMINI_MODELS: ModelInfo[] = [
  { id: 'gemini-3-flash-preview', name: 'Gemini 3.0 Flash (Recommended)', provider: 'gemini' },
  { id: 'gemini-3-pro-preview', name: 'Gemini 3.0 Pro (High Reasoning)', provider: 'gemini' },
  { id: 'gemini-2.5-flash-latest', name: 'Gemini 2.5 Flash', provider: 'gemini' },
  { id: 'gemini-2.5-flash-thinking', name: 'Gemini 2.5 Flash Thinking', provider: 'gemini' },
];

export const clearModelCache = () => {
  const keys = Object.keys(localStorage);
  keys.forEach(key => {
    if (key.startsWith(CACHE_KEY_PREFIX)) {
      localStorage.removeItem(key);
    }
  });
};

export const fetchAvailableModels = async (settings: AISettings): Promise<ModelInfo[]> => {
  const cacheKey = `${CACHE_KEY_PREFIX}${settings.provider}`;
  
  // Try Cache (skip cache for local to ensure we get live status)
  if (settings.provider !== 'local') {
    try {
      const cached = localStorage.getItem(cacheKey);
      if (cached) {
        const { data, timestamp } = JSON.parse(cached);
        if (Date.now() - timestamp < CACHE_DURATION) {
          return data;
        }
      }
    } catch (e) {
      console.warn("Cache read failed", e);
    }
  }

  let models: ModelInfo[] = [];

  if (settings.provider === 'gemini') {
    models = GEMINI_MODELS;
  } else {
    // Fetch for OpenAI, Proxy, or Local
    let endpoint = '';
    
    if (settings.provider === 'local') {
      // For local, default to Ollama/standard local port if not specified
      const base = settings.proxyEndpoint || 'http://localhost:11434/v1';
      endpoint = base.endsWith('/') ? `${base}models` : `${base}/models`;
    } else if (settings.provider === 'openai') {
      endpoint = 'https://api.openai.com/v1/models';
    } else if (settings.proxyEndpoint) {
      if (settings.proxyEndpoint.includes('/chat/completions')) {
        endpoint = settings.proxyEndpoint.replace('/chat/completions', '/models');
      } else {
        endpoint = settings.proxyEndpoint.endsWith('/') 
          ? `${settings.proxyEndpoint}models` 
          : `${settings.proxyEndpoint}/models`;
      }
    }
    
    if (endpoint) {
      try {
        const headers: Record<string, string> = {
          'Content-Type': 'application/json'
        };
        // Add Authorization only if not local (unless needed) or if key exists
        if (settings.provider !== 'local' && process.env.API_KEY) {
           headers['Authorization'] = `Bearer ${process.env.API_KEY}`;
        } else if (process.env.API_KEY && settings.provider === 'local') {
           // Some local proxies might still want a dummy key
           headers['Authorization'] = `Bearer ${process.env.API_KEY}`;
        }

        const response = await fetch(endpoint, {
          method: 'GET',
          headers
        });
        
        if (response.ok) {
          const data = await response.json();
          models = (data.data || [])
            .map((m: any) => ({
              id: m.id,
              name: m.id,
              provider: settings.provider
            }))
            .sort((a: ModelInfo, b: ModelInfo) => a.id.localeCompare(b.id));
        }
      } catch (error) {
        console.error("Failed to fetch models", error);
        return []; 
      }
    }
  }

  if (models.length > 0 && settings.provider !== 'local') {
    localStorage.setItem(cacheKey, JSON.stringify({ data: models, timestamp: Date.now() }));
  }

  return models;
};

const callGemini = async (prompt: string, systemInstruction: string, settings: AISettings): Promise<string> => {
  const ai = new GoogleGenAI({ apiKey: process.env.API_KEY });
  // Pass configuration for maxOutputTokens and thinkingBudget to comply with Gemini API guidelines
  const response: GenerateContentResponse = await ai.models.generateContent({
    model: settings.model || 'gemini-3-flash-preview',
    contents: prompt,
    config: {
      systemInstruction: systemInstruction,
      temperature: settings.temperature,
      maxOutputTokens: settings.maxOutputTokens,
      thinkingConfig: settings.thinkingBudget !== undefined ? { thinkingBudget: settings.thinkingBudget } : undefined,
    }
  });
  return response.text || "";
};

const callOpenAICompatible = async (prompt: string, systemInstruction: string, settings: AISettings): Promise<string> => {
  let endpoint = '';
  if (settings.provider === 'local') {
     endpoint = settings.proxyEndpoint || 'http://localhost:11434/v1/chat/completions';
  } else if (settings.provider === 'openai') {
     endpoint = 'https://api.openai.com/v1/chat/completions';
  } else {
     endpoint = settings.proxyEndpoint || 'https://api.openai.com/v1/chat/completions';
  }
  
  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  };
  
  if (settings.provider !== 'local' && process.env.API_KEY) {
    headers['Authorization'] = `Bearer ${process.env.API_KEY}`;
  } else if (process.env.API_KEY) {
    // Optional for local, but good practice for proxies
    headers['Authorization'] = `Bearer ${process.env.API_KEY}`; 
  }

  try {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        model: settings.model,
        messages: [
          { role: "system", content: systemInstruction },
          { role: "user", content: prompt }
        ],
        temperature: settings.temperature
      })
    });
    const data = await response.json();
    return data.choices?.[0]?.message?.content || "";
  } catch (error) {
    console.error("OpenAI/Proxy/Local Error:", error);
    return "AI 接口调用失败，请检查配置或本地服务状态。";
  }
};

export const generateText = async (prompt: string, systemInstruction: string, settings: AISettings): Promise<string> => {
  if (settings.provider === 'gemini') {
    return callGemini(prompt, systemInstruction, settings);
  } else {
    return callOpenAICompatible(prompt, systemInstruction, settings);
  }
};

export const generateJSON = async (prompt: string, systemInstruction: string, schema: any, settings: AISettings): Promise<any> => {
  if (settings.provider === 'gemini') {
    const ai = new GoogleGenAI({ apiKey: process.env.API_KEY });
    // Pass configuration for maxOutputTokens and thinkingBudget to comply with Gemini API guidelines
    const response = await ai.models.generateContent({
      model: settings.model || 'gemini-3-flash-preview',
      contents: prompt,
      config: {
        systemInstruction: systemInstruction,
        responseMimeType: "application/json",
        responseSchema: schema,
        temperature: settings.temperature,
        maxOutputTokens: settings.maxOutputTokens,
        thinkingConfig: settings.thinkingBudget !== undefined ? { thinkingBudget: settings.thinkingBudget } : undefined,
      }
    });
    return JSON.parse(response.text || "{}");
  } else {
    const jsonPrompt = `${prompt}\n\n请严格按此 JSON 结构返回：${JSON.stringify(schema)}`;
    const text = await callOpenAICompatible(jsonPrompt, systemInstruction, settings);
    try {
      const match = text.match(/\{[\s\S]*\}/);
      const jsonStr = match ? match[0] : text;
      return JSON.parse(jsonStr.trim());
    } catch {
      return null;
    }
  }
};

export async function* generateTextStream(
  prompt: string, 
  systemInstruction: string, 
  settings: AISettings, 
  history: { role: string; parts: { text: string }[] }[] = []
): AsyncGenerator<string> {
  if (settings.provider === 'gemini') {
    const ai = new GoogleGenAI({ apiKey: process.env.API_KEY });
    const chat = ai.chats.create({
      model: settings.model || 'gemini-3-flash-preview',
      history: history,
      config: {
        systemInstruction: systemInstruction,
        temperature: settings.temperature,
        // Pass configuration for maxOutputTokens and thinkingBudget to comply with Gemini API guidelines
        maxOutputTokens: settings.maxOutputTokens,
        thinkingConfig: settings.thinkingBudget !== undefined ? { thinkingBudget: settings.thinkingBudget } : undefined,
      }
    });
    const result = await chat.sendMessageStream({ message: prompt });
    for await (const chunk of result) {
      yield (chunk as GenerateContentResponse).text || "";
    }
  } else {
    let endpoint = '';
    if (settings.provider === 'local') {
       endpoint = settings.proxyEndpoint || 'http://localhost:11434/v1/chat/completions';
    } else if (settings.provider === 'openai') {
       endpoint = 'https://api.openai.com/v1/chat/completions';
    } else {
       endpoint = settings.proxyEndpoint || 'https://api.openai.com/v1/chat/completions';
    }
    
    const messages = [
        { role: "system", content: systemInstruction },
        ...history.map(h => ({ role: h.role === 'model' ? 'assistant' : 'user', content: h.parts[0].text })),
        { role: "user", content: prompt }
    ];

    const headers: Record<string, string> = {
        'Content-Type': 'application/json'
    };
    
    if (settings.provider !== 'local' && process.env.API_KEY) {
        headers['Authorization'] = `Bearer ${process.env.API_KEY}`;
    } else if (process.env.API_KEY) {
        headers['Authorization'] = `Bearer ${process.env.API_KEY}`;
    }

    try {
      const response = await fetch(endpoint, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          model: settings.model,
          messages,
          temperature: settings.temperature,
          stream: true
        })
      });

      if (!response.ok) {
          yield "AI 接口连接失败。";
          return;
      }

      const reader = response.body?.getReader();
      if (!reader) return;

      const decoder = new TextDecoder();
      let buffer = '';
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          const trimmed = line.trim();
          if (!trimmed || trimmed === 'data: [DONE]') continue;
          if (trimmed.startsWith('data: ')) {
            try {
              const data = JSON.parse(trimmed.substring(6));
              const content = data.choices?.[0]?.delta?.content;
              if (content) yield content;
            } catch (e) {}
          }
        }
      }
    } catch (e) {
      console.error("Stream error:", e);
      yield "流式连接中断。";
    }
  }
}

export const fetchInternetInspiration = async (settings: AISettings = DEFAULT_AI_SETTINGS): Promise<{ quote: string, source: string, url?: string }> => {
  if (settings.provider === 'gemini') {
    try {
      const ai = new GoogleGenAI({ apiKey: process.env.API_KEY });
      const response = await ai.models.generateContent({
        model: 'gemini-3-flash-preview',
        contents: '请从互联网中精选一句富有深意、优美且适合激发创作灵感的名言或短句。请直接输出句子内容、作者及出处。要求真实可信，具有文学美感。请以 JSON 格式返回，包含 quote 和 source 两个字段。',
        config: {
          tools: [{ googleSearch: {} }],
          responseMimeType: "application/json",
          responseSchema: {
            type: Type.OBJECT,
            properties: {
              quote: { type: Type.STRING },
              source: { type: Type.STRING }
            },
            required: ["quote", "source"]
          }
        }
      });
      const data = JSON.parse(response.text || "{}");
      const url = response.candidates?.[0]?.groundingMetadata?.groundingChunks?.[0]?.web?.uri;
      return { 
        quote: data.quote || "笔尖下的每一个字，都是灵魂在纸上的呼吸。", 
        source: data.source || "佚名",
        url 
      };
    } catch (e) {
      console.warn("Gemini inspiration fetch failed", e);
    }
  }

  const prompt = '请精选一句富有深意的中文文学名句或创作灵感短句。请直接以 JSON 格式返回：{"quote": "内容", "source": "作者/出处"}';
  const systemInstruction = '你是一位博学的小说家助手。';
  const schema = {
    type: Type.OBJECT,
    properties: {
      quote: { type: Type.STRING },
      source: { type: Type.STRING }
    },
    required: ["quote", "source"]
  };

  try {
    const data = await generateJSON(prompt, systemInstruction, schema, settings);
    return { 
      quote: data?.quote || "笔尖下的每一个字，都是灵魂在纸上的呼吸。", 
      source: data?.source || "佚名"
    };
  } catch (e) {
    return { 
      quote: "每一个不曾起舞的日子，中的对生命的辜负。", 
      source: "尼采" 
    };
  }
};

export const refineNovelCore = async (project: Project): Promise<Partial<Project>> => {
  const settings = project.aiSettings;
  const systemInstruction = `你是一位世界顶尖的小说架构师和文学顾问。你的任务是分析现有的小说草案，并深度完善其核心世界观设定。`;
  const context = `书名：${project.title}\n类型：${project.genre || '未定义'}\n核心冲突：${project.coreConflict || '无'}\n世界规则：${project.worldRules || '无'}`;
  const prompt = `分析以上信息，完善以下核心要素：\n${context}`;

  const schema = {
    type: Type.OBJECT,
    properties: {
      genre: { type: Type.STRING },
      tags: { type: Type.ARRAY, items: { type: Type.STRING } },
      coreConflict: { type: Type.STRING },
      characterArc: { type: Type.STRING },
      ultimateValue: { type: Type.STRING },
      worldRules: { type: Type.STRING },
      characterCore: { type: Type.STRING },
      symbolSettings: { type: Type.STRING }
    },
    required: ["coreConflict", "characterArc", "worldRules", "characterCore"]
  };

  return await generateJSON(prompt, systemInstruction, schema, settings);
};

export const generateNovelBlueprint = async (userInput: string, settings: AISettings): Promise<any> => {
    const systemInstruction = `你是一位世界级的小说架构师。你的任务是根据用户的简单想法，扩展为一个完整的小说世界观蓝图。`;
    const prompt = `用户创意核心：${userInput}`;

    const schema = {
        type: Type.OBJECT,
        properties: {
            title: { type: Type.STRING },
            coreConflict: { type: Type.STRING },
            characterArc: { type: Type.STRING },
            ultimateValue: { type: Type.STRING },
            worldRules: { type: Type.STRING },
            characterCore: { type: Type.STRING },
            symbolSettings: { type: Type.STRING },
            firstVolumeTitle: { type: Type.STRING },
            firstVolumeGoal: { type: Type.STRING },
            firstChapterTitle: { type: Type.STRING },
            firstChapterContent: { type: Type.STRING },
            protagonistName: { type: Type.STRING },
            protagonistDesc: { type: Type.STRING }
        },
        required: ["title", "coreConflict", "firstVolumeTitle", "firstChapterTitle", "protagonistName"]
    };

    return await generateJSON(prompt, systemInstruction, schema, settings);
};

export const generateVolumeOutline = async (
  volumeContext: string,
  structureTemplate: string,
  settings: AISettings
): Promise<any[]> => {
  const systemInstruction = `你是一位精通故事结构的小说策划。你的任务是根据给定的卷信息和情节结构模板，生成一系列章节大纲。`;
  const prompt = `卷背景信息：\n${volumeContext}\n\n目标结构：${structureTemplate}`;

  const schema = {
    type: Type.ARRAY,
    items: {
      type: Type.OBJECT,
      properties: {
        title: { type: Type.STRING },
        chapterGoal: { type: Type.STRING },
        corePlot: { type: Type.STRING },
        hook: { type: Type.STRING }
      },
      required: ["title", "chapterGoal", "corePlot"]
    }
  };

  const result = await generateJSON(prompt, systemInstruction, schema, settings);
  return Array.isArray(result) ? result : [];
};