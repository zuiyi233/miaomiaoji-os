package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"novel-agent-os-backend/internal/config"
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/pkg/logger"

	"gopkg.in/yaml.v3"
	"gorm.io/datatypes"
)

// AIProviderConfig 供应商配置模型
type AIProviderConfig = model.AIProviderConfig

// AIConfigRepository AI配置仓储接口
type AIConfigRepository interface {
	SaveProviderConfig(provider string, payload ProviderConfigPayload) error
	GetProviderConfig(provider string) (*ProviderConfigRecord, error)
	UpdateModelsCache(provider, baseURL, apiKey string, models []ProviderModelInfo) error
	GetModelsCache(provider, baseURL, apiKey string) ([]ProviderModelInfo, error)
	IsCacheValid(provider, baseURL, apiKey string, ttl int) bool
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

	// 环境变量替换：支持 ${ENV_VAR} 格式
	payload.APIKey = expandEnvVar(payload.APIKey)

	// 如果配置文件中有明文密钥（不是环境变量占位符），输出警告
	if payload.APIKey != "" && !strings.HasPrefix(strings.TrimSpace(payload.APIKey), "${") {
		if len(payload.APIKey) > 10 {
			logger.Warn("配置文件中包含明文API密钥，建议使用环境变量",
				logger.String("provider", payload.Provider),
				logger.String("file", path))
		}
	}

	return &payload, nil
}

// expandEnvVar 展开环境变量占位符
func expandEnvVar(value string) string {
	if value == "" {
		return value
	}

	// 匹配 ${VAR_NAME} 格式
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(value, func(match string) string {
		varName := strings.TrimSpace(match[2 : len(match)-1])
		if envValue := os.Getenv(varName); envValue != "" {
			return envValue
		}
		// 如果环境变量不存在，返回空字符串
		return ""
	})
}

func writeProviderConfigFile(path string, payload ProviderConfigPayload) error {
	data, err := yaml.Marshal(payload)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// GetModelsCache 获取模型缓存
func (r *aiConfigRepository) GetModelsCache(provider, baseURL, apiKey string) ([]ProviderModelInfo, error) {
	clean := strings.TrimSpace(provider)
	if clean == "" {
		return nil, errors.New("invalid provider")
	}

	providersPath := config.Get().AI.ProvidersPath
	if providersPath == "" {
		providersPath = "./configs/providers"
	}
	configPath := filepath.Join(providersPath, clean+".yaml")

	payload, err := readProviderConfigFile(configPath)
	if err != nil {
		logger.Debug("读取配置文件缓存失败，尝试从数据库读取",
			logger.String("provider", clean),
			logger.Err(err))

		var record AIProviderConfig
		if err := r.BaseRepository.db.Where("provider = ?", clean).First(&record).Error; err != nil {
			return nil, err
		}

		var modelIDs []string
		if err := json.Unmarshal(record.ModelsCache, &modelIDs); err != nil {
			return nil, err
		}

		models := make([]ProviderModelInfo, 0, len(modelIDs))
		for _, id := range modelIDs {
			models = append(models, ProviderModelInfo{ID: id})
		}
		return models, nil
	}

	models := make([]ProviderModelInfo, 0, len(payload.ModelsCache))
	for _, id := range payload.ModelsCache {
		if id != "" {
			models = append(models, ProviderModelInfo{ID: id})
		}
	}

	return models, nil
}

// IsCacheValid 检查缓存是否有效
func (r *aiConfigRepository) IsCacheValid(provider, baseURL, apiKey string, ttl int) bool {
	clean := strings.TrimSpace(provider)
	if clean == "" {
		return false
	}

	providersPath := config.Get().AI.ProvidersPath
	if providersPath == "" {
		providersPath = "./configs/providers"
	}
	configPath := filepath.Join(providersPath, clean+".yaml")

	payload, err := readProviderConfigFile(configPath)
	if err != nil {
		logger.Debug("读取配置文件失败，缓存无效",
			logger.String("provider", clean),
			logger.Err(err))
		return false
	}

	if payload.UpdatedAt == "" {
		logger.Debug("缓存无更新时间，视为无效",
			logger.String("provider", clean))
		return false
	}

	updatedAt, err := time.Parse(time.RFC3339, payload.UpdatedAt)
	if err != nil {
		logger.Debug("解析缓存时间失败",
			logger.String("provider", clean),
			logger.String("updated_at", payload.UpdatedAt),
			logger.Err(err))
		return false
	}

	expiresAt := updatedAt.Add(time.Duration(ttl) * time.Second)
	isValid := time.Now().Before(expiresAt)

	if !isValid {
		logger.Debug("缓存已过期",
			logger.String("provider", clean),
			logger.String("updated_at", payload.UpdatedAt),
			logger.String("expires_at", expiresAt.Format(time.RFC3339)))
	}

	return isValid
}
