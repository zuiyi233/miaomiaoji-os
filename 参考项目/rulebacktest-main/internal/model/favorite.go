package model

// Favorite 商品收藏模型
type Favorite struct {
	BaseModel
	UserID    uint     `gorm:"uniqueIndex:idx_user_product;not null" json:"user_id"`
	ProductID uint     `gorm:"uniqueIndex:idx_user_product;not null" json:"product_id"`
	Product   *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (Favorite) TableName() string {
	return "favorites"
}

// FavoriteAddReq 添加收藏请求
type FavoriteAddReq struct {
	ProductID uint `json:"product_id" binding:"required"`
}
