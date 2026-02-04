package handler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"
)

// RedemptionCodeHandler 兑换码处理器
type RedemptionCodeHandler struct {
	codeService service.RedemptionCodeService
	userService service.UserService
}

// NewRedemptionCodeHandler 创建兑换码处理器
func NewRedemptionCodeHandler(codeService service.RedemptionCodeService, userService service.UserService) *RedemptionCodeHandler {
	return &RedemptionCodeHandler{
		codeService: codeService,
		userService: userService,
	}
}

// RedeemRequest 兑换请求
type RedeemRequest struct {
	RequestID        string                 `json:"request_id" binding:"required,max=64"`
	IdempotencyKey   string                 `json:"idempotency_key" binding:"omitempty,max=64"`
	Code             string                 `json:"code" binding:"required"`
	DeviceID         string                 `json:"device_id" binding:"omitempty,max=100"`
	ClientTime       string                 `json:"client_time" binding:"omitempty,max=30"`
	AppID            string                 `json:"app_id" binding:"omitempty,max=50"`
	Platform         string                 `json:"platform" binding:"omitempty,max=30"`
	AppVersion       string                 `json:"app_version" binding:"omitempty,max=30"`
	ResultStatus     string                 `json:"result_status" binding:"omitempty,max=20"`
	ResultErrorCode  string                 `json:"result_error_code" binding:"omitempty,max=50"`
	EntitlementDelta map[string]interface{} `json:"entitlement_delta"`
}

// RedeemResponse 兑换响应
type RedeemResponse struct {
	Code          string `json:"code"`
	DurationDays  int    `json:"duration_days"`
	AIAccessUntil string `json:"ai_access_until"`
	UsedCount     int    `json:"used_count"`
	Status        string `json:"status"`
}

// GenerateCodesRequest 批量生成请求
type GenerateCodesRequest struct {
	Prefix       string   `json:"prefix" binding:"omitempty,max=20"`
	Length       int      `json:"length" binding:"required,min=4,max=32"`
	Count        int      `json:"count" binding:"required,min=1,max=1000"`
	ValidityDays int      `json:"validity_days" binding:"required,min=1,max=3650"`
	MaxUses      int      `json:"max_uses" binding:"required,min=1,max=100"`
	CharType     string   `json:"char_type" binding:"omitempty,oneof=alphanum alpha num"`
	Tags         []string `json:"tags"`
	Note         string   `json:"note"`
	Source       string   `json:"source"`
}

// ListCodesRequest 列表请求
type ListCodesRequest struct {
	Status string `form:"status"`
	Search string `form:"search"`
	Page   int    `form:"page"`
	Size   int    `form:"size"`
	Sort   string `form:"sort"`
}

// BatchUpdateRequest 批量更新请求
type BatchUpdateRequest struct {
	Codes  []string `json:"codes" binding:"required"`
	Action string   `json:"action" binding:"required,oneof=disable enable delete renew"`
	Value  int      `json:"value"`
}

// ExportCodes 导出兑换码
func (h *RedemptionCodeHandler) ExportCodes(c *gin.Context) {
	var req ListCodesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}
	if req.Sort != "" && req.Sort != "asc" && req.Sort != "desc" {
		response.Fail(c, errors.CodeInvalidParams, "无效排序参数")
		return
	}
	if req.Status == "rewards" {
		req.Status = "all"
		req.Search = "points_exchange"
	}

	list, _, err := h.codeService.List(repository.RedemptionCodeFilter{
		Status: req.Status,
		Search: req.Search,
		Page:   1,
		Size:   10000,
		Sort:   req.Sort,
	})
	if err != nil {
		response.Fail(c, errors.CodeDatabaseError, "导出失败")
		return
	}

	var sb strings.Builder
	sb.WriteString("Code,Status,Source,Creator,Created At,Expires At,Max Uses,Used Count,Note,Tags\n")
	for _, item := range list {
		expire := ""
		if item.ExpiresAt != nil {
			expire = item.ExpiresAt.Format("2006-01-02")
		}
		tags := []string{}
		if len(item.Tags) > 0 {
			_ = json.Unmarshal(item.Tags, &tags)
		}
		sb.WriteString(strings.Join([]string{
			item.Code,
			item.Status,
			item.Source,
			fmt.Sprintf("%d", item.CreatedBy),
			item.CreatedAt.Format("2006-01-02"),
			expire,
			fmt.Sprintf("%d", item.MaxUses),
			fmt.Sprintf("%d", item.UsedCount),
			"\"" + strings.ReplaceAll(item.Note, "\"", "\"\"") + "\"",
			"\"" + strings.ReplaceAll(strings.Join(tags, ";"), "\"", "\"\"") + "\"",
		}, ","))
		sb.WriteString("\n")
	}

	c.Header("Content-Disposition", "attachment; filename=redemption_codes.csv")
	c.Data(200, "text/csv", []byte(sb.String()))
}

// GenerateCodes 批量生成兑换码
func (h *RedemptionCodeHandler) GenerateCodes(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	var req GenerateCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	charType := req.CharType
	if charType == "" {
		charType = "alphanum"
	}
	items, err := h.codeService.Generate(service.GenerateCodesPayload{
		Prefix:       strings.ToUpper(req.Prefix),
		Length:       req.Length,
		Count:        req.Count,
		ValidityDays: req.ValidityDays,
		MaxUses:      req.MaxUses,
		CharType:     charType,
		Tags:         req.Tags,
		Note:         req.Note,
		Source:       req.Source,
	}, userID)
	if err != nil {
		response.Fail(c, errors.CodeDatabaseError, "生成失败")
		return
	}

	response.SuccessWithData(c, gin.H{
		"list": items,
	})
}

// ListCodes 获取兑换码列表
func (h *RedemptionCodeHandler) ListCodes(c *gin.Context) {
	var req ListCodesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}
	if req.Sort != "" && req.Sort != "asc" && req.Sort != "desc" {
		response.Fail(c, errors.CodeInvalidParams, "无效排序参数")
		return
	}
	if req.Status == "rewards" {
		req.Status = "all"
		req.Search = "points_exchange"
	}

	items, total, err := h.codeService.List(repository.RedemptionCodeFilter{
		Status: req.Status,
		Search: req.Search,
		Page:   req.Page,
		Size:   req.Size,
		Sort:   req.Sort,
	})
	if err != nil {
		response.Fail(c, errors.CodeDatabaseError, "获取失败")
		return
	}

	response.SuccessWithPage(c, items, total, req.Page, req.Size)
}

// BatchUpdateCodes 批量更新兑换码
func (h *RedemptionCodeHandler) BatchUpdateCodes(c *gin.Context) {
	var req BatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}
	if req.Action == "renew" && req.Value <= 0 {
		response.Fail(c, errors.CodeInvalidParams, "续期天数无效")
		return
	}

	if err := h.codeService.UpdateStatus(req.Codes, req.Action, req.Value); err != nil {
		response.Fail(c, errors.CodeDatabaseError, "更新失败")
		return
	}

	response.Success(c)
}

// Redeem 兑换码验证
func (h *RedemptionCodeHandler) Redeem(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Fail(c, errors.CodeUnauthorized, "未登录")
		return
	}

	var req RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.CodeInvalidParams, err.Error())
		return
	}

	code := strings.TrimSpace(req.Code)
	item, durationDays, err := h.codeService.Redeem(service.RedeemPayload{
		RequestID:        req.RequestID,
		IdempotencyKey:   req.IdempotencyKey,
		UserID:           userID,
		DeviceID:         req.DeviceID,
		RedeemCode:       code,
		ClientTime:       req.ClientTime,
		ServerTime:       time.Now().UTC().Format(time.RFC3339),
		AppID:            req.AppID,
		Platform:         req.Platform,
		AppVersion:       req.AppVersion,
		ResultStatus:     req.ResultStatus,
		ResultErrorCode:  req.ResultErrorCode,
		EntitlementDelta: req.EntitlementDelta,
	})
	if err != nil {
		response.Fail(c, errors.CodeValidationError, err.Error())
		return
	}
	if durationDays <= 0 {
		response.Fail(c, errors.CodeValidationError, "兑换码未配置有效期")
		return
	}

	user, err := h.userService.UpdateAIAccess(userID, durationDays, code)
	if err != nil {
		logger.Error("Update AI access failed", logger.Err(err))
		response.Fail(c, errors.CodeDatabaseError, "更新权限失败")
		return
	}

	response.SuccessWithData(c, RedeemResponse{
		Code:          item.Code,
		DurationDays:  durationDays,
		AIAccessUntil: formatTimeRFC3339(user.AIAccessUntil),
		UsedCount:     item.UsedCount,
		Status:        item.Status,
	})
}
