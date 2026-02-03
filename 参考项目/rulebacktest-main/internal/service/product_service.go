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
	productServiceInstance *ProductService
	productServiceOnce     sync.Once
)

// ProductService 商品业务逻辑层
type ProductService struct {
	repo         *repository.ProductRepository
	categoryRepo *repository.CategoryRepository
}

// NewProductService 创建ProductService实例
func NewProductService(repo *repository.ProductRepository, categoryRepo *repository.CategoryRepository) *ProductService {
	return &ProductService{repo: repo, categoryRepo: categoryRepo}
}

// GetProductService 获取ProductService单例
func GetProductService() *ProductService {
	productServiceOnce.Do(func() {
		productServiceInstance = &ProductService{
			repo:         repository.GetProductRepository(),
			categoryRepo: repository.GetCategoryRepository(),
		}
	})
	return productServiceInstance
}

// Create 创建商品
func (s *ProductService) Create(req *model.ProductCreateReq) (*model.Product, error) {
	_, err := s.categoryRepo.GetByID(req.CategoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeNotFound, "分类不存在")
		}
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	product := &model.Product{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		Price:       req.Price,
		Stock:       req.Stock,
		Images:      req.Images,
		Status:      model.StatusEnabled,
	}

	if err := s.repo.Create(product); err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return product, nil
}

// GetByID 根据ID获取商品
func (s *ProductService) GetByID(id uint) (*model.Product, error) {
	product, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return product, nil
}

// List 获取商品列表
func (s *ProductService) List(req *model.ProductListReq) ([]model.Product, int64, error) {
	products, total, err := s.repo.List(req)
	if err != nil {
		return nil, 0, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return products, total, nil
}

// Update 更新商品
func (s *ProductService) Update(id uint, req *model.ProductUpdateReq) (*model.Product, error) {
	product, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if req.CategoryID != nil && *req.CategoryID != product.CategoryID {
		_, err := s.categoryRepo.GetByID(*req.CategoryID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.New(apperrors.CodeNotFound, "分类不存在")
			}
			return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
		}
		product.CategoryID = *req.CategoryID
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.Images != nil {
		product.Images = *req.Images
	}
	if req.Status != nil {
		product.Status = *req.Status
	}

	if err := s.repo.Update(product); err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return product, nil
}

// Delete 删除商品
func (s *ProductService) Delete(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound
		}
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if err := s.repo.Delete(id); err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return nil
}

// UpdateStock 设置商品库存
func (s *ProductService) UpdateStock(id uint, stock int) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound
		}
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if err := s.repo.SetStock(id, stock); err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return nil
}
