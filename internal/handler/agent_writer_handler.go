package handler

import (
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AgentWriterHandler AgentWriter 处理器
type AgentWriterHandler struct {
	agentWriterService *service.AgentWriterService
}

// NewAgentWriterHandler 创建 AgentWriter 处理器
func NewAgentWriterHandler(agentWriterService *service.AgentWriterService) *AgentWriterHandler {
	return &AgentWriterHandler{
		agentWriterService: agentWriterService,
	}
}

// StartWritingTaskRequest 启动写作任务请求
type StartWritingTaskRequest struct {
	ProjectID  uint                     `json:"project_id" binding:"required"`
	DocumentID uint                     `json:"document_id" binding:"required"`
	Prompt     string                   `json:"prompt" binding:"required"`
	Outline    []service.ChapterOutline `json:"outline" binding:"required"`
	Provider   string                   `json:"provider" binding:"required"`
	Path       string                   `json:"path" binding:"required"`
}

// StartWritingTask 启动写作任务
func (h *AgentWriterHandler) StartWritingTask(c *gin.Context) {
	var req StartWritingTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数绑定失败", logger.Err(err))
		response.Fail(c, errors.CodeInvalidParams, "参数错误")
		return
	}

	// 获取用户ID
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未授权")
		return
	}

	// 启动写作任务
	session, err := h.agentWriterService.StartWritingTask(
		req.ProjectID,
		req.DocumentID,
		userID,
		req.Prompt,
		req.Outline,
		req.Provider,
		req.Path,
	)
	if err != nil {
		logger.Error("启动写作任务失败", logger.Err(err))
		response.Fail(c, errors.CodeInternalError, "启动写作任务失败")
		return
	}

	response.SuccessWithData(c, gin.H{
		"session_id": session.ID,
		"status":     session.WorkflowStatus,
		"message":    "写作任务已启动，请通过 SSE 监听进度",
	})
}

// CancelWritingTaskRequest 取消写作任务请求
type CancelWritingTaskRequest struct {
	SessionID uint `json:"session_id" binding:"required"`
}

// CancelWritingTask 取消写作任务
func (h *AgentWriterHandler) CancelWritingTask(c *gin.Context) {
	var req CancelWritingTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数绑定失败", logger.Err(err))
		response.Fail(c, errors.CodeInvalidParams, "参数错误")
		return
	}

	// 取消写作任务
	if err := h.agentWriterService.CancelWritingTask(req.SessionID); err != nil {
		logger.Error("取消写作任务失败", logger.Err(err))
		response.Fail(c, errors.CodeInternalError, err.Error())
		return
	}

	response.SuccessWithData(c, gin.H{
		"session_id": req.SessionID,
		"message":    "写作任务已取消",
	})
}

// GetWritingTaskStatus 查询写作任务状态
func (h *AgentWriterHandler) GetWritingTaskStatus(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "无效的会话ID")
		return
	}

	// 获取会话服务
	sessionService := service.NewSessionService(nil)
	session, err := sessionService.GetSession(uint(sessionID))
	if err != nil {
		logger.Error("获取会话失败", logger.Err(err))
		response.Fail(c, errors.CodeNotFound, "会话不存在")
		return
	}

	response.SuccessWithData(c, gin.H{
		"session_id":      session.ID,
		"workflow_type":   session.WorkflowType,
		"workflow_status": session.WorkflowStatus,
		"workflow_config": session.WorkflowConfig,
		"title":           session.Title,
		"created_at":      session.CreatedAt,
		"updated_at":      session.UpdatedAt,
	})
}
