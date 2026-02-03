package model

import (
	"gorm.io/datatypes"
)

// Session 工作流会话模型
type Session struct {
	BaseModel
	Title     string `gorm:"size:200" json:"title"`
	Mode      string `gorm:"size:50" json:"mode"` // Normal/Fusion/Single/Batch
	ProjectID uint   `gorm:"index;not null" json:"project_id"`
	UserID    uint   `gorm:"index;not null" json:"user_id"`

	// 关联
	Steps []SessionStep `gorm:"foreignKey:SessionID" json:"steps,omitempty"`
}

// TableName 指定表名
func (Session) TableName() string {
	return "sessions"
}

// SessionStep 会话步骤模型
type SessionStep struct {
	BaseModel
	Title      string         `gorm:"size:200" json:"title"`
	Content    string         `gorm:"type:text" json:"content"`
	FormatType string         `gorm:"size:50" json:"format_type"`
	OrderIndex int            `gorm:"index" json:"order_index"`
	Metadata   datatypes.JSON `json:"metadata"` // 摘要、字数等
	SessionID  uint           `gorm:"index;not null" json:"session_id"`

	// 关联
	Session Session `gorm:"foreignKey:SessionID" json:"session,omitempty"`
}

// TableName 指定表名
func (SessionStep) TableName() string {
	return "session_steps"
}
