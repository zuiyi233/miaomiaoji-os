package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// FunctionCallingService Function Calling 服务
type FunctionCallingService interface {
	ExecuteFunctionCallingLoop(ctx context.Context, req ExecuteFunctionCallingLoopRequest) error
	ContinueLoop(ctx context.Context, sessionID, jobID uint) error
}

type functionCallingService struct {
	sessionSvc SessionService
	jobRepo    repository.JobRepository
}

// NewFunctionCallingService 创建 Function Calling 服务
func NewFunctionCallingService(sessionSvc SessionService, jobRepo repository.JobRepository) FunctionCallingService {
	return &functionCallingService{
		sessionSvc: sessionSvc,
		jobRepo:    jobRepo,
	}
}

// ToolCall 工具调用结构
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult 工具执行结果
type ToolResult struct {
	ToolCallID string                 `json:"tool_call_id"`
	Success    bool                   `json:"success"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// AIResponse AI 响应结构
type AIResponse struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ExecuteFunctionCallingLoopRequest 执行 Function Calling 循环请求
type ExecuteFunctionCallingLoopRequest struct {
	SessionID     uint                     `json:"session_id"`
	InitialPrompt string                   `json:"prompt"`
	MaxTurns      int                      `json:"max_turns"`
	Tools         []map[string]interface{} `json:"tools"`
	UserID        uint                     `json:"user_id"`
}

// ExecuteFunctionCallingLoop 执行 Function Calling 多轮对话循环
func (s *functionCallingService) ExecuteFunctionCallingLoop(ctx context.Context, req ExecuteFunctionCallingLoopRequest) error {
	if req.MaxTurns <= 0 {
		req.MaxTurns = 5
	}

	logger.Info("开始 Function Calling 循环",
		logger.Uint("session_id", req.SessionID),
		logger.Int("max_turns", req.MaxTurns),
	)

	// 创建用户输入步骤
	if err := s.sessionSvc.CreateUserStep(req.SessionID, req.InitialPrompt); err != nil {
		logger.Error("创建用户步骤失败", logger.Err(err))
		return err
	}

	currentPrompt := req.InitialPrompt
	for turn := 1; turn <= req.MaxTurns; turn++ {
		logger.Info("执行第 N 轮对话", logger.Int("turn", turn))

		// 调用 AI
		aiResponse, err := s.callAI(ctx, req.SessionID, currentPrompt, req.Tools)
		if err != nil {
			logger.Error("AI 调用失败", logger.Err(err), logger.Int("turn", turn))
			return err
		}

		// 创建 assistant 步骤
		if aiResponse.Content != "" {
			if err := s.sessionSvc.CreateAssistantStep(req.SessionID, aiResponse.Content); err != nil {
				logger.Error("创建 assistant 步骤失败", logger.Err(err))
				return err
			}
		}

		// 检查是否有工具调用
		if len(aiResponse.ToolCalls) == 0 {
			logger.Info("AI 未返回工具调用，循环结束")
			break
		}

		// 执行工具调用
		toolResults, err := s.executeToolCalls(ctx, req.SessionID, req.UserID, aiResponse.ToolCalls)
		if err != nil {
			logger.Error("工具调用执行失败", logger.Err(err))
			return err
		}

		// 检查是否所有工具都失败
		allFailed := true
		for _, result := range toolResults {
			if result.Success {
				allFailed = false
				break
			}
		}
		if allFailed {
			logger.Error("所有工具调用都失败，循环终止")
			return fmt.Errorf("所有工具调用都失败")
		}

		// 构建下一轮提示词
		currentPrompt = s.buildNextPrompt(aiResponse, toolResults)
	}

	logger.Info("Function Calling 循环完成")
	return nil
}

// callAI 调用 AI
func (s *functionCallingService) callAI(ctx context.Context, sessionID uint, prompt string, tools []map[string]interface{}) (*AIResponse, error) {
	// TODO: 实际调用 AI 服务
	// 这里需要集成实际的 AI 调用逻辑
	logger.Debug("调用 AI", logger.Uint("session_id", sessionID))

	// 模拟 AI 响应（实际应该调用 AI 服务）
	response := &AIResponse{
		Content: "这是 AI 的响应",
		ToolCalls: []ToolCall{
			{
				ID:   "call_001",
				Name: "search",
				Arguments: map[string]interface{}{
					"query": "测试查询",
				},
			},
		},
	}

	return response, nil
}

// parseToolCalls 解析 AI 响应中的工具调用
func (s *functionCallingService) parseToolCalls(aiResponse string) ([]ToolCall, error) {
	var response AIResponse
	if err := json.Unmarshal([]byte(aiResponse), &response); err != nil {
		return nil, err
	}
	return response.ToolCalls, nil
}

// executeToolCalls 执行工具调用
func (s *functionCallingService) executeToolCalls(ctx context.Context, sessionID, userID uint, toolCalls []ToolCall) ([]ToolResult, error) {
	results := make([]ToolResult, len(toolCalls))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, toolCall := range toolCalls {
		wg.Add(1)
		go func(idx int, tc ToolCall) {
			defer wg.Done()

			result := s.executeToolCall(ctx, sessionID, userID, tc)

			mu.Lock()
			results[idx] = result
			mu.Unlock()
		}(i, toolCall)
	}

	wg.Wait()
	return results, nil
}

// executeToolCall 执行单个工具调用
func (s *functionCallingService) executeToolCall(ctx context.Context, sessionID, userID uint, toolCall ToolCall) ToolResult {
	logger.Info("执行工具调用",
		logger.String("tool_call_id", toolCall.ID),
		logger.String("tool_name", toolCall.Name),
	)

	// 创建 tool_call 步骤
	if err := s.sessionSvc.CreateToolCallStep(sessionID, toolCall.ID, toolCall.Name, toolCall.Arguments); err != nil {
		logger.Error("创建 tool_call 步骤失败", logger.Err(err))
		return ToolResult{
			ToolCallID: toolCall.ID,
			Success:    false,
			Error:      err.Error(),
		}
	}

	// 创建 Job
	payloadJSON, _ := json.Marshal(toolCall.Arguments)
	jobUUID := uuid.New().String()
	job := &model.Job{
		JobUUID:   jobUUID,
		Type:      model.JobTypePluginInvoke,
		Status:    model.JobStatusQueued,
		UserID:    userID,
		SessionID: sessionID,
		PluginID:  1, // TODO: 根据 toolCall.Name 映射到实际的 PluginID
		Method:    toolCall.Name,
		Payload:   datatypes.JSON(payloadJSON),
	}

	if err := s.jobRepo.Create(job); err != nil {
		logger.Error("创建 Job 失败", logger.Err(err))
		return ToolResult{
			ToolCallID: toolCall.ID,
			Success:    false,
			Error:      err.Error(),
		}
	}

	// 等待 Job 完成
	result := s.waitForJobCompletion(ctx, job.ID, toolCall.ID)
	return result
}

// waitForJobCompletion 等待 Job 完成
func (s *functionCallingService) waitForJobCompletion(ctx context.Context, jobID uint, toolCallID string) ToolResult {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return ToolResult{
				ToolCallID: toolCallID,
				Success:    false,
				Error:      "context cancelled",
			}
		case <-timeout:
			return ToolResult{
				ToolCallID: toolCallID,
				Success:    false,
				Error:      "timeout",
			}
		case <-ticker.C:
			job, err := s.jobRepo.GetByID(jobID)
			if err != nil {
				logger.Error("获取 Job 失败", logger.Err(err))
				continue
			}

			if job.Status == model.JobStatusSucceeded {
				var data map[string]interface{}
				if len(job.Result) > 0 {
					json.Unmarshal(job.Result, &data)
				}
				return ToolResult{
					ToolCallID: toolCallID,
					Success:    true,
					Data:       data,
				}
			} else if job.Status == model.JobStatusFailed {
				return ToolResult{
					ToolCallID: toolCallID,
					Success:    false,
					Error:      job.ErrorMessage,
				}
			}
		}
	}
}

// buildNextPrompt 构建下一轮提示词
func (s *functionCallingService) buildNextPrompt(aiResponse *AIResponse, toolResults []ToolResult) string {
	// 构建包含工具结果的提示词
	prompt := "工具执行结果:\n"
	for _, result := range toolResults {
		if result.Success {
			dataJSON, _ := json.Marshal(result.Data)
			prompt += fmt.Sprintf("- 工具调用 %s 成功: %s\n", result.ToolCallID, string(dataJSON))
		} else {
			prompt += fmt.Sprintf("- 工具调用 %s 失败: %s\n", result.ToolCallID, result.Error)
		}
	}
	return prompt
}

// ContinueLoop 继续 Function Calling 循环（由 Job 完成后调用）
func (s *functionCallingService) ContinueLoop(ctx context.Context, sessionID, jobID uint) error {
	logger.Info("继续 Function Calling 循环",
		logger.Uint("session_id", sessionID),
		logger.Uint("job_id", jobID),
	)

	// TODO: 实现继续循环的逻辑
	// 1. 检查是否还有其他 pending 的 Jobs
	// 2. 如果所有 Jobs 都完成，触发下一轮 AI 调用
	// 3. 如果达到 max_turns，结束循环

	return nil
}
