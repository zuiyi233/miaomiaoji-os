package service

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
)

// DocumentService 文档服务接口
type DocumentService interface {
	Create(projectID uint, title string, content, summary, status string, orderIndex int, timeNode, duration string, targetWordCount int, chapterGoal, corePlot, hook, causeEffect, foreshadowingDetails string, volumeID uint) (*model.Document, error)
	GetByID(id uint) (*model.Document, error)
	ListByProjectID(projectID uint, page, size int) ([]*model.Document, int64, error)
	ListByVolumeID(volumeID uint, page, size int) ([]*model.Document, int64, error)
	Update(id uint, updates map[string]interface{}) (*model.Document, error)
	Delete(id uint) error
	AddBookmark(id uint, title string, position int, note string) error
	RemoveBookmark(id uint, index int) error
	LinkEntity(id, entityID uint, refType string, metadata map[string]interface{}) error
	UnlinkEntity(id, entityID uint) error
	GetEntityRefs(id uint) ([]*model.DocumentEntityRef, error)
}

// documentService 文档服务实现
type documentService struct {
	documentRepo repository.DocumentRepository
	projectRepo  repository.ProjectRepository
	volumeRepo   repository.VolumeRepository
}

// NewDocumentService 创建文档服务实例
func NewDocumentService(documentRepo repository.DocumentRepository, projectRepo repository.ProjectRepository, volumeRepo repository.VolumeRepository) DocumentService {
	return &documentService{
		documentRepo: documentRepo,
		projectRepo:  projectRepo,
		volumeRepo:   volumeRepo,
	}
}

// Create 创建文档
func (s *documentService) Create(projectID uint, title string, content, summary, status string, orderIndex int, timeNode, duration string, targetWordCount int, chapterGoal, corePlot, hook, causeEffect, foreshadowingDetails string, volumeID uint) (*model.Document, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, errors.ErrProjectNotFound
	}

	// 如果指定了卷ID，验证卷是否存在
	if volumeID > 0 {
		_, err := s.volumeRepo.FindByID(volumeID)
		if err != nil {
			return nil, errors.ErrVolumeNotFound
		}
	}

	document := &model.Document{
		ProjectID:            projectID,
		VolumeID:             volumeID,
		Title:                title,
		Content:              content,
		Summary:              summary,
		Status:               status,
		OrderIndex:           orderIndex,
		TimeNode:             timeNode,
		Duration:             duration,
		TargetWordCount:      targetWordCount,
		ChapterGoal:          chapterGoal,
		CorePlot:             corePlot,
		Hook:                 hook,
		CauseEffect:          causeEffect,
		ForeshadowingDetails: foreshadowingDetails,
	}

	if err := s.documentRepo.Create(document); err != nil {
		logger.Error("创建文档失败", logger.Err(err))
		return nil, errors.ErrInternalServer
	}

	return document, nil
}

// GetByID 根据ID获取文档
func (s *documentService) GetByID(id uint) (*model.Document, error) {
	document, err := s.documentRepo.FindByID(id)
	if err != nil {
		return nil, errors.ErrDocumentNotFound
	}
	return document, nil
}

// ListByProjectID 根据项目ID获取文档列表
func (s *documentService) ListByProjectID(projectID uint, page, size int) ([]*model.Document, int64, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, 0, errors.ErrProjectNotFound
	}

	return s.documentRepo.FindByProjectID(projectID, page, size)
}

// ListByVolumeID 根据卷ID获取文档列表
func (s *documentService) ListByVolumeID(volumeID uint, page, size int) ([]*model.Document, int64, error) {
	// 验证卷是否存在
	_, err := s.volumeRepo.FindByID(volumeID)
	if err != nil {
		return nil, 0, errors.ErrVolumeNotFound
	}

	return s.documentRepo.FindByVolumeID(volumeID, page, size)
}

// Update 更新文档
func (s *documentService) Update(id uint, updates map[string]interface{}) (*model.Document, error) {
	document, err := s.documentRepo.FindByID(id)
	if err != nil {
		return nil, errors.ErrDocumentNotFound
	}

	// 应用更新
	if title, ok := updates["title"].(string); ok {
		document.Title = title
	}
	if content, ok := updates["content"].(string); ok {
		document.Content = content
	}
	if summary, ok := updates["summary"].(string); ok {
		document.Summary = summary
	}
	if status, ok := updates["status"].(string); ok {
		document.Status = status
	}
	if orderIndex, ok := updates["order_index"].(int); ok {
		document.OrderIndex = orderIndex
	}
	if timeNode, ok := updates["time_node"].(string); ok {
		document.TimeNode = timeNode
	}
	if duration, ok := updates["duration"].(string); ok {
		document.Duration = duration
	}
	if targetWordCount, ok := updates["target_word_count"].(int); ok {
		document.TargetWordCount = targetWordCount
	}
	if chapterGoal, ok := updates["chapter_goal"].(string); ok {
		document.ChapterGoal = chapterGoal
	}
	if corePlot, ok := updates["core_plot"].(string); ok {
		document.CorePlot = corePlot
	}
	if hook, ok := updates["hook"].(string); ok {
		document.Hook = hook
	}
	if causeEffect, ok := updates["cause_effect"].(string); ok {
		document.CauseEffect = causeEffect
	}
	if foreshadowingDetails, ok := updates["foreshadowing_details"].(string); ok {
		document.ForeshadowingDetails = foreshadowingDetails
	}
	if volumeID, ok := updates["volume_id"].(uint); ok {
		document.VolumeID = volumeID
	}

	if err := s.documentRepo.Update(document); err != nil {
		logger.Error("更新文档失败", logger.Err(err))
		return nil, errors.ErrInternalServer
	}

	return document, nil
}

// Delete 删除文档
func (s *documentService) Delete(id uint) error {
	_, err := s.documentRepo.FindByID(id)
	if err != nil {
		return errors.ErrDocumentNotFound
	}

	if err := s.documentRepo.Delete(id); err != nil {
		logger.Error("删除文档失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// AddBookmark 添加书签
func (s *documentService) AddBookmark(id uint, title string, position int, note string) error {
	_, err := s.documentRepo.FindByID(id)
	if err != nil {
		return errors.ErrDocumentNotFound
	}

	bookmark := model.Bookmark{
		Title:    title,
		Position: position,
		Note:     note,
	}

	if err := s.documentRepo.AddBookmark(id, bookmark); err != nil {
		logger.Error("添加书签失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// RemoveBookmark 移除书签
func (s *documentService) RemoveBookmark(id uint, index int) error {
	_, err := s.documentRepo.FindByID(id)
	if err != nil {
		return errors.ErrDocumentNotFound
	}

	if err := s.documentRepo.RemoveBookmark(id, index); err != nil {
		logger.Error("移除书签失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// LinkEntity 关联实体
func (s *documentService) LinkEntity(id, entityID uint, refType string, metadata map[string]interface{}) error {
	_, err := s.documentRepo.FindByID(id)
	if err != nil {
		return errors.ErrDocumentNotFound
	}

	if err := s.documentRepo.LinkEntity(id, entityID, refType, metadata); err != nil {
		logger.Error("关联实体失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// UnlinkEntity 取消关联实体
func (s *documentService) UnlinkEntity(id, entityID uint) error {
	_, err := s.documentRepo.FindByID(id)
	if err != nil {
		return errors.ErrDocumentNotFound
	}

	if err := s.documentRepo.UnlinkEntity(id, entityID); err != nil {
		logger.Error("取消关联实体失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// GetEntityRefs 获取文档的实体关联
func (s *documentService) GetEntityRefs(id uint) ([]*model.DocumentEntityRef, error) {
	_, err := s.documentRepo.FindByID(id)
	if err != nil {
		return nil, errors.ErrDocumentNotFound
	}

	return s.documentRepo.GetEntityRefs(id)
}
