package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/sse"

	"gorm.io/datatypes"
)

// ChapterOutline 章节大纲
type ChapterOutline struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// AgentWriterConfig 写作工作流配置
type AgentWriterConfig struct {
	ProjectID      uint             `json:"project_id"`
	DocumentID     uint             `json:"document_id"`
	Prompt         string           `json:"prompt"`
	Outline        []ChapterOutline `json:"outline"`
	CurrentChapter int              `json:"current_chapter"`
	TotalChapters  int              `json:"total_chapters"`
	Provider       string           `json:"provider"`
	Path           string           `json:"path"`
}

// AgentWriterService 写作代理服务
type AgentWriterService struct {
	sessionService  SessionService
	documentService DocumentService
	aiConfigService AIConfigService
	cancelFuncs     map[uint]context.CancelFunc
	mu              sync.RWMutex
}

// NewAgentWriterService 创建写作代理服务
func NewAgentWriterService(sessionService SessionService, documentService DocumentService, aiConfigService AIConfigService) *AgentWriterService {
	return &AgentWriterService{
		sessionService:  sessionService,
		documentService: documentService,
		aiConfigService: aiConfigService,
		cancelFuncs:     make(map[uint]context.CancelFunc),
	}
}

// StartWritingTask 启动写作任务
func (s *AgentWriterService) StartWritingTask(projectID, documentID uint, userID uint, prompt string, outline []ChapterOutline, provider, path string) (*model.Session, error) {
	// 构建工作流配置
	config := AgentWriterConfig{
		ProjectID:      projectID,
		DocumentID:     documentID,
		Prompt:         prompt,
		Outline:        outline,
		CurrentChapter: 0,
		TotalChapters:  len(outline),
		Provider:       provider,
		Path:           path,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		logger.Error("序列化工作流配置失败", logger.Err(err))
		return nil, fmt.Errorf("failed to marshal workflow config")
	}

	// 创建 Session
	session := &model.Session{
		Title:          fmt.Sprintf("AgentWriter: %s", prompt),
		Mode:           "AgentWriter",
		ProjectID:      projectID,
		UserID:         userID,
		WorkflowType:   "agent_writer",
		WorkflowStatus: "pending",
		WorkflowConfig: datatypes.JSON(configJSON),
	}

	if err := s.sessionService.CreateSession(session); err != nil {
		logger.Error("创建会话失败", logger.Err(err))
		return nil, fmt.Errorf("failed to create session")
	}

	// 创建初始步骤（用户输入）
	userStep := &model.SessionStep{
		SessionID:  session.ID,
		Title:      "用户输入",
		Content:    fmt.Sprintf("Prompt: %s\n\nOutline: %v", prompt, outline),
		StepType:   "user",
		OrderIndex: 0,
	}
	if err := s.sessionService.CreateStep(userStep); err != nil {
		logger.Error("创建用户步骤失败", logger.Err(err))
	}

	// 异步执行工作流
	go s.executeWritingWorkflow(session.ID)

	return session, nil
}

// executeWritingWorkflow 执行写作工作流
func (s *AgentWriterService) executeWritingWorkflow(sessionID uint) {
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.cancelFuncs[sessionID] = cancel
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.cancelFuncs, sessionID)
		s.mu.Unlock()
	}()

	// 获取 Session
	session, err := s.sessionService.GetSession(sessionID)
	if err != nil {
		logger.Error("获取会话失败", logger.Err(err))
		return
	}

	// 解析配置
	var config AgentWriterConfig
	if err := json.Unmarshal([]byte(session.WorkflowConfig), &config); err != nil {
		logger.Error("解析工作流配置失败", logger.Err(err))
		s.updateWorkflowStatus(sessionID, "error")
		return
	}

	// 更新状态为运行中
	s.updateWorkflowStatus(sessionID, "running")

	hub := sse.GetHub()
	sessionIDStr := fmt.Sprintf("%d", sessionID)

	// 遍历章节生成
	for i, chapter := range config.Outline {
		select {
		case <-ctx.Done():
			logger.Info("工作流被取消", logger.Uint("session_id", sessionID))
			s.updateWorkflowStatus(sessionID, "cancelled")
			return
		default:
		}

		// 更新当前章节
		config.CurrentChapter = i
		s.updateWorkflowConfig(sessionID, config)

		// 推送章节开始事件
		hub.BroadcastToSession(sessionIDStr, sse.NewChapterStartEvent(map[string]interface{}{
			"session_id":     sessionID,
			"chapter_index":  i,
			"chapter_title":  chapter.Title,
			"total_chapters": config.TotalChapters,
		}))

		// 生成章节
		if err := s.generateChapter(ctx, sessionID, config.DocumentID, chapter, config, i); err != nil {
			logger.Error("生成章节失败", logger.Err(err), logger.Int("chapter_index", i))

			// 推送错误事件但继续下一章节
			hub.BroadcastToSession(sessionIDStr, sse.NewStepErrorEvent(map[string]interface{}{
				"session_id":    sessionID,
				"chapter_index": i,
				"error":         err.Error(),
			}))
			continue
		}

		// 推送章节完成事件
		hub.BroadcastToSession(sessionIDStr, sse.NewChapterCompletedEvent(map[string]interface{}{
			"session_id":    sessionID,
			"chapter_index": i,
			"chapter_title": chapter.Title,
		}))
	}

	// 所有章节完成
	s.updateWorkflowStatus(sessionID, "completed")
	hub.BroadcastToSession(sessionIDStr, sse.NewWorkflowCompletedEvent(map[string]interface{}{
		"session_id":     sessionID,
		"total_chapters": config.TotalChapters,
		"document_id":    config.DocumentID,
	}))

	logger.Info("写作工作流完成", logger.Uint("session_id", sessionID))
}

// generateChapter 生成单个章节
func (s *AgentWriterService) generateChapter(ctx context.Context, sessionID, documentID uint, chapter ChapterOutline, config AgentWriterConfig, chapterIndex int) error {
	// 创建章节步骤
	step := &model.SessionStep{
		SessionID:    sessionID,
		Title:        chapter.Title,
		Content:      "",
		IsStreaming:  true,
		StreamStatus: "streaming",
		StepType:     "assistant",
		OrderIndex:   chapterIndex + 1,
	}

	if err := s.sessionService.CreateStep(step); err != nil {
		logger.Error("创建章节步骤失败", logger.Err(err))
		return fmt.Errorf("failed to create chapter step")
	}

	hub := sse.GetHub()
	sessionIDStr := fmt.Sprintf("%d", sessionID)
	var contentBuilder strings.Builder
	chunkCount := 0
	lastUpdateTime := time.Now()

	// 构建 AI 请求体
	requestBody := s.buildChapterRequest(config.Prompt, chapter)

	// 流式生成章节内容
	chunkHandler := func(chunk string) error {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		default:
		}

		contentBuilder.WriteString(chunk)
		chunkCount++

		// 推送进度事件
		hub.BroadcastToSession(sessionIDStr, sse.NewChapterProgressEvent(map[string]interface{}{
			"session_id":    sessionID,
			"step_id":       step.ID,
			"chapter_index": chapterIndex,
			"chunk":         chunk,
		}))

		// 定期更新数据库和文档
		if chunkCount%10 == 0 || time.Since(lastUpdateTime) > 2*time.Second {
			step.Content = contentBuilder.String()
			if err := s.sessionService.UpdateStep(step); err != nil {
				logger.Error("更新步骤内容失败", logger.Err(err))
			}
			lastUpdateTime = time.Now()
		}

		return nil
	}

	// 调用 AI 流式生成
	err := CallAIStream(ctx, s.aiConfigService, config.Provider, config.Path, requestBody, chunkHandler)

	// 最终更新
	step.Content = contentBuilder.String()
	if err != nil {
		step.StreamStatus = "error"
		step.IsStreaming = false
		if updateErr := s.sessionService.UpdateStep(step); updateErr != nil {
			logger.Error("更新步骤状态失败", logger.Err(updateErr))
		}
		return err
	}

	step.StreamStatus = "completed"
	step.IsStreaming = false
	if err := s.sessionService.UpdateStep(step); err != nil {
		logger.Error("更新步骤状态失败", logger.Err(err))
	}

	// 保存到文档
	if err := s.documentService.AppendChapter(documentID, chapter.Title, step.Content); err != nil {
		logger.Error("保存章节到文档失败", logger.Err(err))
		return fmt.Errorf("failed to save chapter to document")
	}

	return nil
}

// buildChapterRequest 构建章节生成请求
func (s *AgentWriterService) buildChapterRequest(prompt string, chapter ChapterOutline) string {
	requestMap := map[string]interface{}{
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是一个专业的小说写作助手，擅长根据大纲生成高质量的章节内容。",
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("根据以下要求生成章节内容：\n\n总体要求：%s\n\n章节标题：%s\n章节描述：%s\n\n请生成完整的章节内容。", prompt, chapter.Title, chapter.Description),
			},
		},
		"stream": true,
	}

	requestJSON, _ := json.Marshal(requestMap)
	return string(requestJSON)
}

// CancelWritingTask 取消写作任务
func (s *AgentWriterService) CancelWritingTask(sessionID uint) error {
	s.mu.Lock()
	cancel, exists := s.cancelFuncs[sessionID]
	s.mu.Unlock()

	if !exists {
		return fmt.Errorf("session not found or already completed")
	}

	cancel()
	s.updateWorkflowStatus(sessionID, "cancelled")

	logger.Info("写作任务已取消", logger.Uint("session_id", sessionID))
	return nil
}

// updateWorkflowStatus 更新工作流状态
func (s *AgentWriterService) updateWorkflowStatus(sessionID uint, status string) {
	session, err := s.sessionService.GetSession(sessionID)
	if err != nil {
		logger.Error("获取会话失败", logger.Err(err))
		return
	}

	session.WorkflowStatus = status
	if err := s.sessionService.UpdateSession(session); err != nil {
		logger.Error("更新工作流状态失败", logger.Err(err))
	}
}

// updateWorkflowConfig 更新工作流配置
func (s *AgentWriterService) updateWorkflowConfig(sessionID uint, config AgentWriterConfig) {
	session, err := s.sessionService.GetSession(sessionID)
	if err != nil {
		logger.Error("获取会话失败", logger.Err(err))
		return
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		logger.Error("序列化工作流配置失败", logger.Err(err))
		return
	}

	session.WorkflowConfig = datatypes.JSON(configJSON)
	if err := s.sessionService.UpdateSession(session); err != nil {
		logger.Error("更新工作流配置失败", logger.Err(err))
	}
}
