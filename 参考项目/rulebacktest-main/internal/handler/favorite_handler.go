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
	favoriteHandlerInstance *FavoriteHandler
	favoriteHandlerOnce     sync.Once
)

// FavoriteHandler 收藏HTTP处理器
type FavoriteHandler struct {
	service *service.FavoriteService
}

// NewFavoriteHandler 创建FavoriteHandler实例
func NewFavoriteHandler(svc *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{service: svc}
}

// GetFavoriteHandler 获取FavoriteHandler单例
func GetFavoriteHandler() *FavoriteHandler {
	favoriteHandlerOnce.Do(func() {
		favoriteHandlerInstance = &FavoriteHandler{
			service: service.GetFavoriteService(),
		}
	})
	return favoriteHandlerInstance
}

// Add 添加收藏
func (h *FavoriteHandler) Add(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	var req model.FavoriteAddReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	if err := h.service.Add(userID, &req); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// Remove 取消收藏
func (h *FavoriteHandler) Remove(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 32)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的商品ID")
		return
	}

	if err := h.service.Remove(userID, uint(productID)); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// List 获取收藏列表
func (h *FavoriteHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	favorites, total, err := h.service.List(userID, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithPage(c, favorites, total, page, pageSize)
}

// Check 检查是否已收藏
func (h *FavoriteHandler) Check(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 32)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的商品ID")
		return
	}

	isFavorite, err := h.service.IsFavorite(userID, uint(productID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, gin.H{"is_favorite": isFavorite})
}

// handleError 统一处理错误
func (h *FavoriteHandler) handleError(c *gin.Context, err error) {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		response.Fail(c, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, apperrors.CodeInternalError, "服务器内部错误")
}
