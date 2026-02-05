package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	projectService service.ProjectService
	fileService    service.FileService
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(projectService service.ProjectService, fileService service.FileService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		fileService:    fileService,
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

// UpsertSnapshotRequest 本地项目快照同步请求
type UpsertSnapshotRequest struct {
	ExternalID string                 `json:"external_id" binding:"required,max=64"`
	Title      string                 `json:"title" binding:"omitempty,max=200"`
	AISettings map[string]interface{} `json:"ai_settings"`
	Snapshot   map[string]interface{} `json:"snapshot" binding:"required"`
}

// BackupSnapshotRequest 项目备份请求
type BackupSnapshotRequest struct {
	ExternalID string                 `json:"external_id" binding:"required,max=64"`
	Title      string                 `json:"title" binding:"omitempty,max=200"`
	AISettings map[string]interface{} `json:"ai_settings"`
	Snapshot   map[string]interface{} `json:"snapshot" binding:"required"`
}

// ProjectResponse 项目响应
type ProjectResponse struct {
	ID            uint                   `json:"id"`
	ExternalID    *string                `json:"external_id,omitempty"`
	Title         string                 `json:"title"`
	Genre         string                 `json:"genre"`
	Tags          []string               `json:"tags"`
	CoreConflict  string                 `json:"core_conflict"`
	CharacterArc  string                 `json:"character_arc"`
	UltimateValue string                 `json:"ultimate_value"`
	WorldRules    string                 `json:"world_rules"`
	AISettings    map[string]interface{} `json:"ai_settings"`
	Snapshot      map[string]interface{} `json:"snapshot,omitempty"`
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
	userID := getUserIDFromContext(c)
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
	id, err := parseUintParam(c, "project_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "无效的项目ID")
		return
	}

	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	project, err := h.projectService.GetByID(id)
	if err != nil {
		logger.Error("Get project failed", logger.Err(err), logger.Uint("project_id", id))
		response.Fail(c, errors.CodeNotFound, "项目不存在")
		return
	}
	if project.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "无权限访问")
		return
	}

	response.SuccessWithData(c, h.toProjectResponse(project))
}

// List 获取项目列表
func (h *ProjectHandler) List(c *gin.Context) {
	userID := getUserIDFromContext(c)
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
	id, err := parseUintParam(c, "project_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "无效的项目ID")
		return
	}

	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	project, err := h.projectService.GetByID(id)
	if err != nil {
		response.Fail(c, errors.CodeNotFound, "项目不存在")
		return
	}
	if project.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "无权限访问")
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

	project, err = h.projectService.Update(id, updates)
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

// UpsertSnapshot 本地项目快照同步
func (h *ProjectHandler) UpsertSnapshot(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	var req UpsertSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	project, err := h.projectService.CreateOrUpdateSnapshot(userID, req.ExternalID, req.Snapshot, req.Title, req.AISettings)
	if err != nil {
		logger.Error("Upsert project snapshot failed", logger.Err(err), logger.Uint("user_id", userID))
		response.Fail(c, errors.CodeDatabaseError, "保存项目快照失败")
		return
	}

	response.SuccessWithData(c, h.toProjectResponse(project))
}

// BackupSnapshot 备份项目到本地文件存储
func (h *ProjectHandler) BackupSnapshot(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	var req BackupSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	project, err := h.projectService.CreateOrUpdateSnapshot(userID, req.ExternalID, req.Snapshot, req.Title, req.AISettings)
	if err != nil {
		logger.Error("Backup project snapshot failed", logger.Err(err), logger.Uint("user_id", userID))
		response.Fail(c, errors.CodeDatabaseError, "保存项目快照失败")
		return
	}

	data, err := json.Marshal(req.Snapshot)
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "项目快照格式错误")
		return
	}

	backupName := fmt.Sprintf("project_%s_%s.json", req.ExternalID, time.Now().Format("20060102_150405"))
	storageKey := fmt.Sprintf("backups/%d/%s/%s", userID, req.ExternalID, backupName)
	file := &model.File{
		FileName:    backupName,
		FileType:    "backup",
		ContentType: "application/json",
		StorageKey:  storageKey,
		UserID:      userID,
		ProjectID:   &project.ID,
		SizeBytes:   int64(len(data)),
	}

	if err := h.fileService.CreateFile(file, bytes.NewReader(data)); err != nil {
		logger.Error("Backup file save failed", logger.Err(err), logger.Uint("user_id", userID))
		response.Fail(c, errors.CodeFileError, "备份写入失败")
		return
	}

	response.SuccessWithData(c, gin.H{
		"file_id":     file.ID,
		"file_name":   file.FileName,
		"storage_key": file.StorageKey,
		"project_id":  project.ID,
	})
}

// GetLatestBackup 获取项目最新备份文件信息
func (h *ProjectHandler) GetLatestBackup(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	projectID, err := parseUintParam(c, "project_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "无效的项目ID")
		return
	}

	project, err := h.projectService.GetByID(projectID)
	if err != nil {
		response.Fail(c, errors.CodeNotFound, "项目不存在")
		return
	}
	if project.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "无权限访问")
		return
	}

	file, err := h.fileService.GetLatestBackupByProject(projectID)
	if err != nil {
		response.Fail(c, errors.CodeNotFound, "未找到备份")
		return
	}

	response.SuccessWithData(c, file)
}

// Delete 删除项目
func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "project_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "无效的项目ID")
		return
	}

	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}
	project, err := h.projectService.GetByID(id)
	if err != nil {
		response.Fail(c, errors.CodeNotFound, "项目不存在")
		return
	}
	if project.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "无权限访问")
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
	id, err := parseUintParam(c, "project_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "无效的项目ID")
		return
	}

	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}
	project, err := h.projectService.GetByID(id)
	if err != nil {
		response.Fail(c, errors.CodeNotFound, "项目不存在")
		return
	}
	if project.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "无权限访问")
		return
	}

	export, err := h.projectService.ExportProject(id)
	if err != nil {
		logger.Error("Export project failed", logger.Err(err), logger.Uint("project_id", id))
		response.Fail(c, errors.CodeNotFound, "项目不存在")
		return
	}

	// 生成JSON导出
	response.SuccessWithData(c, export)
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

	// 解析快照
	var snapshot map[string]interface{}
	if len(project.Snapshot) > 0 {
		json.Unmarshal(project.Snapshot, &snapshot)
	}

	return &ProjectResponse{
		ID:            project.ID,
		ExternalID:    project.ExternalID,
		Title:         project.Title,
		Genre:         project.Genre,
		Tags:          tags,
		CoreConflict:  project.CoreConflict,
		CharacterArc:  project.CharacterArc,
		UltimateValue: project.UltimateValue,
		WorldRules:    project.WorldRules,
		AISettings:    aiSettings,
		Snapshot:      snapshot,
		UserID:        project.UserID,
		CreatedAt:     project.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     project.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
