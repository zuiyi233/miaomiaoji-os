package model

import (
	"time"
)

// User 用户模型
type User struct {
	BaseModel
	Username      string     `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password      string     `gorm:"size:255;not null" json:"-"`
	Nickname      string     `gorm:"size:50" json:"nickname"`
	Email         string     `gorm:"size:100;uniqueIndex" json:"email"`
	Role          string     `gorm:"size:20;default:user" json:"role"` // user/admin
	Status        Status     `gorm:"default:1" json:"status"`
	Points        int        `gorm:"default:0" json:"points"`
	CheckInStreak int        `gorm:"default:0" json:"check_in_streak"`
	LastCheckIn   *time.Time `json:"last_check_in,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
