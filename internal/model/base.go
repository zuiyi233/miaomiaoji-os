package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型（含软删除）
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BaseModelWithoutSoftDelete 基础模型（不含软删除）
type BaseModelWithoutSoftDelete struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Status 状态类型
type Status int

const (
	StatusEnabled  Status = 1
	StatusDisabled Status = 0
)
