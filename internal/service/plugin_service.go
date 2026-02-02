package service

import (
	"time"

	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"

	"gorm.io/datatypes"
)

// PluginService 插件服务接口
type PluginService interface {
	Register(name, description, version, entryPoint, manifest string) (*model.Plugin, error)
	GetByID(id uint) (*model.Plugin, error)
	GetByName(name string) (*model.Plugin, error)
	List(page, size int) ([]*model.Plugin, int64, error)
	ListByStatus(status string, page, size int) ([]*model.Plugin, int64, error)
	UpdateStatus(id uint, status string) error
	UpdateHealth(id uint, healthy bool, latency int) error
	Delete(id uint) error

	// 插件能力
	AddCapability(pluginID uint, name, description, inputSchema, outputSchema string) (*model.PluginCapability, error)
	GetCapability(id uint) (*model.PluginCapability, error)
	ListCapabilities(pluginID uint) ([]*model.PluginCapability, error)
	UpdateCapability(id uint, updates map[string]interface{}) error
	DeleteCapability(id uint) error

	// 调用插件能力
	InvokeCapability(capabilityID uint, input map[string]interface{}, timeout time.Duration) (map[string]interface{}, error)
}

// pluginService 插件服务实现
type pluginService struct {
	pluginRepo repository.PluginRepository
}

// NewPluginService 创建插件服务实例
func NewPluginService(pluginRepo repository.PluginRepository) PluginService {
	return &pluginService{
		pluginRepo: pluginRepo,
	}
}

// Register 注册插件
func (s *pluginService) Register(name, description, version, entryPoint, manifest string) (*model.Plugin, error) {
	// 检查插件名是否已存在
	existing, _ := s.pluginRepo.FindByName(name)
	if existing != nil {
		return nil, errors.ErrAlreadyExists
	}

	plugin := &model.Plugin{
		Name:        name,
		Description: description,
		Version:     version,
		EntryPoint:  entryPoint,
		Manifest:    datatypes.JSON(manifest),
		Status:      "disabled",
		Healthy:     false,
	}

	if err := s.pluginRepo.Create(plugin); err != nil {
		logger.Error("注册插件失败", logger.Err(err))
		return nil, errors.ErrInternalServer
	}

	return plugin, nil
}

// GetByID 根据ID获取插件
func (s *pluginService) GetByID(id uint) (*model.Plugin, error) {
	plugin, err := s.pluginRepo.FindByID(id)
	if err != nil {
		return nil, errors.ErrPluginNotFound
	}
	return plugin, nil
}

// GetByName 根据名称获取插件
func (s *pluginService) GetByName(name string) (*model.Plugin, error) {
	plugin, err := s.pluginRepo.FindByName(name)
	if err != nil {
		return nil, errors.ErrPluginNotFound
	}
	return plugin, nil
}

// List 获取插件列表
func (s *pluginService) List(page, size int) ([]*model.Plugin, int64, error) {
	return s.pluginRepo.List(page, size)
}

// ListByStatus 根据状态获取插件列表
func (s *pluginService) ListByStatus(status string, page, size int) ([]*model.Plugin, int64, error) {
	return s.pluginRepo.ListByStatus(status, page, size)
}

// UpdateStatus 更新插件状态
func (s *pluginService) UpdateStatus(id uint, status string) error {
	plugin, err := s.pluginRepo.FindByID(id)
	if err != nil {
		return errors.ErrPluginNotFound
	}

	// 验证状态值
	validStatuses := map[string]bool{
		"enabled":  true,
		"disabled": true,
	}
	if !validStatuses[status] {
		return errors.ErrValidationError
	}

	plugin.Status = status
	if err := s.pluginRepo.Update(plugin); err != nil {
		logger.Error("更新插件状态失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// UpdateHealth 更新插件健康状态
func (s *pluginService) UpdateHealth(id uint, healthy bool, latency int) error {
	plugin, err := s.pluginRepo.FindByID(id)
	if err != nil {
		return errors.ErrPluginNotFound
	}

	plugin.Healthy = healthy
	plugin.LatencyMs = latency
	now := time.Now()
	plugin.LastHeartbeat = &now

	if err := s.pluginRepo.Update(plugin); err != nil {
		logger.Error("更新插件健康状态失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// Delete 删除插件
func (s *pluginService) Delete(id uint) error {
	_, err := s.pluginRepo.FindByID(id)
	if err != nil {
		return errors.ErrPluginNotFound
	}

	if err := s.pluginRepo.Delete(id); err != nil {
		logger.Error("删除插件失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// AddCapability 添加插件能力
func (s *pluginService) AddCapability(pluginID uint, name, description, inputSchema, outputSchema string) (*model.PluginCapability, error) {
	// 验证插件是否存在
	_, err := s.pluginRepo.FindByID(pluginID)
	if err != nil {
		return nil, errors.ErrPluginNotFound
	}

	// 检查能力名是否已存在
	existing, _ := s.pluginRepo.FindCapabilityByName(pluginID, name)
	if existing != nil {
		return nil, errors.ErrAlreadyExists
	}

	cap := &model.PluginCapability{
		PluginID:     pluginID,
		Name:         name,
		Description:  description,
		InputSchema:  datatypes.JSON(inputSchema),
		OutputSchema: datatypes.JSON(outputSchema),
	}

	if err := s.pluginRepo.CreateCapability(cap); err != nil {
		logger.Error("添加插件能力失败", logger.Err(err))
		return nil, errors.ErrInternalServer
	}

	return cap, nil
}

// GetCapability 获取插件能力
func (s *pluginService) GetCapability(id uint) (*model.PluginCapability, error) {
	cap, err := s.pluginRepo.FindCapabilityByID(id)
	if err != nil {
		return nil, errors.ErrPluginCapNotFound
	}
	return cap, nil
}

// ListCapabilities 获取插件能力列表
func (s *pluginService) ListCapabilities(pluginID uint) ([]*model.PluginCapability, error) {
	// 验证插件是否存在
	_, err := s.pluginRepo.FindByID(pluginID)
	if err != nil {
		return nil, errors.ErrPluginNotFound
	}

	return s.pluginRepo.FindCapabilitiesByPluginID(pluginID)
}

// UpdateCapability 更新插件能力
func (s *pluginService) UpdateCapability(id uint, updates map[string]interface{}) error {
	cap, err := s.pluginRepo.FindCapabilityByID(id)
	if err != nil {
		return errors.ErrPluginCapNotFound
	}

	if description, ok := updates["description"].(string); ok {
		cap.Description = description
	}
	if inputSchema, ok := updates["input_schema"].(string); ok {
		cap.InputSchema = datatypes.JSON(inputSchema)
	}
	if outputSchema, ok := updates["output_schema"].(string); ok {
		cap.OutputSchema = datatypes.JSON(outputSchema)
	}

	if err := s.pluginRepo.UpdateCapability(cap); err != nil {
		logger.Error("更新插件能力失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// DeleteCapability 删除插件能力
func (s *pluginService) DeleteCapability(id uint) error {
	_, err := s.pluginRepo.FindCapabilityByID(id)
	if err != nil {
		return errors.ErrPluginCapNotFound
	}

	if err := s.pluginRepo.DeleteCapability(id); err != nil {
		logger.Error("删除插件能力失败", logger.Err(err))
		return errors.ErrInternalServer
	}

	return nil
}

// InvokeCapability 调用插件能力
func (s *pluginService) InvokeCapability(capabilityID uint, input map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	cap, err := s.pluginRepo.FindCapabilityByID(capabilityID)
	if err != nil {
		return nil, errors.ErrPluginCapNotFound
	}

	// 获取插件信息
	plugin, err := s.pluginRepo.FindByID(cap.PluginID)
	if err != nil {
		return nil, errors.ErrPluginNotFound
	}

	// 检查插件状态
	if plugin.Status != "enabled" {
		return nil, errors.ErrPluginDisabled
	}
	if !plugin.Healthy {
		return nil, errors.ErrPluginOffline
	}

	// TODO: 实际调用插件（通过 HTTP/gRPC 等方式）
	// 这里返回模拟结果
	logger.Info("调用插件能力",
		logger.String("plugin", plugin.Name),
		logger.String("capability", cap.Name),
	)

	return map[string]interface{}{
		"status": "success",
		"data":   input,
	}, nil
}
