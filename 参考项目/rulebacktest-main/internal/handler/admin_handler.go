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
	adminHandlerInstance *AdminHandler
	adminHandlerOnce     sync.Once
)

// AdminHandler 管理员HTTP处理器
type AdminHandler struct {
	orderService   *service.OrderService
	productService *service.ProductService
	userService    *service.UserService
}

// NewAdminHandler 创建AdminHandler实例
func NewAdminHandler(orderSvc *service.OrderService, productSvc *service.ProductService, userSvc *service.UserService) *AdminHandler {
	return &AdminHandler{
		orderService:   orderSvc,
		productService: productSvc,
		userService:    userSvc,
	}
}

// GetAdminHandler 获取AdminHandler单例
func GetAdminHandler() *AdminHandler {
	adminHandlerOnce.Do(func() {
		adminHandlerInstance = &AdminHandler{
			orderService:   service.GetOrderService(),
			productService: service.GetProductService(),
			userService:    service.GetUserService(),
		}
	})
	return adminHandlerInstance
}

// ListOrders 获取所有订单（管理员）
func (h *AdminHandler) ListOrders(c *gin.Context) {
	var req model.AdminOrderListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	orders, total, err := h.orderService.AdminList(&req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	req.SetDefaults()
	response.SuccessWithPage(c, orders, total, req.Page, req.PageSize)
}

// GetOrder 获取订单详情（管理员）
func (h *AdminHandler) GetOrder(c *gin.Context) {
	orderID, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的订单ID")
		return
	}

	order, err := h.orderService.AdminGetByID(orderID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, order)
}

// ShipOrder 订单发货
func (h *AdminHandler) ShipOrder(c *gin.Context) {
	orderID, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的订单ID")
		return
	}

	if err := h.orderService.Ship(orderID); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// CompleteOrder 订单完成
func (h *AdminHandler) CompleteOrder(c *gin.Context) {
	orderID, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的订单ID")
		return
	}

	if err := h.orderService.Complete(orderID); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// UpdateProductStock 更新商品库存
func (h *AdminHandler) UpdateProductStock(c *gin.Context) {
	productID, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的商品ID")
		return
	}

	var req struct {
		Stock int `json:"stock" binding:"min=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	if err := h.productService.UpdateStock(productID, req.Stock); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// parseIDParam 解析ID参数
func (h *AdminHandler) parseIDParam(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// handleError 统一处理错误
func (h *AdminHandler) handleError(c *gin.Context, err error) {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		response.Fail(c, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, apperrors.CodeInternalError, "服务器内部错误")
}
