package model

import "github.com/shopspring/decimal"

// OrderStatus 订单状态
type OrderStatus int8

const (
	OrderStatusPending   OrderStatus = 0 // 待支付
	OrderStatusPaid      OrderStatus = 1 // 已支付
	OrderStatusShipped   OrderStatus = 2 // 已发货
	OrderStatusCompleted OrderStatus = 3 // 已完成
	OrderStatusCancelled OrderStatus = 4 // 已取消
)

// Order 订单模型
type Order struct {
	BaseModel
	OrderNo     string          `gorm:"size:50;uniqueIndex;not null" json:"order_no"`
	UserID      uint            `gorm:"index;not null" json:"user_id"`
	TotalAmount decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status      OrderStatus     `gorm:"default:0" json:"status"`
	Address     string          `gorm:"size:500" json:"address"`
	Remark      string          `gorm:"size:500" json:"remark"`
	Items       []OrderItem     `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

// OrderItem 订单商品项
type OrderItem struct {
	BaseModelWithoutSoftDelete
	OrderID   uint            `gorm:"index;not null" json:"order_id"`
	ProductID uint            `gorm:"index;not null" json:"product_id"`
	Name      string          `gorm:"size:200;not null" json:"name"`
	Price     decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	Quantity  int             `gorm:"not null" json:"quantity"`
	Product   *Product        `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

// OrderCreateReq 创建订单请求
type OrderCreateReq struct {
	Address string `json:"address" binding:"required,max=500"`
	Remark  string `json:"remark" binding:"max=500"`
}

// OrderListReq 订单列表请求
type OrderListReq struct {
	PageQuery
	Status *int8 `form:"status"`
}

// AdminOrderListReq 管理员订单列表请求
type AdminOrderListReq struct {
	PageQuery
	Status  *int8  `form:"status"`
	UserID  *uint  `form:"user_id"`
	OrderNo string `form:"order_no"`
}
