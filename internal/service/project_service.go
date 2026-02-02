package service

import (
	"encoding/json"
	"errors"

	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
)

// ProjectService 项目服务接口
type ProjectService interface {
	Create(userID uint, title, genre string, tags []string, coreConflict, characterArc, ultimateValue, worldRules string, aiSettings map[string]interface{}) (*model.Project, error)
	GetByID(id uint) (*model.Project, error)
	GetByIDWithDetails(id uint) (*model.Project, error)
	ListByUserID(userID uint, page, size int) ([]*model.Project, int64, error)
	Update(id uint, updates map[string]interface{}) (*model.Project, error)
	Delete(id uint) error
}

// projectService 项目服务实现
type projectService struct {
	projectRepo repository.ProjectRepository
}

// NewProjectService 创建项目服务实例
func NewProjectService(projectRepo repository.ProjectRepository) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
	}
}

// Create 创建项目
func (s *projectService) Create(userID uint, title, genre string, tags []string, coreConflict, characterArc, ultimateValue, worldRules string, aiSettings map[string]interface{}) (*model.Project, error) {
	// 序列化标签
	tagsJSON, _ := json.Marshal(tags)
	
	// 序列化AI设置
	aiSettingsJSON, _ := json.Marshal(aiSettings)

	project := &model.Project{
		Title:         title,
		Genre:         genre,
		Tags:          tagsJSON,
		CoreConflict:  coreConflict,
		CharacterArc:  characterArc,
		UltimateValue: ultimateValue,
		WorldRules:    worldRules,
		AISettings:    aiSettingsJSON,
		UserID:        userID,
	}

	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}

	return project, nil
}

// GetByID 根据ID获取项目
func (s *projectService) GetByID(id uint) (*model.Project, error) {
	return s.projectRepo.FindByID(id)
}

// GetByIDWithDetails 根据ID获取项目详情（含关联数据）
func (s *projectService) GetByIDWithDetails(id uint) (*model.Project, error) {
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	
	// TODO: 加载关联数据（volumes, documents, entities, templates）
	// 这里可以通过预加载实现
	
	return project, nil
}

// ListByUserID 获取用户的项目列表
func (s *projectService) ListByUserID(userID uint, page, size int) ([]*model.Project, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	return s.projectRepo.FindByUserID(userID, page, size)
}

// Update 更新项目
func (s *projectService) Update(id uint, updates map[string]interface{}) (*model.Project, error) {
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("project not found")
	}

	// 应用更新
	if title, ok := updates["title"].(string); ok {
		project.Title = title
	}
	if genre, ok := updates["genre"].(string); ok {
		project.Genre = genre
	}
	if tags, ok := updates["tags"].([]string); ok {
		tagsJSON, _ := json.Marshal(tags)
		project.Tags = tagsJSON
	}
	if coreConflict, ok := updates["core_conflict"].(string); ok {
		project.CoreConflict = coreConflict
	}
	if characterArc, ok := updates["character_arc"].(string); ok {
		project.CharacterArc = characterArc
	}
	if ultimateValue, ok := updates["ultimate_value"].(string); ok {
		project.UltimateValue = ultimateValue
	}
	if worldRules, ok := updates["world_rules"].(string); ok {
		project.WorldRules = worldRules
	}
	if aiSettings, ok := updates["ai_settings"].(map[string]interface{}); ok {
		aiSettingsJSON, _ := json.Marshal(aiSettings)
		project.AISettings = aiSettingsJSON
	}

	if err := s.projectRepo.Update(project); err != nil {
		return nil, err
	}

	return project, nil
}

// Delete 删除项目
func (s *projectService) Delete(id uint) error {
	_, err := s.projectRepo.FindByID(id)
	if err != nil {
		return errors.New("project not found")
	}
	return s.projectRepo.Delete(id)
}
