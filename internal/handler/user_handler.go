package handler

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"omitempty,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
}

// ProfileResponse 用户资料响应
type ProfileResponse struct {
	ID            uint   `json:"id"`
	Username      string `json:"username"`
	Nickname      string `json:"nickname"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	Points        int    `json:"points"`
	CheckInStreak int    `json:"check_in_streak"`
}

// GetProfile 获取当前用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		logger.Error("Get user failed", logger.Err(err), logger.Uint("user_id", userID))
		response.Fail(c, errors.CodeNotFound, "用户不存在")
		return
	}

	response.SuccessWithData(c, ProfileResponse{
		ID:            user.ID,
		Username:      user.Username,
		Nickname:      user.Nickname,
		Email:         user.Email,
		Role:          user.Role,
		Points:        user.Points,
		CheckInStreak: user.CheckInStreak,
	})
}

// UpdateProfile 更新用户信息
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		logger.Error("Get user failed", logger.Err(err), logger.Uint("user_id", userID))
		response.Fail(c, errors.CodeNotFound, "用户不存在")
		return
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := h.userService.UpdateUser(user); err != nil {
		logger.Error("Update user failed", logger.Err(err), logger.Uint("user_id", userID))
		response.Fail(c, errors.CodeDatabaseError, "更新失败")
		return
	}

	response.SuccessWithData(c, ProfileResponse{
		ID:            user.ID,
		Username:      user.Username,
		Nickname:      user.Nickname,
		Email:         user.Email,
		Role:          user.Role,
		Points:        user.Points,
		CheckInStreak: user.CheckInStreak,
	})
}

// CheckIn 每日签到
func (h *UserHandler) CheckIn(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	if err := h.userService.CheckIn(userID); err != nil {
		if err.Error() == "already checked in today" {
			response.Fail(c, errors.CodeValidationError, "今日已签到")
			return
		}
		logger.Error("Check in failed", logger.Err(err), logger.Uint("user_id", userID))
		response.Fail(c, errors.CodeInternalError, "签到失败")
		return
	}

	response.SuccessWithData(c, gin.H{"message": "签到成功"})
}

// GetPoints 获取积分详情
func (h *UserHandler) GetPoints(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		logger.Error("Get user failed", logger.Err(err), logger.Uint("user_id", userID))
		response.Fail(c, errors.CodeNotFound, "用户不存在")
		return
	}

	response.SuccessWithData(c, gin.H{
		"points":          user.Points,
		"check_in_streak": user.CheckInStreak,
	})
}

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	Page int `form:"page" binding:"min=1"`
	Size int `form:"size" binding:"min=1,max=100"`
}

// ListUsers 获取用户列表（管理员）
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 10
	}

	users, total, err := h.userService.ListUsers(req.Page, req.Size)
	if err != nil {
		logger.Error("List users failed", logger.Err(err))
		response.Fail(c, errors.CodeDatabaseError, "获取用户列表失败")
		return
	}

	var userList []ProfileResponse
	for _, user := range users {
		userList = append(userList, ProfileResponse{
			ID:            user.ID,
			Username:      user.Username,
			Nickname:      user.Nickname,
			Email:         user.Email,
			Role:          user.Role,
			Points:        user.Points,
			CheckInStreak: user.CheckInStreak,
		})
	}

	response.SuccessWithPage(c, userList, total, req.Page, req.Size)
}

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	Status model.Status `json:"status" binding:"required,oneof=0 1"`
}

// UpdateUserStatus 更新用户状态（管理员）
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "无效的用户ID")
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		logger.Error("Get user failed", logger.Err(err), logger.Uint("user_id", id))
		response.Fail(c, errors.CodeNotFound, "用户不存在")
		return
	}

	user.Status = req.Status
	if err := h.userService.UpdateUser(user); err != nil {
		logger.Error("Update user status failed", logger.Err(err), logger.Uint("user_id", id))
		response.Fail(c, errors.CodeDatabaseError, "更新失败")
		return
	}

	response.Success(c)
}
