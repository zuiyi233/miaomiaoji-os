package model

// Volume 卷模型
type Volume struct {
	BaseModel
	Title                  string     `gorm:"size:200;not null" json:"title"`
	OrderIndex             int        `gorm:"index" json:"order_index"`
	Theme                  string     `gorm:"type:text" json:"theme"`
	CoreGoal               string     `gorm:"type:text" json:"core_goal"`
	Boundaries             string     `gorm:"type:text" json:"boundaries"`
	ChapterLinkageLogic    string     `gorm:"type:text" json:"chapter_linkage_logic"`
	VolumeSpecificSettings string     `gorm:"type:text" json:"volume_specific_settings"`
	PlotRoadmap            string     `gorm:"type:text" json:"plot_roadmap"`
	ProjectID              uint       `gorm:"index;not null" json:"project_id"`
	
	// 关联
	Documents []Document `gorm:"foreignKey:VolumeID" json:"documents,omitempty"`
}

// TableName 指定表名
func (Volume) TableName() string {
	return "volumes"
}
