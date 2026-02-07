package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/sse"
)

// WorkflowStreamService 流式工作流服务
type WorkflowStreamService struct {
	aiConfigService AIConfigService
	sessionRepo     repository.SessionRepository
}

// NewWorkflowStreamService 创建流式工作流服务
func NewWorkflowStreamService(aiConfigService AIConfigService, sessionRepo repository.SessionRepository) *WorkflowStreamService {
	return &WorkflowStreamService{
		aiConfigService: aiConfigService,
		sessionRepo:     sessionRepo,
	}
}

// ExecuteWorkflowStreamRequest 执行流式工作流请求
type ExecuteWorkflowStreamRequest struct {
	SessionID uint
	StepTitle string
	Provider  string
	Path      string
	Body      string
	Timeout   time.Duration
}

// ExecuteWorkflowStreamResponse 执行流式工作流响应
type ExecuteWorkflowStreamResponse struct {
	StepID    uint   `json:"step_id"`
	SessionID uint   `json:"session_id"`
	Message   string `json:"message"`
}

// ExecuteWorkflowStream 执行流式工作流
func (s *WorkflowStreamService) ExecuteWorkflowStream(req ExecuteWorkflowStreamRequest) (*ExecuteWorkflowStreamResponse, error) {
	// 创建 SessionStep
	step := &model.SessionStep{
		SessionID:    req.SessionID,
		Title:        req.StepTitle,
		Content:      "",
		IsStreaming:  true,
		StreamStatus: "streaming",
		OrderIndex:   0,
	}

	if err := s.sessionRepo.CreateStep(step); err != nil {
		logger.Error("failed to create step", logger.Err(err))
		return nil, fmt.Errorf("failed to create step")
	}

	// 异步执行流式调用
	go s.executeStreamInBackground(req, step)

	return &ExecuteWorkflowStreamResponse{
		StepID:    step.ID,
		SessionID: req.SessionID,
		Message:   "Stream started",
	}, nil
}

// executeStreamInBackground 在后台执行流式调用
func (s *WorkflowStreamService) executeStreamInBackground(req ExecuteWorkflowStreamRequest, step *model.SessionStep) {
	ctx := context.Background()
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	hub := sse.GetHub()
	sessionIDStr := fmt.Sprintf("%d", req.SessionID)
	var contentBuilder strings.Builder
	chunkCount := 0
	lastUpdateTime := time.Now()

	chunkHandler := func(chunk string) error {
		contentBuilder.WriteString(chunk)
		chunkCount++

		// 推送 chunk 事件
		hub.BroadcastToSession(sessionIDStr, sse.NewStepChunkEvent(map[string]interface{}{
			"session_id": req.SessionID,
			"step_id":    step.ID,
			"chunk":      chunk,
			"is_final":   false,
		}))

		// 每 10 个 chunk 或每 2 秒更新一次数据库
		if chunkCount%10 == 0 || time.Since(lastUpdateTime) > 2*time.Second {
			step.Content = contentBuilder.String()
			if err := s.sessionRepo.UpdateStep(step); err != nil {
				logger.Error("failed to update step content", logger.Err(err))
			}
			lastUpdateTime = time.Now()
		}

		return nil
	}

	// 执行流式调用
	err := CallAIStream(ctx, s.aiConfigService, req.Provider, req.Path, req.Body, chunkHandler)

	// 最终更新
	step.Content = contentBuilder.String()
	if err != nil {
		logger.Error("stream execution failed", logger.Err(err))
		step.StreamStatus = "error"
		step.IsStreaming = false

		// 推送错误事件
		hub.BroadcastToSession(sessionIDStr, sse.NewStepErrorEvent(map[string]interface{}{
			"session_id": req.SessionID,
			"step_id":    step.ID,
			"error":      err.Error(),
		}))
	} else {
		step.StreamStatus = "completed"
		step.IsStreaming = false

		// 推送完成事件
		hub.BroadcastToSession(sessionIDStr, sse.NewStepCompletedEvent(map[string]interface{}{
			"session_id": req.SessionID,
			"step_id":    step.ID,
			"content":    step.Content,
		}))
	}

	if err := s.sessionRepo.UpdateStep(step); err != nil {
		logger.Error("failed to update step final status", logger.Err(err))
	}
}

// AggregateChunks 聚合 chunks 到 SessionStep
func (s *WorkflowStreamService) AggregateChunks(sessionID uint, stepID uint, chunks []string) error {
	step, err := s.sessionRepo.GetStepByID(stepID)
	if err != nil {
		return fmt.Errorf("step not found")
	}

	if step.SessionID != sessionID {
		return fmt.Errorf("step does not belong to session")
	}

	content := strings.Join(chunks, "")
	step.Content = content

	if err := s.sessionRepo.UpdateStep(step); err != nil {
		return fmt.Errorf("failed to update step: %w", err)
	}

	return nil
}

// CancelStream 取消流式执行
func (s *WorkflowStreamService) CancelStream(stepID uint) error {
	step, err := s.sessionRepo.GetStepByID(stepID)
	if err != nil {
		return fmt.Errorf("step not found")
	}

	if !step.IsStreaming {
		return fmt.Errorf("step is not streaming")
	}

	step.StreamStatus = "cancelled"
	step.IsStreaming = false

	if err := s.sessionRepo.UpdateStep(step); err != nil {
		return fmt.Errorf("failed to cancel stream: %w", err)
	}

	// 推送取消事件
	hub := sse.GetHub()
	hub.BroadcastToSession(fmt.Sprintf("%d", step.SessionID), sse.NewStepErrorEvent(map[string]interface{}{
		"session_id": step.SessionID,
		"step_id":    step.ID,
		"error":      "Stream cancelled",
	}))

	return nil
}

// GetStreamStatus 获取流式状态
func (s *WorkflowStreamService) GetStreamStatus(stepID uint) (string, error) {
	step, err := s.sessionRepo.GetStepByID(stepID)
	if err != nil {
		return "", fmt.Errorf("step not found")
	}

	return step.StreamStatus, nil
}
