package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/datatypes"
	"novel-agent-os-backend/internal/config"
	"novel-agent-os-backend/internal/model"
)

// AIProviderConfig 供应商配置模型
type AIProviderConfig = model.AIProviderConfig

// AIConfigRepository AI配置仓储接口
type AIConfigRepository interface {
	SaveProviderConfig(provider string, payload ProviderConfigPayload) error
	GetProviderConfig(provider string) (*ProviderConfigRecord, error)
	UpdateModelsCache(provider, baseURL, apiKey string, models []ProviderModelInfo) error
}

// ProviderConfigRecord 供应商配置记录
type ProviderConfigRecord struct {
	Provider   string
	BaseURL    string
	APIKey     string
	ModelsPath string
}

// ProviderConfigPayload 配置写入载体
type ProviderConfigPayload struct {
	Provider    string   `yaml:"provider"`
	BaseURL     string   `yaml:"base_url"`
	APIKey      string   `yaml:"api_key"`
	ModelsCache []string `yaml:"models_cache"`
	UpdatedAt   string   `yaml:"updated_at"`
}

// ProviderModelInfo 模型信息
type ProviderModelInfo struct {
	ID string
}

// aiConfigRepository 实现
type aiConfigRepository struct {
	*BaseRepository
}

// NewAIConfigRepository 创建AI配置仓储
func NewAIConfigRepository() AIConfigRepository {
	return &aiConfigRepository{BaseRepository: GetBaseRepository()}
}

// SaveProviderConfig 保存供应商配置
func (r *aiConfigRepository) SaveProviderConfig(provider string, payload ProviderConfigPayload) error {
	clean := strings.TrimSpace(provider)
	if clean == "" {
		return errors.New("invalid provider")
	}

	record := &AIProviderConfig{}
	if err := r.BaseRepository.db.Where("provider = ?", clean).First(record).Error; err != nil {
		record.Provider = clean
	}
	record.BaseURL = payload.BaseURL
	record.APIKey = payload.APIKey
	cacheBytes, _ := json.Marshal(payload.ModelsCache)
	record.ModelsCache = datatypes.JSON(cacheBytes)
	record.UpdatedAtAt = time.Now()

	return r.BaseRepository.db.Save(record).Error
}

// GetProviderConfig 获取供应商配置
func (r *aiConfigRepository) GetProviderConfig(provider string) (*ProviderConfigRecord, error) {
	clean := strings.TrimSpace(provider)
	if clean == "" {
		return nil, errors.New("invalid provider")
	}

	providersPath := config.Get().AI.ProvidersPath
	if providersPath == "" {
		providersPath = "./configs/providers"
	}
	configPath := filepath.Join(providersPath, clean+".yaml")
	result := &ProviderConfigRecord{
		Provider:   clean,
		BaseURL:    "",
		APIKey:     "",
		ModelsPath: configPath,
	}

	envKey := strings.ToUpper(clean)
	result.APIKey = config.GetEnvFirst("NOVEL_AGENT_OS_"+envKey+"_API_KEY", envKey+"_API_KEY", "GOOGLE_API_KEY", "GEMINI_API_KEY")

	if payload, err := readProviderConfigFile(configPath); err == nil {
		result.BaseURL = payload.BaseURL
		if result.APIKey == "" {
			result.APIKey = payload.APIKey
		}
		return result, nil
	}

	var record AIProviderConfig
	if err := r.BaseRepository.db.Where("provider = ?", clean).First(&record).Error; err == nil {
		result.BaseURL = record.BaseURL
		if result.APIKey == "" {
			result.APIKey = record.APIKey
		}
	}

	return result, nil
}

// UpdateModelsCache 更新模型缓存
func (r *aiConfigRepository) UpdateModelsCache(provider, baseURL, apiKey string, models []ProviderModelInfo) error {
	clean := strings.TrimSpace(provider)
	if clean == "" {
		return errors.New("invalid provider")
	}

	providersPath := config.Get().AI.ProvidersPath
	if providersPath == "" {
		providersPath = "./configs/providers"
	}
	configPath := filepath.Join(providersPath, clean+".yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	modelIDs := make([]string, 0, len(models))
	for _, item := range models {
		if item.ID != "" {
			modelIDs = append(modelIDs, item.ID)
		}
	}

	payload := ProviderConfigPayload{
		Provider:    clean,
		BaseURL:     baseURL,
		APIKey:      apiKey,
		ModelsCache: modelIDs,
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	if err := writeProviderConfigFile(configPath, payload); err != nil {
		return err
	}

	record := &AIProviderConfig{}
	if err := r.BaseRepository.db.Where("provider = ?", clean).First(record).Error; err != nil {
		record.Provider = clean
	}
	record.BaseURL = baseURL
	record.APIKey = apiKey
	cacheBytes, _ := json.Marshal(modelIDs)
	record.ModelsCache = datatypes.JSON(cacheBytes)
	record.UpdatedAtAt = time.Now()

	return r.BaseRepository.db.Save(record).Error
}

func readProviderConfigFile(path string) (*ProviderConfigPayload, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var payload ProviderConfigPayload
	if err := yaml.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

func writeProviderConfigFile(path string, payload ProviderConfigPayload) error {
	data, err := yaml.Marshal(payload)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
