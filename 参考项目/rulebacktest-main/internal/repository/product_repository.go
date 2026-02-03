package repository

import (
	"sync"

	"gorm.io/gorm"
	"rulebacktest/internal/model"
	"rulebacktest/pkg/database"
)

var (
	productRepoInstance *ProductRepository
	productRepoOnce     sync.Once
)

// ProductRepository 商品数据访问层
type ProductRepository struct {
	*BaseRepository
}

// NewProductRepository 创建ProductRepository实例
func NewProductRepository(base *BaseRepository) *ProductRepository {
	return &ProductRepository{BaseRepository: base}
}

// GetProductRepository 获取ProductRepository单例
func GetProductRepository() *ProductRepository {
	productRepoOnce.Do(func() {
		productRepoInstance = &ProductRepository{
			BaseRepository: &BaseRepository{db: database.GetDB()},
		}
	})
	return productRepoInstance
}

// GetByID 根据ID获取商品
func (r *ProductRepository) GetByID(id uint) (*model.Product, error) {
	var product model.Product
	err := r.db.Preload("Category").First(&product, id).Error
	return &product, err
}

// List 获取商品列表
func (r *ProductRepository) List(req *model.ProductListReq) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	query := r.db.Model(&model.Product{})

	if req.CategoryID > 0 {
		query = query.Where("category_id = ?", req.CategoryID)
	}
	if req.Keyword != "" {
		query = query.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	req.SetDefaults()
	err := query.Preload("Category").
		Scopes(r.Paginate(req.Page, req.PageSize)).
		Order("id DESC").
		Find(&products).Error

	return products, total, err
}

// Create 创建商品
func (r *ProductRepository) Create(product *model.Product) error {
	return r.db.Create(product).Error
}

// Update 更新商品
func (r *ProductRepository) Update(product *model.Product) error {
	return r.db.Save(product).Error
}

// Delete 删除商品
func (r *ProductRepository) Delete(id uint) error {
	return r.db.Delete(&model.Product{}, id).Error
}

// UpdateStock 更新库存（增量）
func (r *ProductRepository) UpdateStock(productID uint, quantity int) error {
	return r.db.Model(&model.Product{}).
		Where("id = ?", productID).
		Update("stock", gorm.Expr("stock + ?", quantity)).Error
}

// SetStock 设置库存（绝对值）
func (r *ProductRepository) SetStock(productID uint, stock int) error {
	return r.db.Model(&model.Product{}).
		Where("id = ?", productID).
		Update("stock", stock).Error
}

// DeductStock 扣减库存
func (r *ProductRepository) DeductStock(tx *gorm.DB, productID uint, quantity int) error {
	result := tx.Model(&model.Product{}).
		Where("id = ? AND stock >= ?", productID, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetByIDForUpdate 获取商品并加锁
func (r *ProductRepository) GetByIDForUpdate(tx *gorm.DB, id uint) (*model.Product, error) {
	var product model.Product
	err := tx.Set("gorm:query_option", "FOR UPDATE").First(&product, id).Error
	return &product, err
}
