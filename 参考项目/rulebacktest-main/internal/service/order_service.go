package service

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"rulebacktest/internal/model"
	"rulebacktest/internal/repository"
	apperrors "rulebacktest/pkg/errors"
	"rulebacktest/pkg/logger"
)

var (
	orderServiceInstance *OrderService
	orderServiceOnce     sync.Once
)

// OrderService 订单业务逻辑层
type OrderService struct {
	repo        *repository.OrderRepository
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
}

// NewOrderService 创建OrderService实例
func NewOrderService(repo *repository.OrderRepository, cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *OrderService {
	return &OrderService{repo: repo, cartRepo: cartRepo, productRepo: productRepo}
}

// GetOrderService 获取OrderService单例
func GetOrderService() *OrderService {
	orderServiceOnce.Do(func() {
		orderServiceInstance = &OrderService{
			repo:        repository.GetOrderRepository(),
			cartRepo:    repository.GetCartRepository(),
			productRepo: repository.GetProductRepository(),
		}
	})
	return orderServiceInstance
}

// Create 创建订单
func (s *OrderService) Create(userID uint, req *model.OrderCreateReq) (*model.Order, error) {
	carts, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if len(carts) == 0 {
		return nil, apperrors.New(apperrors.CodeInvalidParams, "购物车为空")
	}

	var order *model.Order
	err = s.repo.Transaction(func(tx *gorm.DB) error {
		totalAmount := decimal.Zero
		orderItems := make([]model.OrderItem, 0, len(carts))

		for _, cart := range carts {
			product, err := s.productRepo.GetByIDForUpdate(tx, cart.ProductID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return apperrors.New(apperrors.CodeNotFound, fmt.Sprintf("商品 %d 不存在", cart.ProductID))
				}
				return err
			}

			if !product.Status.IsEnabled() {
				return apperrors.New(apperrors.CodeInvalidParams, fmt.Sprintf("商品 %s 已下架", product.Name))
			}

			if product.Stock < cart.Quantity {
				return apperrors.New(apperrors.CodeInvalidParams, fmt.Sprintf("商品 %s 库存不足", product.Name))
			}

			if err := s.productRepo.DeductStock(tx, product.ID, cart.Quantity); err != nil {
				return err
			}

			itemAmount := product.Price.Mul(decimal.NewFromInt(int64(cart.Quantity)))
			totalAmount = totalAmount.Add(itemAmount)

			orderItems = append(orderItems, model.OrderItem{
				ProductID: product.ID,
				Name:      product.Name,
				Price:     product.Price,
				Quantity:  cart.Quantity,
			})
		}

		order = &model.Order{
			OrderNo:     generateOrderNo(),
			UserID:      userID,
			TotalAmount: totalAmount,
			Status:      model.OrderStatusPending,
			Address:     req.Address,
			Remark:      req.Remark,
		}

		if err := s.repo.CreateWithTx(tx, order); err != nil {
			return err
		}

		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}

		if err := s.repo.CreateOrderItems(tx, orderItems); err != nil {
			return err
		}

		if err := s.cartRepo.DeleteByUserIDWithTx(tx, userID); err != nil {
			return err
		}

		order.Items = orderItems
		return nil
	})

	if err != nil {
		logger.Error("创建订单失败", logger.Uint("user_id", userID), logger.Err(err))
		if apperrors.IsAppError(err) {
			return nil, err
		}
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return order, nil
}

// GetByID 根据ID获取订单
func (s *OrderService) GetByID(userID, orderID uint) (*model.Order, error) {
	order, err := s.repo.GetByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if order.UserID != userID {
		return nil, apperrors.ErrForbidden
	}

	return order, nil
}

// List 获取订单列表
func (s *OrderService) List(userID uint, req *model.OrderListReq) ([]model.Order, int64, error) {
	orders, total, err := s.repo.ListByUserID(userID, req)
	if err != nil {
		return nil, 0, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return orders, total, nil
}

// Cancel 取消订单
func (s *OrderService) Cancel(userID, orderID uint) error {
	order, err := s.repo.GetByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound
		}
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if order.UserID != userID {
		return apperrors.ErrForbidden
	}

	if order.Status != model.OrderStatusPending {
		return apperrors.New(apperrors.CodeInvalidParams, "只能取消待支付订单")
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		for _, item := range order.Items {
			if err := tx.Model(&model.Product{}).
				Where("id = ?", item.ProductID).
				Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return err
			}
		}

		return tx.Model(&model.Order{}).
			Where("id = ?", orderID).
			Update("status", model.OrderStatusCancelled).Error
	})

	if err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return nil
}

// Pay 支付订单
func (s *OrderService) Pay(userID, orderID uint) error {
	order, err := s.repo.GetByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound
		}
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if order.UserID != userID {
		return apperrors.ErrForbidden
	}

	if order.Status != model.OrderStatusPending {
		return apperrors.New(apperrors.CodeInvalidParams, "订单状态不正确")
	}

	if err := s.repo.UpdateStatus(orderID, model.OrderStatusPaid); err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return nil
}

// generateOrderNo 生成订单号
func generateOrderNo() string {
	return time.Now().Format("20060102150405") + uuid.New().String()[:8]
}

// AdminList 管理员获取订单列表
func (s *OrderService) AdminList(req *model.AdminOrderListReq) ([]model.Order, int64, error) {
	orders, total, err := s.repo.AdminList(req)
	if err != nil {
		return nil, 0, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return orders, total, nil
}

// AdminGetByID 管理员获取订单详情
func (s *OrderService) AdminGetByID(orderID uint) (*model.Order, error) {
	order, err := s.repo.GetByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return order, nil
}

// Ship 订单发货
func (s *OrderService) Ship(orderID uint) error {
	order, err := s.repo.GetByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound
		}
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if order.Status != model.OrderStatusPaid {
		return apperrors.New(apperrors.CodeInvalidParams, "只能发货已支付订单")
	}

	if err := s.repo.UpdateStatus(orderID, model.OrderStatusShipped); err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	logger.Info("订单已发货",
		logger.Uint("order_id", orderID),
		logger.String("order_no", order.OrderNo),
	)
	return nil
}

// Complete 管理员完成订单
func (s *OrderService) Complete(orderID uint) error {
	order, err := s.repo.GetByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound
		}
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if order.Status != model.OrderStatusShipped {
		return apperrors.New(apperrors.CodeInvalidParams, "只能完成已发货订单")
	}

	if err := s.repo.UpdateStatus(orderID, model.OrderStatusCompleted); err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	logger.Info("订单已完成",
		logger.Uint("order_id", orderID),
		logger.String("order_no", order.OrderNo),
	)
	return nil
}

// ConfirmReceipt 用户确认收货
func (s *OrderService) ConfirmReceipt(userID, orderID uint) error {
	order, err := s.repo.GetByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound
		}
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if order.UserID != userID {
		return apperrors.ErrForbidden
	}

	if order.Status != model.OrderStatusShipped {
		return apperrors.New(apperrors.CodeInvalidParams, "只能确认已发货订单")
	}

	if err := s.repo.UpdateStatus(orderID, model.OrderStatusCompleted); err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	logger.Info("用户确认收货",
		logger.Uint("user_id", userID),
		logger.Uint("order_id", orderID),
		logger.String("order_no", order.OrderNo),
	)
	return nil
}
