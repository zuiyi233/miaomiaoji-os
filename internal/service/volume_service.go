package service

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
)

// VolumeService 卷服务接口
type VolumeService interface {
	Create(projectID uint, title string, orderIndex int, theme, coreGoal, boundaries, chapterLinkageLogic, volumeSpecificSettings, plotRoadmap string) (*model.Volume, error)
	GetByID(id uint) (*model.Volume, error)
	ListByProjectID(projectID uint, page, size int) ([]*model.Volume, int64, error)
	Update(id uint, updates map[string]interface{}) (*model.Volume, error)
	Delete(id uint) error
	ReorderVolumes(projectID uint, volumeIDs []uint) error
}

// volumeService 卷服务实现
type volumeService struct {
	volumeRepo  repository.VolumeRepository
	projectRepo repository.ProjectRepository
}

// NewVolumeService 创建卷服务实例
func NewVolumeService(volumeRepo repository.VolumeRepository, projectRepo repository.ProjectRepository) VolumeService {
	return &volumeService{
		volumeRepo:  volumeRepo,
		projectRepo: projectRepo,
	}
}

// Create 创建卷
func (s *volumeService) Create(projectID uint, title string, orderIndex int, theme, coreGoal, boundaries, chapterLinkageLogic, volumeSpecificSettings, plotRoadmap string) (*model.Volume, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, errors.ErrProjectNotFound
	}

	volume := &model.Volume{
		ProjectID:              projectID,
		Title:                  title,
		OrderIndex:             orderIndex,
		Theme:                  theme,
		CoreGoal:               coreGoal,
		Boundaries:             boundaries,
		ChapterLinkageLogic:    chapterLinkageLogic,
		VolumeSpecificSettings: volumeSpecificSettings,
		PlotRoadmap:            plotRoadmap,
	}

	if err := s.volumeRepo.Create(volume); err != nil {
		logger.Error("创建卷失败", logger.Err(err))
		return nil, errors.ErrInternalServer
	}

	return volume, nil
}

// GetByID 根据ID获取卷
func (s *volumeService) GetByID(id uint) (*model.Volume, error) {
	volume, err := s.volumeRepo.FindByID(id)
	if err != nil {
		return nil, errors.ErrVolumeNotFound
	}
	return volume, nil
}

// ListByProjectID 根据项目ID获取卷列表
func (s *volumeService) ListByProjectID(projectID uint, page, size int) ([]*model.Volume, int64, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, 0, errors.ErrProjectNotFound
	}

	return s.volumeRepo.FindByProjectID(projectID, page, size)
}

// Update 更新卷
func (s *volumeService) Update(id uint, updates map[string]interface{}) (*model.Volume, error) {
	volume, err := s.volumeRepo.FindByID(id)
	if err != nil {
		return nil, errors.ErrVolumeNotFound
	}

	// 应用更新
	if title, ok := updates["title"].(string); ok {
		volume.Title = title
	}
	if orderIndex, ok := updates["order_index"].(int); ok {
		volume.OrderIndex = orderIndex
	}
	if theme, ok := updates["theme"].(string); ok {
		volume.Theme = theme
	}
	if coreGoal, ok := updates["core_goal"].(string); ok {
		volume.CoreGoal = coreGoal
	}
	if boundaries, ok := updates["boundaries"].(string); ok {
		volume.Boundaries = boundaries
	}
	if chapterLinkageLogic, ok := updates["chapter_linkage_logic"].(string); ok {
		volume.ChapterLinkageLogic = chapterLinkageLogic
	}
	if volumeSpecificSettings, ok := updates["volume_specific_settings"].(string); ok {
		volume.VolumeSpecificSettings = volumeSpecificSettings
	}
	if plotRoadmap, ok := updates["plot_roadmap"].(string); ok {
		volume.PlotRoadmap = plotRoadmap
	}

	if err := s.volumeRepo.Update(volume); err != nil {
		logger.Error("更新卷失败", logger.Err(err))
		return nil, errors.ErrInternalServer
	}

	return volume, nil
}

// Delete 删除卷
func (s *volumeService) Delete(id uint) error {
	_, err := s.volumeRepo.FindByID(id)
	if err != nil {
		return errors.ErrVolumeNotFound
	}

	if err := s.volumeRepo.Delete(id); err != nil {
		logger.Error("删除卷失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// ReorderVolumes 批量重新排序卷
func (s *volumeService) ReorderVolumes(projectID uint, volumeIDs []uint) error {
	// 验证项目是否存在
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return errors.ErrProjectNotFound
	}

	if err := s.volumeRepo.ReorderVolumes(projectID, volumeIDs); err != nil {
		logger.Error("重新排序卷失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}
