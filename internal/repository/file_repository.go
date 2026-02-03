package repository

import (
	"novel-agent-os-backend/internal/model"

	"gorm.io/gorm"
)

type FileRepository interface {
	Create(file *model.File) error
	GetByID(id uint) (*model.File, error)
	GetByStorageKey(key string) (*model.File, error)
	Update(file *model.File) error
	Delete(id uint) error
	ListByUserID(userID uint, page, pageSize int) ([]*model.File, int64, error)
	ListByProjectID(projectID uint, page, pageSize int) ([]*model.File, int64, error)
	ListByType(fileType string, page, pageSize int) ([]*model.File, int64, error)
}

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(file *model.File) error {
	return r.db.Create(file).Error
}

func (r *fileRepository) GetByID(id uint) (*model.File, error) {
	var file model.File
	err := r.db.First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepository) GetByStorageKey(key string) (*model.File, error) {
	var file model.File
	err := r.db.Where("storage_key = ?", key).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepository) Update(file *model.File) error {
	return r.db.Save(file).Error
}

func (r *fileRepository) Delete(id uint) error {
	return r.db.Delete(&model.File{}, id).Error
}

func (r *fileRepository) ListByUserID(userID uint, page, pageSize int) ([]*model.File, int64, error) {
	var files []*model.File
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.File{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("user_id = ?", userID).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&files).Error

	return files, total, err
}

func (r *fileRepository) ListByProjectID(projectID uint, page, pageSize int) ([]*model.File, int64, error) {
	var files []*model.File
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.File{}).Where("project_id = ?", projectID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("project_id = ?", projectID).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&files).Error

	return files, total, err
}

func (r *fileRepository) ListByType(fileType string, page, pageSize int) ([]*model.File, int64, error) {
	var files []*model.File
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.File{}).Where("file_type = ?", fileType).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("file_type = ?", fileType).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&files).Error

	return files, total, err
}
