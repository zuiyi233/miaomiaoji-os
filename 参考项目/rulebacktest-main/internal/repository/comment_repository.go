package repository

import (
	"sync"

	"gorm.io/gorm"
	"rulebacktest/internal/model"
	"rulebacktest/pkg/database"
)

var (
	commentRepoInstance *CommentRepository
	commentRepoOnce     sync.Once
)

// CommentRepository 评论数据访问层
type CommentRepository struct {
	*BaseRepository
}

// NewCommentRepository 创建CommentRepository实例
func NewCommentRepository(base *BaseRepository) *CommentRepository {
	return &CommentRepository{BaseRepository: base}
}

// GetCommentRepository 获取CommentRepository单例
func GetCommentRepository() *CommentRepository {
	commentRepoOnce.Do(func() {
		commentRepoInstance = &CommentRepository{
			BaseRepository: &BaseRepository{db: database.GetDB()},
		}
	})
	return commentRepoInstance
}

// Create 创建评论
func (r *CommentRepository) Create(comment *model.Comment) error {
	return r.DB().Create(comment).Error
}

// GetByID 根据ID获取评论
func (r *CommentRepository) GetByID(id uint) (*model.Comment, error) {
	var comment model.Comment
	err := r.DB().Preload("User").First(&comment, id).Error
	return &comment, err
}

// ListByTarget 获取目标评论列表（只获取顶级评论）
func (r *CommentRepository) ListByTarget(commentType model.CommentType, targetID uint, page, pageSize int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64

	query := r.DB().Model(&model.Comment{}).
		Where("type = ? AND target_id = ? AND parent_id = 0", commentType, targetID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Preload("ReplyTo").Order("id ASC").Limit(3)
		}).
		Scopes(r.Paginate(page, pageSize)).
		Order("id DESC").
		Find(&comments).Error

	return comments, total, err
}

// ListReplies 获取评论的回复列表
func (r *CommentRepository) ListReplies(parentID uint, page, pageSize int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64

	query := r.DB().Model(&model.Comment{}).Where("parent_id = ?", parentID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").Preload("ReplyTo").
		Scopes(r.Paginate(page, pageSize)).
		Order("id ASC").
		Find(&comments).Error

	return comments, total, err
}

// Delete 删除评论
func (r *CommentRepository) Delete(id uint) error {
	return r.DB().Delete(&model.Comment{}, id).Error
}

// DeleteByParentID 删除子评论
func (r *CommentRepository) DeleteByParentID(parentID uint) error {
	return r.DB().Where("parent_id = ?", parentID).Delete(&model.Comment{}).Error
}

// CountByTarget 统计目标评论数
func (r *CommentRepository) CountByTarget(commentType model.CommentType, targetID uint) (int64, error) {
	var count int64
	err := r.DB().Model(&model.Comment{}).
		Where("type = ? AND target_id = ?", commentType, targetID).
		Count(&count).Error
	return count, err
}
