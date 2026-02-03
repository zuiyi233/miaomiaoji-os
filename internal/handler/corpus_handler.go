package handler

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CorpusHandler struct {
	corpusService service.CorpusService
}

func NewCorpusHandler(corpusService service.CorpusService) *CorpusHandler {
	return &CorpusHandler{
		corpusService: corpusService,
	}
}

type CreateCorpusStoryRequest struct {
	Title    string `json:"title" binding:"required"`
	Genre    string `json:"genre" binding:"required"`
	FilePath string `json:"file_path" binding:"required"`
}

type UpdateCorpusStoryRequest struct {
	Title    string `json:"title"`
	Genre    string `json:"genre"`
	FilePath string `json:"file_path"`
}

func (h *CorpusHandler) CreateStory(c *gin.Context) {
	var req CreateCorpusStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	story := &model.CorpusStory{
		Title:    req.Title,
		Genre:    req.Genre,
		FilePath: req.FilePath,
	}

	if err := h.corpusService.CreateStory(story); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to create corpus story")
		return
	}

	response.SuccessWithData(c, story)
}

func (h *CorpusHandler) GetStory(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid story ID")
		return
	}

	story, err := h.corpusService.GetStory(id)
	if err != nil {
		response.Fail(c, errors.CodeNotFound, "Corpus story not found")
		return
	}

	response.SuccessWithData(c, story)
}

func (h *CorpusHandler) UpdateStory(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid story ID")
		return
	}

	var req UpdateCorpusStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	story, err := h.corpusService.GetStory(id)
	if err != nil {
		response.Fail(c, errors.CodeNotFound, "Corpus story not found")
		return
	}

	if req.Title != "" {
		story.Title = req.Title
	}
	if req.Genre != "" {
		story.Genre = req.Genre
	}
	if req.FilePath != "" {
		story.FilePath = req.FilePath
	}

	if err := h.corpusService.UpdateStory(story); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to update corpus story")
		return
	}

	response.SuccessWithData(c, story)
}

func (h *CorpusHandler) DeleteStory(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid story ID")
		return
	}

	if err := h.corpusService.DeleteStory(id); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to delete corpus story")
		return
	}

	response.Success(c)
}

func (h *CorpusHandler) ListStories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	stories, total, err := h.corpusService.ListStories(page, pageSize)
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to list corpus stories")
		return
	}

	response.SuccessWithData(c, gin.H{
		"stories":   stories,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *CorpusHandler) ListStoriesByGenre(c *gin.Context) {
	genre := c.Query("genre")
	if genre == "" {
		response.Fail(c, errors.CodeInvalidParams, "genre is required")
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

	stories, total, err := h.corpusService.ListStoriesByGenre(genre, page, pageSize)
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to list corpus stories by genre")
		return
	}

	response.SuccessWithData(c, gin.H{
		"stories":   stories,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"genre":     genre,
	})
}

func (h *CorpusHandler) SearchStories(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.Fail(c, errors.CodeInvalidParams, "keyword is required")
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

	stories, total, err := h.corpusService.SearchStories(keyword, page, pageSize)
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to search corpus stories")
		return
	}

	response.SuccessWithData(c, gin.H{
		"stories":   stories,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"keyword":   keyword,
	})
}
