package service

import (
	"encoding/json"
	stderrors "errors"

	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
)

// EntityService 实体服务接口
type EntityService interface {
	Create(projectID uint, entityType, title, subtitle, content, voiceStyle, importance string, customFields []model.EntityCustomField) (*model.Entity, error)
	GetByID(id uint) (*model.Entity, error)
	ListByProjectID(projectID uint, page, size int) ([]*model.Entity, int64, error)
	ListByType(projectID uint, entityType string, page, size int) ([]*model.Entity, int64, error)
	ListByTag(projectID uint, tag string, page, size int) ([]*model.Entity, int64, error)
	ListByTypeAndTag(projectID uint, entityType, tag string, page, size int) ([]*model.Entity, int64, error)
	Update(id uint, updates map[string]interface{}) (*model.Entity, error)
	Delete(id uint) error
	AddTag(id uint, tag string) error
	RemoveTag(id uint, tag string) error
	CreateLink(id, targetID uint, linkType, relationName string) error
	DeleteLink(id, targetID uint) error
}

// entityService 实体服务实现
type entityService struct {
	entityRepo  repository.EntityRepository
	projectRepo repository.ProjectRepository
}

// NewEntityService 创建实体服务实例
func NewEntityService(entityRepo repository.EntityRepository, projectRepo repository.ProjectRepository) EntityService {
	return &entityService{
		entityRepo:  entityRepo,
		projectRepo: projectRepo,
	}
}

// Create 创建实体
func (s *entityService) Create(projectID uint, entityType, title, subtitle, content, voiceStyle, importance string, customFields []model.EntityCustomField) (*model.Entity, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, errors.ErrProjectNotFound
	}

	// 验证实体类型
	validTypes := map[string]bool{
		"character":    true,
		"setting":      true,
		"organization": true,
		"item":         true,
		"magic":        true,
		"event":        true,
	}
	if !validTypes[entityType] {
		return nil, errors.ErrValidationError
	}

	// 序列化自定义字段
	customFieldsJSON, _ := json.Marshal(customFields)

	entity := &model.Entity{
		ProjectID:    projectID,
		EntityType:   entityType,
		Title:        title,
		Subtitle:     subtitle,
		Content:      content,
		VoiceStyle:   voiceStyle,
		Importance:   importance,
		CustomFields: customFieldsJSON,
	}

	if err := s.entityRepo.Create(entity); err != nil {
		logger.Error("创建实体失败", logger.Err(err))
		return nil, errors.ErrInternalServer
	}

	return entity, nil
}

// GetByID 根据ID获取实体
func (s *entityService) GetByID(id uint) (*model.Entity, error) {
	entity, err := s.entityRepo.FindByID(id)
	if err != nil {
		return nil, errors.ErrEntityNotFound
	}
	return entity, nil
}

// ListByProjectID 根据项目ID获取实体列表
func (s *entityService) ListByProjectID(projectID uint, page, size int) ([]*model.Entity, int64, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, 0, errors.ErrProjectNotFound
	}

	return s.entityRepo.FindByProjectID(projectID, page, size)
}

// ListByType 根据类型获取实体列表
func (s *entityService) ListByType(projectID uint, entityType string, page, size int) ([]*model.Entity, int64, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, 0, errors.ErrProjectNotFound
	}

	return s.entityRepo.FindByType(projectID, entityType, page, size)
}

// ListByTag 根据标签获取实体列表
func (s *entityService) ListByTag(projectID uint, tag string, page, size int) ([]*model.Entity, int64, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, 0, errors.ErrProjectNotFound
	}

	return s.entityRepo.FindByTag(projectID, tag, page, size)
}

// ListByTypeAndTag 根据类型和标签获取实体列表
func (s *entityService) ListByTypeAndTag(projectID uint, entityType, tag string, page, size int) ([]*model.Entity, int64, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, 0, errors.ErrProjectNotFound
	}

	return s.entityRepo.FindByTypeAndTag(projectID, entityType, tag, page, size)
}

// Update 更新实体
func (s *entityService) Update(id uint, updates map[string]interface{}) (*model.Entity, error) {
	entity, err := s.entityRepo.FindByID(id)
	if err != nil {
		return nil, errors.ErrEntityNotFound
	}

	// 应用更新
	if entityType, ok := updates["entity_type"].(string); ok {
		entity.EntityType = entityType
	}
	if title, ok := updates["title"].(string); ok {
		entity.Title = title
	}
	if subtitle, ok := updates["subtitle"].(string); ok {
		entity.Subtitle = subtitle
	}
	if content, ok := updates["content"].(string); ok {
		entity.Content = content
	}
	if voiceStyle, ok := updates["voice_style"].(string); ok {
		entity.VoiceStyle = voiceStyle
	}
	if importance, ok := updates["importance"].(string); ok {
		entity.Importance = importance
	}
	if customFields, ok := updates["custom_fields"].([]model.EntityCustomField); ok {
		customFieldsJSON, _ := json.Marshal(customFields)
		entity.CustomFields = customFieldsJSON
	}

	if err := s.entityRepo.Update(entity); err != nil {
		logger.Error("更新实体失败", logger.Err(err))
		return nil, errors.ErrInternalServer
	}

	return entity, nil
}

// Delete 删除实体
func (s *entityService) Delete(id uint) error {
	_, err := s.entityRepo.FindByID(id)
	if err != nil {
		return errors.ErrEntityNotFound
	}

	if err := s.entityRepo.Delete(id); err != nil {
		logger.Error("删除实体失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// AddTag 添加标签
func (s *entityService) AddTag(id uint, tag string) error {
	_, err := s.entityRepo.FindByID(id)
	if err != nil {
		return errors.ErrEntityNotFound
	}

	if err := s.entityRepo.AddTag(id, tag); err != nil {
		logger.Error("添加标签失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// RemoveTag 移除标签
func (s *entityService) RemoveTag(id uint, tag string) error {
	_, err := s.entityRepo.FindByID(id)
	if err != nil {
		return errors.ErrEntityNotFound
	}

	if err := s.entityRepo.RemoveTag(id, tag); err != nil {
		logger.Error("移除标签失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// CreateLink 创建实体关联
func (s *entityService) CreateLink(id, targetID uint, linkType, relationName string) error {
	// 验证源实体是否存在
	_, err := s.entityRepo.FindByID(id)
	if err != nil {
		return errors.ErrEntityNotFound
	}

	// 验证目标实体是否存在
	_, err = s.entityRepo.FindByID(targetID)
	if err != nil {
		return errors.ErrEntityNotFound
	}

	// 检测循环引用
	if s.detectCycle(id, targetID) {
		logger.Error("检测到循环引用", logger.Uint("source_id", id), logger.Uint("target_id", targetID))
		return stderrors.New("circular reference detected")
	}

	if err := s.entityRepo.CreateLink(id, targetID, linkType, relationName); err != nil {
		logger.Error("创建实体关联失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// detectCycle 使用DFS检测循环引用
func (s *entityService) detectCycle(sourceID, targetID uint) bool {
	// 获取所有链接
	links, err := s.entityRepo.GetLinks(targetID)
	if err != nil {
		return false
	}

	visited := make(map[uint]bool)
	return s.hasPathTo(links, targetID, sourceID, visited)
}

// hasPathTo DFS搜索是否存在从current到target的路径
func (s *entityService) hasPathTo(links []*model.EntityLink, current, target uint, visited map[uint]bool) bool {
	if current == target {
		return true
	}

	if visited[current] {
		return false
	}
	visited[current] = true

	// 获取当前节点的所有出边
	currentLinks, _ := s.entityRepo.GetLinks(current)
	for _, link := range currentLinks {
		if link.TargetID == target {
			return true
		}
		if !visited[link.TargetID] {
			if s.hasPathTo(nil, link.TargetID, target, visited) {
				return true
			}
		}
	}

	return false
}

// DeleteLink 删除实体关联
func (s *entityService) DeleteLink(id, targetID uint) error {
	_, err := s.entityRepo.FindByID(id)
	if err != nil {
		return errors.ErrEntityNotFound
	}

	if err := s.entityRepo.DeleteLink(id, targetID); err != nil {
		logger.Error("删除实体关联失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}
