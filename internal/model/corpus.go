package model

import (
	"gorm.io/datatypes"
)

// CorpusStory 语料库故事模型
type CorpusStory struct {
	BaseModel
	Title     string         `gorm:"size:200;not null" json:"title"`
	Genre     string         `gorm:"size:50;index" json:"genre"`
	FilePath  string         `gorm:"size:500" json:"file_path"`
	FileSize  int64          `json:"file_size"`
	WordCount int            `json:"word_count"`
	Metadata  datatypes.JSON `json:"metadata"`
}

// TableName 指定表名
func (CorpusStory) TableName() string {
	return "corpus_stories"
}
