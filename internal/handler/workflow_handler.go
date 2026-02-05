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
	documentService service.DocumentService
}

func NewWorkflowHandler(workflowService service.WorkflowService, sessionService service.SessionService, documentService service.DocumentService) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
		sessionService:  sessionService,
		documentService: documentService,
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

type ChapterWriteBack struct {
	Mode       string `json:"mode"`
	SetStatus  string `json:"set_status"`
	SetSummary bool   `json:"set_summary"`
}

type ChapterGenerateRequest struct {
	ProjectID  uint            `json:"project_id" binding:"required"`
	SessionID  uint            `json:"session_id"`
	DocumentID uint            `json:"document_id"`
	VolumeID   uint            `json:"volume_id"`
	Title      string          `json:"title"`
	OrderIndex int             `json:"order_index"`
	Provider   string          `json:"provider" binding:"required"`
	Path       string          `json:"path" binding:"required"`
	Body       string          `json:"body" binding:"required"`
	WriteBack  ChapterWriteBack `json:"write_back"`
}

type ChapterAnalyzeRequest struct {
	ProjectID  uint            `json:"project_id" binding:"required"`
	SessionID  uint            `json:"session_id"`
	DocumentID uint            `json:"document_id" binding:"required"`
	Provider   string          `json:"provider" binding:"required"`
	Path       string          `json:"path" binding:"required"`
	Body       string          `json:"body" binding:"required"`
	WriteBack  ChapterWriteBack `json:"write_back"`
}

type ChapterRewriteRequest struct {
	ProjectID   uint            `json:"project_id" binding:"required"`
	SessionID   uint            `json:"session_id"`
	DocumentID  uint            `json:"document_id" binding:"required"`
	RewriteMode string          `json:"rewrite_mode"`
	Provider    string          `json:"provider" binding:"required"`
	Path        string          `json:"path" binding:"required"`
	Body        string          `json:"body" binding:"required"`
	WriteBack   ChapterWriteBack `json:"write_back"`
}

type ChapterBatchItem struct {
	Title      string `json:"title"`
	OrderIndex int    `json:"order_index"`
	Outline    string `json:"outline"`
}

type ChapterBatchRequest struct {
	ProjectID    uint             `json:"project_id" binding:"required"`
	SessionID    uint             `json:"session_id"`
	VolumeID     uint             `json:"volume_id"`
	Items        []ChapterBatchItem `json:"items" binding:"required"`
	Provider     string           `json:"provider" binding:"required"`
	Path         string           `json:"path" binding:"required"`
	BodyTemplate string           `json:"body_template" binding:"required"`
	WriteBack    ChapterWriteBack  `json:"write_back"`
}

func (h *WorkflowHandler) RunWorld(c *gin.Context) {
	h.runWorkflow(c, "world", "世界观生成")
}

func (h *WorkflowHandler) RunPolish(c *gin.Context) {
	h.runWorkflow(c, "polish", "章节润色")
}

func (h *WorkflowHandler) RunChapterGenerate(c *gin.Context) {
	var req ChapterGenerateRequest
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
		title = "章节生成 " + time.Now().Format("2006-01-02 15:04")
	}
	if req.Title == "" {
		req.Title = title
	}

	result, err := h.workflowService.RunChapterGenerate(service.ChapterGenerateRequest{
		UserID:       userID,
		ProjectID:    req.ProjectID,
		Session:      sess,
		SessionTitle: title,
		DocumentID:   req.DocumentID,
		VolumeID:     req.VolumeID,
		Title:        req.Title,
		OrderIndex:   req.OrderIndex,
		Provider:     req.Provider,
		Path:         req.Path,
		Body:         req.Body,
		WriteBack: service.ChapterWriteBack{
			Mode:       req.WriteBack.Mode,
			SetStatus:  req.WriteBack.SetStatus,
			SetSummary: req.WriteBack.SetSummary,
		},
	})
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to run workflow")
		return
	}

	response.SuccessWithData(c, gin.H{
		"session":  result.Session,
		"document": result.Document,
		"steps":    result.Steps,
		"content":  result.Content,
		"raw":      result.Raw,
	})
}

func (h *WorkflowHandler) RunChapterAnalyze(c *gin.Context) {
	var req ChapterAnalyzeRequest
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

	title := "章节分析 " + time.Now().Format("2006-01-02 15:04")
	result, err := h.workflowService.RunChapterAnalyze(service.ChapterAnalyzeRequest{
		UserID:       userID,
		ProjectID:    req.ProjectID,
		Session:      sess,
		SessionTitle: title,
		DocumentID:   req.DocumentID,
		Provider:     req.Provider,
		Path:         req.Path,
		Body:         req.Body,
		WriteBack: service.ChapterWriteBack{
			SetStatus:  req.WriteBack.SetStatus,
			SetSummary: req.WriteBack.SetSummary,
		},
	})
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to run workflow")
		return
	}

	response.SuccessWithData(c, gin.H{
		"session":  result.Session,
		"document": result.Document,
		"content":  result.Content,
		"raw":      result.Raw,
	})
}

func (h *WorkflowHandler) RunChapterRewrite(c *gin.Context) {
	var req ChapterRewriteRequest
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

	title := "章节重写 " + time.Now().Format("2006-01-02 15:04")
	result, err := h.workflowService.RunChapterRewrite(service.ChapterRewriteRequest{
		UserID:       userID,
		ProjectID:    req.ProjectID,
		Session:      sess,
		SessionTitle: title,
		DocumentID:   req.DocumentID,
		RewriteMode:  req.RewriteMode,
		Provider:     req.Provider,
		Path:         req.Path,
		Body:         req.Body,
		WriteBack: service.ChapterWriteBack{
			Mode:      req.WriteBack.Mode,
			SetStatus: req.WriteBack.SetStatus,
		},
	})
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to run workflow")
		return
	}

	response.SuccessWithData(c, gin.H{
		"session":  result.Session,
		"document": result.Document,
		"content":  result.Content,
		"raw":      result.Raw,
	})
}

func (h *WorkflowHandler) RunChapterBatch(c *gin.Context) {
	var req ChapterBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}
	if len(req.Items) == 0 {
		response.Fail(c, errors.CodeInvalidParams, "Items required")
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

	title := "批量生成 " + time.Now().Format("2006-01-02 15:04")
	items := make([]service.ChapterBatchItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, service.ChapterBatchItem{
			Title:      item.Title,
			OrderIndex: item.OrderIndex,
			Outline:    item.Outline,
		})
	}

	result, err := h.workflowService.RunChapterBatch(service.ChapterBatchRequest{
		UserID:       userID,
		ProjectID:    req.ProjectID,
		Session:      sess,
		SessionTitle: title,
		VolumeID:     req.VolumeID,
		Items:        items,
		Provider:     req.Provider,
		Path:         req.Path,
		BodyTemplate: req.BodyTemplate,
		WriteBack: service.ChapterWriteBack{
			SetStatus:  req.WriteBack.SetStatus,
			SetSummary: req.WriteBack.SetSummary,
		},
	})
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to run workflow")
		return
	}

	response.SuccessWithData(c, gin.H{
		"session":   result.Session,
		"documents": result.Documents,
	})
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
