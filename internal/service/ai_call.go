package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"novel-agent-os-backend/pkg/logger"
)

// callAI 统一的 AI 调用封装
func callAI(aiConfigService AIConfigService, provider, path, body string) (json.RawMessage, string, error) {
	if path == "" || strings.Contains(path, "..") {
		return nil, "", fmt.Errorf("invalid path")
	}

	providerCfg, err := aiConfigService.GetProviderConfigRaw(provider)
	if err != nil {
		return nil, "", fmt.Errorf("provider not found")
	}
	if err := ValidateAIProxyTarget(provider, providerCfg.BaseURL, path); err != nil {
		return nil, "", err
	}

	base := strings.TrimRight(providerCfg.BaseURL, "/")
	url := base + "/" + strings.TrimLeft(path, "/")

	client := &http.Client{Timeout: 60 * time.Second}
	proxyReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		return nil, "", fmt.Errorf("proxy request failed")
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("Accept", "application/json")
	if providerCfg.APIKey != "" {
		if provider == "gemini" {
			proxyReq.Header.Set("x-goog-api-key", providerCfg.APIKey)
		} else {
			proxyReq.Header.Set("Authorization", "Bearer "+providerCfg.APIKey)
		}
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		logger.Error("workflow proxy request failed", logger.Err(err))
		return nil, "", fmt.Errorf("proxy request failed")
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read response failed")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("upstream error: %s", string(raw))
	}

	content := extractAIText(raw)
	return json.RawMessage(raw), content, nil
}

// extractAIText 尝试从主流响应中提取文本
func extractAIText(raw []byte) string {
	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}

	// Gemini: candidates[0].content.parts[0].text
	if candidates, ok := payload["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if cand, ok := candidates[0].(map[string]interface{}); ok {
			if content, ok := cand["content"].(map[string]interface{}); ok {
				if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
					if part, ok := parts[0].(map[string]interface{}); ok {
						if text, ok := part["text"].(string); ok {
							return text
						}
					}
				}
			}
		}
	}

	// OpenAI: choices[0].message.content
	if choices, ok := payload["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if text, ok := message["content"].(string); ok {
					return text
				}
			}
		}
	}

	return ""
}

// CallAIStream 流式调用 AI 接口
func CallAIStream(ctx context.Context, aiConfigService AIConfigService, provider, path, body string, chunkHandler func(chunk string) error) error {
	if path == "" || strings.Contains(path, "..") {
		return fmt.Errorf("invalid path")
	}

	providerCfg, err := aiConfigService.GetProviderConfigRaw(provider)
	if err != nil {
		return fmt.Errorf("provider not found")
	}
	if err := ValidateAIProxyTarget(provider, providerCfg.BaseURL, path); err != nil {
		return err
	}

	base := strings.TrimRight(providerCfg.BaseURL, "/")
	url := base + "/" + strings.TrimLeft(path, "/")

	client := &http.Client{Timeout: 0}
	proxyReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("proxy request failed")
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("Accept", "text/event-stream")
	if providerCfg.APIKey != "" {
		if provider == "gemini" {
			proxyReq.Header.Set("x-goog-api-key", providerCfg.APIKey)
		} else {
			proxyReq.Header.Set("Authorization", "Bearer "+providerCfg.APIKey)
		}
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		logger.Error("stream request failed", logger.Err(err))
		return fmt.Errorf("stream request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upstream error: %s", string(raw))
	}

	return parseOpenAISSE(ctx, resp.Body, chunkHandler)
}

// parseOpenAISSE 解析 OpenAI SSE 格式
func parseOpenAISSE(ctx context.Context, reader io.Reader, chunkHandler func(chunk string) error) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return nil
		}

		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			logger.Warn("failed to parse SSE chunk", logger.String("data", data))
			continue
		}

		// 提取 OpenAI 格式的 content
		if choices, ok := payload["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if content, ok := delta["content"].(string); ok && content != "" {
						if err := chunkHandler(content); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}
