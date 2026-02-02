package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"novel-agent-os/internal/service"
	"novel-agent-os/pkg/errors"
	"novel-agent-os/pkg/logger"
	"novel-agent-os/pkg/response"
)

// SessionHandler 会话处理器
type SessionHandler struct {
	sessionService service.SessionService
}

// NewSessionHandler 创建会话处理器
func NewSessionHandler(sessionService service.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// RegisterRoutes 注册路由
func (h *SessionHandler) RegisterRoutes(r *gin.RouterGroup) {
	sessions := r.Group("/sessions")
	{
		sessions.POST("", h.Create)
		sessions.GET("", h.List)
		sessions.GET("/:id", h.Get)
		sessions.PUT("/:id", h.Update)
		sessions.DELETE("/:id", h.Delete)

		// SessionStep 路由
		sessions.POST("/:id/steps", h.AddStep)
		sessions.GET("/:id/steps", h.ListSteps)
		sessions.PUT("/steps/reorder", h.ReorderSteps)
		sessions.GET("/steps/:step_id", h.GetStep)
		sessions.PUT("/steps/:step_id", h.UpdateStep)
		sessions.DELETE("/steps/:step_id", h.DeleteStep)
	}
}

// Create 创建会话
func (h *SessionHandler) Create(c *gin.Context) {
	var req struct {
		Title     string `json:"title" binding:"required"`
		Mode      string `json:"mode" binding:"required"`
		ProjectID uint   `json:"project_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	// 从上下文获取当前用户ID
	userID, _ := c.Get("userID")
	uid, _ := userID.(uint)

	session, err := h.sessionService.Create(req.Title, req.Mode, req.ProjectID, uid)
	if err != nil {
		logger.Error("创建会话失败", logger.Err(err))
		response.Error(c, errors.ErrInternalServer)
		return
	}

	response.Success(c, session)
}

// Get 获取会话详情
func (h *SessionHandler) Get(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	session, err := h.sessionService.GetByID(id)
	if err != nil {
		logger.Error("获取会话失败", logger.Err(err))
		response.Error(c, errors.ErrNotFound)
		return
	}

	response.Success(c, session)
}

// List 列��会话
func (h *SessionHandler) List(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	var sessions []*model.Session
	var total int64
	var err error

	if projectIDStr != "" {
		projectID, _ := strconv.ParseUint(projectIDStr, 10, 64)
		sessions, total, err = h.sessionService.ListByProject(uint(projectID), page, size)
	} else {
		// 从上下文获取当前用户ID
		userID, _ := c.Get("userID")
		uid, _ := userID.(uint)
		sessions, total, err = h.sessionService.ListByUser(uid, page, size)
	}

	if err != nil {
		logger.Error("列会话失败", logger.Err(err))
		response.Error(c, errors.ErrInternalServer)
		return
	}

	response.SuccessWithPage(c, sessions, total, page, size)
}

// Update 更新会话
func (h *SessionHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req struct {
		Title string `json:"title"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	session, err := h.sessionService.Update(id, req.Title)
	if err != nil {
		logger.Error("更新会话失败", logger.Err(err))
		response.Error(c, errors.ErrInternalServer)
		return
	}

	response.Success(c, session)
}

// Delete 删除会话
func (h *SessionHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.sessionService.Delete(id); err != nil {
		logger.Error("删除会话失败", logger.Err(err))
		response.Error(c, errors.ErrInternalServer)
		return
	}

	response.Success(c, nil)
}

// AddStep 添加步骤
func (h *SessionHandler) AddStep(c *gin.Context) {
	sessionID, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams