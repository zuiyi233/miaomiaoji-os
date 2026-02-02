package handler

import (
	"strconv"
	"time"

	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// PluginHandler 插件处理器
type PluginHandler struct {
	pluginService service.PluginService
}

// NewPluginHandler 创建插件处理器
func NewPluginHandler(pluginService service.PluginService) *PluginHandler {
	return &PluginHandler{
		pluginService: pluginService,
	}
}

// RegisterPluginRequest 注册插件请求
type RegisterPluginRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description"`
	Version     string `json:"version" binding:"required"`
	EntryPoint  string `json:"entry_point" binding:"required"`
	Manifest    string `json:"manifest"`
}

// UpdatePluginStatusRequest 更新插件状态请求
type UpdatePluginStatusRequest struct {
	Status string `json:"status" binding:"required"` // enabled/disabled
}

// UpdatePluginHealthRequest 更新插件健康状态请求
type UpdatePluginHealthRequest struct {
	Healthy bool `json:"healthy"`
	Latency int  `json:"latency_ms"`
}

// AddCapabilityRequest 添加插件能力请求
type AddCapabilityRequest struct {
	Name         string `json:"name" binding:"required,max=100"`
	Description  string `json:"description"`
	InputSchema  string `json:"input_schema"`
	OutputSchema string `json:"output_schema"`
}

// UpdateCapabilityRequest 更新插件能力请求
type UpdateCapabilityRequest struct {
	Description  string `json:"description"`
	InputSchema  string `json:"input_schema"`
	OutputSchema string `json:"output_schema"`
}

// InvokeCapabilityRequest 调用插件能力请求
type InvokeCapabilityRequest struct {
	Input   map[string]interface{} `json:"input"`
	Timeout int                    `json:"timeout_ms"` // 毫秒
}

// Register 注册插件
func (h *PluginHandler) Register(c *gin.Context) {
	var req RegisterPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("注册插件请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	plugin, err := h.pluginService.Register(
		req.Name,
		req.Description,
		req.Version,
		req.EntryPoint,
		req.Manifest,
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, plugin)
}

// List 获取插件列表
func (h *PluginHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	var plugins []*model.Plugin
	var total int64
	var err error

	if status != "" {
		plugins, total, err = h.pluginService.ListByStatus(status, page, size)
	} else {
		plugins, total, err = h.pluginService.List(page, size)
	}

	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPage(c, plugins, total, page, size)
}

// GetByID 根据ID获取插件
func (h *PluginHandler) GetByID(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	plugin, err := h.pluginService.GetByID(id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, plugin)
}

// UpdateStatus 更新插件状态
func (h *PluginHandler) UpdateStatus(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req UpdatePluginStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新插件状态请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.pluginService.UpdateStatus(id, req.Status); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// UpdateHealth 更新插件健康状态
func (h *PluginHandler) UpdateHealth(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req UpdatePluginHealthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新插件健康状态请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.pluginService.UpdateHealth(id, req.Healthy, req.Latency); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// Delete 删除插件
func (h *PluginHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.pluginService.Delete(id); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// AddCapability 添加插件能力
func (h *PluginHandler) AddCapability(c *gin.Context) {
	pluginID, err := parseUintParam(c, "plugin_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req AddCapabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("添加插件能力请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	cap, err := h.pluginService.AddCapability(
		pluginID,
		req.Name,
		req.Description,
		req.InputSchema,
		req.OutputSchema,
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, cap)
}

// ListCapabilities 获取插件能力列表
func (h *PluginHandler) ListCapabilities(c *gin.Context) {
	pluginID, err := parseUintParam(c, "plugin_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	caps, err := h.pluginService.ListCapabilities(pluginID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, caps)
}

// GetCapability 获取插件能力
func (h *PluginHandler) GetCapability(c *gin.Context) {
	capabilityID, err := parseUintParam(c, "capability_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	cap, err := h.pluginService.GetCapability(capabilityID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, cap)
}

// UpdateCapability 更新插件能力
func (h *PluginHandler) UpdateCapability(c *gin.Context) {
	capabilityID, err := parseUintParam(c, "capability_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req UpdateCapabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新插件能力请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	updates := make(map[string]interface{})
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.InputSchema != "" {
		updates["input_schema"] = req.InputSchema
	}
	if req.OutputSchema != "" {
		updates["output_schema"] = req.OutputSchema
	}

	if err := h.pluginService.UpdateCapability(capabilityID, updates); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// DeleteCapability 删除插件能力
func (h *PluginHandler) DeleteCapability(c *gin.Context) {
	capabilityID, err := parseUintParam(c, "capability_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.pluginService.DeleteCapability(capabilityID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// InvokeCapability 调用插件能力
func (h *PluginHandler) InvokeCapability(c *gin.Context) {
	capabilityID, err := parseUintParam(c, "capability_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req InvokeCapabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("调用插件能力请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	timeout := time.Duration(req.Timeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 30 * time.Second // 默认30秒超时
	}

	result, err := h.pluginService.InvokeCapability(capabilityID, req.Input, timeout)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, result)
}
