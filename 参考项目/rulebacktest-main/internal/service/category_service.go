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
	categoryServiceInstance *CategoryService
	categoryServiceOnce     sync.Once
)

// CategoryService 分类业务逻辑层
type CategoryService struct {
	repo *repository.CategoryRepository
}

// NewCategoryService 创建CategoryService实例
func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

// GetCategoryService 获取CategoryService单例
func GetCategoryService() *CategoryService {
	categoryServiceOnce.Do(func() {
		categoryServiceInstance = &CategoryService{
			repo: repository.GetCategoryRepository(),
		}
	})
	return categoryServiceInstance
}

// Create 创建分类
func (s *CategoryService) Create(req *model.CategoryCreateReq) (*model.Category, error) {
	exists, err := s.repo.ExistsByName(req.Name, 0)
	if err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	if exists {
		return nil, apperrors.New(apperrors.CodeConflict, "分类名称已存在")
	}

	if req.ParentID > 0 {
		_, err := s.repo.GetByID(req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.New(apperrors.CodeNotFound, "父分类不存在")
			}
			return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
		}
	}

	category := &model.Category{
		Name:     req.Name,
		ParentID: req.ParentID,
		Sort:     req.Sort,
		Status:   model.StatusEnabled,
	}

	if err := s.repo.Create(category); err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return category, nil
}

// GetByID 根据ID获取分类
func (s *CategoryService) GetByID(id uint) (*model.Category, error) {
	category, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return category, nil
}

// List 获取分类列表
func (s *CategoryService) List() ([]model.Category, error) {
	categories, err := s.repo.List()
	if err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return categories, nil
}

// ListByParentID 获取子分类列表
func (s *CategoryService) ListByParentID(parentID uint) ([]model.Category, error) {
	categories, err := s.repo.ListByParentID(parentID)
	if err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return categories, nil
}

// Update 更新分类
func (s *CategoryService) Update(id uint, req *model.CategoryUpdateReq) (*model.Category, error) {
	category, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if req.Name != "" && req.Name != category.Name {
		exists, err := s.repo.ExistsByName(req.Name, id)
		if err != nil {
			return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
		}
		if exists {
			return nil, apperrors.New(apperrors.CodeConflict, "分类名称已存在")
		}
		category.Name = req.Name
	}

	if req.Sort != nil {
		category.Sort = *req.Sort
	}
	if req.Status != nil {
		category.Status = *req.Status
	}

	if err := s.repo.Update(category); err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return category, nil
}

// Delete 删除分类
func (s *CategoryService) Delete(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound
		}
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	hasProducts, err := s.repo.HasProducts(id)
	if err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	if hasProducts {
		return apperrors.New(apperrors.CodeConflict, "该分类下存在商品，无法删除")
	}

	if err := s.repo.Delete(id); err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return nil
}
