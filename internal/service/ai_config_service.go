package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
	"novel-agent-os-backend/internal/config"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/pkg/logger"
)

// AIConfigService AI配置服务接口
type AIConfigService interface {
	UpdateProviderConfig(provider, baseURL, apiKey string) error
	GetProviderConfig(provider string) (*ProviderConfig, error)
	TestProvider(provider string) error
}

// ProviderConfig 供应商配置
type ProviderConfig struct {
	Provider    string   `yaml:"provider" json:"provider"`
	BaseURL     string   `yaml:"base_url" json:"base_url"`
	APIKey      string   `yaml:"api_key" json:"api_key"`
	ModelsCache []string `yaml:"models_cache" json:"models_cache"`
	UpdatedAt   string   `yaml:"updated_at" json:"updated_at"`
}

type aiConfigService struct {
	mu         sync.Mutex
	configRepo repository.AIConfigRepository
}

// NewAIConfigService 创建AI配置服务
func NewAIConfigService(configRepo repository.AIConfigRepository) AIConfigService {
	return &aiConfigService{
		configRepo: configRepo,
	}
}

// UpdateProviderConfig 更新供应商配置（写入config.yaml + DB）
func (s *aiConfigService) UpdateProviderConfig(provider, baseURL, apiKey string) error {
	cleanProvider := strings.TrimSpace(provider)
	if cleanProvider == "" {
		return errors.New("invalid provider")
	}
	cleanBaseURL := strings.TrimSpace(baseURL)
	if cleanBaseURL == "" {
		return errors.New("invalid base url")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	providersPath := config.Get().AI.ProvidersPath
	if providersPath == "" {
		providersPath = "./configs/providers"
	}
	configPath := filepath.Join(providersPath, cleanProvider+".yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	payload := ProviderConfig{
		Provider:  cleanProvider,
		BaseURL:   cleanBaseURL,
		APIKey:    strings.TrimSpace(apiKey),
		UpdatedAt: config.NowUTCString(),
	}

	data, err := yaml.Marshal(payload)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}

	if err := s.configRepo.SaveProviderConfig(cleanProvider, repository.ProviderConfigPayload{
		Provider:    payload.Provider,
		BaseURL:     payload.BaseURL,
		APIKey:      payload.APIKey,
		ModelsCache: payload.ModelsCache,
		UpdatedAt:   payload.UpdatedAt,
	}); err != nil {
		return err
	}

	if err := config.Reload(); err != nil {
		logger.Warn("config reload failed", logger.Err(err))
	}

	return nil
}

// GetProviderConfig 获取供应商配置
func (s *aiConfigService) GetProviderConfig(provider string) (*ProviderConfig, error) {
	cleanProvider := strings.TrimSpace(provider)
	if cleanProvider == "" {
		return nil, errors.New("invalid provider")
	}

	record, err := s.configRepo.GetProviderConfig(cleanProvider)
	if err != nil {
		return nil, err
	}

	return &ProviderConfig{
		Provider:  record.Provider,
		BaseURL:   record.BaseURL,
		APIKey:    maskAPIKey(record.APIKey),
		UpdatedAt: "",
	}, nil
}

// TestProvider 测试供应商连接
func (s *aiConfigService) TestProvider(provider string) error {
	cleanProvider := strings.TrimSpace(provider)
	if cleanProvider == "" {
		return errors.New("invalid provider")
	}

	record, err := s.configRepo.GetProviderConfig(cleanProvider)
	if err != nil {
		return err
	}

	if record.BaseURL == "" {
		return errors.New("base url missing")
	}

	_, err = fetchModelsFromProvider(cleanProvider, record.BaseURL, record.APIKey)
	return err
}

func maskAPIKey(value string) string {
	if len(value) <= 6 {
		return value
	}
	return value[:3] + "***" + value[len(value)-3:]
}
