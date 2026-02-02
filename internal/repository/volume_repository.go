package repository

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/pkg/database"
)

// VolumeRepository 卷数据访问接口
type VolumeRepository interface {
	Create(volume *model.Volume) error
	FindByID(id uint) (*model.Volume, error)
	FindByProjectID(projectID uint, page, size int) ([]*model.Volume, int64, error)
	Update(volume *model.Volume) error
	Delete(id uint) error
	FindByProjectIDWithDocuments(projectID uint) ([]*model.Volume, error)
	UpdateOrderIndex(id uint, orderIndex int) error
	ReorderVolumes(projectID uint, volumeIDs []uint) error
}

// volumeRepository 卷数据访问实现
type volumeRepository struct{}

// NewVolumeRepository 创建卷仓库实例
func NewVolumeRepository() VolumeRepository {
	return &volumeRepository{}
}

// Create 创建卷
func (r *volumeRepository) Create(volume *model.Volume) error {
	return database.GetDB().Create(volume).Error
}

// FindByID 根据ID查找卷
func (r *volumeRepository) FindByID(id uint) (*model.Volume, error) {
	var volume model.Volume
	err := database.GetDB().Preload("Documents").First(&volume, id).Error
	if err != nil {
		return nil, err
	}
	return &volume, nil
}

// FindByProjectID 根据项目ID查找卷列表
func (r *volumeRepository) FindByProjectID(projectID uint, page, size int) ([]*model.Volume, int64, error) {
	var volumes []*model.Volume
	var total int64

	db := database.GetDB().Model(&model.Volume{}).Where("project_id = ?", projectID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("order_index ASC, id ASC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&volumes).Error; err != nil {
		return nil, 0, err
	}

	return volumes, total, nil
}

// Update 更新卷
func (r *volumeRepository) Update(volume *model.Volume) error {
	return database.GetDB().Save(volume).Error
}

// Delete 删除卷（软删除）
func (r *volumeRepository) Delete(id uint) error {
	return database.GetDB().Delete(&model.Volume{}, id).Error
}

// FindByProjectIDWithDocuments 根据项目ID查找卷及其文档
func (r *volumeRepository) FindByProjectIDWithDocuments(projectID uint) ([]*model.Volume, error) {
	var volumes []*model.Volume
	err := database.GetDB().
		Where("project_id = ?", projectID).
		Order("order_index ASC, id ASC").
		Preload("Documents").
		Find(&volumes).Error
	return volumes, err
}

// UpdateOrderIndex 更新卷的顺序索引
func (r *volumeRepository) UpdateOrderIndex(id uint, orderIndex int) error {
	return database.GetDB().Model(&model.Volume{}).Where("id = ?", id).Update("order_index", orderIndex).Error
}

// ReorderVolumes 批量重新排序卷
func (r *volumeRepository) ReorderVolumes(projectID uint, volumeIDs []uint) error {
	tx := database.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for index, volumeID := range volumeIDs {
		if err := tx.Model(&model.Volume{}).
			Where("id = ? AND project_id = ?", volumeID, projectID).
			Update("order_index", index+1).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
