package handler

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	sessionService service.SessionService
}

func NewSessionHandler(sessionService service.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

type CreateSessionRequest struct {
	Title     string `json:"title" binding:"required"`
	Mode      string `json:"mode" binding:"required"`
	ProjectID uint   `json:"project_id" binding:"required"`
}

type UpdateSessionRequest struct {
	Title string `json:"title"`
	Mode  string `json:"mode"`
}

type CreateStepRequest struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	FormatType string `json:"format_type"`
	OrderIndex int    `json:"order_index"`
}

type UpdateStepRequest struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	FormatType string `json:"format_type"`
	OrderIndex int    `json:"order_index"`
}

func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "Unauthorized")
		return
	}

	session := &model.Session{
		Title:     req.Title,
		Mode:      req.Mode,
		ProjectID: req.ProjectID,
		UserID:    userID,
	}

	if err := h.sessionService.CreateSession(session); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to create session")
		return
	}

	response.SuccessWithData(c, session)
}

func (h *SessionHandler) GetSession(c *gin.Context) {
	id, err := parseUintParam(c, "session_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid session ID")
		return
	}

	session, err := h.sessionService.GetSession(id)
	if err != nil {
		response.Fail(c, errors.CodeSessionNotFound, "Session not found")
		return
	}

	userID := getUserIDFromContext(c)
	if session.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
		return
	}

	response.SuccessWithData(c, session)
}

func (h *SessionHandler) UpdateSession(c *gin.Context) {
	id, err := parseUintParam(c, "session_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid session ID")
		return
	}

	var req UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	session, err := h.sessionService.GetSession(id)
	if err != nil {
		response.Fail(c, errors.CodeSessionNotFound, "Session not found")
		return
	}

	userID := getUserIDFromContext(c)
	if session.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
		return
	}

	if req.Title != "" {
		session.Title = req.Title
	}
	if req.Mode != "" {
		session.Mode = req.Mode
	}

	if err := h.sessionService.UpdateSession(session); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to update session")
		return
	}

	response.SuccessWithData(c, session)
}

func (h *SessionHandler) DeleteSession(c *gin.Context) {
	id, err := parseUintParam(c, "session_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid session ID")
		return
	}

	session, err := h.sessionService.GetSession(id)
	if err != nil {
		response.Fail(c, errors.CodeSessionNotFound, "Session not found")
		return
	}

	userID := getUserIDFromContext(c)
	if session.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
		return
	}

	if err := h.sessionService.DeleteSession(id); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to delete session")
		return
	}

	response.Success(c)
}

func (h *SessionHandler) ListSessions(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "Unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	sessions, total, err := h.sessionService.ListSessions(userID, page, pageSize)
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to list sessions")
		return
	}

	response.SuccessWithData(c, gin.H{
		"sessions":  sessions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *SessionHandler) ListSessionsByProject(c *gin.Context) {
	projectID, err := parseUintParam(c, "project_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid project ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	sessions, total, err := h.sessionService.ListSessionsByProject(projectID, page, pageSize)
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to list sessions")
		return
	}

	response.SuccessWithData(c, gin.H{
		"sessions":  sessions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *SessionHandler) CreateStep(c *gin.Context) {
	sessionID, err := parseUintParam(c, "session_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid session ID")
		return
	}

	var req CreateStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	session, err := h.sessionService.GetSession(sessionID)
	if err != nil {
		response.Fail(c, errors.CodeSessionNotFound, "Session not found")
		return
	}

	userID := getUserIDFromContext(c)
	if session.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
		return
	}

	step := &model.SessionStep{
		Title:      req.Title,
		Content:    req.Content,
		FormatType: req.FormatType,
		OrderIndex: req.OrderIndex,
		SessionID:  sessionID,
	}

	if err := h.sessionService.CreateStep(step); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to create step")
		return
	}

	response.SuccessWithData(c, step)
}

func (h *SessionHandler) GetStep(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid step ID")
		return
	}

	step, err := h.sessionService.GetStep(id)
	if err != nil {
		response.Fail(c, errors.CodeSessionStepNotFound, "Step not found")
		return
	}

	session, err := h.sessionService.GetSession(step.SessionID)
	if err != nil {
		response.Fail(c, errors.CodeSessionNotFound, "Session not found")
		return
	}

	userID := getUserIDFromContext(c)
	if session.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
		return
	}

	response.SuccessWithData(c, step)
}

func (h *SessionHandler) UpdateStep(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid step ID")
		return
	}

	var req UpdateStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	step, err := h.sessionService.GetStep(id)
	if err != nil {
		response.Fail(c, errors.CodeSessionStepNotFound, "Step not found")
		return
	}

	session, err := h.sessionService.GetSession(step.SessionID)
	if err != nil {
		response.Fail(c, errors.CodeSessionNotFound, "Session not found")
		return
	}

	userID := getUserIDFromContext(c)
	if session.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
		return
	}

	if req.Title != "" {
		step.Title = req.Title
	}
	if req.Content != "" {
		step.Content = req.Content
	}
	if req.FormatType != "" {
		step.FormatType = req.FormatType
	}
	if req.OrderIndex > 0 {
		step.OrderIndex = req.OrderIndex
	}

	if err := h.sessionService.UpdateStep(step); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to update step")
		return
	}

	response.SuccessWithData(c, step)
}

func (h *SessionHandler) DeleteStep(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid step ID")
		return
	}

	step, err := h.sessionService.GetStep(id)
	if err != nil {
		response.Fail(c, errors.CodeSessionStepNotFound, "Step not found")
		return
	}

	session, err := h.sessionService.GetSession(step.SessionID)
	if err != nil {
		response.Fail(c, errors.CodeSessionNotFound, "Session not found")
		return
	}

	userID := getUserIDFromContext(c)
	if session.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
		return
	}

	if err := h.sessionService.DeleteStep(id); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to delete step")
		return
	}

	response.Success(c)
}

func (h *SessionHandler) ListSteps(c *gin.Context) {
	sessionID, err := parseUintParam(c, "session_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid session ID")
		return
	}

	session, err := h.sessionService.GetSession(sessionID)
	if err != nil {
		response.Fail(c, errors.CodeSessionNotFound, "Session not found")
		return
	}

	userID := getUserIDFromContext(c)
	if session.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
		return
	}

	steps, err := h.sessionService.ListSteps(sessionID)
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to list steps")
		return
	}

	response.SuccessWithData(c, steps)
}
