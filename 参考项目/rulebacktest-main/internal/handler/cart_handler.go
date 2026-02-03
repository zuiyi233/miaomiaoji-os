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
	cartHandlerInstance *CartHandler
	cartHandlerOnce     sync.Once
)

// CartHandler 购物车HTTP处理器
type CartHandler struct {
	service *service.CartService
}

// NewCartHandler 创建CartHandler实例
func NewCartHandler(svc *service.CartService) *CartHandler {
	return &CartHandler{service: svc}
}

// GetCartHandler 获取CartHandler单例
func GetCartHandler() *CartHandler {
	cartHandlerOnce.Do(func() {
		cartHandlerInstance = &CartHandler{
			service: service.GetCartService(),
		}
	})
	return cartHandlerInstance
}

// Add 添加商品到购物车
func (h *CartHandler) Add(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	var req model.CartAddReq
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

// List 获取购物车列表
func (h *CartHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	resp, err := h.service.List(userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, resp)
}

// Update 更新购物车数量
func (h *CartHandler) Update(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	cartID, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的购物车ID")
		return
	}

	var req model.CartUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	if err := h.service.Update(userID, cartID, &req); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// Delete 删除购物车项
func (h *CartHandler) Delete(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	cartID, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的购物车ID")
		return
	}

	if err := h.service.Delete(userID, cartID); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// Clear 清空购物车
func (h *CartHandler) Clear(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	if err := h.service.Clear(userID); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// parseIDParam 解析ID参数
func (h *CartHandler) parseIDParam(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// handleError 统一处理错误
func (h *CartHandler) handleError(c *gin.Context, err error) {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		response.Fail(c, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, apperrors.CodeInternalError, "服务器内部错误")
}
