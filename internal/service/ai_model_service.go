package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"novel-agent-os-backend/internal/repository"
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

// ListModels 获取模型列表
func (s *aiModelService) ListModels(provider string) ([]ModelInfo, error) {
	cleanProvider := strings.TrimSpace(provider)
	if cleanProvider == "" {
		return nil, errors.New("invalid provider")
	}

	cfg, err := s.configRepo.GetProviderConfig(cleanProvider)
	if err != nil {
		return nil, err
	}

	models, err := fetchModelsFromProvider(cleanProvider, cfg.BaseURL, cfg.APIKey)
	if err != nil {
		return nil, err
	}

	cacheModels := make([]repository.ProviderModelInfo, 0, len(models))
	for _, item := range models {
		cacheModels = append(cacheModels, repository.ProviderModelInfo{ID: item.ID})
	}

	if err := s.configRepo.UpdateModelsCache(cleanProvider, cfg.BaseURL, cfg.APIKey, cacheModels); err != nil {
		return nil, err
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
