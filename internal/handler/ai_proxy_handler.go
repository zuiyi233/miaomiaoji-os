package handler

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"
)

// AIProxyHandler AI代理处理器
type AIProxyHandler struct {
	configService service.AIConfigService
}

// NewAIProxyHandler 创建AI代理处理器
func NewAIProxyHandler(configService service.AIConfigService) *AIProxyHandler {
	return &AIProxyHandler{configService: configService}
}

// ProxyRequest 代理请求
type ProxyRequest struct {
	Provider string `json:"provider" binding:"required"`
	Path     string `json:"path" binding:"required"`
	Body     string `json:"body" binding:"required"`
}

// Proxy AI代理
func (h *AIProxyHandler) Proxy(c *gin.Context) {
	var req ProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	if req.Path == "" || strings.Contains(req.Path, "..") {
		response.Fail(c, errors.CodeInvalidParams, "invalid path")
		return
	}

	providerCfg, err := h.configService.GetProviderConfigRaw(req.Provider)
	if err != nil {
		response.Fail(c, errors.CodeNotFound, "provider not found")
		return
	}
	if err := service.ValidateAIProxyTarget(req.Provider, providerCfg.BaseURL, req.Path); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	base := strings.TrimRight(providerCfg.BaseURL, "/")
	url := base + "/" + strings.TrimLeft(req.Path, "/")

	client := &http.Client{Timeout: 60 * time.Second}
	proxyReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(req.Body))
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "proxy request failed")
		return
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("Accept", "application/json")
	if providerCfg.APIKey != "" {
		if req.Provider == "gemini" {
			proxyReq.Header.Set("x-goog-api-key", providerCfg.APIKey)
		} else {
			proxyReq.Header.Set("Authorization", "Bearer "+providerCfg.APIKey)
		}
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		logger.Error("proxy request failed", logger.Err(err))
		response.Fail(c, errors.CodeExternalAPIError, "代理请求失败")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Fail(c, errors.CodeExternalAPIError, "读取响应失败")
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		response.Fail(c, errors.CodeExternalAPIError, string(body))
		return
	}

	c.Data(resp.StatusCode, "application/json", body)
}
