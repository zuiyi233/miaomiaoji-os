package repository

import (
	"novel-agent-os-backend/internal/model"
)

// RedemptionCodeRepository 兑换码仓储接口
type RedemptionCodeRepository interface {
	Create(code *model.RedemptionCode) error
	FindByCode(code string) (*model.RedemptionCode, error)
	Update(code *model.RedemptionCode) error
	RecordUse(use *model.RedemptionCodeUse) error
	CountUserUses(codeID uint, userID uint) (int64, error)
	List(filter RedemptionCodeFilter) ([]*model.RedemptionCode, int64, error)
	UpdateStatus(codes []string, status string) error
	DeleteByCodes(codes []string) error
	RenewByCodes(codes []string, days int) error
}

// redemptionCodeRepository 兑换码仓储实现
type redemptionCodeRepository struct {
	*BaseRepository
}

// RedemptionCodeFilter 兑换码过滤
type RedemptionCodeFilter struct {
	Status string
	Search string
	Page   int
	Size   int
	Sort   string
}

// NewRedemptionCodeRepository 创建兑换码仓储实例
func NewRedemptionCodeRepository() RedemptionCodeRepository {
	return &redemptionCodeRepository{
		BaseRepository: GetBaseRepository(),
	}
}

// Create 创建兑换码
func (r *redemptionCodeRepository) Create(code *model.RedemptionCode) error {
	return r.BaseRepository.db.Create(code).Error
}

// FindByCode 根据编码查找
func (r *redemptionCodeRepository) FindByCode(code string) (*model.RedemptionCode, error) {
	var item model.RedemptionCode
	if err := r.BaseRepository.db.Where("code = ?", code).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// Update 更新兑换码
func (r *redemptionCodeRepository) Update(code *model.RedemptionCode) error {
	return r.BaseRepository.db.Save(code).Error
}

// RecordUse 记录兑换码使用
func (r *redemptionCodeRepository) RecordUse(use *model.RedemptionCodeUse) error {
	return r.BaseRepository.db.Create(use).Error
}

// CountUserUses 统计用户使用次数
func (r *redemptionCodeRepository) CountUserUses(codeID uint, userID uint) (int64, error) {
	var total int64
	if err := r.BaseRepository.db.Model(&model.RedemptionCodeUse{}).
		Where("code_id = ? AND user_id = ?", codeID, userID).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// List 获取兑换码列表
func (r *redemptionCodeRepository) List(filter RedemptionCodeFilter) ([]*model.RedemptionCode, int64, error) {
	var codes []*model.RedemptionCode
	var total int64

	query := r.BaseRepository.db.Model(&model.RedemptionCode{})
	if filter.Status != "" && filter.Status != "all" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("code LIKE ? OR note LIKE ? OR source LIKE ?", like, like, like)
	}
	if filter.Status == "expired" {
		query = query.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now())
	}
	if filter.Status == "active" {
		query = query.Where("status = ?", "active").
			Where("(expires_at IS NULL OR expires_at >= ?)", time.Now())
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	size := filter.Size
	if size < 1 {
		size = 20
	}
	offset := (page - 1) * size

	sort := "created_at DESC"
	if filter.Sort == "asc" {
		sort = "created_at ASC"
	}
	if err := query.Order(sort).Offset(offset).Limit(size).Find(&codes).Error; err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}

// UpdateStatus 批量更新状态
func (r *redemptionCodeRepository) UpdateStatus(codes []string, status string) error {
	if len(codes) == 0 {
		return nil
	}
	return r.BaseRepository.db.Model(&model.RedemptionCode{}).Where("code IN ?", codes).Update("status", status).Error
}

// DeleteByCodes 批量删除
func (r *redemptionCodeRepository) DeleteByCodes(codes []string) error {
	if len(codes) == 0 {
		return nil
	}
	return r.BaseRepository.db.Where("code IN ?", codes).Delete(&model.RedemptionCode{}).Error
}

// RenewByCodes 批量续期
func (r *redemptionCodeRepository) RenewByCodes(codes []string, days int) error {
	if len(codes) == 0 || days <= 0 {
		return nil
	}

	var items []*model.RedemptionCode
	if err := r.BaseRepository.db.Where("code IN ?", codes).Find(&items).Error; err != nil {
		return err
	}

	for _, item := range items {
		if item.ExpiresAt == nil {
			continue
		}
		newTime := item.ExpiresAt.AddDate(0, 0, days)
		item.ExpiresAt = &newTime
		if err := r.BaseRepository.db.Save(item).Error; err != nil {
			return err
		}
	}

	return nil
}
