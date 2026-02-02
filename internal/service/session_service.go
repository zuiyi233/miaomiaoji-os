package service

import (
	"encoding/json"

	"gorm.io/datatypes"
	"novel-agent-os/internal/model"
	"novel-agent-os/internal/repository"
)

// SessionService 会话服务接口
type SessionService interface {
	Create(title, mode string, projectID, userID uint) (*model.Session, error)
	GetByID(id uint) (*model.Session, error)
	ListByProject(projectID uint, page, size int) ([]*model.Session, int64, error)
	ListByUser(userID uint, page, size int) ([]*model.Session, int64, error)
	Update(id uint, title string) (*model.Session, error)
	Delete(id uint) error

	// SessionStep 相关
	AddStep(sessionID uint, title, content, formatType string, metadata map[string]interface{}) (*model.SessionStep, error)
	GetStep(id uint) (*model.SessionStep, error)
	ListSteps(sessionID uint) ([]*model.SessionStep, error)
	UpdateStep(id uint, title, content string, metadata map[string]interface{}) (*model.SessionStep, error)
	DeleteStep(id uint) error
	ReorderSteps(sessionID uint, stepIDs []uint) error
}

// sessionService 会话服务实现
type sessionService struct {
	sessionRepo repository.SessionRepository
}

// NewSessionService 创建会话服务实例
func NewSessionService(sessionRepo repository.SessionRepository) SessionService {
	return &sessionService{
		sessionRepo: sessionRepo,
	}
}

func (s *sessionService) Create(title, mode string, projectID, userID uint) (*model.Session, error) {
	session := &model.Session{
		Title:     title,
		Mode:      mode,
		ProjectID: projectID,
		UserID:    userID,
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *sessionService) GetByID(id uint) (*model.Session, error) {
	return s.sessionRepo.FindByID(id)
}

func (s *sessionService) ListByProject(projectID uint, page, size int) ([]*model.Session, int64, error) {
	return s.sessionRepo.ListByProject(projectID, page, size)
}

func (s *sessionService) ListByUser(userID uint, page, size int) ([]*model.Session, int64, error) {
	return s.sessionRepo.ListByUser(userID, page, size)
}

func (s *sessionService) Update(id uint, title string) (*model.Session, error) {
	session, err := s.sessionRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	session.Title = title
	if err := s.sessionRepo.Update(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *sessionService) Delete(id uint) error {
	// 先删除关联的步骤
	if err := s.sessionRepo.DeleteStepsBySession(id); err != nil {
		return err
	}
	return s.sessionRepo.Delete(id)
}

// SessionStep 实现

func (s *sessionService) AddStep(sessionID uint, title, content, formatType string, metadata map[string]interface{}) (*model.SessionStep, error) {
	// 获取当前最大顺序
	maxIndex, err := s.sessionRepo.GetMaxOrderIndex(sessionID)
	if err != nil {
		return nil, err
	}

	// 转换 metadata 为 JSON
	var metadataJSON datatypes.JSON
	if metadata != nil {
		jsonBytes, err := json.Marshal(metadata)
		if err != nil {
			return nil, err
		}
		metadataJSON = datatypes.JSON(jsonBytes)
	}

	step := &model.SessionStep{
		Title:      title,
		Content:    content,
		FormatType: formatType,
		OrderIndex: maxIndex + 1,
		Metadata:   metadataJSON,
		SessionID:  sessionID,
	}

	if err := s.sessionRepo.CreateStep(step); err != nil {
		return nil, err
	}

	return step, nil
}

func (s *sessionService) GetStep(id uint) (*model.SessionStep, error) {
	return s.sessionRepo.FindStepByID(id)
}

func (s *sessionService) ListSteps(sessionID uint) ([]*model.SessionStep, error) {
	return s.sessionRepo.ListStepsBySession(sessionID)
}

func (s *sessionService) UpdateStep(id uint, title, content string, metadata map[string]interface{}) (*model.SessionStep, error) {
	step, err := s.sessionRepo.FindStepByID(id)
	if err != nil {
		return nil, err
	}

	if title != "" {
		step.Title = title
	}
	if content != "" {
		step.Content = content
	}
	if metadata != nil {
		jsonBytes, err := json.Marshal(metadata)
		if err != nil {
			return nil, err
		}
		step.Metadata = datatypes.JSON(jsonBytes)
	}

	if err := s.sessionRepo.UpdateStep(step); err != nil {
		return nil, err
	}

	return step, nil
}

func (s *sessionService) DeleteStep(id uint) error {
	return s.sessionRepo.DeleteStep(id)
}

func (s *sessionService) ReorderSteps(sessionID uint, stepIDs []uint) error {
	steps, err := s.sessionRepo.ListStepsBySession(sessionID)
	if err != nil {
		return err
	}

	// 创建 ID 到 step 的映射
	stepMap := make(map[uint]*model.SessionStep)
	for _, step := range steps {
		stepMap[step.ID] = step
	}

	// 按新顺序更新 order_index
	for i, stepID := range stepIDs {
		if step, exists := stepMap[stepID]; exists {
			step.OrderIndex = i + 1
			if err := s.sessionRepo.UpdateStep(step); err != nil {
				return err
			}
		}
	}

	return nil
}
