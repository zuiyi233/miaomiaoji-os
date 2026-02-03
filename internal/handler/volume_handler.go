package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"
)

// VolumeHandler 卷处理器
type VolumeHandler struct {
	volumeService service.VolumeService
}

// NewVolumeHandler 创建卷处理器
func NewVolumeHandler(volumeService service.VolumeService) *VolumeHandler {
	return &VolumeHandler{
		volumeService: volumeService,
	}
}

// CreateVolumeRequest 创建卷请求
type CreateVolumeRequest struct {
	Title                  string `json:"title" binding:"required,max=200"`
	OrderIndex             int    `json:"order_index" binding:"required"`
	Theme                  string `json:"theme"`
	CoreGoal               string `json:"core_goal"`
	Boundaries             string `json:"boundaries"`
	ChapterLinkageLogic    string `json:"chapter_linkage_logic"`
	VolumeSpecificSettings string `json:"volume_specific_settings"`
	PlotRoadmap            string `json:"plot_roadmap"`
}

// UpdateVolumeRequest 更新卷请求
type UpdateVolumeRequest struct {
	Title                  string `json:"title"`
	OrderIndex             *int   `json:"order_index,omitempty"`
	Theme                  string `json:"theme"`
	CoreGoal               string `json:"core_goal"`
	Boundaries             string `json:"boundaries"`
	ChapterLinkageLogic    string `json:"chapter_linkage_logic"`
	VolumeSpecificSettings string `json:"volume_specific_settings"`
	PlotRoadmap            string `json:"plot_roadmap"`
}

// ReorderVolumesRequest 重新排序卷请求
type ReorderVolumesRequest struct {
	VolumeIDs []uint `json:"volume_ids" binding:"required"`
}

// Create 创建卷
func (h *VolumeHandler) Create(c *gin.Context) {
	projectID, err := parseUintParam(c, "project_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req CreateVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("创建卷请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	volume, err := h.volumeService.Create(
		projectID,
		req.Title,
		req.OrderIndex,
		req.Theme,
		req.CoreGoal,
		req.Boundaries,
		req.ChapterLinkageLogic,
		req.VolumeSpecificSettings,
		req.PlotRoadmap,
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, volume)
}

// List 获取卷列表
func (h *VolumeHandler) List(c *gin.Context) {
	projectID, err := parseUintParam(c, "project_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
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

	volumes, total, err := h.volumeService.ListByProjectID(projectID, page, size)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPage(c, volumes, total, page, size)
}

// GetByID 根据ID获取卷
func (h *VolumeHandler) GetByID(c *gin.Context) {
	id, err := parseUintParam(c, "volume_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	volume, err := h.volumeService.GetByID(id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, volume)
}

// Update 更新卷
func (h *VolumeHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "volume_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req UpdateVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新卷请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.OrderIndex != nil {
		updates["order_index"] = *req.OrderIndex
	}
	if req.Theme != "" {
		updates["theme"] = req.Theme
	}
	if req.CoreGoal != "" {
		updates["core_goal"] = req.CoreGoal
	}
	if req.Boundaries != "" {
		updates["boundaries"] = req.Boundaries
	}
	if req.ChapterLinkageLogic != "" {
		updates["chapter_linkage_logic"] = req.ChapterLinkageLogic
	}
	if req.VolumeSpecificSettings != "" {
		updates["volume_specific_settings"] = req.VolumeSpecificSettings
	}
	if req.PlotRoadmap != "" {
		updates["plot_roadmap"] = req.PlotRoadmap
	}

	volume, err := h.volumeService.Update(id, updates)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, volume)
}

// Delete 删除卷
func (h *VolumeHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "volume_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.volumeService.Delete(id); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// Reorder 重新排序卷
func (h *VolumeHandler) Reorder(c *gin.Context) {
	projectID, err := parseUintParam(c, "project_id")
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req ReorderVolumesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("重新排序卷请求参数错误", logger.Err(err))
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.volumeService.ReorderVolumes(projectID, req.VolumeIDs); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}
