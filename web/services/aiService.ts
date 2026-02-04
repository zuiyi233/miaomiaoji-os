
import { Type } from "@google/genai";
import inspirationRaw from '../assets/inspiration.md?raw';
import { AISettings, EntityType, ModelInfo, Project } from "../types";
import { clearModelCache as clearModelCacheDB, getModelCache, setModelCache } from "./db";
import { apiRequest, getApiBaseUrl } from "./apiClient";

// Base interface for the response
export interface AIResponse {
  text: string;
  raw?: any;
}

export type WorkflowPayload = {
  provider: string;
  path: string;
  body: string;
};

const CACHE_KEY_PREFIX = 'novel_agent_models_';
const CACHE_DURATION = 300 * 60 * 60 * 1000; // 300 hours as per request

// New Default Configuration: Removed specific hardcoded proxy, defaults to Gemini or standard OpenAI
export const DEFAULT_AI_SETTINGS: AISettings = {
  provider: 'gemini',
  model: 'gemini-3-flash-preview',
  proxyEndpoint: '', // Clean default
  temperature: 0.9
};

const getProviderConfigFlag = (provider: string): string | null => {
	try {
		return localStorage.getItem(`nao_ai_provider_configured_${provider}`);
	} catch {
		return null;
	}
};

const setProviderConfigFlag = (provider: string, value: 'true' | 'false') => {
	try {
		localStorage.setItem(`nao_ai_provider_configured_${provider}`, value);
	} catch {
		return;
	}
};

const isProviderConfigured = (settings: AISettings): boolean => {
	if (!settings.model) return false;
	if (settings.provider === 'gemini') {
		const flag = getProviderConfigFlag('gemini');
		return flag !== 'false';
	}
	if (settings.provider === 'local' || settings.provider === 'proxy' || settings.provider === 'openai' || settings.provider === 'openrouter' || settings.provider === 'anthropic') {
		return Boolean(settings.proxyEndpoint && settings.proxyEndpoint.trim());
	}
	return true;
};

// Hardcoded recommended models for Gemini to ensure stability and guideline compliance
const GEMINI_MODELS: ModelInfo[] = [
  { id: 'gemini-3-flash-preview', name: 'Gemini 3.0 Flash (Recommended)', provider: 'gemini' },
  { id: 'gemini-3-pro-preview', name: 'Gemini 3.0 Pro (High Reasoning)', provider: 'gemini' },
  { id: 'gemini-2.5-flash-latest', name: 'Gemini 2.5 Flash', provider: 'gemini' },
  { id: 'gemini-2.5-flash-thinking', name: 'Gemini 2.5 Flash Thinking', provider: 'gemini' },
];

export const clearModelCache = async (provider?: string) => {
	if (provider) {
		await clearModelCacheDB(`${CACHE_KEY_PREFIX}${provider}`);
		return;
	}
	await clearModelCacheDB();
};

export const fetchAvailableModels = async (settings: AISettings): Promise<ModelInfo[]> => {
	const cacheKey = `${CACHE_KEY_PREFIX}${settings.provider}`;

	if (!isProviderConfigured(settings)) {
		return [];
	}

	if (settings.provider !== 'local') {
		try {
			const cached = await getModelCache(cacheKey);
			if (cached && Date.now() - cached.timestamp < CACHE_DURATION) {
				return cached.data as ModelInfo[];
			}
		} catch (e) {
			console.warn("Cache read failed", e);
		}
	}

	let models: ModelInfo[] = [];
	if (settings.provider === 'gemini') {
		models = GEMINI_MODELS;
	} else {
		try {
			const data = await apiRequest<{ models: ModelInfo[] }>(`/api/v1/ai/models?provider=${settings.provider}`, {
				method: 'GET'
			});
			models = (data.models || []).map((m) => ({
				id: m.id,
				name: m.name || m.id,
				provider: settings.provider
			}));
		} catch (error) {
			console.error("Failed to fetch models", error);
			return [];
		}
	}

	if (models.length > 0 && settings.provider !== 'local') {
		await setModelCache({ provider: cacheKey, data: models, timestamp: Date.now() });
	}

	return models;
};

const callGemini = async (prompt: string, systemInstruction: string, settings: AISettings): Promise<string> => {
	if (!isProviderConfigured(settings)) {
		return "";
	}
	const modelId = settings.model || 'gemini-3-flash-preview';
	const body = JSON.stringify({
		contents: [{ role: "user", parts: [{ text: prompt }] }],
		generationConfig: {
			temperature: settings.temperature,
			maxOutputTokens: settings.maxOutputTokens,
		},
		systemInstruction: systemInstruction ? { parts: [{ text: systemInstruction }] } : undefined
	});
	try {
		const data = await apiRequest<any>(`/api/v1/ai/proxy`, {
			method: 'POST',
			body: JSON.stringify({ provider: 'gemini', path: `v1beta/models/${modelId}:generateContent`, body })
		});
		return data?.candidates?.[0]?.content?.parts?.[0]?.text || "";
	} catch (error) {
		if ((error as { code?: number })?.code === 10008) {
			setProviderConfigFlag('gemini', 'false');
		}
		return "";
	}
};

const callOpenAICompatible = async (prompt: string, systemInstruction: string, settings: AISettings): Promise<string> => {
	if (!isProviderConfigured(settings)) {
		return "AI 供应商未配置，请先在设置中填写接口地址与模型。";
	}
	const body = JSON.stringify({
		model: settings.model,
		messages: [
			{ role: "system", content: systemInstruction },
			{ role: "user", content: prompt }
		],
		temperature: settings.temperature
	});

	try {
		const data = await apiRequest<any>(`/api/v1/ai/proxy`, {
			method: 'POST',
			body: JSON.stringify({
				provider: settings.provider,
				path: settings.provider === 'local' ? 'v1/chat/completions' : 'v1/chat/completions',
				body
			})
		});
		return data.choices?.[0]?.message?.content || "";
	} catch (error) {
		if (isProviderConfigured(settings)) {
			console.error("OpenAI/Proxy/Local Error:", error);
		}
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
		if (!isProviderConfigured(settings)) return null;
		const modelId = settings.model || 'gemini-3-flash-preview';
		const body = JSON.stringify({
			contents: [{ role: "user", parts: [{ text: prompt }] }],
			generationConfig: {
				temperature: settings.temperature,
				maxOutputTokens: settings.maxOutputTokens,
				responseMimeType: "application/json",
				responseSchema: schema,
			},
			systemInstruction: systemInstruction ? { parts: [{ text: systemInstruction }] } : undefined
		});
		try {
			const data = await apiRequest<any>(`/api/v1/ai/proxy`, {
				method: 'POST',
				body: JSON.stringify({ provider: 'gemini', path: `v1beta/models/${modelId}:generateContent`, body })
			});
			const text = data?.candidates?.[0]?.content?.parts?.[0]?.text || "";
			return JSON.parse(text || "{}");
		} catch (error) {
			if ((error as { code?: number })?.code === 10008) {
				setProviderConfigFlag('gemini', 'false');
			}
			return null;
		}
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

export const buildWorkflowPayload = (
  prompt: string,
  systemInstruction: string,
  settings: AISettings,
  schema?: any
): WorkflowPayload => {
  if (settings.provider === 'gemini') {
    const modelId = settings.model || 'gemini-3-flash-preview';
    const body = JSON.stringify({
      contents: [{ role: 'user', parts: [{ text: prompt }] }],
      generationConfig: {
        temperature: settings.temperature,
        maxOutputTokens: settings.maxOutputTokens,
        ...(schema
          ? { responseMimeType: 'application/json', responseSchema: schema }
          : {}),
      },
      systemInstruction: systemInstruction ? { parts: [{ text: systemInstruction }] } : undefined,
    });
    return {
      provider: 'gemini',
      path: `v1beta/models/${modelId}:generateContent`,
      body,
    };
  }

  const jsonPrompt = schema ? `${prompt}\n\n请严格按此 JSON 结构返回：${JSON.stringify(schema)}` : prompt;
  const body = JSON.stringify({
    model: settings.model,
    messages: [
      { role: 'system', content: systemInstruction },
      { role: 'user', content: jsonPrompt },
    ],
    temperature: settings.temperature,
  });

  return {
    provider: settings.provider,
    path: 'v1/chat/completions',
    body,
  };
};

export async function* generateTextStream(
  prompt: string, 
  systemInstruction: string, 
  settings: AISettings, 
  history: { role: string; parts: { text: string }[] }[] = []
): AsyncGenerator<string> {
	if (settings.provider === 'gemini') {
		const modelId = settings.model || 'gemini-3-flash-preview';
		const body = JSON.stringify({
			contents: [{ role: "user", parts: [{ text: prompt }] }],
			generationConfig: {
				temperature: settings.temperature,
				maxOutputTokens: settings.maxOutputTokens,
			},
			systemInstruction: systemInstruction ? { parts: [{ text: systemInstruction }] } : undefined
		});

		const baseUrl = getApiBaseUrl();
		const streamUrl = baseUrl ? `${baseUrl}/api/v1/ai/proxy/stream` : '/api/v1/ai/proxy/stream';
		const response = await fetch(streamUrl, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ provider: 'gemini', path: `v1beta/models/${modelId}:streamGenerateContent?alt=sse`, body })
		});

		if (!response.ok || !response.body) {
			yield "流式连接失败。";
			return;
		}

		const reader = response.body.getReader();
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
						const chunk = data?.candidates?.[0]?.content?.parts?.[0]?.text || '';
						if (chunk) yield chunk;
					} catch {}
				}
			}
		}
	} else {
    if (!isProviderConfigured(settings)) {
      yield "AI 供应商未配置，请先在设置中填写接口地址与模型。";
      return;
    }
    const messages = [
        { role: "system", content: systemInstruction },
        ...history.map(h => ({ role: h.role === 'model' ? 'assistant' : 'user', content: h.parts[0].text })),
        { role: "user", content: prompt }
    ];

    const body = JSON.stringify({
      model: settings.model,
      messages,
      temperature: settings.temperature,
      stream: true
    });

    try {
		const baseUrl = getApiBaseUrl();
		const streamUrl = baseUrl ? `${baseUrl}/api/v1/ai/proxy/stream` : '/api/v1/ai/proxy/stream';
		const response = await fetch(streamUrl, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ provider: settings.provider, path: 'v1/chat/completions', body })
		});

      if (!response.ok || !response.body) {
          yield "AI 接口连接失败。";
          return;
      }

      const reader = response.body.getReader();
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

type InspirationEntry = { quote: string; source: string };

const parseInspiration = (): InspirationEntry[] => {
  return inspirationRaw
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line.startsWith('- '))
    .map((line) => line.replace(/^-\s*/, ''))
    .map((line) => {
      const [quotePart, sourcePart] = line.split('—').map((part) => part.trim());
      return {
        quote: quotePart || '',
        source: sourcePart || '未知',
      };
    })
    .filter((entry) => entry.quote);
};

const INSPIRATION_POOL = parseInspiration();

export const fetchInternetInspiration = async (): Promise<{ quote: string, source: string, url?: string }> => {
  if (INSPIRATION_POOL.length === 0) {
    return { quote: "笔尖下的每一个字，都是灵魂在纸上的呼吸。", source: "佚名" };
  }
  const index = Math.floor(Math.random() * INSPIRATION_POOL.length);
  const selected = INSPIRATION_POOL[index];
  return { quote: selected.quote, source: selected.source };
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
