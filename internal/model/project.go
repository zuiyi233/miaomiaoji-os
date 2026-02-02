package model

import (
	"gorm.io/datatypes"
)

// Project 项目模型
type Project struct {
	BaseModel
	Title         string         `gorm:"size:200;not null" json:"title"`
	Genre         string         `gorm:"size:50" json:"genre"`
	Tags          datatypes.JSON `json:"tags"`
	CoreConflict  string         `gorm:"type:text" json:"core_conflict"`
	CharacterArc  string         `gorm:"type:text" json:"character_arc"`
	UltimateValue string         `gorm:"type:text" json:"ultimate_value"`
	WorldRules    string         `gorm:"type:text" json:"world_rules"`
	AISettings    datatypes.JSON `json:"ai_settings"`
	UserID        uint           `gorm:"index;not null" json:"user_id"`
	
	// 关联
	Volumes   []Volume   `gorm:"foreignKey:ProjectID" json:"volumes,omitempty"`
	Documents []Document `gorm:"foreignKey:ProjectID" json:"documents,omitempty"`
	Entities  []Entity   `gorm:"foreignKey:ProjectID" json:"entities,omitempty"`
	Templates []Template `gorm:"foreignKey:ProjectID" json:"templates,omitempty"`
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}
