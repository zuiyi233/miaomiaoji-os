package repository

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/pkg/database"
)

// TemplateRepository 模板数据访问接口
type TemplateRepository interface {
	Create(template *model.Template) error
	FindByID(id uint) (*model.Template, error)
	FindByProjectID(projectID uint, page, size int) ([]*model.Template, int64, error)
	FindSystemTemplates(page, size int) ([]*model.Template, int64, error)
	FindByCategory(projectID uint, category string, page, size int) ([]*model.Template, int64, error)
	Update(template *model.Template) error
	Delete(id uint) error
}

// templateRepository 模板数据访问实现
type templateRepository struct{}

// NewTemplateRepository 创建模板仓库实例
func NewTemplateRepository() TemplateRepository {
	return &templateRepository{}
}

// Create 创建模板
func (r *templateRepository) Create(template *model.Template) error {
	return database.GetDB().Create(template).Error
}

// FindByID 根据ID查找模板
func (r *templateRepository) FindByID(id uint) (*model.Template, error) {
	var template model.Template
	err := database.GetDB().First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// FindByProjectID 根据项目ID查找模板列表
func (r *templateRepository) FindByProjectID(projectID uint, page, size int) ([]*model.Template, int64, error) {
	var templates []*model.Template
	var total int64

	db := database.GetDB().Model(&model.Template{}).Where("project_id = ?", projectID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// FindSystemTemplates 查找系统模板列表
func (r *templateRepository) FindSystemTemplates(page, size int) ([]*model.Template, int64, error) {
	var templates []*model.Template
	var total int64

	db := database.GetDB().Model(&model.Template{}).Where("project_id = ?", 0)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// FindByCategory 根据分类查找模板列表
func (r *templateRepository) FindByCategory(projectID uint, category string, page, size int) ([]*model.Template, int64, error) {
	var templates []*model.Template
	var total int64

	db := database.GetDB().Model(&model.Template{}).
		Where("(project_id = ? OR project_id = 0) AND category = ?", projectID, category)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// Update 更新模板
func (r *templateRepository) Update(template *model.Template) error {
	return database.GetDB().Save(template).Error
}

// Delete 删除模板（软删除）
func (r *templateRepository) Delete(id uint) error {
	return database.GetDB().Delete(&model.Template{}, id).Error
}
