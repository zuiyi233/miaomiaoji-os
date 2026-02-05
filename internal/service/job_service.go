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
	"novel-agent-os-backend/pkg/sse"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type JobService interface {
	CreatePluginInvokeJob(userID uint, sessionID uint, projectID *uint, pluginID uint, method string, payload map[string]interface{}, authorizationHeader string) (*model.Job, error)
	CreatePluginInvokeJobFromSession(userID uint, sessionID uint, pluginID uint, method string, payload map[string]interface{}, authorizationHeader string) (*model.Job, error)
	GetJobByUUID(jobUUID string) (*model.Job, error)
	CancelJob(userID uint, jobUUID string) (*model.Job, error)
}

type jobService struct {
	jobRepo     repository.JobRepository
	sessionRepo repository.SessionRepository
	pluginSvc   PluginService
	sessionSvc  SessionService

	queue chan string

	mu        sync.RWMutex
	authByJob map[string]string
	cancelBy  map[string]context.CancelFunc
}

func NewJobService(jobRepo repository.JobRepository, sessionRepo repository.SessionRepository, pluginSvc PluginService, sessionSvc SessionService) JobService {
	s := &jobService{
		jobRepo:     jobRepo,
		sessionRepo: sessionRepo,
		pluginSvc:   pluginSvc,
		sessionSvc:  sessionSvc,
		queue:       make(chan string, 1000),
		authByJob:   make(map[string]string),
		cancelBy:    make(map[string]context.CancelFunc),
	}

	// MVP：进程内后台 worker
	go s.worker()

	return s
}

func (s *jobService) CreatePluginInvokeJob(userID uint, sessionID uint, projectID *uint, pluginID uint, method string, payload map[string]interface{}, authorizationHeader string) (*model.Job, error) {
	// 校验 session 归属
	sess, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}
	if sess.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload failed: %w", err)
	}

	jobUUID := uuid.New().String()
	job := &model.Job{
		JobUUID:   jobUUID,
		Type:      model.JobTypePluginInvoke,
		Status:    model.JobStatusQueued,
		Progress:  0,
		UserID:    userID,
		SessionID: sessionID,
		ProjectID: projectID,
		PluginID:  pluginID,
		Method:    method,
		Payload:   datatypes.JSON(payloadJSON),
	}

	if err := s.jobRepo.Create(job); err != nil {
		return nil, err
	}

	if authorizationHeader != "" {
		s.mu.Lock()
		s.authByJob[jobUUID] = authorizationHeader
		s.mu.Unlock()
	}

	// SSE：job.created
	s.broadcastJobEvent(job.SessionID, sse.EventType("job.created"), map[string]interface{}{
		"job_uuid":   job.JobUUID,
		"status":     job.Status,
		"progress":   job.Progress,
		"plugin_id":  job.PluginID,
		"session_id": job.SessionID,
	})

	// 入队
	select {
	case s.queue <- jobUUID:
	default:
		logger.Warn("job queue full, dropping job", logger.String("job_uuid", jobUUID))
	}

	return job, nil
}

// CreatePluginInvokeJobFromSession 便捷方法：根据 session_id 自动补 project_id
func (s *jobService) CreatePluginInvokeJobFromSession(userID uint, sessionID uint, pluginID uint, method string, payload map[string]interface{}, authorizationHeader string) (*model.Job, error) {
	sess, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}
	pid := sess.ProjectID
	return s.CreatePluginInvokeJob(userID, sessionID, &pid, pluginID, method, payload, authorizationHeader)
}

func (s *jobService) GetJobByUUID(jobUUID string) (*model.Job, error) {
	return s.jobRepo.GetByUUID(jobUUID)
}

func (s *jobService) CancelJob(userID uint, jobUUID string) (*model.Job, error) {
	job, err := s.jobRepo.GetByUUID(jobUUID)
	if err != nil {
		return nil, err
	}
	if job.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	// 若已结束则直接返回
	if job.Status == model.JobStatusSucceeded || job.Status == model.JobStatusFailed || job.Status == model.JobStatusCanceled {
		return job, nil
	}

	s.mu.Lock()
	if cancel, ok := s.cancelBy[jobUUID]; ok {
		cancel()
	}
	s.mu.Unlock()

	job.Status = model.JobStatusCanceled
	job.Progress = 0
	now := time.Now()
	job.FinishedAt = &now
	if err := s.jobRepo.Update(job); err != nil {
		return nil, err
	}

	s.broadcastJobEvent(job.SessionID, sse.EventType("job.canceled"), map[string]interface{}{
		"job_uuid":   job.JobUUID,
		"status":     job.Status,
		"progress":   job.Progress,
		"plugin_id":  job.PluginID,
		"session_id": job.SessionID,
	})

	return job, nil
}

func (s *jobService) worker() {
	for jobUUID := range s.queue {
		job, err := s.jobRepo.GetByUUID(jobUUID)
		if err != nil {
			logger.Error("load job failed", logger.Err(err), logger.String("job_uuid", jobUUID))
			continue
		}
		if job.Status != model.JobStatusQueued {
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		s.mu.Lock()
		s.cancelBy[jobUUID] = cancel
		authHeader := s.authByJob[jobUUID]
		s.mu.Unlock()

		now := time.Now()
		job.Status = model.JobStatusRunning
		job.Progress = 10
		job.StartedAt = &now
		_ = s.jobRepo.Update(job)

		s.broadcastJobEvent(job.SessionID, sse.EventType("job.started"), map[string]interface{}{
			"job_uuid":   job.JobUUID,
			"status":     job.Status,
			"progress":   job.Progress,
			"plugin_id":  job.PluginID,
			"session_id": job.SessionID,
		})

		payloadMap := map[string]interface{}{}
		_ = json.Unmarshal(job.Payload, &payloadMap)

		// 这里复用现有插件调用（内部仍是30s timeout）
		res, invokeErr := s.pluginSvc.InvokePlugin(ctx, job.PluginID, job.Method, payloadMap, authHeader)
		if invokeErr != nil {
			job.Status = model.JobStatusFailed
			job.Progress = 0
			job.ErrorMessage = invokeErr.Error()
			end := time.Now()
			job.FinishedAt = &end
			_ = s.jobRepo.Update(job)
			s.broadcastJobEvent(job.SessionID, sse.EventType("job.failed"), map[string]interface{}{
				"job_uuid":   job.JobUUID,
				"status":     job.Status,
				"progress":   job.Progress,
				"plugin_id":  job.PluginID,
				"session_id": job.SessionID,
				"error":      job.ErrorMessage,
			})
			cancel()
			s.cleanupJobMemory(jobUUID)
			continue
		}

		job.Progress = 80
		_ = s.jobRepo.Update(job)
		s.broadcastJobEvent(job.SessionID, sse.EventType("job.progress"), map[string]interface{}{
			"job_uuid":   job.JobUUID,
			"status":     job.Status,
			"progress":   job.Progress,
			"plugin_id":  job.PluginID,
			"session_id": job.SessionID,
		})

		resultJSON, _ := json.Marshal(res)
		job.Result = datatypes.JSON(resultJSON)
		job.Status = model.JobStatusSucceeded
		job.Progress = 100
		end := time.Now()
		job.FinishedAt = &end
		_ = s.jobRepo.Update(job)

		// A+B：同时追加 SessionStep（沉淀工作流产物）
		step := &model.SessionStep{
			Title:      fmt.Sprintf("plugin:%d %s", job.PluginID, job.Method),
			Content:    string(resultJSON),
			FormatType: "plugin_result",
			SessionID:  job.SessionID,
		}
		_ = s.sessionSvc.CreateStepAutoOrder(step)

		s.broadcastJobEvent(job.SessionID, sse.EventTypeStepAppended, map[string]interface{}{
			"step_id":   step.ID,
			"title":     step.Title,
			"content":   step.Content,
			"job_uuid":  job.JobUUID,
			"plugin_id": job.PluginID,
			"timestamp": time.Now().Format(time.RFC3339),
		})

		s.broadcastJobEvent(job.SessionID, sse.EventType("job.succeeded"), map[string]interface{}{
			"job_uuid":   job.JobUUID,
			"status":     job.Status,
			"progress":   job.Progress,
			"plugin_id":  job.PluginID,
			"session_id": job.SessionID,
		})

		cancel()
		s.cleanupJobMemory(jobUUID)
	}
}

func (s *jobService) cleanupJobMemory(jobUUID string) {
	s.mu.Lock()
	delete(s.authByJob, jobUUID)
	delete(s.cancelBy, jobUUID)
	s.mu.Unlock()
}

func (s *jobService) broadcastJobEvent(sessionID uint, eventType sse.EventType, data interface{}) {
	hub := sse.GetHub()
	hub.BroadcastToSession(fmt.Sprintf("%d", sessionID), sse.Event{Type: eventType, Data: data, Timestamp: time.Now()})
}
