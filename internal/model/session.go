package model

import (
	"gorm.io/datatypes"
)

// Session 工作流会话模型
type Session struct {
	BaseModel
	Title          string         `gorm:"size:200" json:"title"`
	Mode           string         `gorm:"size:50" json:"mode"` // Normal/Fusion/Single/Batch
	ProjectID      uint           `gorm:"index;not null" json:"project_id"`
	UserID         uint           `gorm:"index;not null" json:"user_id"`
	WorkflowType   string         `gorm:"size:50;index" json:"workflow_type"`   // agent_writer/function_calling/stream
	WorkflowStatus string         `gorm:"size:20;index" json:"workflow_status"` // pending/running/completed/error/cancelled
	WorkflowConfig datatypes.JSON `json:"workflow_config"`                      // 工作流配置（存储 outline、prompt 等）

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
	Title        string         `gorm:"size:200" json:"title"`
	Content      string         `gorm:"type:text" json:"content"`
	FormatType   string         `gorm:"size:50" json:"format_type"`
	OrderIndex   int            `gorm:"index" json:"order_index"`
	Metadata     datatypes.JSON `json:"metadata"` // 摘要、字数等
	SessionID    uint           `gorm:"index;not null" json:"session_id"`
	IsStreaming  bool           `gorm:"default:false" json:"is_streaming"`  // 是否为流式步骤
	StreamStatus string         `gorm:"size:20" json:"stream_status"`       // 流式状态: streaming/completed/error
	StepType     string         `gorm:"size:20;index" json:"step_type"`     // 步骤类型: user/assistant/tool_call/tool_result
	ToolCallID   string         `gorm:"size:100;index" json:"tool_call_id"` // 工具调用ID，关联tool_call和tool_result

	// 关联
	Session Session `gorm:"foreignKey:SessionID" json:"session,omitempty"`
}

// TableName 指定表名
func (SessionStep) TableName() string {
	return "session_steps"
}
