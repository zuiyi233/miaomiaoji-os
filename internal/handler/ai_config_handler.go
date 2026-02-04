package handler

import (
	"github.com/gin-gonic/gin"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"
)

// AIConfigHandler AI配置处理器
type AIConfigHandler struct {
	configService service.AIConfigService
}

// NewAIConfigHandler 创建AI配置处理器
func NewAIConfigHandler(configService service.AIConfigService) *AIConfigHandler {
	return &AIConfigHandler{
		configService: configService,
	}
}

// UpdateProviderRequest 更新供应商配置请求
type UpdateProviderRequest struct {
	Provider string `json:"provider" binding:"required"`
	BaseURL  string `json:"base_url" binding:"required"`
	APIKey   string `json:"api_key" binding:"omitempty"`
}

// TestProviderRequest 测试连接请求
type TestProviderRequest struct {
	Provider string `json:"provider" binding:"required"`
}

// GetProviderRequest 获取供应商配置请求
type GetProviderRequest struct {
	Provider string `form:"provider" binding:"required"`
}

// UpdateProvider 更新供应商配置
func (h *AIConfigHandler) UpdateProvider(c *gin.Context) {
	var req UpdateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	if err := h.configService.UpdateProviderConfig(req.Provider, req.BaseURL, req.APIKey); err != nil {
		logger.Error("Update provider config failed", logger.Err(err))
		response.Fail(c, errors.CodeInternalError, "更新配置失败")
		return
	}

	response.Success(c)
}

// GetProvider 获取供应商配置
func (h *AIConfigHandler) GetProvider(c *gin.Context) {
	var req GetProviderRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	item, err := h.configService.GetProviderConfig(req.Provider)
	if err != nil {
		response.Fail(c, errors.CodeNotFound, "供应商不存在")
		return
	}

	response.SuccessWithData(c, item)
}

// TestProvider 测试供应商连接
func (h *AIConfigHandler) TestProvider(c *gin.Context) {
	var req TestProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	if err := h.configService.TestProvider(req.Provider); err != nil {
		response.Fail(c, errors.CodeExternalAPIError, err.Error())
		return
	}

	response.Success(c)
}
