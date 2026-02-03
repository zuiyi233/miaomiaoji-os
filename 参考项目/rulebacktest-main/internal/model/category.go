package model

// Category 商品分类模型
type Category struct {
	BaseModel
	Name     string `gorm:"size:50;not null" json:"name"`
	ParentID uint   `gorm:"default:0" json:"parent_id"`
	Sort     int    `gorm:"default:0" json:"sort"`
	Status   Status `gorm:"default:1" json:"status"`
}

func (Category) TableName() string {
	return "categories"
}

// CategoryCreateReq 创建分类请求
type CategoryCreateReq struct {
	Name     string `json:"name" binding:"required,max=50"`
	ParentID uint   `json:"parent_id"`
	Sort     int    `json:"sort"`
}

// CategoryUpdateReq 更新分类请求
type CategoryUpdateReq struct {
	Name   string  `json:"name" binding:"omitempty,max=50"`
	Sort   *int    `json:"sort"`
	Status *Status `json:"status"`
}
