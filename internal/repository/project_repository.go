package repository

import (
	"novel-agent-os-backend/internal/model"
)

// ProjectRepository 项目仓储接口
type ProjectRepository interface {
	Create(project *model.Project) error
	FindByID(id uint) (*model.Project, error)
	FindByUserID(userID uint, page, size int) ([]*model.Project, int64, error)
	Update(project *model.Project) error
	Delete(id uint) error
}

// projectRepository 项目仓储实现
type projectRepository struct {
	*BaseRepository
}

// NewProjectRepository 创建项目仓储实例
func NewProjectRepository() ProjectRepository {
	return &projectRepository{
		BaseRepository: NewBaseRepository(),
	}
}

// Create 创建项目
func (r *projectRepository) Create(project *model.Project) error {
	return r.DB.Create(project).Error
}

// FindByID 根据ID查找项目
func (r *projectRepository) FindByID(id uint) (*model.Project, error) {
	var project model.Project
	if err := r.DB.First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// FindByUserID 根据用户ID查找项目列表
func (r *projectRepository) FindByUserID(userID uint, page, size int) ([]*model.Project, int64, error) {
	var projects []*model.Project
	var total int64

	if err := r.DB.Model(&model.Project{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := r.DB.Where("user_id = ?", userID).Offset(offset).Limit(size).Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// Update 更新项目
func (r *projectRepository) Update(project *model.Project) error {
	return r.DB.Save(project).Error
}

// Delete 删除项目（软删除）
func (r *projectRepository) Delete(id uint) error {
	return r.DB.Delete(&model.Project{}, id).Error
}
