package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"
)

// DocumentHandler 文档处理器
type DocumentHandler struct {
	documentService service.DocumentService
	projectService  service.ProjectService
	volumeService   service.VolumeService
}

// NewDocumentHandler 创建文档处理器
func NewDocumentHandler(documentService service.DocumentService, projectService service.ProjectService, volumeService service.VolumeService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
		projectService:  projectService,
		volumeService:   volumeService,
	}
}

func (h *DocumentHandler) ensureProjectOwner(c *gin.Context, projectID uint) bool {
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

func (h *DocumentHandler) ensureDocumentOwner(c *gin.Context, documentID uint) (*serviceSafeDocument, bool) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Error(c, errors.ErrUnauthorized)
		return nil, false
	}
	document, err := h.documentService.GetByID(documentID)
	if err != nil {
		response.Error(c, err)
		return nil, false
	}
	project, err := h.projectService.GetByID(document.ProjectID)
	if err != nil {
		response.Error(c, errors.ErrProjectNotFound)
		return nil, false
	}
	if project.UserID != userID {
		response.Error(c, errors.ErrForbidden)
		return nil, false
	}
	return &serviceSafeDocument{DocumentID: document.ID, ProjectID: document.ProjectID, VolumeID: document.VolumeID}, true
}

func (h *DocumentHandler) ensureVolumeOwner(c *gin.Context, volumeID uint) bool {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Error(c, errors.ErrUnauthorized)
		return false
	}
	volume, err := h.volumeService.GetByID(volumeID)
	if err != nil {
		response.Error(c, err)
		return false
	}
	project, err := h.projectService.GetByID(volume.ProjectID)
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

// serviceSafeDocument 避免 handler 直接依赖 model
type serviceSafeDocument struct {
	DocumentID uint
	ProjectID  uint
	VolumeID   uint
}

// CreateDocumentRequest 创建文档请求
type CreateDocumentRequest struct {
	Title                string `json:"title" binding:"required,max=200"`
	Content              string `json:"content"`
	Summary              string `json:"summary"`
	Status               string `json:"status"`
	OrderIndex           int    `json:"order_index" binding:"required"`
	TimeNode             string `json:"time_node"`
	Duration             string `json:"duration"`
	TargetWordCount      int    `json:"target_word_count"`
	ChapterGoal          string `json:"chapter_goal"`
	CorePlot             string `json:"core_plot"`
	Hook                 string `json:"hook"`
	CauseEffect          string `json:"cause_effect"`
	ForeshadowingDetails string `json:"foreshadowing_details"`
	VolumeID             uint   `json:"volume_id"`
}

// UpdateDocumentRequest 更新文档请求
type UpdateDocumentRequest struct {
	Title                string `json:"title"`
	Content              string `json:"content"`
	Summary              string `json:"summary"`
	Status               string `json:"status"`
	OrderIndex           *int   `json:"order_index,omitempty"`
	TimeNode             string `json:"time_node"`
	Duration             string `json:"duration"`
	TargetWordCount      *int   `json:"target_word_count,omitempty"`
	ChapterGoal          string `json:"chapter_goal"`
	CorePlot             string `json:"core_plot"`
	Hook                 string `json:"hook"`
	CauseEffect          string `json:"cause_effect"`
	ForeshadowingDetails string `json:"foreshadowing_details"`
	VolumeID             *uint  `json:"volume_id,omitempty"`
}

// AddBookmarkRequest 添加书签请求
type AddBookmarkRequest struct {
	Title    string `json:"title" binding:"required"`
	Position int    `json:"position" binding:"required"`
	Note     string `json:"note"`
}

// LinkEntityRequest 关联实体请求
type LinkEntityRequest struct {
	EntityID uint                   `json:"entity_id" binding:"required"`
	RefType  string                 `json:"ref_type"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Create 创建文档
func (h *DocumentHandler) Create(c *gin.Context) {
	projectID, err := parseUintParam(c, "project_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if !h.ensureProjectOwner(c, projectID) {
		return
	}

	var req CreateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("创建文档请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	status := req.Status
	if status == "" {
		status = "草稿"
	}

	document, err := h.documentService.Create(
		projectID,
		req.Title,
		req.Content,
		req.Summary,
		status,
		req.OrderIndex,
		req.TimeNode,
		req.Duration,
		req.TargetWordCount,
		req.ChapterGoal,
		req.CorePlot,
		req.Hook,
		req.CauseEffect,
		req.ForeshadowingDetails,
		req.VolumeID,
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, document)
}

// ListByProject 根据项目获取文档列表
func (h *DocumentHandler) ListByProject(c *gin.Context) {
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

	documents, total, err := h.documentService.ListByProjectID(projectID, page, size)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPage(c, documents, total, page, size)
}

// ListByVolume 根据卷获取文档列表
func (h *DocumentHandler) ListByVolume(c *gin.Context) {
	volumeID, err := parseUintParam(c, "volume_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if !h.ensureVolumeOwner(c, volumeID) {
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

	documents, total, err := h.documentService.ListByVolumeID(volumeID, page, size)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPage(c, documents, total, page, size)
}

// GetByID 根据ID获取文档
func (h *DocumentHandler) GetByID(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if _, ok := h.ensureDocumentOwner(c, id); !ok {
		return
	}
	document, err := h.documentService.GetByID(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithData(c, document)
}

// Update 更新文档
func (h *DocumentHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if _, ok := h.ensureDocumentOwner(c, id); !ok {
		return
	}

	var req UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新文档请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Summary != "" {
		updates["summary"] = req.Summary
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.OrderIndex != nil {
		updates["order_index"] = *req.OrderIndex
	}
	if req.TimeNode != "" {
		updates["time_node"] = req.TimeNode
	}
	if req.Duration != "" {
		updates["duration"] = req.Duration
	}
	if req.TargetWordCount != nil {
		updates["target_word_count"] = *req.TargetWordCount
	}
	if req.ChapterGoal != "" {
		updates["chapter_goal"] = req.ChapterGoal
	}
	if req.CorePlot != "" {
		updates["core_plot"] = req.CorePlot
	}
	if req.Hook != "" {
		updates["hook"] = req.Hook
	}
	if req.CauseEffect != "" {
		updates["cause_effect"] = req.CauseEffect
	}
	if req.ForeshadowingDetails != "" {
		updates["foreshadowing_details"] = req.ForeshadowingDetails
	}
	if req.VolumeID != nil {
		updates["volume_id"] = *req.VolumeID
	}

	document, err := h.documentService.Update(id, updates)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, document)
}

// Delete 删除文档
func (h *DocumentHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if _, ok := h.ensureDocumentOwner(c, id); !ok {
		return
	}

	if err := h.documentService.Delete(id); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// AddBookmark 添加书签
func (h *DocumentHandler) AddBookmark(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if _, ok := h.ensureDocumentOwner(c, id); !ok {
		return
	}

	var req AddBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("添加书签请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.documentService.AddBookmark(id, req.Title, req.Position, req.Note); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// RemoveBookmark 移除书签
func (h *DocumentHandler) RemoveBookmark(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if _, ok := h.ensureDocumentOwner(c, id); !ok {
		return
	}

	index, err := strconv.Atoi(c.Param("index"))
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.documentService.RemoveBookmark(id, index); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// LinkEntity 关联实体
func (h *DocumentHandler) LinkEntity(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if _, ok := h.ensureDocumentOwner(c, id); !ok {
		return
	}

	var req LinkEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("关联实体请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	refType := req.RefType
	if refType == "" {
		refType = "mention"
	}

	if err := h.documentService.LinkEntity(id, req.EntityID, refType, req.Metadata); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// UnlinkEntity 取消关联实体
func (h *DocumentHandler) UnlinkEntity(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}
	if _, ok := h.ensureDocumentOwner(c, id); !ok {
		return
	}

	entityID, err := parseUintParam(c, "entity_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.documentService.UnlinkEntity(id, entityID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}
