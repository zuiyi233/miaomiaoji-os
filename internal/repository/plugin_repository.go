package repository

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/pkg/database"
)

// PluginRepository 插件数据访问接口
type PluginRepository interface {
	Create(plugin *model.Plugin) error
	FindByID(id uint) (*model.Plugin, error)
	FindByName(name string) (*model.Plugin, error)
	List(page, size int) ([]*model.Plugin, int64, error)
	ListByStatus(status string, page, size int) ([]*model.Plugin, int64, error)
	Update(plugin *model.Plugin) error
	Delete(id uint) error

	// 插件能力
	CreateCapability(cap *model.PluginCapability) error
	FindCapabilityByID(id uint) (*model.PluginCapability, error)
	FindCapabilitiesByPluginID(pluginID uint) ([]*model.PluginCapability, error)
	FindCapabilityByName(pluginID uint, name string) (*model.PluginCapability, error)
	UpdateCapability(cap *model.PluginCapability) error
	DeleteCapability(id uint) error
}

// pluginRepository 插件数据访问实现
type pluginRepository struct{}

// NewPluginRepository 创建插件仓库实例
func NewPluginRepository() PluginRepository {
	return &pluginRepository{}
}

// Create 创建插件
func (r *pluginRepository) Create(plugin *model.Plugin) error {
	return database.GetDB().Create(plugin).Error
}

// FindByID 根据ID查找插件
func (r *pluginRepository) FindByID(id uint) (*model.Plugin, error) {
	var plugin model.Plugin
	err := database.GetDB().Preload("Capabilities").First(&plugin, id).Error
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

// FindByName 根据名称查找插件
func (r *pluginRepository) FindByName(name string) (*model.Plugin, error) {
	var plugin model.Plugin
	err := database.GetDB().Where("name = ?", name).First(&plugin).Error
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

// List 获取插件列表
func (r *pluginRepository) List(page, size int) ([]*model.Plugin, int64, error) {
	var plugins []*model.Plugin
	var total int64

	db := database.GetDB().Model(&model.Plugin{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&plugins).Error; err != nil {
		return nil, 0, err
	}

	return plugins, total, nil
}

// ListByStatus 根据状态获取插件列表
func (r *pluginRepository) ListByStatus(status string, page, size int) ([]*model.Plugin, int64, error) {
	var plugins []*model.Plugin
	var total int64

	db := database.GetDB().Model(&model.Plugin{}).Where("status = ?", status)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&plugins).Error; err != nil {
		return nil, 0, err
	}

	return plugins, total, nil
}

// Update 更新插件
func (r *pluginRepository) Update(plugin *model.Plugin) error {
	return database.GetDB().Save(plugin).Error
}

// Delete 删除插件
func (r *pluginRepository) Delete(id uint) error {
	return database.GetDB().Delete(&model.Plugin{}, id).Error
}

// CreateCapability 创建插件能力
func (r *pluginRepository) CreateCapability(cap *model.PluginCapability) error {
	return database.GetDB().Create(cap).Error
}

// FindCapabilityByID 根据ID查找插件能力
func (r *pluginRepository) FindCapabilityByID(id uint) (*model.PluginCapability, error) {
	var cap model.PluginCapability
	err := database.GetDB().First(&cap, id).Error
	if err != nil {
		return nil, err
	}
	return &cap, nil
}

// FindCapabilitiesByPluginID 根据插件ID查找能力列表
func (r *pluginRepository) FindCapabilitiesByPluginID(pluginID uint) ([]*model.PluginCapability, error) {
	var caps []*model.PluginCapability
	err := database.GetDB().
		Where("plugin_id = ?", pluginID).
		Find(&caps).Error
	return caps, err
}

// FindCapabilityByName 根据名称查找插件能力
func (r *pluginRepository) FindCapabilityByName(pluginID uint, name string) (*model.PluginCapability, error) {
	var cap model.PluginCapability
	err := database.GetDB().
		Where("plugin_id = ? AND name = ?", pluginID, name).
		First(&cap).Error
	if err != nil {
		return nil, err
	}
	return &cap, nil
}

// UpdateCapability 更新插件能力
func (r *pluginRepository) UpdateCapability(cap *model.PluginCapability) error {
	return database.GetDB().Save(cap).Error
}

// DeleteCapability 删除插件能力
func (r *pluginRepository) DeleteCapability(id uint) error {
	return database.GetDB().Delete(&model.PluginCapability{}, id).Error
}
