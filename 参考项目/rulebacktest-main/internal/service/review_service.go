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
	reviewServiceInstance *ReviewService
	reviewServiceOnce     sync.Once
)

// ReviewService 评价业务逻辑层
type ReviewService struct {
	repo      *repository.ReviewRepository
	orderRepo *repository.OrderRepository
}

// NewReviewService 创建ReviewService实例
func NewReviewService(repo *repository.ReviewRepository, orderRepo *repository.OrderRepository) *ReviewService {
	return &ReviewService{repo: repo, orderRepo: orderRepo}
}

// GetReviewService 获取ReviewService单例
func GetReviewService() *ReviewService {
	reviewServiceOnce.Do(func() {
		reviewServiceInstance = &ReviewService{
			repo:      repository.GetReviewRepository(),
			orderRepo: repository.GetOrderRepository(),
		}
	})
	return reviewServiceInstance
}

// Create 创建评价
func (s *ReviewService) Create(userID uint, req *model.ReviewCreateReq) (*model.Review, error) {
	// 验证订单
	order, err := s.orderRepo.GetByID(req.OrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeNotFound, "订单不存在")
		}
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	// 验证订单归属
	if order.UserID != userID {
		return nil, apperrors.ErrForbidden
	}

	// 验证订单状态（只有已完成的订单才能评价）
	if order.Status != model.OrderStatusCompleted {
		return nil, apperrors.New(apperrors.CodeInvalidParams, "只能评价已完成的订单")
	}

	// 验证商品是否在订单中
	var found bool
	for _, item := range order.Items {
		if item.ProductID == req.ProductID {
			found = true
			break
		}
	}
	if !found {
		return nil, apperrors.New(apperrors.CodeInvalidParams, "商品不在订单中")
	}

	// 检查是否已评价
	exists, err := s.repo.ExistsByOrderAndProduct(userID, req.OrderID, req.ProductID)
	if err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	if exists {
		return nil, apperrors.New(apperrors.CodeConflict, "已评价过该商品")
	}

	review := &model.Review{
		UserID:    userID,
		OrderID:   req.OrderID,
		ProductID: req.ProductID,
		Rating:    req.Rating,
		Content:   req.Content,
		Images:    req.Images,
	}

	if err := s.repo.Create(review); err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return review, nil
}

// ListByProduct 获取商品评价列表
func (s *ReviewService) ListByProduct(productID uint, page, pageSize int) ([]model.Review, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reviews, total, err := s.repo.ListByProductID(productID, page, pageSize)
	if err != nil {
		return nil, 0, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return reviews, total, nil
}

// ListByUser 获取用户评价列表
func (s *ReviewService) ListByUser(userID uint, page, pageSize int) ([]model.Review, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reviews, total, err := s.repo.ListByUserID(userID, page, pageSize)
	if err != nil {
		return nil, 0, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return reviews, total, nil
}

// GetProductRating 获取商品评分统计
func (s *ReviewService) GetProductRating(productID uint) (float64, int64, error) {
	avgRating, count, err := s.repo.GetProductRating(productID)
	if err != nil {
		return 0, 0, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return avgRating, count, nil
}
