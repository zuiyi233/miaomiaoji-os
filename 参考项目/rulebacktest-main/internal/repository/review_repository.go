package repository

import (
	"sync"

	"rulebacktest/internal/model"
	"rulebacktest/pkg/database"
)

var (
	reviewRepoInstance *ReviewRepository
	reviewRepoOnce     sync.Once
)

// ReviewRepository 评价数据访问层
type ReviewRepository struct {
	*BaseRepository
}

// NewReviewRepository 创建ReviewRepository实例
func NewReviewRepository(base *BaseRepository) *ReviewRepository {
	return &ReviewRepository{BaseRepository: base}
}

// GetReviewRepository 获取ReviewRepository单例
func GetReviewRepository() *ReviewRepository {
	reviewRepoOnce.Do(func() {
		reviewRepoInstance = &ReviewRepository{
			BaseRepository: &BaseRepository{db: database.GetDB()},
		}
	})
	return reviewRepoInstance
}

// Create 创建评价
func (r *ReviewRepository) Create(review *model.Review) error {
	return r.DB().Create(review).Error
}

// GetByID 根据ID获取评价
func (r *ReviewRepository) GetByID(id uint) (*model.Review, error) {
	var review model.Review
	err := r.DB().Preload("User").Preload("Product").First(&review, id).Error
	return &review, err
}

// ListByProductID 获取商品评价列表
func (r *ReviewRepository) ListByProductID(productID uint, page, pageSize int) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64

	query := r.DB().Model(&model.Review{}).Where("product_id = ?", productID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").
		Scopes(r.Paginate(page, pageSize)).
		Order("id DESC").
		Find(&reviews).Error

	return reviews, total, err
}

// ListByUserID 获取用户评价列表
func (r *ReviewRepository) ListByUserID(userID uint, page, pageSize int) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64

	query := r.DB().Model(&model.Review{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Product").
		Scopes(r.Paginate(page, pageSize)).
		Order("id DESC").
		Find(&reviews).Error

	return reviews, total, err
}

// ExistsByOrderAndProduct 检查订单商品是否已评价
func (r *ReviewRepository) ExistsByOrderAndProduct(userID, orderID, productID uint) (bool, error) {
	var count int64
	err := r.DB().Model(&model.Review{}).
		Where("user_id = ? AND order_id = ? AND product_id = ?", userID, orderID, productID).
		Count(&count).Error
	return count > 0, err
}

// GetProductRating 获取商品平均评分
func (r *ReviewRepository) GetProductRating(productID uint) (float64, int64, error) {
	var result struct {
		AvgRating float64
		Count     int64
	}
	err := r.DB().Model(&model.Review{}).
		Select("AVG(rating) as avg_rating, COUNT(*) as count").
		Where("product_id = ?", productID).
		Scan(&result).Error
	return result.AvgRating, result.Count, err
}
