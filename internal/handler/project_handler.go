package handler

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"novel-agent-os-backend/internal/middleware"
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	projectService service.ProjectService
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(projectService service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Title         string                 `json:"title" binding:"required,max=200"`
	Genre         string                 `json:"genre" binding:"omitempty,max=50"`
	Tags          []string               `json:"tags"`
	CoreConflict  string                 `json:"core_conflict"`
	CharacterArc  string                 `json:"character_arc"`
	UltimateValue string                 `json:"ultimate_value"`
	WorldRules    string                 `json:"world_rules"`
	AISettings    map[string]interface{} `json:"ai_settings"`
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	Title         string                 `json:"title" binding:"omitempty,max=200"`
	Genre         string                 `json:"genre" binding:"omitempty,max=50"`
	Tags          []string               `json:"tags"`
	CoreConflict  string                 `json:"core_conflict"`
	CharacterArc  string                 `json:"character_arc"`
	UltimateValue string                 `json:"ultimate_value"`
	WorldRules    string                 `json:"world_rules"`
	AISettings    map[string]interface{} `json:"ai_settings"`
}

// ProjectResponse 项目响应
type ProjectResponse struct {
	ID            uint                   `json:"id"`
	Title         string                 `json:"title"`
	Genre         string                 `json:"genre"`
	Tags          []string               `json:"tags"`
	CoreConflict  string                 `json:"core_conflict"`
	CharacterArc  string                 `json:"character_arc"`
	UltimateValue string                 `json:"ultimate_value"`
	WorldRules    string                 `json:"world_rules"`
	AISettings    map[string]interface{} `json:"ai_settings"`
	UserID        uint                   `json:"user_id"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// ListProjectsRequest 项目列表请求
type ListProjectsRequest struct {
	Page int `form:"page" binding:"min=1"`
	Size int `form:"size" binding:"min=1,max=100"`
}

// Create 创建项目
func (h *ProjectHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	project, err := h.projectService.Create(
		userID,
		req.Title,
		req.Genre,
		req.Tags,
		req.CoreConflict,
		req.CharacterArc,
		req.UltimateValue,
		req.WorldRules,
		req.AISettings,
	)
	if err != nil {
		logger.Error("Create project failed", logger.Err(err), logger.Uint("user_id", userID))
		response.Fail(c, errors.CodeDatabaseError, "创建项目失败")
		return
	}

	response.SuccessWithData(c, h.toProjectResponse(project))
}

// GetByID 获取项目详情
func (h *ProjectHandler) GetByID(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "无效的项目ID")
		return
	}

	project, err := h.projectService.GetByID(id)
	if err != nil {
		logger.Error("Get project failed", logger.Err(err), logger.Uint("project_id", id))
		response.Fail(c, errors.CodeNotFound, "项目不存在")
		return
	}

	response.SuccessWithData(c, h.toProjectResponse(project))
}

// List 获取项目列表
func (h *ProjectHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	var req ListProjectsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 10
	}

	projects, total, err := h.projectService.ListByUserID(userID, req.Page, req.Size)
	if err != nil {
		logger.Error("List projects failed", logger.Err(err), logger.Uint("user_id", userID))
		response.Fail(c, errors.CodeDatabaseError, "获取项目列表失败")
		return
	}

	var list []ProjectResponse
	for _, project := range projects {
		list = append(list, *h.toProjectResponse(project))
	}

	response.SuccessWithPage(c, list, total, req.Page, req.Size)
}

// Update 更新项目
func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "无效的项目ID")
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Genre != "" {
		updates["genre"] = req.Genre
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}
	if req.CoreConflict != "" {
		updates["core_conflict"] = req.CoreConflict
	}
	if req.CharacterArc != "" {
		updates["character_arc"] = req.CharacterArc
	}
	if req.UltimateValue != "" {
		updates["ultimate_value"] = req.UltimateValue
	}
	if req.WorldRules != "" {
		updates["world_rules"] = req.WorldRules
	}
	if req.AISettings != nil {
		updates["ai_settings"] = req.AISettings
	}

	project, err := h.projectService.Update(id, updates)
	if err != nil {
		if err.Error() == "project not found" {
			response.Fail(c, errors.CodeNotFound, "项目不存在")
			return
		}
		logger.Error("Update project failed", logger.Err(err), logger.Uint("project_id", id))
		response.Fail(c, errors.CodeDatabaseError, "更新项目失败")
		return
	}

	response.SuccessWithData(c, h.toProjectResponse(project))
}

// Delete 删除项目
func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "无效的项目ID")
		return
	}

	if err := h.projectService.Delete(id); err != nil {
		if err.Error() == "project not found" {
			response.Fail(c, errors.CodeNotFound, "项目不存在")
			return
		}
		logger.Error("Delete project failed", logger.Err(err), logger.Uint("project_id", id))
		response.Fail(c, errors.CodeDatabaseError, "删除项目失败")
		return
	}

	response.Success(c)
}

// Export 导出项目
func (h *ProjectHandler) Export(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "无效的项目ID")
		return
	}

	project, err := h.projectService.GetByIDWithDetails(id)
	if err != nil {
		logger.Error("Get project failed", logger.Err(err), logger.Uint("project_id", id))
		response.Fail(c, errors.CodeNotFound, "项目不存在")
		return
	}

	// TODO: 导出项目为JSON格式
	response.SuccessWithData(c, h.toProjectResponse(project))
}

// toProjectResponse 转换为项目响应
func (h *ProjectHandler) toProjectResponse(project *model.Project) *ProjectResponse {
	// 解析标签
	var tags []string
	if len(project.Tags) > 0 {
		json.Unmarshal(project.Tags, &tags)
	}

	// 解析AI设置
	var aiSettings map[string]interface{}
	if len(project.AISettings) > 0 {
		json.Unmarshal(project.AISettings, &aiSettings)
	}

	return &ProjectResponse{
		ID:            project.ID,
		Title:         project.Title,
		Genre:         project.Genre,
		Tags:          tags,
		CoreConflict:  project.CoreConflict,
		CharacterArc:  project.CharacterArc,
		UltimateValue: project.UltimateValue,
		WorldRules:    project.WorldRules,
		AISettings:    aiSettings,
		UserID:        project.UserID,
		CreatedAt:     project.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     project.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}


