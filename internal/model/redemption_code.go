package model

import (
	"time"

	"gorm.io/datatypes"
)

// RedemptionCode 兑换码模型
type RedemptionCode struct {
	BaseModel
	Code         string         `gorm:"size:60;uniqueIndex;not null" json:"code"`
	Status       string         `gorm:"size:20;default:active" json:"status"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
	MaxUses      int            `gorm:"default:1" json:"max_uses"`
	UsedCount    int            `gorm:"default:0" json:"used_count"`
	DurationDays int            `gorm:"default:0" json:"duration_days"`
	CreatedBy    uint           `gorm:"index" json:"created_by"`
	BatchID      string         `gorm:"size:50" json:"batch_id"`
	Prefix       string         `gorm:"size:20" json:"prefix"`
	Tags         datatypes.JSON `json:"tags"`
	Note         string         `gorm:"type:text" json:"note"`
	Source       string         `gorm:"size:30" json:"source"`
}

// TableName 指定表名
func (RedemptionCode) TableName() string {
	return "redemption_codes"
}

// RedemptionCodeUse 兑换码使用记录
type RedemptionCodeUse struct {
	BaseModelWithoutSoftDelete
	CodeID   uint   `gorm:"index;not null" json:"code_id"`
	UserID   uint   `gorm:"index;not null" json:"user_id"`
	DeviceID string `gorm:"size:100" json:"device_id"`
}

// TableName 指定表名
func (RedemptionCodeUse) TableName() string {
	return "redemption_code_uses"
}
