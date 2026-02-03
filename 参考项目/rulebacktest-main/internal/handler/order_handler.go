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
	orderHandlerInstance *OrderHandler
	orderHandlerOnce     sync.Once
)

// OrderHandler 订单HTTP处理器
type OrderHandler struct {
	service *service.OrderService
}

// NewOrderHandler 创建OrderHandler实例
func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{service: svc}
}

// GetOrderHandler 获取OrderHandler单例
func GetOrderHandler() *OrderHandler {
	orderHandlerOnce.Do(func() {
		orderHandlerInstance = &OrderHandler{
			service: service.GetOrderService(),
		}
	})
	return orderHandlerInstance
}

// Create 创建订单
func (h *OrderHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	var req model.OrderCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	order, err := h.service.Create(userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, order)
}

// GetByID 获取订单详情
func (h *OrderHandler) GetByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	orderID, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的订单ID")
		return
	}

	order, err := h.service.GetByID(userID, orderID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, order)
}

// List 获取订单列表
func (h *OrderHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	var req model.OrderListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	orders, total, err := h.service.List(userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	req.SetDefaults()
	response.SuccessWithPage(c, orders, total, req.Page, req.PageSize)
}

// Cancel 取消订单
func (h *OrderHandler) Cancel(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	orderID, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的订单ID")
		return
	}

	if err := h.service.Cancel(userID, orderID); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// Pay 支付订单
func (h *OrderHandler) Pay(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	orderID, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的订单ID")
		return
	}

	if err := h.service.Pay(userID, orderID); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// ConfirmReceipt 确认收货
func (h *OrderHandler) ConfirmReceipt(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	orderID, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的订单ID")
		return
	}

	if err := h.service.ConfirmReceipt(userID, orderID); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// parseIDParam 解析ID参数
func (h *OrderHandler) parseIDParam(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// handleError 统一处理错误
func (h *OrderHandler) handleError(c *gin.Context, err error) {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		response.Fail(c, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, apperrors.CodeInternalError, "服务器内部错误")
}
