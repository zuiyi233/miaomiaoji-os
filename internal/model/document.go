package model

import (
	"gorm.io/datatypes"
)

// Bookmark 书签结构
type Bookmark struct {
	Title    string `json:"title"`
	Position int    `json:"position"`
	Note     string `json:"note"`
}

// Document 文档/章节模型
type Document struct {
	BaseModel
	Title                string         `gorm:"size:200;not null" json:"title"`
	Content              string         `gorm:"type:text" json:"content"`
	Summary              string         `gorm:"type:text" json:"summary"`
	Status               string         `gorm:"size:20;default:草稿" json:"status"` // 草稿/修改中/完成
	OrderIndex           int            `gorm:"index" json:"order_index"`
	Bookmarks            datatypes.JSON `json:"bookmarks"` // Bookmark[]
	TimeNode             string         `gorm:"size:100" json:"time_node"`
	Duration             string         `gorm:"size:50" json:"duration"`
	TargetWordCount      int            `json:"target_word_count"`
	ChapterGoal          string         `gorm:"type:text" json:"chapter_goal"`
	CorePlot             string         `gorm:"type:text" json:"core_plot"`
	Hook                 string         `gorm:"type:text" json:"hook"`
	CauseEffect          string         `gorm:"type:text" json:"cause_effect"`
	ForeshadowingDetails string         `gorm:"type:text" json:"foreshadowing_details"`
	ProjectID            uint           `gorm:"index;not null" json:"project_id"`
	VolumeID             uint           `gorm:"index" json:"volume_id"`

	// 关联
	EntityRefs []DocumentEntityRef `gorm:"foreignKey:DocumentID" json:"entity_refs,omitempty"`
}

// TableName 指定表名
func (Document) TableName() string {
	return "documents"
}

// DocumentEntityRef 文档与实体关联表（替代JSONB）
type DocumentEntityRef struct {
	BaseModelWithoutSoftDelete
	DocumentID uint           `gorm:"index;not null" json:"document_id"`
	EntityID   uint           `gorm:"index;not null" json:"entity_id"`
	RefType    string         `gorm:"size:20;default:mention" json:"ref_type"`
	Metadata   datatypes.JSON `json:"metadata"`
	
	// 关联
	Document Document `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
	Entity   Entity   `gorm:"foreignKey:EntityID" json:"entity,omitempty"`
}

// TableName 指定表名
func (DocumentEntityRef) TableName() string {
	return "document_entity_refs"
}
