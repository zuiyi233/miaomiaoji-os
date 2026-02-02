package repository

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/pkg/database"
)

// EntityRepository 实体数据访问接口
type EntityRepository interface {
	Create(entity *model.Entity) error
	FindByID(id uint) (*model.Entity, error)
	FindByProjectID(projectID uint, page, size int) ([]*model.Entity, int64, error)
	FindByType(projectID uint, entityType string, page, size int) ([]*model.Entity, int64, error)
	FindByTag(projectID uint, tag string, page, size int) ([]*model.Entity, int64, error)
	FindByTypeAndTag(projectID uint, entityType, tag string, page, size int) ([]*model.Entity, int64, error)
	Update(entity *model.Entity) error
	Delete(id uint) error
	AddTag(entityID uint, tag string) error
	RemoveTag(entityID uint, tag string) error
	GetTags(entityID uint) ([]*model.EntityTag, error)
	CreateLink(sourceID, targetID uint, linkType, relationName string) error
	DeleteLink(sourceID, targetID uint) error
	GetLinks(sourceID uint) ([]*model.EntityLink, error)
	IncrementReferenceCount(entityID uint) error
	DecrementReferenceCount(entityID uint) error
}

// entityRepository 实体数据访问实现
type entityRepository struct{}

// NewEntityRepository 创建实体仓库实例
func NewEntityRepository() EntityRepository {
	return &entityRepository{}
}

// Create 创建实体
func (r *entityRepository) Create(entity *model.Entity) error {
	return database.GetDB().Create(entity).Error
}

// FindByID 根据ID查找实体
func (r *entityRepository) FindByID(id uint) (*model.Entity, error) {
	var entity model.Entity
	err := database.GetDB().Preload("Tags").Preload("Links").First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindByProjectID 根据项目ID查找实体列表
func (r *entityRepository) FindByProjectID(projectID uint, page, size int) ([]*model.Entity, int64, error) {
	var entities []*model.Entity
	var total int64

	db := database.GetDB().Model(&model.Entity{}).Where("project_id = ?", projectID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Preload("Tags").
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// FindByType 根据类型查找实体列表
func (r *entityRepository) FindByType(projectID uint, entityType string, page, size int) ([]*model.Entity, int64, error) {
	var entities []*model.Entity
	var total int64

	db := database.GetDB().Model(&model.Entity{}).
		Where("project_id = ? AND entity_type = ?", projectID, entityType)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Preload("Tags").
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// FindByTag 根据标签查找实体列表
func (r *entityRepository) FindByTag(projectID uint, tag string, page, size int) ([]*model.Entity, int64, error) {
	var entities []*model.Entity
	var total int64

	subQuery := database.GetDB().Model(&model.EntityTag{}).
		Select("entity_id").
		Where("tag = ?", tag)

	db := database.GetDB().Model(&model.Entity{}).
		Where("project_id = ? AND id IN (?)", projectID, subQuery)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Preload("Tags").
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// FindByTypeAndTag 根据类型和标签查找实体列表
func (r *entityRepository) FindByTypeAndTag(projectID uint, entityType, tag string, page, size int) ([]*model.Entity, int64, error) {
	var entities []*model.Entity
	var total int64

	subQuery := database.GetDB().Model(&model.EntityTag{}).
		Select("entity_id").
		Where("tag = ?", tag)

	db := database.GetDB().Model(&model.Entity{}).
		Where("project_id = ? AND entity_type = ? AND id IN (?)", projectID, entityType, subQuery)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Preload("Tags").
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// Update 更新实体
func (r *entityRepository) Update(entity *model.Entity) error {
	return database.GetDB().Save(entity).Error
}

// Delete 删除实体（软删除）
func (r *entityRepository) Delete(id uint) error {
	return database.GetDB().Delete(&model.Entity{}, id).Error
}

// AddTag 添加标签
func (r *entityRepository) AddTag(entityID uint, tag string) error {
	entityTag := &model.EntityTag{
		EntityID: entityID,
		Tag:      tag,
	}
	return database.GetDB().Create(entityTag).Error
}

// RemoveTag 移除标签
func (r *entityRepository) RemoveTag(entityID uint, tag string) error {
	return database.GetDB().
		Where("entity_id = ? AND tag = ?", entityID, tag).
		Delete(&model.EntityTag{}).Error
}

// GetTags 获取实体的所有标签
func (r *entityRepository) GetTags(entityID uint) ([]*model.EntityTag, error) {
	var tags []*model.EntityTag
	err := database.GetDB().
		Where("entity_id = ?", entityID).
		Find(&tags).Error
	return tags, err
}

// CreateLink 创建实体关联
func (r *entityRepository) CreateLink(sourceID, targetID uint, linkType, relationName string) error {
	link := &model.EntityLink{
		SourceID:     sourceID,
		TargetID:     targetID,
		Type:         linkType,
		RelationName: relationName,
	}
	return database.GetDB().Create(link).Error
}

// DeleteLink 删除实体关联
func (r *entityRepository) DeleteLink(sourceID, targetID uint) error {
	return database.GetDB().
		Where("source_id = ? AND target_id = ?", sourceID, targetID).
		Delete(&model.EntityLink{}).Error
}

// GetLinks 获取实体的所有关联
func (r *entityRepository) GetLinks(sourceID uint) ([]*model.EntityLink, error) {
	var links []*model.EntityLink
	err := database.GetDB().
		Where("source_id = ?", sourceID).
		Preload("Target").
		Find(&links).Error
	return links, err
}

// IncrementReferenceCount 增加引用计数
func (r *entityRepository) IncrementReferenceCount(entityID uint) error {
	return database.GetDB().Model(&model.Entity{}).
		Where("id = ?", entityID).
		UpdateColumn("reference_count", database.GetDB().Raw("reference_count + 1")).Error
}

// DecrementReferenceCount 减少引用计数
func (r *entityRepository) DecrementReferenceCount(entityID uint) error {
	return database.GetDB().Model(&model.Entity{}).
		Where("id = ?", entityID).
		UpdateColumn("reference_count", database.GetDB().Raw("reference_count - 1")).Error
}
