package repository

import (
	"sync"

	"rulebacktest/internal/model"
	"rulebacktest/pkg/database"
)

var (
	favoriteRepoInstance *FavoriteRepository
	favoriteRepoOnce     sync.Once
)

// FavoriteRepository 收藏数据访问层
type FavoriteRepository struct {
	*BaseRepository
}

// NewFavoriteRepository 创建FavoriteRepository实例
func NewFavoriteRepository(base *BaseRepository) *FavoriteRepository {
	return &FavoriteRepository{BaseRepository: base}
}

// GetFavoriteRepository 获取FavoriteRepository单例
func GetFavoriteRepository() *FavoriteRepository {
	favoriteRepoOnce.Do(func() {
		favoriteRepoInstance = &FavoriteRepository{
			BaseRepository: &BaseRepository{db: database.GetDB()},
		}
	})
	return favoriteRepoInstance
}

// Create 添加收藏
func (r *FavoriteRepository) Create(favorite *model.Favorite) error {
	return r.DB().Create(favorite).Error
}

// Delete 取消收藏
func (r *FavoriteRepository) Delete(userID, productID uint) error {
	return r.DB().Where("user_id = ? AND product_id = ?", userID, productID).
		Delete(&model.Favorite{}).Error
}

// GetByUserID 获取用户收藏列表
func (r *FavoriteRepository) GetByUserID(userID uint, page, pageSize int) ([]model.Favorite, int64, error) {
	var favorites []model.Favorite
	var total int64

	query := r.DB().Model(&model.Favorite{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Product").Preload("Product.Category").
		Scopes(r.Paginate(page, pageSize)).
		Order("id DESC").
		Find(&favorites).Error

	return favorites, total, err
}

// Exists 检查是否已收藏
func (r *FavoriteRepository) Exists(userID, productID uint) (bool, error) {
	var count int64
	err := r.DB().Model(&model.Favorite{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count).Error
	return count > 0, err
}

// GetByUserAndProduct 根据用户和商品获取收藏
func (r *FavoriteRepository) GetByUserAndProduct(userID, productID uint) (*model.Favorite, error) {
	var favorite model.Favorite
	err := r.DB().Where("user_id = ? AND product_id = ?", userID, productID).First(&favorite).Error
	return &favorite, err
}
