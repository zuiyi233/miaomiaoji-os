package model

import "github.com/shopspring/decimal"

// Cart 购物车模型
type Cart struct {
	BaseModelWithoutSoftDelete
	UserID    uint     `gorm:"index;not null" json:"user_id"`
	ProductID uint     `gorm:"index;not null" json:"product_id"`
	Quantity  int      `gorm:"default:1" json:"quantity"`
	Product   *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (Cart) TableName() string {
	return "carts"
}

// CartAddReq 添加购物车请求
type CartAddReq struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

// CartUpdateReq 更新购物车请求
type CartUpdateReq struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// CartItem 购物车商品项
type CartItem struct {
	ID        uint            `json:"id"`
	ProductID uint            `json:"product_id"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
	Images    string          `json:"images"`
	Quantity  int             `json:"quantity"`
	Stock     int             `json:"stock"`
}

// CartListResp 购物车列表响应
type CartListResp struct {
	Items      []CartItem      `json:"items"`
	TotalPrice decimal.Decimal `json:"total_price"`
	TotalCount int             `json:"total_count"`
}
