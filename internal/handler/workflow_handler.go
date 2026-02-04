package handler

import (
	"time"

	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// WorkflowHandler 工作流执行处理器
type WorkflowHandler struct {
	workflowService service.WorkflowService
	sessionService  service.SessionService
}

func NewWorkflowHandler(workflowService service.WorkflowService, sessionService service.SessionService) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
		sessionService:  sessionService,
	}
}

type RunWorkflowRequest struct {
	ProjectID uint   `json:"project_id" binding:"required"`
	SessionID uint   `json:"session_id"`
	Title     string `json:"title"`
	StepTitle string `json:"step_title"`
	Provider  string `json:"provider" binding:"required"`
	Path      string `json:"path" binding:"required"`
	Body      string `json:"body" binding:"required"`
}

func (h *WorkflowHandler) RunWorld(c *gin.Context) {
	h.runWorkflow(c, "world", "世界观生成")
}

func (h *WorkflowHandler) RunPolish(c *gin.Context) {
	h.runWorkflow(c, "polish", "章节润色")
}

func (h *WorkflowHandler) runWorkflow(c *gin.Context, formatType string, defaultTitle string) {
	var req RunWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "Unauthorized")
		return
	}

	var sess *model.Session
	if req.SessionID > 0 {
		existing, err := h.sessionService.GetSession(req.SessionID)
		if err != nil {
			response.Fail(c, errors.CodeSessionNotFound, "Session not found")
			return
		}
		if existing.UserID != userID {
			response.Fail(c, errors.CodeForbidden, "Access denied")
			return
		}
		sess = existing
	}

	title := req.Title
	if title == "" {
		title = defaultTitle + " " + time.Now().Format("2006-01-02 15:04")
	}
	stepTitle := req.StepTitle
	if stepTitle == "" {
		stepTitle = defaultTitle
	}

	runReq := service.RunWorkflowRequest{
		UserID:       userID,
		ProjectID:    req.ProjectID,
		SessionTitle: title,
		Mode:         formatType,
		StepTitle:    stepTitle,
		FormatType:   formatType,
		Provider:     req.Provider,
		Path:         req.Path,
		Body:         req.Body,
		Session:      sess,
	}

	result, err := h.workflowService.RunStep(runReq)
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to run workflow")
		return
	}

	response.SuccessWithData(c, gin.H{
		"session": result.Session,
		"step":    result.Step,
		"content": result.Content,
		"raw":     result.Raw,
	})
}
