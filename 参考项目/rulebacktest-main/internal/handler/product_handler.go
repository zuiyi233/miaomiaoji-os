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
	productHandlerInstance *ProductHandler
	productHandlerOnce     sync.Once
)

// ProductHandler 商品HTTP处理器
type ProductHandler struct {
	service *service.ProductService
}

// NewProductHandler 创建ProductHandler实例
func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{service: svc}
}

// GetProductHandler 获取ProductHandler单例
func GetProductHandler() *ProductHandler {
	productHandlerOnce.Do(func() {
		productHandlerInstance = &ProductHandler{
			service: service.GetProductService(),
		}
	})
	return productHandlerInstance
}

// Create 创建商品
func (h *ProductHandler) Create(c *gin.Context) {
	var req model.ProductCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	product, err := h.service.Create(&req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, product)
}

// GetByID 获取商品详情
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的商品ID")
		return
	}

	product, err := h.service.GetByID(id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, product)
}

// List 获取商品列表
func (h *ProductHandler) List(c *gin.Context) {
	var req model.ProductListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	products, total, err := h.service.List(&req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	req.SetDefaults()
	response.SuccessWithPage(c, products, total, req.Page, req.PageSize)
}

// Update 更新商品
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的商品ID")
		return
	}

	var req model.ProductUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	product, err := h.service.Update(id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, product)
}

// Delete 删除商品
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的商品ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// parseIDParam 解析ID参数
func (h *ProductHandler) parseIDParam(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// handleError 统一处理错误
func (h *ProductHandler) handleError(c *gin.Context, err error) {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		response.Fail(c, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, apperrors.CodeInternalError, "服务器内部错误")
}
