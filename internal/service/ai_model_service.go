package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"novel-agent-os-backend/internal/config"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/pkg/logger"
)

// AIModelService AI模型服务接口
type AIModelService interface {
	ListModels(provider string) ([]ModelInfo, error)
}

// ModelInfo 模型信息
type ModelInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

type aiModelService struct {
	configRepo repository.AIConfigRepository
}

// NewAIModelService 创建AI模型服务
func NewAIModelService(configRepo repository.AIConfigRepository) AIModelService {
	return &aiModelService{
		configRepo: configRepo,
	}
}

// ListModels 获取模型列表（带缓存兜底）
func (s *aiModelService) ListModels(provider string) ([]ModelInfo, error) {
	cleanProvider := strings.TrimSpace(provider)
	if cleanProvider == "" {
		return nil, errors.New("invalid provider")
	}

	cfg, err := s.configRepo.GetProviderConfig(cleanProvider)
	if err != nil {
		return nil, err
	}

	cacheTTL := config.Get().AI.ModelsCacheTTL
	useStaleCacheOnError := config.Get().AI.UseStaleCacheOnError

	// 检查缓存是否有效
	if s.configRepo.IsCacheValid(cleanProvider, cfg.BaseURL, cfg.APIKey, cacheTTL) {
		logger.Debug("模型缓存命中",
			logger.String("provider", cleanProvider))

		cachedModels, err := s.configRepo.GetModelsCache(cleanProvider, cfg.BaseURL, cfg.APIKey)
		if err == nil && len(cachedModels) > 0 {
			models := make([]ModelInfo, 0, len(cachedModels))
			for _, item := range cachedModels {
				models = append(models, ModelInfo{
					ID:       item.ID,
					Name:     item.ID,
					Provider: cleanProvider,
				})
			}
			return models, nil
		}
		logger.Warn("缓存读取失败，尝试请求上游",
			logger.String("provider", cleanProvider),
			logger.Err(err))
	} else {
		logger.Info("模型缓存已过期，请求上游",
			logger.String("provider", cleanProvider))
	}

	// 请求上游API
	models, err := fetchModelsFromProvider(cleanProvider, cfg.BaseURL, cfg.APIKey)
	if err != nil {
		logger.Warn("上游请求失败",
			logger.String("provider", cleanProvider),
			logger.Err(err))

		// 上游失败时尝试使用过期缓存
		if useStaleCacheOnError {
			cachedModels, cacheErr := s.configRepo.GetModelsCache(cleanProvider, cfg.BaseURL, cfg.APIKey)
			if cacheErr == nil && len(cachedModels) > 0 {
				logger.Info("使用过期缓存兜底",
					logger.String("provider", cleanProvider),
					logger.Int("models_count", len(cachedModels)))

				models := make([]ModelInfo, 0, len(cachedModels))
				for _, item := range cachedModels {
					models = append(models, ModelInfo{
						ID:       item.ID,
						Name:     item.ID,
						Provider: cleanProvider,
					})
				}
				return models, nil
			}
			logger.Error("无可用缓存",
				logger.String("provider", cleanProvider),
				logger.Err(cacheErr))
		}

		return nil, err
	}

	// 更新缓存
	cacheModels := make([]repository.ProviderModelInfo, 0, len(models))
	for _, item := range models {
		cacheModels = append(cacheModels, repository.ProviderModelInfo{ID: item.ID})
	}

	if err := s.configRepo.UpdateModelsCache(cleanProvider, cfg.BaseURL, cfg.APIKey, cacheModels); err != nil {
		logger.Warn("更新缓存失败",
			logger.String("provider", cleanProvider),
			logger.Err(err))
	} else {
		logger.Debug("缓存已更新",
			logger.String("provider", cleanProvider),
			logger.Int("models_count", len(models)))
	}

	return models, nil
}

type openAIModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

type openRouterModelsResponse struct {
	Data []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"data"`
}

type anthropicModelsResponse struct {
	Models []struct {
		ID string `json:"id"`
	} `json:"models"`
}

func fetchModelsFromProvider(provider, baseURL, apiKey string) ([]ModelInfo, error) {
	endpoint := buildModelsEndpoint(provider, baseURL)
	if endpoint == "" {
		return nil, errors.New("invalid provider endpoint")
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		if provider == "gemini" {
			req.Header.Set("x-goog-api-key", apiKey)
		} else if provider == "anthropic" {
			req.Header.Set("x-api-key", apiKey)
			req.Header.Set("anthropic-version", "2023-06-01")
		} else {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("provider request failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if provider == "gemini" {
		return parseGeminiModels(provider, body)
	}
	if provider == "openrouter" {
		return parseOpenRouterModels(provider, body)
	}
	if provider == "anthropic" {
		return parseAnthropicModels(provider, body)
	}

	var parsed openAIModelsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	models := make([]ModelInfo, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		if item.ID == "" {
			continue
		}
		models = append(models, ModelInfo{ID: item.ID, Name: item.ID, Provider: provider})
	}
	return models, nil
}

type geminiModelsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func parseGeminiModels(provider string, body []byte) ([]ModelInfo, error) {
	var parsed geminiModelsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	models := make([]ModelInfo, 0, len(parsed.Models))
	for _, item := range parsed.Models {
		if item.Name == "" {
			continue
		}
		id := strings.TrimPrefix(item.Name, "models/")
		name := id
		models = append(models, ModelInfo{ID: id, Name: name, Provider: provider})
	}
	return models, nil
}

func parseOpenRouterModels(provider string, body []byte) ([]ModelInfo, error) {
	var parsed openRouterModelsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	models := make([]ModelInfo, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		if item.ID == "" {
			continue
		}
		name := item.Name
		if name == "" {
			name = item.ID
		}
		models = append(models, ModelInfo{ID: item.ID, Name: name, Provider: provider})
	}
	return models, nil
}

func parseAnthropicModels(provider string, body []byte) ([]ModelInfo, error) {
	var parsed anthropicModelsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	models := make([]ModelInfo, 0, len(parsed.Models))
	for _, item := range parsed.Models {
		if item.ID == "" {
			continue
		}
		models = append(models, ModelInfo{ID: item.ID, Name: item.ID, Provider: provider})
	}
	return models, nil
}

func buildModelsEndpoint(provider, baseURL string) string {
	if baseURL == "" {
		return ""
	}

	cleanBase := strings.TrimRight(baseURL, "/")
	if provider == "gemini" {
		if strings.Contains(cleanBase, "/v1beta") {
			return cleanBase + "/models"
		}
		return cleanBase + "/v1beta/models"
	}
	if provider == "anthropic" {
		return cleanBase + "/v1/models"
	}

	if strings.Contains(cleanBase, "/chat/completions") {
		return strings.Replace(cleanBase, "/chat/completions", "/models", 1)
	}

	if strings.HasSuffix(cleanBase, "/models") {
		return cleanBase
	}
	return cleanBase + "/models"
}
