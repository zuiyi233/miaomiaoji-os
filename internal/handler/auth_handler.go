package handler

import (
	"github.com/gin-gonic/gin"
	"novel-agent-os-backend/internal/middleware"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService service.UserService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userService service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Email    string `json:"email" binding:"omitempty,email"`
	Nickname string `json:"nickname" binding:"omitempty,max=50"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token              string   `json:"token"`
	ExpiresIn          int      `json:"expires_in"`
	MustChangePassword bool     `json:"must_change_password"`
	User               UserInfo `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Points   int    `json:"points"`
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	user, err := h.userService.Register(req.Username, req.Password, req.Email, req.Nickname)
	if err != nil {
		logger.Error("Register failed", logger.Err(err), logger.String("username", req.Username))
		if err.Error() == "username already exists" {
			response.Fail(c, errors.CodeAlreadyExists, "用户名已存在")
			return
		}
		if err.Error() == "email already exists" {
			response.Fail(c, errors.CodeAlreadyExists, "邮箱已被注册")
			return
		}
		response.Fail(c, errors.CodeInternalError, "注册失败")
		return
	}

	// 生成JWT Token
	token, err := middleware.GenerateToken(user.ID, user.Role)
	if err != nil {
		logger.Error("Generate token failed", logger.Err(err))
		response.Fail(c, errors.CodeInternalError, "生成Token失败")
		return
	}

	response.SuccessWithData(c, AuthResponse{
		Token:              token,
		ExpiresIn:          24 * 3600, // 24小时
		MustChangePassword: user.MustChangePassword,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Email:    user.Email,
			Role:     user.Role,
			Points:   user.Points,
		},
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	user, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		logger.Warn("Login failed", logger.String("username", req.Username), logger.Err(err))
		response.Fail(c, errors.CodeUnauthorized, "用户名或密码错误")
		return
	}

	// 生成JWT Token
	token, err := middleware.GenerateToken(user.ID, user.Role)
	if err != nil {
		logger.Error("Generate token failed", logger.Err(err))
		response.Fail(c, errors.CodeInternalError, "生成Token失败")
		return
	}

	response.SuccessWithData(c, AuthResponse{
		Token:              token,
		ExpiresIn:          24 * 3600, // 24小时
		MustChangePassword: user.MustChangePassword,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Email:    user.Email,
			Role:     user.Role,
			Points:   user.Points,
		},
	})
}

// Logout 用户退出
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT无状态，客户端删除token即可
	response.Success(c)
}

// Refresh 刷新Token
func (h *AuthHandler) Refresh(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	role, _ := c.Get("userRole")
	roleStr, _ := role.(string)
	uid, _ := userID.(uint)

	if uid == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	// 生成新Token
	token, err := middleware.GenerateToken(uid, roleStr)
	if err != nil {
		logger.Error("Generate token failed", logger.Err(err))
		response.Fail(c, errors.CodeInternalError, "生成Token失败")
		return
	}

	response.SuccessWithData(c, gin.H{
		"token":      token,
		"expires_in": 24 * 3600,
	})
}
