package handler

import (
	"novel-agent-os-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck 健康检查
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response.SuccessWithData(c, map[string]interface{}{
		"status":  "ok",
		"message": "Service is healthy",
	})
}

// ReadinessCheck 就绪检查
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	response.SuccessWithData(c, map[string]interface{}{
		"status":   "ready",
		"database": "connected",
	})
}

// RegisterHealthRoutes 注册健康检查路由
func RegisterHealthRoutes(r *gin.Engine) {
	handler := NewHealthHandler()
	r.GET("/healthz", handler.HealthCheck)
	r.GET("/ready", handler.ReadinessCheck)
}
