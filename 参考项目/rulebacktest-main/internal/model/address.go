package model

// Address 收货地址模型
type Address struct {
	BaseModel
	UserID      uint   `gorm:"index;not null" json:"user_id"`
	Name        string `gorm:"size:50;not null" json:"name"`
	Phone       string `gorm:"size:20;not null" json:"phone"`
	Province    string `gorm:"size:50;not null" json:"province"`
	City        string `gorm:"size:50;not null" json:"city"`
	District    string `gorm:"size:50;not null" json:"district"`
	Detail      string `gorm:"size:255;not null" json:"detail"`
	IsDefault   bool   `gorm:"default:false" json:"is_default"`
}

func (Address) TableName() string {
	return "addresses"
}

// AddressCreateReq 创建地址请求
type AddressCreateReq struct {
	Name      string `json:"name" binding:"required,max=50"`
	Phone     string `json:"phone" binding:"required,max=20"`
	Province  string `json:"province" binding:"required,max=50"`
	City      string `json:"city" binding:"required,max=50"`
	District  string `json:"district" binding:"required,max=50"`
	Detail    string `json:"detail" binding:"required,max=255"`
	IsDefault bool   `json:"is_default"`
}

// AddressUpdateReq 更新地址请求
type AddressUpdateReq struct {
	Name      string `json:"name" binding:"omitempty,max=50"`
	Phone     string `json:"phone" binding:"omitempty,max=20"`
	Province  string `json:"province" binding:"omitempty,max=50"`
	City      string `json:"city" binding:"omitempty,max=50"`
	District  string `json:"district" binding:"omitempty,max=50"`
	Detail    string `json:"detail" binding:"omitempty,max=255"`
	IsDefault *bool  `json:"is_default"`
}

// FullAddress 获取完整地址
func (a *Address) FullAddress() string {
	return a.Province + a.City + a.District + a.Detail
}
