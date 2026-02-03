package service

import (
	"errors"
	"sync"

	"gorm.io/gorm"
	"rulebacktest/internal/model"
	"rulebacktest/internal/repository"
	apperrors "rulebacktest/pkg/errors"
)

var (
	addressServiceInstance *AddressService
	addressServiceOnce     sync.Once
)

// AddressService 地址业务逻辑层
type AddressService struct {
	repo *repository.AddressRepository
}

// NewAddressService 创建AddressService实例
func NewAddressService(repo *repository.AddressRepository) *AddressService {
	return &AddressService{repo: repo}
}

// GetAddressService 获取AddressService单例
func GetAddressService() *AddressService {
	addressServiceOnce.Do(func() {
		addressServiceInstance = &AddressService{
			repo: repository.GetAddressRepository(),
		}
	})
	return addressServiceInstance
}

// Create 创建地址
func (s *AddressService) Create(userID uint, req *model.AddressCreateReq) (*model.Address, error) {
	address := &model.Address{
		UserID:    userID,
		Name:      req.Name,
		Phone:     req.Phone,
		Province:  req.Province,
		City:      req.City,
		District:  req.District,
		Detail:    req.Detail,
		IsDefault: req.IsDefault,
	}

	if req.IsDefault {
		if err := s.repo.ClearDefault(userID); err != nil {
			return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
		}
	}

	if err := s.repo.Create(address); err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return address, nil
}

// GetByID 根据ID获取地址
func (s *AddressService) GetByID(userID, addressID uint) (*model.Address, error) {
	address, err := s.repo.GetByID(addressID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if address.UserID != userID {
		return nil, apperrors.ErrForbidden
	}

	return address, nil
}

// List 获取用户地址列表
func (s *AddressService) List(userID uint) ([]model.Address, error) {
	addresses, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return addresses, nil
}

// GetDefault 获取用户默认地址
func (s *AddressService) GetDefault(userID uint) (*model.Address, error) {
	address, err := s.repo.GetDefaultByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}
	return address, nil
}

// Update 更新地址
func (s *AddressService) Update(userID, addressID uint, req *model.AddressUpdateReq) (*model.Address, error) {
	address, err := s.repo.GetByID(addressID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if address.UserID != userID {
		return nil, apperrors.ErrForbidden
	}

	if req.Name != "" {
		address.Name = req.Name
	}
	if req.Phone != "" {
		address.Phone = req.Phone
	}
	if req.Province != "" {
		address.Province = req.Province
	}
	if req.City != "" {
		address.City = req.City
	}
	if req.District != "" {
		address.District = req.District
	}
	if req.Detail != "" {
		address.Detail = req.Detail
	}
	if req.IsDefault != nil && *req.IsDefault && !address.IsDefault {
		if err := s.repo.ClearDefault(userID); err != nil {
			return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
		}
		address.IsDefault = true
	} else if req.IsDefault != nil && !*req.IsDefault {
		address.IsDefault = false
	}

	if err := s.repo.Update(address); err != nil {
		return nil, apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return address, nil
}

// Delete 删除地址
func (s *AddressService) Delete(userID, addressID uint) error {
	address, err := s.repo.GetByID(addressID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound
		}
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if address.UserID != userID {
		return apperrors.ErrForbidden
	}

	if err := s.repo.Delete(addressID); err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return nil
}

// SetDefault 设置默认地址
func (s *AddressService) SetDefault(userID, addressID uint) error {
	address, err := s.repo.GetByID(addressID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound
		}
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	if address.UserID != userID {
		return apperrors.ErrForbidden
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Address{}).
			Where("user_id = ? AND is_default = ?", userID, true).
			Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&model.Address{}).
			Where("id = ?", addressID).
			Update("is_default", true).Error
	})

	if err != nil {
		return apperrors.WrapWithCode(apperrors.CodeDatabaseError, err)
	}

	return nil
}
