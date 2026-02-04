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
	fileService service.FileService
}

func NewFileHandler(fileService service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
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

	file, err := h.fileService.GetFile(id)
	if err != nil {
		response.Fail(c, errors.CodeFileNotFound, "File not found")
		return
	}

	userID := getUserIDFromContext(c)
	if file.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
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

	file, err := h.fileService.GetFile(id)
	if err != nil {
		response.Fail(c, errors.CodeFileNotFound, "File not found")
		return
	}

	userID := getUserIDFromContext(c)
	if file.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
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

	file, err := h.fileService.GetFile(id)
	if err != nil {
		response.Fail(c, errors.CodeFileNotFound, "File not found")
		return
	}

	userID := getUserIDFromContext(c)
	if file.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
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

	file, err := h.fileService.GetFile(id)
	if err != nil {
		response.Fail(c, errors.CodeFileNotFound, "File not found")
		return
	}

	userID := getUserIDFromContext(c)
	if file.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
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
