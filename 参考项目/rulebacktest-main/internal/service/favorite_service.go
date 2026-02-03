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
	favoriteServiceInstance *FavoriteService
	favoriteServiceOnce     sync.Once
)

// FavoriteService 收藏业务逻辑层
type FavoriteService struct {
	repo        *repository.FavoriteRepository
	productRepo *repository.ProductRepository
}

// NewFavoriteService 创建FavoriteService实例
func NewFavoriteService(repo *repository.FavoriteRepository, productRepo *repository.ProductRepository) *FavoriteService {
	return &FavoriteService{repo: repo, productRepo: productRepo}
}

// GetFavoriteService 获取FavoriteService单例
func GetFavoriteService() *FavoriteService {
	favoriteServiceOnce.Do(func() {
		favoriteServiceInstance = &FavoriteService{
			repo:        repository.GetFavoriteRepository(),
			productRepo: repository.GetProductRepository(),
		}
	})
	return favoriteServiceInstance
}

// Add 添加收藏
func (s *FavoriteService) Add(userID uint, req *model.FavoriteAddReq) error {
	// 检查商品是否存在
	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.CodeNotFound, "商品不存在")
		}
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if !product.Status.IsEnabled() {
		return apperrors.New(apperrors.CodeInvalidParams, "商品已下架")
	}

	// 检查是否已收藏
	exists, err := s.repo.Exists(userID, req.ProductID)
	if err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	if exists {
		return apperrors.New(apperrors.CodeConflict, "已收藏该商品")
	}

	favorite := &model.Favorite{
		UserID:    userID,
		ProductID: req.ProductID,
	}

	if err := s.repo.Create(favorite); err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return nil
}

// Remove 取消收藏
func (s *FavoriteService) Remove(userID, productID uint) error {
	exists, err := s.repo.Exists(userID, productID)
	if err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	if !exists {
		return apperrors.New(apperrors.CodeNotFound, "未收藏该商品")
	}

	if err := s.repo.Delete(userID, productID); err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return nil
}

// List 获取收藏列表
func (s *FavoriteService) List(userID uint, page, pageSize int) ([]model.Favorite, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	favorites, total, err := s.repo.GetByUserID(userID, page, pageSize)
	if err != nil {
		return nil, 0, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return favorites, total, nil
}

// IsFavorite 检查是否已收藏
func (s *FavoriteService) IsFavorite(userID, productID uint) (bool, error) {
	exists, err := s.repo.Exists(userID, productID)
	if err != nil {
		return false, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return exists, nil
}
