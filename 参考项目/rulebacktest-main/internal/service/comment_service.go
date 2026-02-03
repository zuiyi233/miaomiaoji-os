package service

import (
	"errors"
	"sync"

	"gorm.io/gorm"
	"rulebacktest/internal/model"
	"rulebacktest/internal/repository"
	apperrors "rulebacktest/pkg/errors"
)

var (
	commentServiceInstance *CommentService
	commentServiceOnce     sync.Once
)

// CommentService 评论业务逻辑层
type CommentService struct {
	repo        *repository.CommentRepository
	productRepo *repository.ProductRepository
	reviewRepo  *repository.ReviewRepository
}

// NewCommentService 创建CommentService实例
func NewCommentService(repo *repository.CommentRepository, productRepo *repository.ProductRepository, reviewRepo *repository.ReviewRepository) *CommentService {
	return &CommentService{repo: repo, productRepo: productRepo, reviewRepo: reviewRepo}
}

// GetCommentService 获取CommentService单例
func GetCommentService() *CommentService {
	commentServiceOnce.Do(func() {
		commentServiceInstance = &CommentService{
			repo:        repository.GetCommentRepository(),
			productRepo: repository.GetProductRepository(),
			reviewRepo:  repository.GetReviewRepository(),
		}
	})
	return commentServiceInstance
}

// Create 创建评论
func (s *CommentService) Create(userID uint, req *model.CommentCreateReq) (*model.Comment, error) {
	// 验证目标是否存在
	if err := s.validateTarget(req.Type, req.TargetID); err != nil {
		return nil, err
	}

	// 如果是回复，验证父评论是否存在
	if req.ParentID > 0 {
		parent, err := s.repo.GetByID(req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.New(apperrors.CodeNotFound, "父评论不存在")
			}
			return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
		}
		// 回复必须属于同一目标
		if parent.Type != req.Type || parent.TargetID != req.TargetID {
			return nil, apperrors.New(apperrors.CodeInvalidParams, "回复目标不匹配")
		}
	}

	comment := &model.Comment{
		UserID:   userID,
		Type:     req.Type,
		TargetID: req.TargetID,
		ParentID: req.ParentID,
		ReplyUID: req.ReplyUID,
		Content:  req.Content,
	}

	if err := s.repo.Create(comment); err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	// 重新获取以加载关联
	return s.repo.GetByID(comment.ID)
}

// validateTarget 验证评论目标是否存在
func (s *CommentService) validateTarget(commentType model.CommentType, targetID uint) error {
	switch commentType {
	case model.CommentTypeProduct:
		_, err := s.productRepo.GetByID(targetID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.New(apperrors.CodeNotFound, "商品不存在")
			}
			return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
		}
	case model.CommentTypeReview:
		_, err := s.reviewRepo.GetByID(targetID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.New(apperrors.CodeNotFound, "评价不存在")
			}
			return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
		}
	default:
		return apperrors.New(apperrors.CodeInvalidParams, "无效的评论类型")
	}
	return nil
}

// List 获取评论列表
func (s *CommentService) List(req *model.CommentListReq) ([]model.Comment, int64, error) {
	req.SetDefaults()
	comments, total, err := s.repo.ListByTarget(req.Type, req.TargetID, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return comments, total, nil
}

// ListReplies 获取回复列表
func (s *CommentService) ListReplies(parentID uint, page, pageSize int) ([]model.Comment, int64, error) {
	comments, total, err := s.repo.ListReplies(parentID, page, pageSize)
	if err != nil {
		return nil, 0, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return comments, total, nil
}

// Delete 删除评论
func (s *CommentService) Delete(userID, commentID uint) error {
	comment, err := s.repo.GetByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound
		}
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if comment.UserID != userID {
		return apperrors.ErrForbidden
	}

	// 删除子评论
	if err := s.repo.DeleteByParentID(commentID); err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	// 删除评论
	if err := s.repo.Delete(commentID); err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return nil
}

// Count 统计评论数
func (s *CommentService) Count(commentType model.CommentType, targetID uint) (int64, error) {
	count, err := s.repo.CountByTarget(commentType, targetID)
	if err != nil {
		return 0, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return count, nil
}
