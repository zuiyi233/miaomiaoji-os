package model

import (
	"time"

	"gorm.io/datatypes"
)

// JobType 任务类型
type JobType string

const (
	JobTypePluginInvoke JobType = "plugin_invoke"
)

// JobStatus 任务状态
type JobStatus string

const (
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusSucceeded JobStatus = "succeeded"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCanceled  JobStatus = "canceled"
)

// Job 异步任务模型
// 对外暴露 job_uuid，内部仍使用自增主键 id
type Job struct {
	BaseModel
	JobUUID string `gorm:"size:36;not null;uniqueIndex" json:"job_uuid"`

	Type   JobType   `gorm:"size:50;not null" json:"type"`
	Status JobStatus `gorm:"size:20;not null;index" json:"status"`

	Progress int `json:"progress"`

	UserID    uint  `gorm:"index;not null" json:"user_id"`
	SessionID uint  `gorm:"index;not null" json:"session_id"`
	ProjectID *uint `gorm:"index" json:"project_id,omitempty"`

	PluginID uint           `gorm:"index;not null" json:"plugin_id"`
	Method   string         `gorm:"size:100;not null" json:"method"`
	Payload  datatypes.JSON `json:"payload"`

	Result       datatypes.JSON `json:"result"`
	ErrorMessage string         `gorm:"type:text" json:"error_message"`

	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// JobPublic 对外返回结构（隐藏内部自增主键）
type JobPublic struct {
	JobUUID      string         `json:"job_uuid"`
	Type         JobType        `json:"type"`
	Status       JobStatus      `json:"status"`
	Progress     int            `json:"progress"`
	SessionID    uint           `json:"session_id"`
	ProjectID    *uint          `json:"project_id,omitempty"`
	PluginID     uint           `json:"plugin_id"`
	Method       string         `json:"method"`
	Result       datatypes.JSON `json:"result"`
	ErrorMessage string         `json:"error_message"`
	StartedAt    *time.Time     `json:"started_at,omitempty"`
	FinishedAt   *time.Time     `json:"finished_at,omitempty"`
}

func (j *Job) ToPublic() JobPublic {
	return JobPublic{
		JobUUID:      j.JobUUID,
		Type:         j.Type,
		Status:       j.Status,
		Progress:     j.Progress,
		SessionID:    j.SessionID,
		ProjectID:    j.ProjectID,
		PluginID:     j.PluginID,
		Method:       j.Method,
		Result:       j.Result,
		ErrorMessage: j.ErrorMessage,
		StartedAt:    j.StartedAt,
		FinishedAt:   j.FinishedAt,
	}
}

func (Job) TableName() string {
	return "jobs"
}
