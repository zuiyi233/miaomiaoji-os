package model

import (
	"time"

	"gorm.io/datatypes"
)

// AIProviderConfig 供应商配置模型
type AIProviderConfig struct {
	BaseModel
	Provider    string         `gorm:"size:30;uniqueIndex;not null" json:"provider"`
	BaseURL     string         `gorm:"type:text" json:"base_url"`
	APIKey      string         `gorm:"type:text" json:"api_key"`
	ModelsCache datatypes.JSON `json:"models_cache"`
	UpdatedAtAt time.Time      `json:"updated_at_at"`
}

// TableName 指定表名
func (AIProviderConfig) TableName() string {
	return "ai_provider_configs"
}
