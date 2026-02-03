package service

import (
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
)

type SessionService interface {
	CreateSession(session *model.Session) error
	GetSession(id uint) (*model.Session, error)
	UpdateSession(session *model.Session) error
	DeleteSession(id uint) error
	ListSessions(userID uint, page, pageSize int) ([]*model.Session, int64, error)
	ListSessionsByProject(projectID uint, page, pageSize int) ([]*model.Session, int64, error)

	CreateStep(step *model.SessionStep) error
	CreateStepAutoOrder(step *model.SessionStep) error
	GetStep(id uint) (*model.SessionStep, error)
	UpdateStep(step *model.SessionStep) error
	DeleteStep(id uint) error
	ListSteps(sessionID uint) ([]*model.SessionStep, error)
}

type sessionService struct {
	sessionRepo repository.SessionRepository
}

func NewSessionService(sessionRepo repository.SessionRepository) SessionService {
	return &sessionService{
		sessionRepo: sessionRepo,
	}
}

func (s *sessionService) CreateSession(session *model.Session) error {
	return s.sessionRepo.Create(session)
}

func (s *sessionService) GetSession(id uint) (*model.Session, error) {
	return s.sessionRepo.GetByID(id)
}

func (s *sessionService) UpdateSession(session *model.Session) error {
	return s.sessionRepo.Update(session)
}

func (s *sessionService) DeleteSession(id uint) error {
	return s.sessionRepo.Delete(id)
}

func (s *sessionService) ListSessions(userID uint, page, pageSize int) ([]*model.Session, int64, error) {
	return s.sessionRepo.ListByUserID(userID, page, pageSize)
}

func (s *sessionService) ListSessionsByProject(projectID uint, page, pageSize int) ([]*model.Session, int64, error) {
	return s.sessionRepo.ListByProjectID(projectID, page, pageSize)
}

func (s *sessionService) CreateStep(step *model.SessionStep) error {
	return s.sessionRepo.CreateStep(step)
}

func (s *sessionService) CreateStepAutoOrder(step *model.SessionStep) error {
	max, err := s.sessionRepo.GetMaxStepOrderIndex(step.SessionID)
	if err != nil {
		return err
	}
	step.OrderIndex = max + 1
	return s.sessionRepo.CreateStep(step)
}

func (s *sessionService) GetStep(id uint) (*model.SessionStep, error) {
	return s.sessionRepo.GetStepByID(id)
}

func (s *sessionService) UpdateStep(step *model.SessionStep) error {
	return s.sessionRepo.UpdateStep(step)
}

func (s *sessionService) DeleteStep(id uint) error {
	return s.sessionRepo.DeleteStep(id)
}

func (s *sessionService) ListSteps(sessionID uint) ([]*model.SessionStep, error) {
	return s.sessionRepo.ListStepsBySessionID(sessionID)
}
