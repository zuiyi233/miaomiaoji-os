package model

// CommentType 评论类型
type CommentType int8

const (
	CommentTypeProduct CommentType = 1 // 商品评论
	CommentTypeReview  CommentType = 2 // 评价回复
)

// Comment 评论模型
type Comment struct {
	BaseModel
	UserID   uint        `gorm:"index;not null" json:"user_id"`
	Type     CommentType `gorm:"not null" json:"type"`
	TargetID uint        `gorm:"index;not null" json:"target_id"`
	ParentID uint        `gorm:"index;default:0" json:"parent_id"`
	ReplyUID uint        `gorm:"default:0" json:"reply_uid"`
	Content  string      `gorm:"size:500;not null" json:"content"`
	User     *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ReplyTo  *User       `gorm:"foreignKey:ReplyUID" json:"reply_to,omitempty"`
	Replies  []Comment   `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

func (Comment) TableName() string {
	return "comments"
}

// CommentCreateReq 创建评论请求
type CommentCreateReq struct {
	Type     CommentType `json:"type" binding:"required,oneof=1 2"`
	TargetID uint        `json:"target_id" binding:"required"`
	ParentID uint        `json:"parent_id"`
	ReplyUID uint        `json:"reply_uid"`
	Content  string      `json:"content" binding:"required,max=500"`
}

// CommentListReq 评论列表请求
type CommentListReq struct {
	PageQuery
	Type     CommentType `form:"type" binding:"required,oneof=1 2"`
	TargetID uint        `form:"target_id" binding:"required"`
}
