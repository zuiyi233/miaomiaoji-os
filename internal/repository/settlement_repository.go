package repository

import (
	"novel-agent-os-backend/internal/model"

	"gorm.io/gorm"
)

type SettlementRepository interface {
	Create(entry *model.SettlementEntry) error
	GetByID(id uint) (*model.SettlementEntry, error)
	Update(entry *model.SettlementEntry) error
	Delete(id uint) error
	ListByUserID(userID uint, page, pageSize int) ([]*model.SettlementEntry, int64, error)
	ListByWorldID(worldID string, page, pageSize int) ([]*model.SettlementEntry, int64, error)
	ListByChapterID(chapterID string, page, pageSize int) ([]*model.SettlementEntry, int64, error)
	ListByLoopStage(loopStage string, page, pageSize int) ([]*model.SettlementEntry, int64, error)
	Filter(worldID, chapterID, loopStage string, page, pageSize int) ([]*model.SettlementEntry, int64, error)
	GetTotalPointsByUserID(userID uint) (int, error)
}

type settlementRepository struct {
	db *gorm.DB
}

func NewSettlementRepository(db *gorm.DB) SettlementRepository {
	return &settlementRepository{db: db}
}

func (r *settlementRepository) Create(entry *model.SettlementEntry) error {
	return r.db.Create(entry).Error
}

func (r *settlementRepository) GetByID(id uint) (*model.SettlementEntry, error) {
	var entry model.SettlementEntry
	err := r.db.First(&entry, id).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *settlementRepository) Update(entry *model.SettlementEntry) error {
	return r.db.Save(entry).Error
}

func (r *settlementRepository) Delete(id uint) error {
	return r.db.Delete(&model.SettlementEntry{}, id).Error
}

func (r *settlementRepository) ListByUserID(userID uint, page, pageSize int) ([]*model.SettlementEntry, int64, error) {
	var entries []*model.SettlementEntry
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.SettlementEntry{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("user_id = ?", userID).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&entries).Error

	return entries, total, err
}

func (r *settlementRepository) ListByWorldID(worldID string, page, pageSize int) ([]*model.SettlementEntry, int64, error) {
	var entries []*model.SettlementEntry
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.SettlementEntry{}).Where("world_id = ?", worldID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("world_id = ?", worldID).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&entries).Error

	return entries, total, err
}

func (r *settlementRepository) ListByChapterID(chapterID string, page, pageSize int) ([]*model.SettlementEntry, int64, error) {
	var entries []*model.SettlementEntry
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.SettlementEntry{}).Where("chapter_id = ?", chapterID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("chapter_id = ?", chapterID).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&entries).Error

	return entries, total, err
}

func (r *settlementRepository) ListByLoopStage(loopStage string, page, pageSize int) ([]*model.SettlementEntry, int64, error) {
	var entries []*model.SettlementEntry
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.SettlementEntry{}).Where("loop_stage = ?", loopStage).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("loop_stage = ?", loopStage).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&entries).Error

	return entries, total, err
}

func (r *settlementRepository) Filter(worldID, chapterID, loopStage string, page, pageSize int) ([]*model.SettlementEntry, int64, error) {
	var entries []*model.SettlementEntry
	var total int64

	offset := (page - 1) * pageSize

	query := r.db.Model(&model.SettlementEntry{})

	if worldID != "" {
		query = query.Where("world_id = ?", worldID)
	}
	if chapterID != "" {
		query = query.Where("chapter_id = ?", chapterID)
	}
	if loopStage != "" {
		query = query.Where("loop_stage = ?", loopStage)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&entries).Error

	return entries, total, err
}

func (r *settlementRepository) GetTotalPointsByUserID(userID uint) (int, error) {
	var total int
	err := r.db.Model(&model.SettlementEntry{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(points_delta), 0)").
		Scan(&total).Error
	return total, err
}
