package handler

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileService    service.FileService
	projectService service.ProjectService
}

func NewFileHandler(fileService service.FileService, projectService service.ProjectService) *FileHandler {
	return &FileHandler{
		fileService:    fileService,
		projectService: projectService,
	}
}

func (h *FileHandler) ensureProjectOwner(c *gin.Context, projectID uint) bool {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "Unauthorized")
		return false
	}
	project, err := h.projectService.GetByID(projectID)
	if err != nil {
		response.Fail(c, errors.CodeNotFound, "Project not found")
		return false
	}
	if project.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
		return false
	}
	return true
}

func (h *FileHandler) ensureFileOwner(c *gin.Context, fileID uint) (*model.File, bool) {
	file, err := h.fileService.GetFile(fileID)
	if err != nil {
		response.Fail(c, errors.CodeFileNotFound, "File not found")
		return nil, false
	}

	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "Unauthorized")
		return nil, false
	}

	// 关联项目的文件，以项目 owner 权限为准
	if file.ProjectID != nil {
		if !h.ensureProjectOwner(c, *file.ProjectID) {
			return nil, false
		}
		return file, true
	}

	// 非项目文件，保持按 userID 校验
	if file.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
		return nil, false
	}
	return file, true
}

type CreateFileRequest struct {
	FileName    string `json:"file_name" binding:"required"`
	FileType    string `json:"file_type" binding:"required"`
	ContentType string `json:"content_type"`
	StorageKey  string `json:"storage_key"`
	ProjectID   *uint  `json:"project_id"`
}

type UpdateFileRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

func (h *FileHandler) CreateFile(c *gin.Context) {
	var req CreateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "Unauthorized")
		return
	}

	file := &model.File{
		FileName:    req.FileName,
		FileType:    req.FileType,
		ContentType: req.ContentType,
		StorageKey:  req.StorageKey,
		UserID:      userID,
		ProjectID:   req.ProjectID,
	}
	if req.ProjectID != nil {
		if !h.ensureProjectOwner(c, *req.ProjectID) {
			return
		}
	}

	if err := h.fileService.CreateFile(file, c.Request.Body); err != nil {
		response.Fail(c, errors.CodeFileUploadFailed, "Failed to create file")
		return
	}

	response.SuccessWithData(c, file)
}

func (h *FileHandler) GetFile(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid file ID")
		return
	}

	file, ok := h.ensureFileOwner(c, id)
	if !ok {
		return
	}

	response.SuccessWithData(c, file)
}

func (h *FileHandler) UpdateFile(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid file ID")
		return
	}

	var req UpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	file, ok := h.ensureFileOwner(c, id)
	if !ok {
		return
	}

	if req.FileName != "" {
		file.FileName = req.FileName
	}
	if req.ContentType != "" {
		file.ContentType = req.ContentType
	}

	if err := h.fileService.UpdateFile(file); err != nil {
		response.Fail(c, errors.CodeFileError, "Failed to update file")
		return
	}

	response.SuccessWithData(c, file)
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid file ID")
		return
	}

	if _, ok := h.ensureFileOwner(c, id); !ok {
		return
	}

	if err := h.fileService.DeleteFile(id); err != nil {
		response.Fail(c, errors.CodeFileError, "Failed to delete file")
		return
	}

	response.Success(c)
}

func (h *FileHandler) ListFiles(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "Unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	files, total, err := h.fileService.ListFiles(userID, page, pageSize)
	if err != nil {
		response.Fail(c, errors.CodeFileError, "Failed to list files")
		return
	}

	response.SuccessWithData(c, gin.H{
		"files":     files,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *FileHandler) ListFilesByProject(c *gin.Context) {
	projectID, err := parseUintParam(c, "project_id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid project ID")
		return
	}
	if !h.ensureProjectOwner(c, projectID) {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	fileType := c.Query("file_type")
	if fileType == "" {
		fileType = "backup"
	}

	files, total, err := h.fileService.ListFilesByProject(projectID, page, pageSize)
	if err != nil {
		response.Fail(c, errors.CodeFileError, "Failed to list files")
		return
	}
	if fileType != "" {
		filtered := make([]*model.File, 0, len(files))
		for _, item := range files {
			if item.FileType == fileType {
				filtered = append(filtered, item)
			}
		}
		files = filtered
		if int64(len(files)) < total {
			total = int64(len(files))
		}
	}

	response.SuccessWithData(c, gin.H{
		"files":     files,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *FileHandler) DownloadFile(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid file ID")
		return
	}

	file, ok := h.ensureFileOwner(c, id)
	if !ok {
		return
	}

	reader, err := h.fileService.DownloadFile(id)
	if err != nil {
		response.Fail(c, errors.CodeFileDownloadFailed, "Failed to download file")
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", "attachment; filename="+file.FileName)
	c.Header("Content-Type", file.ContentType)
	c.DataFromReader(200, file.SizeBytes, file.ContentType, reader, nil)
}
