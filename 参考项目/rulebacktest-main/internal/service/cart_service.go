package service

import (
	"errors"
	"sync"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"rulebacktest/internal/model"
	"rulebacktest/internal/repository"
	apperrors "rulebacktest/pkg/errors"
)

var (
	cartServiceInstance *CartService
	cartServiceOnce     sync.Once
)

// CartService 购物车业务逻辑层
type CartService struct {
	repo        *repository.CartRepository
	productRepo *repository.ProductRepository
}

// NewCartService 创建CartService实例
func NewCartService(repo *repository.CartRepository, productRepo *repository.ProductRepository) *CartService {
	return &CartService{repo: repo, productRepo: productRepo}
}

// GetCartService 获取CartService单例
func GetCartService() *CartService {
	cartServiceOnce.Do(func() {
		cartServiceInstance = &CartService{
			repo:        repository.GetCartRepository(),
			productRepo: repository.GetProductRepository(),
		}
	})
	return cartServiceInstance
}

// Add 添加商品到购物车
func (s *CartService) Add(userID uint, req *model.CartAddReq) error {
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

	if product.Stock < req.Quantity {
		return apperrors.New(apperrors.CodeInvalidParams, "库存不足")
	}

	cart, err := s.repo.GetByUserAndProduct(userID, req.ProductID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if cart != nil && cart.ID > 0 {
		newQuantity := cart.Quantity + req.Quantity
		if product.Stock < newQuantity {
			return apperrors.New(apperrors.CodeInvalidParams, "库存不足")
		}
		return s.repo.UpdateQuantity(cart.ID, newQuantity)
	}

	newCart := &model.Cart{
		UserID:    userID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}

	if err := s.repo.Create(newCart); err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return nil
}

// List 获取购物车列表
func (s *CartService) List(userID uint) (*model.CartListResp, error) {
	carts, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	items := make([]model.CartItem, 0, len(carts))
	totalPrice := decimal.Zero
	totalCount := 0

	for _, cart := range carts {
		if cart.Product == nil {
			continue
		}
		item := model.CartItem{
			ID:        cart.ID,
			ProductID: cart.ProductID,
			Name:      cart.Product.Name,
			Price:     cart.Product.Price,
			Images:    cart.Product.Images,
			Quantity:  cart.Quantity,
			Stock:     cart.Product.Stock,
		}
		items = append(items, item)
		totalPrice = totalPrice.Add(cart.Product.Price.Mul(decimal.NewFromInt(int64(cart.Quantity))))
		totalCount += cart.Quantity
	}

	return &model.CartListResp{
		Items:      items,
		TotalPrice: totalPrice,
		TotalCount: totalCount,
	}, nil
}

// Update 更新购物车数量
func (s *CartService) Update(userID, cartID uint, req *model.CartUpdateReq) error {
	cart, err := s.repo.GetByID(cartID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound
		}
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if cart.UserID != userID {
		return apperrors.ErrForbidden
	}

	product, err := s.productRepo.GetByID(cart.ProductID)
	if err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if product.Stock < req.Quantity {
		return apperrors.New(apperrors.CodeInvalidParams, "库存不足")
	}

	return s.repo.UpdateQuantity(cartID, req.Quantity)
}

// Delete 删除购物车项
func (s *CartService) Delete(userID, cartID uint) error {
	cart, err := s.repo.GetByID(cartID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound
		}
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if cart.UserID != userID {
		return apperrors.ErrForbidden
	}

	return s.repo.Delete(cartID)
}

// Clear 清空购物车
func (s *CartService) Clear(userID uint) error {
	return s.repo.DeleteByUserID(userID)
}

// GetRepo 获取Repository
func (s *CartService) GetRepo() *repository.CartRepository {
	return s.repo
}
