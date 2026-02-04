package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"
)

// AIModelHandler AI模型处理器
type AIModelHandler struct {
	modelService service.AIModelService
}

// NewAIModelHandler 创建AI模型处理器
func NewAIModelHandler(modelService service.AIModelService) *AIModelHandler {
	return &AIModelHandler{
		modelService: modelService,
	}
}

// ListModels 获取模型列表
func (h *AIModelHandler) ListModels(c *gin.Context) {
	provider := strings.TrimSpace(c.Query("provider"))
	if provider == "" {
		response.Fail(c, errors.CodeInvalidParams, "provider is required")
		return
	}

	models, err := h.modelService.ListModels(provider)
	if err != nil {
		logger.Error("List models failed", logger.Err(err))
		response.Fail(c, errors.CodeExternalAPIError, "获取模型列表失败")
		return
	}

	response.SuccessWithData(c, gin.H{
		"models": models,
	})
}
