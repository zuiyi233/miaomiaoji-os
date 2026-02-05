package handler

import (
	"strconv"

	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// TemplateHandler 模板处理器
type TemplateHandler struct {
	templateService service.TemplateService
	projectService  service.ProjectService
}

// NewTemplateHandler 创建模板处理器
func NewTemplateHandler(templateService service.TemplateService, projectService service.ProjectService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
		projectService:  projectService,
	}
}

func (h *TemplateHandler) ensureProjectOwner(c *gin.Context, projectID uint) bool {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Error(c, errors.ErrUnauthorized)
		return false
	}
	project, err := h.projectService.GetByID(projectID)
	if err != nil {
		response.Error(c, errors.ErrProjectNotFound)
		return false
	}
	if project.UserID != userID {
		response.Error(c, errors.ErrForbidden)
		return false
	}
	return true
}

func (h *TemplateHandler) ensureTemplateOwner(c *gin.Context, templateID uint, write bool) (*model.Template, bool) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Error(c, errors.ErrUnauthorized)
		return nil, false
	}

	tmpl, err := h.templateService.GetByID(templateID)
	if err != nil {
		response.Error(c, err)
		return nil, false
	}

	// 系统模板（ProjectID=0）：允许读，不允许改/删
	if tmpl.ProjectID == 0 {
		if write {
			response.Error(c, errors.ErrForbidden)
			return nil, false
		}
		return tmpl, true
	}

	project, err := h.projectService.GetByID(tmpl.ProjectID)
	if err != nil {
		response.Error(c, errors.ErrProjectNotFound)
		return nil, false
	}
	if project.UserID != userID {
		response.Error(c, errors.ErrForbidden)
		return nil, false
	}
	return tmpl, true
}

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Template    string `json:"template" binding:"required"`
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Template    string `json:"template"`
}

// Create 创建模板
func (h *TemplateHandler) Create(c *gin.Context) {
	projectID, err := parseUintParam(c, "project_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if !h.ensureProjectOwner(c, projectID) {
		return
	}

	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("创建模板请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	template, err := h.templateService.Create(
		projectID,
		req.Name,
		req.Description,
		req.Category,
		req.Template,
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, template)
}

// ListByProject 根据项目获取模板列表
func (h *TemplateHandler) ListByProject(c *gin.Context) {
	projectID, err := parseUintParam(c, "project_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if !h.ensureProjectOwner(c, projectID) {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	templates, total, err := h.templateService.ListByProjectID(projectID, page, size)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPage(c, templates, total, page, size)
}

// ListSystem 获取系统模板列表
func (h *TemplateHandler) ListSystem(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	templates, total, err := h.templateService.ListSystemTemplates(page, size)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPage(c, templates, total, page, size)
}

// GetByID 根据ID获取模板
func (h *TemplateHandler) GetByID(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	template, ok := h.ensureTemplateOwner(c, id, false)
	if !ok {
		return
	}

	response.SuccessWithData(c, template)
}

// Update 更新模板
func (h *TemplateHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if _, ok := h.ensureTemplateOwner(c, id, true); !ok {
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新模板请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.Template != "" {
		updates["template"] = req.Template
	}

	template, err := h.templateService.Update(id, updates)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, template)
}

// Delete 删除模板
func (h *TemplateHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if _, ok := h.ensureTemplateOwner(c, id, true); !ok {
		return
	}

	if err := h.templateService.Delete(id); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}
