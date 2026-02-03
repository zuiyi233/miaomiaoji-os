package repository

import (
	"sync"

	"gorm.io/gorm"
	"rulebacktest/internal/model"
)

var (
	addressRepoInstance *AddressRepository
	addressRepoOnce     sync.Once
)

// AddressRepository 地址数据访问层
type AddressRepository struct {
	*BaseRepository
}

// NewAddressRepository 创建AddressRepository实例
func NewAddressRepository(base *BaseRepository) *AddressRepository {
	return &AddressRepository{BaseRepository: base}
}

// GetAddressRepository 获取AddressRepository单例
func GetAddressRepository() *AddressRepository {
	addressRepoOnce.Do(func() {
		addressRepoInstance = &AddressRepository{
			BaseRepository: GetBaseRepository(),
		}
	})
	return addressRepoInstance
}

// Create 创建地址
func (r *AddressRepository) Create(address *model.Address) error {
	return r.DB().Create(address).Error
}

// GetByID 根据ID获取地址
func (r *AddressRepository) GetByID(id uint) (*model.Address, error) {
	var address model.Address
	if err := r.DB().First(&address, id).Error; err != nil {
		return nil, err
	}
	return &address, nil
}

// GetByUserID 获取用户的所有地址
func (r *AddressRepository) GetByUserID(userID uint) ([]model.Address, error) {
	var addresses []model.Address
	if err := r.DB().Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&addresses).Error; err != nil {
		return nil, err
	}
	return addresses, nil
}

// GetDefaultByUserID 获取用户默认地址
func (r *AddressRepository) GetDefaultByUserID(userID uint) (*model.Address, error) {
	var address model.Address
	if err := r.DB().Where("user_id = ? AND is_default = ?", userID, true).
		First(&address).Error; err != nil {
		return nil, err
	}
	return &address, nil
}

// Update 更新地址
func (r *AddressRepository) Update(address *model.Address) error {
	return r.DB().Save(address).Error
}

// Delete 删除地址
func (r *AddressRepository) Delete(id uint) error {
	return r.DB().Delete(&model.Address{}, id).Error
}

// ClearDefault 清除用户的默认地址
func (r *AddressRepository) ClearDefault(userID uint) error {
	return r.DB().Model(&model.Address{}).
		Where("user_id = ? AND is_default = ?", userID, true).
		Update("is_default", false).Error
}

// Transaction 事务处理
func (r *AddressRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.DB().Transaction(fn)
}
