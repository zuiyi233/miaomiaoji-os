package repository

import (
	"sync"

	"gorm.io/gorm"
	"rulebacktest/internal/model"
	"rulebacktest/pkg/database"
)

var (
	orderRepoInstance *OrderRepository
	orderRepoOnce     sync.Once
)

// OrderRepository 订单数据访问层
type OrderRepository struct {
	*BaseRepository
}

// NewOrderRepository 创建OrderRepository实例
func NewOrderRepository(base *BaseRepository) *OrderRepository {
	return &OrderRepository{BaseRepository: base}
}

// GetOrderRepository 获取OrderRepository单例
func GetOrderRepository() *OrderRepository {
	orderRepoOnce.Do(func() {
		orderRepoInstance = &OrderRepository{
			BaseRepository: &BaseRepository{db: database.GetDB()},
		}
	})
	return orderRepoInstance
}

// GetByID 根据ID获取订单
func (r *OrderRepository) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Items").First(&order, id).Error
	return &order, err
}

// GetByOrderNo 根据订单号获取订单
func (r *OrderRepository) GetByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Items").Where("order_no = ?", orderNo).First(&order).Error
	return &order, err
}

// ListByUserID 获取用户订单列表
func (r *OrderRepository) ListByUserID(userID uint, req *model.OrderListReq) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{}).Where("user_id = ?", userID)

	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	req.SetDefaults()
	err := query.Preload("Items").
		Scopes(r.Paginate(req.Page, req.PageSize)).
		Order("id DESC").
		Find(&orders).Error

	return orders, total, err
}

// Create 创建订单
func (r *OrderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

// CreateWithTx 使用事务创建订单
func (r *OrderRepository) CreateWithTx(tx *gorm.DB, order *model.Order) error {
	return tx.Create(order).Error
}

// Update 更新订单
func (r *OrderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

// UpdateStatus 更新订单状态
func (r *OrderRepository) UpdateStatus(orderID uint, status model.OrderStatus) error {
	return r.db.Model(&model.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

// Delete 删除订单
func (r *OrderRepository) Delete(id uint) error {
	return r.db.Delete(&model.Order{}, id).Error
}

// Transaction 事务
func (r *OrderRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// CreateOrderItems 创建订单项
func (r *OrderRepository) CreateOrderItems(tx *gorm.DB, items []model.OrderItem) error {
	return tx.Create(&items).Error
}

// AdminList 管理员获取订单列表
func (r *OrderRepository) AdminList(req *model.AdminOrderListReq) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{})

	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.OrderNo != "" {
		query = query.Where("order_no LIKE ?", "%"+req.OrderNo+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	req.SetDefaults()
	err := query.Preload("Items").
		Scopes(r.Paginate(req.Page, req.PageSize)).
		Order("id DESC").
		Find(&orders).Error

	return orders, total, err
}
