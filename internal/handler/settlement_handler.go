package handler

import (
	"encoding/json"
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type SettlementHandler struct {
	settlementService service.SettlementService
}

func NewSettlementHandler(settlementService service.SettlementService) *SettlementHandler {
	return &SettlementHandler{
		settlementService: settlementService,
	}
}

type CreateSettlementRequest struct {
	WorldID     string                 `json:"world_id" binding:"required"`
	ChapterID   string                 `json:"chapter_id" binding:"required"`
	LoopStage   string                 `json:"loop_stage" binding:"required"`
	PointsDelta int                    `json:"points_delta" binding:"required"`
	Payload     map[string]interface{} `json:"payload"`
}

type UpdateSettlementRequest struct {
	PointsDelta int                    `json:"points_delta"`
	Payload     map[string]interface{} `json:"payload"`
}

func (h *SettlementHandler) CreateEntry(c *gin.Context) {
	var req CreateSettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "Unauthorized")
		return
	}

	var payload datatypes.JSON
	if req.Payload != nil {
		b, err := json.Marshal(req.Payload)
		if err != nil {
			response.Fail(c, errors.CodeInvalidParams, "Invalid payload")
			return
		}
		payload = datatypes.JSON(b)
	}

	entry := &model.SettlementEntry{
		WorldID:     req.WorldID,
		ChapterID:   req.ChapterID,
		LoopStage:   req.LoopStage,
		PointsDelta: req.PointsDelta,
		Payload:     payload,
		UserID:      userID,
	}

	if err := h.settlementService.CreateEntry(entry); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to create settlement entry")
		return
	}

	response.SuccessWithData(c, entry)
}

func (h *SettlementHandler) GetEntry(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid entry ID")
		return
	}

	entry, err := h.settlementService.GetEntry(id)
	if err != nil {
		response.Fail(c, errors.CodeNotFound, "Settlement entry not found")
		return
	}

	userID := getUserIDFromContext(c)
	if entry.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
		return
	}

	response.SuccessWithData(c, entry)
}

func (h *SettlementHandler) UpdateEntry(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid entry ID")
		return
	}

	var req UpdateSettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid request body")
		return
	}

	entry, err := h.settlementService.GetEntry(id)
	if err != nil {
		response.Fail(c, errors.CodeNotFound, "Settlement entry not found")
		return
	}

	userID := getUserIDFromContext(c)
	if entry.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
		return
	}

	if req.PointsDelta != 0 {
		entry.PointsDelta = req.PointsDelta
	}
	if req.Payload != nil {
		b, err := json.Marshal(req.Payload)
		if err != nil {
			response.Fail(c, errors.CodeInvalidParams, "Invalid payload")
			return
		}
		entry.Payload = datatypes.JSON(b)
	}

	if err := h.settlementService.UpdateEntry(entry); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to update settlement entry")
		return
	}

	response.SuccessWithData(c, entry)
}

func (h *SettlementHandler) DeleteEntry(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.Fail(c, errors.CodeInvalidParams, "Invalid entry ID")
		return
	}

	entry, err := h.settlementService.GetEntry(id)
	if err != nil {
		response.Fail(c, errors.CodeNotFound, "Settlement entry not found")
		return
	}

	userID := getUserIDFromContext(c)
	if entry.UserID != userID {
		response.Fail(c, errors.CodeForbidden, "Access denied")
		return
	}

	if err := h.settlementService.DeleteEntry(id); err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to delete settlement entry")
		return
	}

	response.Success(c)
}

func (h *SettlementHandler) ListEntries(c *gin.Context) {
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

	entries, total, err := h.settlementService.ListEntries(userID, page, pageSize)
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to list settlement entries")
		return
	}

	response.SuccessWithData(c, gin.H{
		"entries":   entries,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *SettlementHandler) FilterEntries(c *gin.Context) {
	worldID := c.Query("world_id")
	chapterID := c.Query("chapter_id")
	loopStage := c.Query("loop_stage")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	entries, total, err := h.settlementService.FilterEntries(worldID, chapterID, loopStage, page, pageSize)
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to filter settlement entries")
		return
	}

	response.SuccessWithData(c, gin.H{
		"entries":   entries,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"filters": gin.H{
			"world_id":   worldID,
			"chapter_id": chapterID,
			"loop_stage": loopStage,
		},
	})
}

func (h *SettlementHandler) GetTotalPoints(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "Unauthorized")
		return
	}

	total, err := h.settlementService.GetTotalPoints(userID)
	if err != nil {
		response.Fail(c, errors.CodeInternalError, "Failed to get total points")
		return
	}

	response.SuccessWithData(c, gin.H{
		"total_points": total,
	})
}
