package service

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
)

// TemplateService 模板服务接口
type TemplateService interface {
	Create(projectID uint, name, description, category, template string) (*model.Template, error)
	GetByID(id uint) (*model.Template, error)
	ListByProjectID(projectID uint, page, size int) ([]*model.Template, int64, error)
	ListSystemTemplates(page, size int) ([]*model.Template, int64, error)
	ListByCategory(projectID uint, category string, page, size int) ([]*model.Template, int64, error)
	Update(id uint, updates map[string]interface{}) (*model.Template, error)
	Delete(id uint) error
}

// templateService 模板服务实现
type templateService struct {
	templateRepo repository.TemplateRepository
	projectRepo  repository.ProjectRepository
}

// NewTemplateService 创建模板服务实例
func NewTemplateService(templateRepo repository.TemplateRepository, projectRepo repository.ProjectRepository) TemplateService {
	return &templateService{
		templateRepo: templateRepo,
		projectRepo:  projectRepo,
	}
}

// Create 创建模板
func (s *templateService) Create(projectID uint, name, description, category, template string) (*model.Template, error) {
	// 验证项目是否存在（projectID为0表示系统模板）
	if projectID > 0 {
		_, err := s.projectRepo.FindByID(projectID)
		if err != nil {
			return nil, errors.ErrProjectNotFound
		}
	}

	// 验证分类
	validCategories := map[string]bool{
		"logic":     true,
		"style":     true,
		"content":   true,
		"character": true,
	}
	if category != "" && !validCategories[category] {
		return nil, errors.ErrValidationError
	}

	tmpl := &model.Template{
		ProjectID:   projectID,
		Name:        name,
		Description: description,
		Category:    category,
		Template:    template,
	}

	if err := s.templateRepo.Create(tmpl); err != nil {
		logger.Error("创建模板失败", logger.Err(err))
		return nil, errors.ErrInternalServer
	}

	return tmpl, nil
}

// GetByID 根据ID获取模板
func (s *templateService) GetByID(id uint) (*model.Template, error) {
	template, err := s.templateRepo.FindByID(id)
	if err != nil {
		return nil, errors.ErrTemplateNotFound
	}
	return template, nil
}

// ListByProjectID 根据项目ID获取模板列表
func (s *templateService) ListByProjectID(projectID uint, page, size int) ([]*model.Template, int64, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, 0, errors.ErrProjectNotFound
	}

	return s.templateRepo.FindByProjectID(projectID, page, size)
}

// ListSystemTemplates 获取系统模板列表
func (s *templateService) ListSystemTemplates(page, size int) ([]*model.Template, int64, error) {
	return s.templateRepo.FindSystemTemplates(page, size)
}

// ListByCategory 根据分类获取模板列表
func (s *templateService) ListByCategory(projectID uint, category string, page, size int) ([]*model.Template, int64, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, 0, errors.ErrProjectNotFound
	}

	return s.templateRepo.FindByCategory(projectID, category, page, size)
}

// Update 更新模板
func (s *templateService) Update(id uint, updates map[string]interface{}) (*model.Template, error) {
	template, err := s.templateRepo.FindByID(id)
	if err != nil {
		return nil, errors.ErrTemplateNotFound
	}

	// 应用更新
	if name, ok := updates["name"].(string); ok {
		template.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		template.Description = description
	}
	if category, ok := updates["category"].(string); ok {
		template.Category = category
	}
	if tmpl, ok := updates["template"].(string); ok {
		template.Template = tmpl
	}

	if err := s.templateRepo.Update(template); err != nil {
		logger.Error("更新模板失败", logger.Err(err))
		return nil, errors.ErrInternalServer
	}

	return template, nil
}

// Delete 删除模板
func (s *templateService) Delete(id uint) error {
	_, err := s.templateRepo.FindByID(id)
	if err != nil {
		return errors.ErrTemplateNotFound
	}

	if err := s.templateRepo.Delete(id); err != nil {
		logger.Error("删除模板失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}
