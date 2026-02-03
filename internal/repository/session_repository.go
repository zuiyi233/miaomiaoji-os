package repository

import (
	"novel-agent-os-backend/internal/model"

	"gorm.io/gorm"
)

type SessionRepository interface {
	Create(session *model.Session) error
	GetByID(id uint) (*model.Session, error)
	Update(session *model.Session) error
	Delete(id uint) error
	ListByUserID(userID uint, page, pageSize int) ([]*model.Session, int64, error)
	ListByProjectID(projectID uint, page, pageSize int) ([]*model.Session, int64, error)

	CreateStep(step *model.SessionStep) error
	GetMaxStepOrderIndex(sessionID uint) (int, error)
	GetStepByID(id uint) (*model.SessionStep, error)
	UpdateStep(step *model.SessionStep) error
	DeleteStep(id uint) error
	ListStepsBySessionID(sessionID uint) ([]*model.SessionStep, error)
}

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(session *model.Session) error {
	return r.db.Create(session).Error
}

func (r *sessionRepository) GetByID(id uint) (*model.Session, error) {
	var session model.Session
	err := r.db.Preload("Steps").First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) Update(session *model.Session) error {
	return r.db.Save(session).Error
}

func (r *sessionRepository) Delete(id uint) error {
	return r.db.Delete(&model.Session{}, id).Error
}

func (r *sessionRepository) ListByUserID(userID uint, page, pageSize int) ([]*model.Session, int64, error) {
	var sessions []*model.Session
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.Session{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("user_id = ?", userID).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&sessions).Error

	return sessions, total, err
}

func (r *sessionRepository) ListByProjectID(projectID uint, page, pageSize int) ([]*model.Session, int64, error) {
	var sessions []*model.Session
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.Session{}).Where("project_id = ?", projectID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("project_id = ?", projectID).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&sessions).Error

	return sessions, total, err
}

func (r *sessionRepository) CreateStep(step *model.SessionStep) error {
	return r.db.Create(step).Error
}

func (r *sessionRepository) GetMaxStepOrderIndex(sessionID uint) (int, error) {
	var max int
	// COALESCE 避免空表返回 NULL
	err := r.db.Model(&model.SessionStep{}).
		Where("session_id = ?", sessionID).
		Select("COALESCE(MAX(order_index), 0)").
		Scan(&max).Error
	if err != nil {
		return 0, err
	}
	return max, nil
}

func (r *sessionRepository) GetStepByID(id uint) (*model.SessionStep, error) {
	var step model.SessionStep
	err := r.db.First(&step, id).Error
	if err != nil {
		return nil, err
	}
	return &step, nil
}

func (r *sessionRepository) UpdateStep(step *model.SessionStep) error {
	return r.db.Save(step).Error
}

func (r *sessionRepository) DeleteStep(id uint) error {
	return r.db.Delete(&model.SessionStep{}, id).Error
}

func (r *sessionRepository) ListStepsBySessionID(sessionID uint) ([]*model.SessionStep, error) {
	var steps []*model.SessionStep
	err := r.db.Where("session_id = ?", sessionID).
		Order("order_index ASC").
		Find(&steps).Error
	return steps, err
}
