package model

// File 文件元信息模型
type File struct {
	BaseModel
	FileName    string `gorm:"size:255;not null" json:"file_name"`
	FileType    string `gorm:"size:20;index;not null" json:"file_type"` // upload/export
	ContentType string `gorm:"size:100" json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	StorageKey  string `gorm:"size:500;uniqueIndex;not null" json:"storage_key"`
	SHA256      string `gorm:"size:64;index" json:"sha256"`
	UserID      uint   `gorm:"index;not null" json:"user_id"`
	ProjectID   *uint  `gorm:"index" json:"project_id,omitempty"`
}

// TableName 指定表名
func (File) TableName() string {
	return "files"
}
