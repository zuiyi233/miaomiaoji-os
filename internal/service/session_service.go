package service

import (
	"encoding/json"
	"fmt"
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

	// Function Calling 相关方法
	CreateUserStep(sessionID uint, content string) error
	CreateAssistantStep(sessionID uint, content string) error
	CreateToolCallStep(sessionID uint, toolCallID, toolName string, arguments map[string]interface{}) error
	CreateToolResultStep(sessionID uint, toolCallID string, result map[string]interface{}, metadata map[string]interface{}) error
	GetSessionHistory(sessionID uint, includeToolCalls bool) ([]*model.SessionStep, error)
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

// CreateUserStep 创建用户输入步骤
func (s *sessionService) CreateUserStep(sessionID uint, content string) error {
	max, err := s.sessionRepo.GetMaxStepOrderIndex(sessionID)
	if err != nil {
		return err
	}

	step := &model.SessionStep{
		SessionID:  sessionID,
		StepType:   "user",
		Content:    content,
		OrderIndex: max + 1,
	}

	return s.sessionRepo.CreateStep(step)
}

// CreateAssistantStep 创建 AI 响应步骤
func (s *sessionService) CreateAssistantStep(sessionID uint, content string) error {
	max, err := s.sessionRepo.GetMaxStepOrderIndex(sessionID)
	if err != nil {
		return err
	}

	step := &model.SessionStep{
		SessionID:  sessionID,
		StepType:   "assistant",
		Content:    content,
		OrderIndex: max + 1,
	}

	return s.sessionRepo.CreateStep(step)
}

// CreateToolCallStep 创建工具调用步骤
func (s *sessionService) CreateToolCallStep(sessionID uint, toolCallID, toolName string, arguments map[string]interface{}) error {
	max, err := s.sessionRepo.GetMaxStepOrderIndex(sessionID)
	if err != nil {
		return err
	}

	metadataMap := map[string]interface{}{
		"tool_call": map[string]interface{}{
			"id":        toolCallID,
			"name":      toolName,
			"arguments": arguments,
		},
	}

	metadataJSON, err := json.Marshal(metadataMap)
	if err != nil {
		return err
	}

	step := &model.SessionStep{
		SessionID:  sessionID,
		StepType:   "tool_call",
		ToolCallID: toolCallID,
		Content:    fmt.Sprintf("调用工具: %s", toolName),
		Metadata:   metadataJSON,
		OrderIndex: max + 1,
	}

	return s.sessionRepo.CreateStep(step)
}

// CreateToolResultStep 创建工具结果步骤
func (s *sessionService) CreateToolResultStep(sessionID uint, toolCallID string, result map[string]interface{}, metadata map[string]interface{}) error {
	max, err := s.sessionRepo.GetMaxStepOrderIndex(sessionID)
	if err != nil {
		return err
	}

	metadataMap := map[string]interface{}{
		"tool_result": result,
	}
	if metadata != nil {
		for k, v := range metadata {
			metadataMap[k] = v
		}
	}

	metadataJSON, err := json.Marshal(metadataMap)
	if err != nil {
		return err
	}

	resultJSON, _ := json.Marshal(result)
	step := &model.SessionStep{
		SessionID:  sessionID,
		StepType:   "tool_result",
		ToolCallID: toolCallID,
		Content:    string(resultJSON),
		Metadata:   metadataJSON,
		OrderIndex: max + 1,
	}

	return s.sessionRepo.CreateStep(step)
}

// GetSessionHistory 获取会话历史
func (s *sessionService) GetSessionHistory(sessionID uint, includeToolCalls bool) ([]*model.SessionStep, error) {
	steps, err := s.sessionRepo.ListStepsBySessionID(sessionID)
	if err != nil {
		return nil, err
	}

	if includeToolCalls {
		return steps, nil
	}

	// 过滤掉 tool_call 和 tool_result 类型的步骤
	filtered := make([]*model.SessionStep, 0)
	for _, step := range steps {
		if step.StepType != "tool_call" && step.StepType != "tool_result" {
			filtered = append(filtered, step)
		}
	}

	return filtered, nil
}
