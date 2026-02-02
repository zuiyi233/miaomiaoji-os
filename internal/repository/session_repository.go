package repository

import (
	"gorm.io/gorm"
	"novel-agent-os/internal/model"
)

// SessionRepository 会话数据访问接口
type SessionRepository interface {
	Create(session *model.Session) error
	FindByID(id uint) (*model.Session, error)
	ListByProject(projectID uint, page, size int) ([]*model.Session, int64, error)
	ListByUser(userID uint, page, size int) ([]*model.Session, int64, error)
	Update(session *model.Session) error
	Delete(id uint) error

	// SessionStep 相关
	CreateStep(step *model.SessionStep) error
	FindStepByID(id uint) (*model.SessionStep, error)
	ListStepsBySession(sessionID uint) ([]*model.SessionStep, error)
	UpdateStep(step *model.SessionStep) error
	DeleteStep(id uint) error
	DeleteStepsBySession(sessionID uint) error
	GetMaxOrderIndex(sessionID uint) (int, error)
}

// sessionRepository 会话数据访问实现
type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository 创建会话仓库实例
func NewSessionRepository() SessionRepository {
	return &sessionRepository{
		db: GetDB(),
	}
}

func (r *sessionRepository) Create(session *model.Session) error {
	return r.db.Create(session).Error
}

func (r *sessionRepository) FindByID(id uint) (*model.Session, error) {
	var session model.Session
	err := r.db.Preload("Steps").First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) ListByProject(projectID uint, page, size int) ([]*model.Session, int64, error) {
	var sessions []*model.Session
	var total int64

	offset := (page - 1) * size

	err := r.db.Model(&model.Session{}).Where("project_id = ?", projectID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("project_id = ?", projectID).
		Order("created_at DESC").
		Offset(offset).Limit(size).
		Find(&sessions).Error

	return sessions, total, err
}

func (r *sessionRepository) ListByUser(userID uint, page, size int) ([]*model.Session, int64, error) {
	var sessions []*model.Session
	var total int64

	offset := (page - 1) * size

	err := r.db.Model(&model.Session{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(size).
		Find(&sessions).Error

	return sessions, total, err
}

func (r *sessionRepository) Update(session *model.Session) error {
	return r.db.Save(session).Error
}

func (r *sessionRepository) Delete(id uint) error {
	return r.db.Delete(&model.Session{}, id).Error
}

// SessionStep 实现

func (r *sessionRepository) CreateStep(step *model.SessionStep) error {
	return r.db.Create(step).Error
}

func (r *sessionRepository) FindStepByID(id uint) (*model.SessionStep, error) {
	var step model.SessionStep
	err := r.db.First(&step, id).Error
	if err != nil {
		return nil, err
	}
	return &step, nil
}

func (r *sessionRepository) ListStepsBySession(sessionID uint) ([]*model.SessionStep, error) {
	var steps []*model.SessionStep
	err := r.db.Where("session_id = ?", sessionID).
		Order("order_index ASC, created_at ASC").
		Find(&steps).Error
	return steps, err
}

func (r *sessionRepository) UpdateStep(step *model.SessionStep) error {
	return r.db.Save(step).Error
}

func (r *sessionRepository) DeleteStep(id uint) error {
	return r.db.Delete(&model.SessionStep{}, id).Error
}

func (r *sessionRepository) DeleteStepsBySession(sessionID uint) error {
	return r.db.Where("session_id = ?", sessionID).Delete(&model.SessionStep{}).Error
}

func (r *sessionRepository) GetMaxOrderIndex(sessionID uint) (int, error) {
	var maxIndex int
	err := r.db.Model(&model.SessionStep{}).
		Where("session_id = ?", sessionID).
		Select("COALESCE(MAX(order_index), 0)").
		Scan(&maxIndex).Error
	return maxIndex, err
}
