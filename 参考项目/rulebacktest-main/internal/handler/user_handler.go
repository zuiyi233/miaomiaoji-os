package handler

import (
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"rulebacktest/internal/model"
	"rulebacktest/internal/service"
	apperrors "rulebacktest/pkg/errors"
	"rulebacktest/pkg/response"
)

var (
	userHandlerInstance *UserHandler
	userHandlerOnce     sync.Once
)

// UserHandler 用户HTTP处理器
type UserHandler struct {
	service *service.UserService
}

// NewUserHandler 创建UserHandler实例
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{service: svc}
}

// GetUserHandler 获取UserHandler单例
func GetUserHandler() *UserHandler {
	userHandlerOnce.Do(func() {
		userHandlerInstance = &UserHandler{
			service: service.GetUserService(),
		}
	})
	return userHandlerInstance
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req model.UserRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	user, err := h.service.Register(&req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, user)
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req model.UserLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	resp, err := h.service.Login(&req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, resp)
}

// GetProfile 获取当前用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	user, err := h.service.GetByID(userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, user)
}

// UpdateProfile 更新用户信息
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	var req model.UserUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	user, err := h.service.Update(userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, user)
}

// GetByID 根据ID获取用户
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的用户ID")
		return
	}

	user, err := h.service.GetByID(id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, user)
}

// parseIDParam 解析ID参数
func (h *UserHandler) parseIDParam(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// handleError 统一处理错误
func (h *UserHandler) handleError(c *gin.Context, err error) {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		response.Fail(c, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, apperrors.CodeInternalError, "服务器内部错误")
}

// GetService 获取Service
func (h *UserHandler) GetService() *service.UserService {
	return h.service
}
