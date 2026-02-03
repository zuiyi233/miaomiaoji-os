package model

import (
	"gorm.io/datatypes"
)

// EntityCustomField 实体自定义字段
type EntityCustomField struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
	Type  string `json:"type"` // text/number/date/boolean
}

// Entity 实体/世界观卡模型
type Entity struct {
	BaseModel
	EntityType     string         `gorm:"size:20;not null;index" json:"entity_type"` // character/setting/organization/item/magic/event
	Title          string         `gorm:"size:200;not null" json:"title"`
	Subtitle       string         `gorm:"size:200" json:"subtitle"`
	Content        string         `gorm:"type:text" json:"content"`
	VoiceStyle     string         `gorm:"size:100" json:"voice_style"`
	Importance     string         `gorm:"size:20;default:secondary" json:"importance"` // main/secondary/minor
	CustomFields   datatypes.JSON `json:"custom_fields"`                               // EntityCustomField[]
	ReferenceCount int            `gorm:"default:0" json:"reference_count"`
	ProjectID      uint           `gorm:"index;not null" json:"project_id"`

	// 关联
	Tags  []EntityTag         `gorm:"foreignKey:EntityID" json:"tags,omitempty"`
	Links []EntityLink        `gorm:"foreignKey:SourceID" json:"links,omitempty"`
	Refs  []DocumentEntityRef `gorm:"foreignKey:EntityID" json:"refs,omitempty"`
}

// TableName 指定表名
func (Entity) TableName() string {
	return "entities"
}

// EntityTag 实体标签表（替代JSONB）
type EntityTag struct {
	BaseModelWithoutSoftDelete
	EntityID uint   `gorm:"index;not null" json:"entity_id"`
	Tag      string `gorm:"size:50;index;not null" json:"tag"`

	// 关联
	Entity Entity `gorm:"foreignKey:EntityID" json:"entity,omitempty"`
}

// TableName 指定表名
func (EntityTag) TableName() string {
	return "entity_tags"
}

// EntityLink 实体关联模型
type EntityLink struct {
	BaseModelWithoutSoftDelete
	SourceID     uint   `gorm:"index;not null" json:"source_id"`
	TargetID     uint   `gorm:"index;not null" json:"target_id"`
	Type         string `gorm:"size:20" json:"type"`
	RelationName string `gorm:"size:50" json:"relation_name"`

	// 关联
	Source Entity `gorm:"foreignKey:SourceID" json:"source,omitempty"`
	Target Entity `gorm:"foreignKey:TargetID" json:"target,omitempty"`
}

// TableName 指定表名
func (EntityLink) TableName() string {
	return "entity_links"
}
