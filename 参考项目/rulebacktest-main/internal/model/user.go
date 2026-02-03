package model

// 用户角色常量
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User 用户模型
type User struct {
	BaseModel
	Username string `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password string `gorm:"size:255;not null" json:"-"`
	Nickname string `gorm:"size:50" json:"nickname"`
	Email    string `gorm:"size:100;uniqueIndex" json:"email"`
	Phone    string `gorm:"size:20" json:"phone"`
	Avatar   string `gorm:"size:255" json:"avatar"`
	Role     string `gorm:"size:20;default:user" json:"role"`
	Status   Status `gorm:"default:1" json:"status"`
}

func (User) TableName() string {
	return "users"
}

// UserRegisterReq 用户注册请求
type UserRegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Nickname string `json:"nickname" binding:"max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty,max=20"`
}

// UserLoginReq 用户登录请求
type UserLoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserLoginResp 用户登录响应
type UserLoginResp struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// UserUpdateReq 用户更新请求
type UserUpdateReq struct {
	Nickname string `json:"nickname" binding:"max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty,max=20"`
	Avatar   string `json:"avatar" binding:"omitempty,max=255"`
}
