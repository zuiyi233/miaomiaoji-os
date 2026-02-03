package handler

import (
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"rulebacktest/internal/model"
	"rulebacktest/internal/service"
	apperrors "rulebacktest/pkg/errors"
	"rulebacktest/pkg/response"
)

var (
	reviewHandlerInstance *ReviewHandler
	reviewHandlerOnce     sync.Once
)

// ReviewHandler 评价HTTP处理器
type ReviewHandler struct {
	service *service.ReviewService
}

// NewReviewHandler 创建ReviewHandler实例
func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: svc}
}

// GetReviewHandler 获取ReviewHandler单例
func GetReviewHandler() *ReviewHandler {
	reviewHandlerOnce.Do(func() {
		reviewHandlerInstance = &ReviewHandler{
			service: service.GetReviewService(),
		}
	})
	return reviewHandlerInstance
}

// Create 创建评价
func (h *ReviewHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	var req model.ReviewCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	review, err := h.service.Create(userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, review)
}

// ListByProduct 获取商品评价列表
func (h *ReviewHandler) ListByProduct(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的商品ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	reviews, total, err := h.service.ListByProduct(uint(productID), page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithPage(c, reviews, total, page, pageSize)
}

// ListByUser 获取用户评价列表
func (h *ReviewHandler) ListByUser(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	reviews, total, err := h.service.ListByUser(userID, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithPage(c, reviews, total, page, pageSize)
}

// GetProductRating 获取商品评分统计
func (h *ReviewHandler) GetProductRating(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的商品ID")
		return
	}

	avgRating, count, err := h.service.GetProductRating(uint(productID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, gin.H{
		"avg_rating":   avgRating,
		"review_count": count,
	})
}

// handleError 统一处理错误
func (h *ReviewHandler) handleError(c *gin.Context, err error) {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		response.Fail(c, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, apperrors.CodeInternalError, "服务器内部错误")
}
