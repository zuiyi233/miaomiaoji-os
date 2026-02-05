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

// EntityHandler 实体处理器
type EntityHandler struct {
	entityService  service.EntityService
	projectService service.ProjectService
}

// NewEntityHandler 创建实体处理器
func NewEntityHandler(entityService service.EntityService, projectService service.ProjectService) *EntityHandler {
	return &EntityHandler{
		entityService:  entityService,
		projectService: projectService,
	}
}

func (h *EntityHandler) ensureProjectOwner(c *gin.Context, projectID uint) bool {
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

func (h *EntityHandler) ensureEntityOwner(c *gin.Context, entityID uint) (*model.Entity, bool) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Error(c, errors.ErrUnauthorized)
		return nil, false
	}
	entity, err := h.entityService.GetByID(entityID)
	if err != nil {
		response.Error(c, err)
		return nil, false
	}
	project, err := h.projectService.GetByID(entity.ProjectID)
	if err != nil {
		response.Error(c, errors.ErrProjectNotFound)
		return nil, false
	}
	if project.UserID != userID {
		response.Error(c, errors.ErrForbidden)
		return nil, false
	}
	return entity, true
}

func (h *EntityHandler) ensureEntityLinkOwner(c *gin.Context, sourceID, targetID uint) bool {
	source, ok := h.ensureEntityOwner(c, sourceID)
	if !ok {
		return false
	}
	target, ok := h.ensureEntityOwner(c, targetID)
	if !ok {
		return false
	}
	if source.ProjectID != target.ProjectID {
		response.Error(c, errors.ErrForbidden)
		return false
	}
	return true
}

// CreateEntityRequest 创建实体请求
type CreateEntityRequest struct {
	EntityType   string                    `json:"entity_type" binding:"required"`
	Title        string                    `json:"title" binding:"required,max=200"`
	Subtitle     string                    `json:"subtitle"`
	Content      string                    `json:"content"`
	VoiceStyle   string                    `json:"voice_style"`
	Importance   string                    `json:"importance"`
	CustomFields []model.EntityCustomField `json:"custom_fields"`
}

// UpdateEntityRequest 更新实体请求
type UpdateEntityRequest struct {
	EntityType   string                    `json:"entity_type"`
	Title        string                    `json:"title"`
	Subtitle     string                    `json:"subtitle"`
	Content      string                    `json:"content"`
	VoiceStyle   string                    `json:"voice_style"`
	Importance   string                    `json:"importance"`
	CustomFields []model.EntityCustomField `json:"custom_fields"`
}

// AddTagRequest 添加标签请求
type AddTagRequest struct {
	Tag string `json:"tag" binding:"required,max=50"`
}

// CreateLinkRequest 创建关联请求
type CreateLinkRequest struct {
	TargetID     uint   `json:"target_id" binding:"required"`
	Type         string `json:"type"`
	RelationName string `json:"relation_name"`
}

// Create 创建实体
func (h *EntityHandler) Create(c *gin.Context) {
	projectID, err := parseUintParam(c, "project_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if !h.ensureProjectOwner(c, projectID) {
		return
	}

	var req CreateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("创建实体请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	importance := req.Importance
	if importance == "" {
		importance = "secondary"
	}

	entity, err := h.entityService.Create(
		projectID,
		req.EntityType,
		req.Title,
		req.Subtitle,
		req.Content,
		req.VoiceStyle,
		importance,
		req.CustomFields,
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, entity)
}

// List 获取实体列表
func (h *EntityHandler) List(c *gin.Context) {
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
	entityType := c.Query("type")
	tag := c.Query("tag")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	var entities []*model.Entity
	var total int64

	// 根据查询参数选择不同的查询方式
	if entityType != "" && tag != "" {
		entities, total, err = h.entityService.ListByTypeAndTag(projectID, entityType, tag, page, size)
	} else if entityType != "" {
		entities, total, err = h.entityService.ListByType(projectID, entityType, page, size)
	} else if tag != "" {
		entities, total, err = h.entityService.ListByTag(projectID, tag, page, size)
	} else {
		entities, total, err = h.entityService.ListByProjectID(projectID, page, size)
	}

	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPage(c, entities, total, page, size)
}

// GetByID 根据ID获取实体
func (h *EntityHandler) GetByID(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	entity, ok := h.ensureEntityOwner(c, id)
	if !ok {
		return
	}

	response.SuccessWithData(c, entity)
}

// Update 更新实体
func (h *EntityHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if _, ok := h.ensureEntityOwner(c, id); !ok {
		return
	}

	var req UpdateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新实体请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	updates := make(map[string]interface{})
	if req.EntityType != "" {
		updates["entity_type"] = req.EntityType
	}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Subtitle != "" {
		updates["subtitle"] = req.Subtitle
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.VoiceStyle != "" {
		updates["voice_style"] = req.VoiceStyle
	}
	if req.Importance != "" {
		updates["importance"] = req.Importance
	}
	if req.CustomFields != nil {
		updates["custom_fields"] = req.CustomFields
	}

	entity, err := h.entityService.Update(id, updates)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, entity)
}

// Delete 删除实体
func (h *EntityHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if _, ok := h.ensureEntityOwner(c, id); !ok {
		return
	}

	if err := h.entityService.Delete(id); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// AddTag 添加标签
func (h *EntityHandler) AddTag(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if _, ok := h.ensureEntityOwner(c, id); !ok {
		return
	}

	var req AddTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("添加标签请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.entityService.AddTag(id, req.Tag); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// RemoveTag 移除标签
func (h *EntityHandler) RemoveTag(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if _, ok := h.ensureEntityOwner(c, id); !ok {
		return
	}

	tag := c.Param("tag")
	if tag == "" {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.entityService.RemoveTag(id, tag); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// CreateLink 创建实体关联
func (h *EntityHandler) CreateLink(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("创建实体关联请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if !h.ensureEntityLinkOwner(c, id, req.TargetID) {
		return
	}

	if err := h.entityService.CreateLink(id, req.TargetID, req.Type, req.RelationName); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// DeleteLink 删除实体关联
func (h *EntityHandler) DeleteLink(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	targetID, err := parseUintParam(c, "target_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if !h.ensureEntityLinkOwner(c, id, targetID) {
		return
	}

	if err := h.entityService.DeleteLink(id, targetID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}
