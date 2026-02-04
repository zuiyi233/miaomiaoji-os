package model

import (
	"time"
)

// User 用户模型
type User struct {
	BaseModel
	Username       string     `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password       string     `gorm:"size:255;not null" json:"-"`
	Nickname       string     `gorm:"size:50" json:"nickname"`
	Email          string     `gorm:"size:100;uniqueIndex" json:"email"`
	Role           string     `gorm:"size:20;default:user" json:"role"` // user/admin
	Status         Status     `gorm:"default:1" json:"status"`
	Points         int        `gorm:"default:0" json:"points"`
	CheckInStreak  int        `gorm:"default:0" json:"check_in_streak"`
	LastCheckIn    *time.Time `json:"last_check_in,omitempty"`
	AIAccessUntil  *time.Time `json:"ai_access_until,omitempty"`
	InviteCodeUsed string     `gorm:"size:100" json:"invite_code_used"`
	LastDeviceID   string     `gorm:"size:100" json:"last_device_id"`
	LastActiveAt   *time.Time `json:"last_active_at,omitempty"`
	// MustChangePassword 是否需要修改密码（默认管理员首次登录提示修改）
	MustChangePassword bool `gorm:"default:false" json:"must_change_password"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
