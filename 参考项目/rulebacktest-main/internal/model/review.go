package model

// Review 商品评价模型
type Review struct {
	BaseModel
	UserID    uint     `gorm:"index;not null" json:"user_id"`
	OrderID   uint     `gorm:"index;not null" json:"order_id"`
	ProductID uint     `gorm:"index;not null" json:"product_id"`
	Rating    int8     `gorm:"not null" json:"rating"`
	Content   string   `gorm:"size:1000" json:"content"`
	Images    string   `gorm:"size:1000" json:"images"`
	User      *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Product   *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (Review) TableName() string {
	return "reviews"
}

// ReviewCreateReq 创建评价请求
type ReviewCreateReq struct {
	OrderID   uint   `json:"order_id" binding:"required"`
	ProductID uint   `json:"product_id" binding:"required"`
	Rating    int8   `json:"rating" binding:"required,min=1,max=5"`
	Content   string `json:"content" binding:"max=1000"`
	Images    string `json:"images" binding:"max=1000"`
}

// ReviewListReq 评价列表请求
type ReviewListReq struct {
	PageQuery
	ProductID uint `form:"product_id"`
}
