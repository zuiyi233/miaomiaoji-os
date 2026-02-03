package model

import "github.com/shopspring/decimal"

// Product 商品模型
type Product struct {
	BaseModel
	Name        string          `gorm:"size:200;not null" json:"name"`
	Description string          `gorm:"type:text" json:"description"`
	CategoryID  uint            `gorm:"index;not null" json:"category_id"`
	Price       decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock       int             `gorm:"default:0" json:"stock"`
	Images      string          `gorm:"type:text" json:"images"`
	Status      Status          `gorm:"default:1" json:"status"`
	Category    *Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (Product) TableName() string {
	return "products"
}

// ProductCreateReq 创建商品请求
type ProductCreateReq struct {
	Name        string          `json:"name" binding:"required,max=200"`
	Description string          `json:"description"`
	CategoryID  uint            `json:"category_id" binding:"required"`
	Price       decimal.Decimal `json:"price" binding:"required"`
	Stock       int             `json:"stock" binding:"min=0"`
	Images      string          `json:"images"`
}

// ProductUpdateReq 更新商品请求
type ProductUpdateReq struct {
	Name        string           `json:"name" binding:"omitempty,max=200"`
	Description *string          `json:"description"`
	CategoryID  *uint            `json:"category_id"`
	Price       *decimal.Decimal `json:"price"`
	Stock       *int             `json:"stock"`
	Images      *string          `json:"images"`
	Status      *Status          `json:"status"`
}

// ProductListReq 商品列表请求
type ProductListReq struct {
	PageQuery
	CategoryID uint   `form:"category_id"`
	Keyword    string `form:"keyword"`
	Status     *int8  `form:"status"`
}
