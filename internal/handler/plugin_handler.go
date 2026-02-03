package handler

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PluginHandler struct {
	pluginService service.PluginService
	jobService    service.JobService
}

func NewPluginHandler(pluginService service.PluginService, jobService service.JobService) *PluginHandler {
	return &PluginHandler{
		pluginService: pluginService,
		jobService:    jobService,
	}
}

type CreatePluginRequest struct {
	Name        string `json:"name" binding:"required"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	EntryPoint  string `json:"entry_point"`
}

type UpdatePluginRequest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	EntryPoint  string `json:"entry_point"`
	IsEnabled   *bool  `json:"is_enabled"`
}

type CreateCapabilityRequest struct {
	CapID       string `json:"cap_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

func (h *PluginHandler) CreatePlugin(c *gin.Context) {
	var req CreatePluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	plugin := &model.Plugin{
		Name:        req.Name,
		Version:     req.Version,
		Author:      req.Author,
		Description: req.Description,
		Endpoint:    req.Endpoint,
		EntryPoint:  req.EntryPoint,
		IsEnabled:   false,
		Status:      "disabled",
		Healthy:     false,
	}

	if err := h.pluginService.CreatePlugin(plugin); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to create plugin")
		return
	}

	response.SuccessWithData(c, plugin)
}

func (h *PluginHandler) GetPlugin(c *gin.Context) {
	id, err := parseUintParam(c, "plugin_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid plugin ID")
		return
	}

	plugin, err := h.pluginService.GetPlugin(id)
	if err != nil {
		response.Fail(c, errors.CodePluginNotFound, "Plugin not found")
		return
	}

	response.SuccessWithData(c, plugin)
}

func (h *PluginHandler) UpdatePlugin(c *gin.Context) {
	id, err := parseUintParam(c, "plugin_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid plugin ID")
		return
	}

	var req UpdatePluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	plugin, err := h.pluginService.GetPlugin(id)
	if err != nil {
		response.Fail(c, errors.CodePluginNotFound, "Plugin not found")
		return
	}

	if req.Name != "" {
		plugin.Name = req.Name
	}
	if req.Version != "" {
		plugin.Version = req.Version
	}
	if req.Author != "" {
		plugin.Author = req.Author
	}
	if req.Description != "" {
		plugin.Description = req.Description
	}
	if req.Endpoint != "" {
		plugin.Endpoint = req.Endpoint
	}
	if req.EntryPoint != "" {
		plugin.EntryPoint = req.EntryPoint
	}
	if req.IsEnabled != nil {
		plugin.IsEnabled = *req.IsEnabled
	}

	if err := h.pluginService.UpdatePlugin(plugin); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to update plugin")
		return
	}

	response.SuccessWithData(c, plugin)
}

func (h *PluginHandler) DeletePlugin(c *gin.Context) {
	id, err := parseUintParam(c, "plugin_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid plugin ID")
		return
	}

	if err := h.pluginService.DeletePlugin(id); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to delete plugin")
		return
	}

	response.Success(c)
}

func (h *PluginHandler) ListPlugins(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	plugins, total, err := h.pluginService.ListPlugins(page, pageSize)
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to list plugins")
		return
	}

	response.SuccessWithData(c, gin.H{
		"plugins":   plugins,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *PluginHandler) EnablePlugin(c *gin.Context) {
	id, err := parseUintParam(c, "plugin_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid plugin ID")
		return
	}

	if err := h.pluginService.EnablePlugin(id); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to enable plugin")
		return
	}

	response.Success(c)
}

func (h *PluginHandler) DisablePlugin(c *gin.Context) {
	id, err := parseUintParam(c, "plugin_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid plugin ID")
		return
	}

	if err := h.pluginService.DisablePlugin(id); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to disable plugin")
		return
	}

	response.Success(c)
}

func (h *PluginHandler) PingPlugin(c *gin.Context) {
	id, err := parseUintParam(c, "plugin_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid plugin ID")
		return
	}

	if err := h.pluginService.PingPlugin(id); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to ping plugin")
		return
	}

	response.Success(c)
}

func (h *PluginHandler) AddCapability(c *gin.Context) {
	pluginID, err := parseUintParam(c, "plugin_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid plugin ID")
		return
	}

	var req CreateCapabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	capability := &model.PluginCapability{
		PluginID:    pluginID,
		CapID:       req.CapID,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Icon:        req.Icon,
	}

	if err := h.pluginService.AddCapability(capability); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to add capability")
		return
	}

	response.SuccessWithData(c, capability)
}

func (h *PluginHandler) GetCapabilities(c *gin.Context) {
	pluginID, err := parseUintParam(c, "plugin_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid plugin ID")
		return
	}

	capabilities, err := h.pluginService.GetCapabilities(pluginID)
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to get capabilities")
		return
	}

	response.SuccessWithData(c, capabilities)
}

func (h *PluginHandler) RemoveCapability(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid capability ID")
		return
	}

	if err := h.pluginService.RemoveCapability(id); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to remove capability")
		return
	}

	response.Success(c)
}

type InvokePluginRequest struct {
	Method  string                 `json:"method" binding:"required"`
	Payload map[string]interface{} `json:"payload"`
}

type InvokePluginAsyncRequest struct {
	SessionID uint                   `json:"session_id" binding:"required"`
	Method    string                 `json:"method" binding:"required"`
	Payload   map[string]interface{} `json:"payload"`
}

func (h *PluginHandler) InvokePlugin(c *gin.Context) {
	id, err := parseUintParam(c, "plugin_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid plugin ID")
		return
	}

	var req InvokePluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	authorizationHeader := c.GetHeader("Authorization")
	result, err := h.pluginService.InvokePlugin(c.Request.Context(), id, req.Method, req.Payload, authorizationHeader)
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to invoke plugin")
		return
	}

	response.SuccessWithData(c, result)
}

func (h *PluginHandler) InvokePluginAsync(c *gin.Context) {
	pluginID, err := parseUintParam(c, "plugin_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid plugin ID")
		return
	}

	var req InvokePluginAsyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "Unauthorized")
		return
	}

	authorizationHeader := c.GetHeader("Authorization")
	job, err := h.jobService.CreatePluginInvokeJobFromSession(userID, req.SessionID, pluginID, req.Method, req.Payload, authorizationHeader)
	if err != nil {
		// 约定：service 内用字符串错误区分，保持简单
		if err.Error() == "access denied" {
			response.Fail(c, errors.CodeForbidden, "Access denied")
			return
		}
		response.Fail(c, errors.CodeJobCreateFailed, "Failed to create job")
		return
	}

	c.Header("Location", "/api/v1/jobs/"+job.JobUUID)
	c.JSON(202, response.Response{Code: errors.CodeSuccess, Message: "success", Data: gin.H{"job_uuid": job.JobUUID, "status": job.Status}})
}
