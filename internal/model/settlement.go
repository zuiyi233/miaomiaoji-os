package model

import (
	"gorm.io/datatypes"
)

// SettlementEntry 结算记录模型
type SettlementEntry struct {
	BaseModel
	WorldID     string         `gorm:"size:100;index;not null" json:"world_id"`
	ChapterID   string         `gorm:"size:100;index;not null" json:"chapter_id"`
	LoopStage   string         `gorm:"size:30;index;not null" json:"loop_stage"`
	PointsDelta int            `json:"points_delta"`
	Payload     datatypes.JSON `json:"payload"`
	UserID      uint           `gorm:"index;not null" json:"user_id"`
}

// TableName 指定表名
func (SettlementEntry) TableName() string {
	return "settlement_entries"
}
