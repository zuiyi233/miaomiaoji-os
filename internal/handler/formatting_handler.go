package handler

import (
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type FormattingHandler struct {
	formattingService service.FormattingService
}

func NewFormattingHandler(formattingService service.FormattingService) *FormattingHandler {
	return &FormattingHandler{
		formattingService: formattingService,
	}
}

type FormatTextRequest struct {
	Text  string `json:"text" binding:"required"`
	Style string `json:"style" binding:"required"`
}

func (h *FormattingHandler) FormatText(c *gin.Context) {
	var req FormatTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	formatted, err := h.formattingService.FormatText(req.Text, req.Style)
	if err != nil {
		response.Fail(c, errors.CodeFormattingFailed, "Failed to format text")
		return
	}

	response.SuccessWithData(c, gin.H{
		"original":  req.Text,
		"formatted": formatted,
		"style":     req.Style,
	})
}

func (h *FormattingHandler) GetAvailableStyles(c *gin.Context) {
	styles := h.formattingService.GetAvailableStyles()
	response.SuccessWithData(c, gin.H{
		"styles": styles,
	})
}

type QualityHandler struct {
	qualityGateService service.QualityGateService
}

func NewQualityHandler(qualityGateService service.QualityGateService) *QualityHandler {
	return &QualityHandler{
		qualityGateService: qualityGateService,
	}
}

type CheckQualityRequest struct {
	Content string `json:"content" binding:"required"`
}

func (h *QualityHandler) CheckQuality(c *gin.Context) {
	var req CheckQualityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	result, err := h.qualityGateService.CheckQuality(req.Content)
	if err != nil {
		response.Fail(c, errors.CodeQualityCheckFailed, "Failed to check quality")
		return
	}

	response.SuccessWithData(c, result)
}

func (h *QualityHandler) GetThresholds(c *gin.Context) {
	thresholds := h.qualityGateService.GetThresholds()
	response.SuccessWithData(c, gin.H{
		"thresholds": thresholds,
	})
}
