package repository

import (
	"sync"

	"rulebacktest/internal/model"
	"rulebacktest/pkg/database"
)

var (
	categoryRepoInstance *CategoryRepository
	categoryRepoOnce     sync.Once
)

// CategoryRepository 分类数据访问层
type CategoryRepository struct {
	*BaseRepository
}

// NewCategoryRepository 创建CategoryRepository实例
func NewCategoryRepository(base *BaseRepository) *CategoryRepository {
	return &CategoryRepository{BaseRepository: base}
}

// GetCategoryRepository 获取CategoryRepository单例
func GetCategoryRepository() *CategoryRepository {
	categoryRepoOnce.Do(func() {
		categoryRepoInstance = &CategoryRepository{
			BaseRepository: &BaseRepository{db: database.GetDB()},
		}
	})
	return categoryRepoInstance
}

// GetByID 根据ID获取分类
func (r *CategoryRepository) GetByID(id uint) (*model.Category, error) {
	var category model.Category
	err := r.db.First(&category, id).Error
	return &category, err
}

// List 获取分类列表
func (r *CategoryRepository) List() ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Where("status = ?", model.StatusEnabled).
		Order("sort ASC, id ASC").
		Find(&categories).Error
	return categories, err
}

// ListByParentID 根据父级ID获取子分类
func (r *CategoryRepository) ListByParentID(parentID uint) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Where("parent_id = ? AND status = ?", parentID, model.StatusEnabled).
		Order("sort ASC, id ASC").
		Find(&categories).Error
	return categories, err
}

// ExistsByName 检查分类名是否存在
func (r *CategoryRepository) ExistsByName(name string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.Model(&model.Category{}).Where("name = ?", name)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// Create 创建分类
func (r *CategoryRepository) Create(category *model.Category) error {
	return r.db.Create(category).Error
}

// Update 更新分类
func (r *CategoryRepository) Update(category *model.Category) error {
	return r.db.Save(category).Error
}

// Delete 删除分类
func (r *CategoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Category{}, id).Error
}

// HasProducts 检查分类下是否有商品
func (r *CategoryRepository) HasProducts(categoryID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Product{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count > 0, err
}
