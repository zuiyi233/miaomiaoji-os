// Package service 项目服务
package service

import (
	"encoding/json"
	"errors"

	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/pkg/logger"

	"gorm.io/gorm"
)

// ProjectService 项目服务接口
type ProjectService interface {
	Create(userID uint, title, genre string, tags []string, coreConflict, characterArc, ultimateValue, worldRules string, aiSettings map[string]interface{}) (*model.Project, error)
	CreateOrUpdateSnapshot(userID uint, externalID string, snapshot map[string]interface{}, title string, aiSettings map[string]interface{}) (*model.Project, error)
	GetByID(id uint) (*model.Project, error)
	GetByIDWithDetails(id uint) (*ProjectDetailResult, error)
	ListByUserID(userID uint, page, size int) ([]*model.Project, int64, error)
	Update(id uint, updates map[string]interface{}) (*model.Project, error)
	Delete(id uint) error
	ExportProject(id uint) (*ProjectExport, error)
}

// ProjectDetailResult 项目详情结果
type ProjectDetailResult struct {
	Project   *model.Project    `json:"project"`
	Volumes   []*model.Volume   `json:"volumes"`
	Documents []*model.Document `json:"documents"`
	Entities  []*model.Entity   `json:"entities"`
	Templates []*model.Template `json:"templates"`
}

// ProjectExport 项目导出数据结构
type ProjectExport struct {
	Project   *model.Project         `json:"project"`
	Volumes   []*model.Volume        `json:"volumes"`
	Documents []*model.Document      `json:"documents"`
	Entities  []*model.Entity        `json:"entities"`
	Templates []*model.Template      `json:"templates"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// projectService 项目服务实现
type projectService struct {
	projectRepo  repository.ProjectRepository
	volumeRepo   repository.VolumeRepository
	documentRepo repository.DocumentRepository
	entityRepo   repository.EntityRepository
	templateRepo repository.TemplateRepository
}

// NewProjectService 创建项目服务实例
func NewProjectService(
	projectRepo repository.ProjectRepository,
	volumeRepo repository.VolumeRepository,
	documentRepo repository.DocumentRepository,
	entityRepo repository.EntityRepository,
	templateRepo repository.TemplateRepository,
) ProjectService {
	return &projectService{
		projectRepo:  projectRepo,
		volumeRepo:   volumeRepo,
		documentRepo: documentRepo,
		entityRepo:   entityRepo,
		templateRepo: templateRepo,
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

// CreateOrUpdateSnapshot 创建或更新项目快照
func (s *projectService) CreateOrUpdateSnapshot(userID uint, externalID string, snapshot map[string]interface{}, title string, aiSettings map[string]interface{}) (*model.Project, error) {
	if externalID == "" {
		return nil, errors.New("external_id required")
	}

	snapshotJSON, _ := json.Marshal(snapshot)
	aiSettingsJSON, _ := json.Marshal(aiSettings)

	project, err := s.projectRepo.FindByUserAndExternalID(userID, externalID)
	if err == nil {
		project.Snapshot = snapshotJSON
		if aiSettings != nil {
			project.AISettings = aiSettingsJSON
		}
		if title != "" {
			project.Title = title
		}
		if err := s.projectRepo.Update(project); err != nil {
			return nil, err
		}
		return project, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	project = &model.Project{
		Title:      title,
		Snapshot:   snapshotJSON,
		UserID:     userID,
		ExternalID: &externalID,
	}
	if aiSettings != nil {
		project.AISettings = aiSettingsJSON
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
func (s *projectService) GetByIDWithDetails(id uint) (*ProjectDetailResult, error) {
	// 加载项目基本信息
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	logger.Debug("加载项目关联数据", logger.Uint("project_id", id))

	// 加载卷列表
	volumes, _, _ := s.volumeRepo.FindByProjectID(id, 1, 1000)

	// 加载文档列表
	documents, _, _ := s.documentRepo.FindByProjectID(id, 1, 1000)

	// 加载实体列表
	entities, _, _ := s.entityRepo.FindByProjectID(id, 1, 1000)

	// 加载模板列表
	templates, _, _ := s.templateRepo.FindByProjectID(id, 1, 1000)

	result := &ProjectDetailResult{
		Project:   project,
		Volumes:   volumes,
		Documents: documents,
		Entities:  entities,
		Templates: templates,
	}

	logger.Info("项目详情加载完成",
		logger.Uint("project_id", id),
		logger.Int("volumes_count", len(volumes)),
		logger.Int("documents_count", len(documents)),
		logger.Int("entities_count", len(entities)),
		logger.Int("templates_count", len(templates)))

	return result, nil
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

// ExportProject 导出项目
func (s *projectService) ExportProject(id uint) (*ProjectExport, error) {
	// 获取项目详情
	details, err := s.GetByIDWithDetails(id)
	if err != nil {
		return nil, err
	}

	logger.Info("开始导出项目", logger.Uint("project_id", id))

	export := &ProjectExport{
		Project:   details.Project,
		Volumes:   details.Volumes,
		Documents: details.Documents,
		Entities:  details.Entities,
		Templates: details.Templates,
		Metadata: map[string]interface{}{
			"version":         "1.0",
			"export_type":     "full",
			"volumes_count":   len(details.Volumes),
			"documents_count": len(details.Documents),
			"entities_count":  len(details.Entities),
			"templates_count": len(details.Templates),
		},
	}

	logger.Info("项目导出完成",
		logger.Uint("project_id", id),
		logger.Int("total_items", len(details.Volumes)+len(details.Documents)+len(details.Entities)+len(details.Templates)))

	return export, nil
}
