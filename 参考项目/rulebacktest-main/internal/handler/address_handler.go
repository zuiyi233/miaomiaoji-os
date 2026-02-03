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
	addressHandlerInstance *AddressHandler
	addressHandlerOnce     sync.Once
)

// AddressHandler 地址HTTP处理器
type AddressHandler struct {
	service *service.AddressService
}

// NewAddressHandler 创建AddressHandler实例
func NewAddressHandler(svc *service.AddressService) *AddressHandler {
	return &AddressHandler{service: svc}
}

// GetAddressHandler 获取AddressHandler单例
func GetAddressHandler() *AddressHandler {
	addressHandlerOnce.Do(func() {
		addressHandlerInstance = &AddressHandler{
			service: service.GetAddressService(),
		}
	})
	return addressHandlerInstance
}

// Create 创建地址
func (h *AddressHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	var req model.AddressCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	address, err := h.service.Create(userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, address)
}

// GetByID 获取地址详情
func (h *AddressHandler) GetByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	addressID, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的地址ID")
		return
	}

	address, err := h.service.GetByID(userID, addressID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, address)
}

// List 获取地址列表
func (h *AddressHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	addresses, err := h.service.List(userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, addresses)
}

// GetDefault 获取默认地址
func (h *AddressHandler) GetDefault(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	address, err := h.service.GetDefault(userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, address)
}

// Update 更新地址
func (h *AddressHandler) Update(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	addressID, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的地址ID")
		return
	}

	var req model.AddressUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	address, err := h.service.Update(userID, addressID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, address)
}

// Delete 删除地址
func (h *AddressHandler) Delete(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	addressID, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的地址ID")
		return
	}

	if err := h.service.Delete(userID, addressID); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// SetDefault 设置默认地址
func (h *AddressHandler) SetDefault(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	addressID, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的地址ID")
		return
	}

	if err := h.service.SetDefault(userID, addressID); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// parseIDParam 解析ID参数
func (h *AddressHandler) parseIDParam(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// handleError 统一处理错误
func (h *AddressHandler) handleError(c *gin.Context, err error) {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		response.Fail(c, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, apperrors.CodeInternalError, "服务器内部错误")
}
