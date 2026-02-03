package repository

import (
	"sync"

	"gorm.io/gorm"
	"rulebacktest/internal/model"
	"rulebacktest/pkg/database"
)

var (
	cartRepoInstance *CartRepository
	cartRepoOnce     sync.Once
)

// CartRepository 购物车数据访问层
type CartRepository struct {
	*BaseRepository
}

// NewCartRepository 创建CartRepository实例
func NewCartRepository(base *BaseRepository) *CartRepository {
	return &CartRepository{BaseRepository: base}
}

// GetCartRepository 获取CartRepository单例
func GetCartRepository() *CartRepository {
	cartRepoOnce.Do(func() {
		cartRepoInstance = &CartRepository{
			BaseRepository: &BaseRepository{db: database.GetDB()},
		}
	})
	return cartRepoInstance
}

// GetByID 根据ID获取购物车项
func (r *CartRepository) GetByID(id uint) (*model.Cart, error) {
	var cart model.Cart
	err := r.db.Preload("Product").First(&cart, id).Error
	return &cart, err
}

// GetByUserID 获取用户的购物车
func (r *CartRepository) GetByUserID(userID uint) ([]model.Cart, error) {
	var carts []model.Cart
	err := r.db.Preload("Product").
		Where("user_id = ?", userID).
		Find(&carts).Error
	return carts, err
}

// GetByUserAndProduct 获取用户特定商品的购物车项
func (r *CartRepository) GetByUserAndProduct(userID, productID uint) (*model.Cart, error) {
	var cart model.Cart
	err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&cart).Error
	return &cart, err
}

// Create 添加购物车
func (r *CartRepository) Create(cart *model.Cart) error {
	return r.db.Create(cart).Error
}

// Update 更新购物车
func (r *CartRepository) Update(cart *model.Cart) error {
	return r.db.Save(cart).Error
}

// UpdateQuantity 更新数量
func (r *CartRepository) UpdateQuantity(id uint, quantity int) error {
	return r.db.Model(&model.Cart{}).Where("id = ?", id).Update("quantity", quantity).Error
}

// Delete 删除购物车项
func (r *CartRepository) Delete(id uint) error {
	return r.db.Delete(&model.Cart{}, id).Error
}

// DeleteByUserID 清空用户购物车
func (r *CartRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.Cart{}).Error
}

// DeleteByUserIDWithTx 使用事务清空用户购物车
func (r *CartRepository) DeleteByUserIDWithTx(tx *gorm.DB, userID uint) error {
	return tx.Where("user_id = ?", userID).Delete(&model.Cart{}).Error
}

// CountByUserID 获取用户购物车商品数量
func (r *CartRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Cart{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
