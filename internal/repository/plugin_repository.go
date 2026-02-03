package repository

import (
	"novel-agent-os-backend/internal/model"
	"time"

	"gorm.io/gorm"
)

type PluginRepository interface {
	Create(plugin *model.Plugin) error
	GetByID(id uint) (*model.Plugin, error)
	GetByName(name string) (*model.Plugin, error)
	Update(plugin *model.Plugin) error
	Delete(id uint) error
	List(page, pageSize int) ([]*model.Plugin, int64, error)
	ListEnabled() ([]*model.Plugin, error)
	UpdateStatus(id uint, status string) error
	UpdateHealth(id uint, healthy bool, latencyMs int) error
	UpdateLastPing(id uint) error

	CreateCapability(capability *model.PluginCapability) error
	GetCapabilitiesByPluginID(pluginID uint) ([]*model.PluginCapability, error)
	DeleteCapability(id uint) error
}

type pluginRepository struct {
	db *gorm.DB
}

func NewPluginRepository(db *gorm.DB) PluginRepository {
	return &pluginRepository{db: db}
}

func (r *pluginRepository) Create(plugin *model.Plugin) error {
	return r.db.Create(plugin).Error
}

func (r *pluginRepository) GetByID(id uint) (*model.Plugin, error) {
	var plugin model.Plugin
	err := r.db.Preload("Capabilities").First(&plugin, id).Error
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

func (r *pluginRepository) GetByName(name string) (*model.Plugin, error) {
	var plugin model.Plugin
	err := r.db.Preload("Capabilities").Where("name = ?", name).First(&plugin).Error
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

func (r *pluginRepository) Update(plugin *model.Plugin) error {
	return r.db.Save(plugin).Error
}

func (r *pluginRepository) Delete(id uint) error {
	return r.db.Delete(&model.Plugin{}, id).Error
}

func (r *pluginRepository) List(page, pageSize int) ([]*model.Plugin, int64, error) {
	var plugins []*model.Plugin
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.Plugin{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Capabilities").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&plugins).Error

	return plugins, total, err
}

func (r *pluginRepository) ListEnabled() ([]*model.Plugin, error) {
	var plugins []*model.Plugin
	err := r.db.Where("is_enabled = ?", true).
		Preload("Capabilities").
		Find(&plugins).Error
	return plugins, err
}

func (r *pluginRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Plugin{}).Where("id = ?", id).Update("status", status).Error
}

func (r *pluginRepository) UpdateHealth(id uint, healthy bool, latencyMs int) error {
	return r.db.Model(&model.Plugin{}).Where("id = ?", id).Updates(map[string]interface{}{
		"healthy":    healthy,
		"latency_ms": latencyMs,
	}).Error
}

func (r *pluginRepository) UpdateLastPing(id uint) error {
	now := time.Now()
	return r.db.Model(&model.Plugin{}).Where("id = ?", id).Update("last_ping", &now).Error
}

func (r *pluginRepository) CreateCapability(capability *model.PluginCapability) error {
	return r.db.Create(capability).Error
}

func (r *pluginRepository) GetCapabilitiesByPluginID(pluginID uint) ([]*model.PluginCapability, error) {
	var capabilities []*model.PluginCapability
	err := r.db.Where("plugin_id = ?", pluginID).Find(&capabilities).Error
	return capabilities, err
}

func (r *pluginRepository) DeleteCapability(id uint) error {
	return r.db.Delete(&model.PluginCapability{}, id).Error
}
