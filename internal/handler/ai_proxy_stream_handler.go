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

// AIProxyStreamHandler AI代理流式处理器
type AIProxyStreamHandler struct {
	configService service.AIConfigService
}

// NewAIProxyStreamHandler 创建AI代理流式处理器
func NewAIProxyStreamHandler(configService service.AIConfigService) *AIProxyStreamHandler {
	return &AIProxyStreamHandler{configService: configService}
}

// ProxyStreamRequest 代理流式请求
type ProxyStreamRequest struct {
	Provider string `json:"provider" binding:"required"`
	Path     string `json:"path" binding:"required"`
	Body     string `json:"body" binding:"required"`
}

// ProxyStream AI代理流式
func (h *AIProxyStreamHandler) ProxyStream(c *gin.Context) {
	var req ProxyStreamRequest
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

	client := &http.Client{Timeout: 0}
	proxyReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(req.Body))
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "proxy request failed")
		return
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("Accept", "text/event-stream")
	if providerCfg.APIKey != "" {
		if req.Provider == "gemini" {
			proxyReq.Header.Set("x-goog-api-key", providerCfg.APIKey)
		} else {
			proxyReq.Header.Set("Authorization", "Bearer "+providerCfg.APIKey)
		}
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		logger.Error("proxy stream request failed", logger.Err(err))
		response.Fail(c, errors.CodeExternalAPIError, "代理请求失败")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		response.Fail(c, errors.CodeExternalAPIError, string(body))
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	buffer := make([]byte, 4096)
	for {
		read, err := resp.Body.Read(buffer)
		if read > 0 {
			_, _ = c.Writer.Write(buffer[:read])
			c.Writer.Flush()
		}
		if err != nil {
			if err != io.EOF {
				logger.Warn("proxy stream read failed", logger.Err(err))
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
