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
	commentHandlerInstance *CommentHandler
	commentHandlerOnce     sync.Once
)

// CommentHandler 评论HTTP处理器
type CommentHandler struct {
	service *service.CommentService
}

// NewCommentHandler 创建CommentHandler实例
func NewCommentHandler(svc *service.CommentService) *CommentHandler {
	return &CommentHandler{service: svc}
}

// GetCommentHandler 获取CommentHandler单例
func GetCommentHandler() *CommentHandler {
	commentHandlerOnce.Do(func() {
		commentHandlerInstance = &CommentHandler{
			service: service.GetCommentService(),
		}
	})
	return commentHandlerInstance
}

// Create 创建评论
func (h *CommentHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	var req model.CommentCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	comment, err := h.service.Create(userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, comment)
}

// List 获取评论列表
func (h *CommentHandler) List(c *gin.Context) {
	var req model.CommentListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, err.Error())
		return
	}

	comments, total, err := h.service.List(&req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	req.SetDefaults()
	response.SuccessWithPage(c, comments, total, req.Page, req.PageSize)
}

// ListReplies 获取回复列表
func (h *CommentHandler) ListReplies(c *gin.Context) {
	parentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的评论ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	comments, total, err := h.service.ListReplies(uint(parentID), page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithPage(c, comments, total, page, pageSize)
}

// Delete 删除评论
func (h *CommentHandler) Delete(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Fail(c, apperrors.CodeUnauthorized, "未登录")
		return
	}

	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的评论ID")
		return
	}

	if err := h.service.Delete(userID, uint(commentID)); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c)
}

// Count 获取评论数量
func (h *CommentHandler) Count(c *gin.Context) {
	typeVal, err := strconv.ParseInt(c.Query("type"), 10, 8)
	if err != nil || (typeVal != 1 && typeVal != 2) {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的评论类型")
		return
	}

	targetID, err := strconv.ParseUint(c.Query("target_id"), 10, 32)
	if err != nil {
		response.Fail(c, apperrors.CodeInvalidParams, "无效的目标ID")
		return
	}

	count, err := h.service.Count(model.CommentType(typeVal), uint(targetID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithData(c, gin.H{"count": count})
}

// handleError 统一处理错误
func (h *CommentHandler) handleError(c *gin.Context, err error) {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		response.Fail(c, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, apperrors.CodeInternalError, "服务器内部错误")
}
