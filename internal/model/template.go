package model

// Template AI模板模型
type Template struct {
	BaseModel
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Category    string `gorm:"size:20" json:"category"` // logic/style/content/character
	Template    string `gorm:"type:text" json:"template"`
	ProjectID   uint   `gorm:"index" json:"project_id"` // 0表示系统模板
}

// TableName 指定表名
func (Template) TableName() string {
	return "templates"
}
