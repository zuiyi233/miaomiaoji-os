package model

import (
	"time"

	"gorm.io/datatypes"
)

// Plugin 插件模型
type Plugin struct {
	BaseModel
	Name          string         `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Version       string         `gorm:"size:20" json:"version"`
	Author        string         `gorm:"size:100" json:"author"`
	Description   string         `gorm:"type:text" json:"description"`
	Endpoint      string         `gorm:"size:500" json:"endpoint"`
	EntryPoint    string         `gorm:"size:500" json:"entry_point"`
	Manifest      datatypes.JSON `json:"manifest"`
	IsEnabled     bool           `gorm:"default:false" json:"is_enabled"`
	Status        string         `gorm:"size:20;default:disabled" json:"status"` // enabled/disabled/error/unknown
	Healthy       bool           `gorm:"default:false" json:"healthy"`
	LastPing      *time.Time     `json:"last_ping,omitempty"`
	LastHeartbeat *time.Time     `json:"last_heartbeat,omitempty"`
	LatencyMs     int            `json:"latency_ms"`
	Config        datatypes.JSON `json:"config"`

	// 关联
	Capabilities []PluginCapability `gorm:"foreignKey:PluginID" json:"capabilities,omitempty"`
}

// TableName 指定表名
func (Plugin) TableName() string {
	return "plugins"
}

// PluginCapability 插件能力模型
type PluginCapability struct {
	BaseModel
	PluginID     uint           `gorm:"index;not null" json:"plugin_id"`
	CapID        string         `gorm:"size:50" json:"cap_id"`
	Name         string         `gorm:"size:100" json:"name"`
	Type         string         `gorm:"size:30" json:"type"` // text_processor/data_provider/ui_extension/logic_checker/generator
	Description  string         `gorm:"type:text" json:"description"`
	Icon         string         `gorm:"size:100" json:"icon"`
	InputSchema  datatypes.JSON `json:"input_schema"`
	OutputSchema datatypes.JSON `json:"output_schema"`

	// 关联
	Plugin Plugin `gorm:"foreignKey:PluginID" json:"plugin,omitempty"`
}

// TableName 指定表名
func (PluginCapability) TableName() string {
	return "plugin_capabilities"
}
