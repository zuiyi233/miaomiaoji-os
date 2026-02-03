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
	categoryHandlerInstance *CategoryHandler
	categoryHandlerOnce     sync.Once
)

// CategoryHandler 分类HTTP处理器
type CategoryHandler struct {
	service *service.CategoryService
}

// NewCategoryHandler 创建CategoryHandler实例
func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: svc}
}

// GetCategoryHandler 获取CategoryHandler单例
func GetCategoryHandler() *CategoryHandler {
	categoryHandlerOnce.Do(func() {
		categoryHandlerInstance = &CategoryHandler{
			service: service.GetCategoryService(),
		}
	})
	return categoryHandlerInstance
}

// Create 创建分类
func (h *CategoryHandler) Create(c *gin.Context) {
	var req model.CategoryCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	category, err := h.service.Create(&req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, category)
}

// GetByID 获取分类详情
func (h *CategoryHandler) GetByID(c *gin.Context) {
	id, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的分类ID")
		return
	}

	category, err := h.service.GetByID(id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, category)
}

// List 获取分类列表
func (h *CategoryHandler) List(c *gin.Context) {
	parentIDStr := c.Query("parent_id")
	if parentIDStr != "" {
		parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err != nil {
			response.Fail(c, apperrors.CodeInvalidParams, "无效的父分类ID")
			return
		}
		categories, err := h.service.ListByParentID(uint(parentID))
		if err != nil {
			h.handleError(c, err)
			return
		}
		response.SuccessWithData(c, categories)
		return
	}

	categories, err := h.service.List()
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, categories)
}

// Update 更新分类
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的分类ID")
		return
	}

	var req model.CategoryUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	category, err := h.service.Update(id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, category)
}

// Delete 删除分类
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := h.parseIDParam(c)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的分类ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// parseIDParam 解析ID参数
func (h *CategoryHandler) parseIDParam(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// handleError 统一处理错误
func (h *CategoryHandler) handleError(c *gin.Context, err error) {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		response.Fail(c, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, apperrors.CodeInternalError, "服务器内部错误")
}
