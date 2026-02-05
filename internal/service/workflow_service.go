package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/pkg/sse"

	"gorm.io/datatypes"
)

// RunWorkflowRequest 工作流执行请求
type RunWorkflowRequest struct {
	UserID       uint
	ProjectID    uint
	Session      *model.Session
	SessionTitle string
	Mode         string
	StepTitle    string
	FormatType   string
	Provider     string
	Path         string
	Body         string
}

// RunWorkflowResult 工作流执行结果
type RunWorkflowResult struct {
	Session *model.Session
	Step    *model.SessionStep
	Content string
	Raw     json.RawMessage
}

type WorkflowService interface {
	RunStep(req RunWorkflowRequest) (*RunWorkflowResult, error)
	RunChapterGenerate(req ChapterGenerateRequest) (*ChapterGenerateResult, error)
	RunChapterAnalyze(req ChapterAnalyzeRequest) (*ChapterAnalyzeResult, error)
	RunChapterRewrite(req ChapterRewriteRequest) (*ChapterRewriteResult, error)
	RunChapterBatch(req ChapterBatchRequest) (*ChapterBatchResult, error)
}

type workflowService struct {
	aiConfigService AIConfigService
	sessionService  SessionService
	documentService DocumentService
}

func NewWorkflowService(aiConfigService AIConfigService, sessionService SessionService, documentService DocumentService) WorkflowService {
	return &workflowService{
		aiConfigService: aiConfigService,
		sessionService:  sessionService,
		documentService: documentService,
	}
}

func (s *workflowService) RunStep(req RunWorkflowRequest) (*RunWorkflowResult, error) {
	session, err := s.ensureSession(req.Session, req.SessionTitle, req.Mode, req.ProjectID, req.UserID)
	if err != nil {
		return nil, err
	}

	raw, content, err := callAI(s.aiConfigService, req.Provider, req.Path, req.Body)
	if err != nil {
		return nil, err
	}

	if content == "" {
		content = string(raw)
	}

	step, err := s.appendStep(session.ID, req.StepTitle, content, req.FormatType, nil)
	if err != nil {
		return nil, err
	}

	return &RunWorkflowResult{
		Session: session,
		Step:    step,
		Content: content,
		Raw:     raw,
	}, nil
}

// ChapterWriteBack 章节写回配置
type ChapterWriteBack struct {
	Mode       string `json:"mode"`
	SetStatus  string `json:"set_status"`
	SetSummary bool   `json:"set_summary"`
}

// ChapterGenerateRequest 章节生成请求
type ChapterGenerateRequest struct {
	UserID       uint
	ProjectID    uint
	Session      *model.Session
	SessionTitle string
	DocumentID   uint
	VolumeID     uint
	Title        string
	OrderIndex   int
	Provider     string
	Path         string
	Body         string
	WriteBack    ChapterWriteBack
}

// ChapterGenerateResult 章节生成结果
type ChapterGenerateResult struct {
	Session  *model.Session
	Document *model.Document
	Steps    []*model.SessionStep
	Content  string
	Raw      json.RawMessage
}

// ChapterAnalyzeRequest 章节分析请求
type ChapterAnalyzeRequest struct {
	UserID       uint
	ProjectID    uint
	Session      *model.Session
	SessionTitle string
	DocumentID   uint
	Provider     string
	Path         string
	Body         string
	WriteBack    ChapterWriteBack
}

// ChapterAnalyzeResult 章节分析结果
type ChapterAnalyzeResult struct {
	Session  *model.Session
	Document *model.Document
	Content  string
	Raw      json.RawMessage
}

// ChapterRewriteRequest 章节重写请求
type ChapterRewriteRequest struct {
	UserID       uint
	ProjectID    uint
	Session      *model.Session
	SessionTitle string
	DocumentID   uint
	RewriteMode  string
	Provider     string
	Path         string
	Body         string
	WriteBack    ChapterWriteBack
}

// ChapterRewriteResult 章节重写结果
type ChapterRewriteResult struct {
	Session  *model.Session
	Document *model.Document
	Content  string
	Raw      json.RawMessage
}

// ChapterBatchItem 批量章节条目
type ChapterBatchItem struct {
	Title      string `json:"title"`
	OrderIndex int    `json:"order_index"`
	Outline    string `json:"outline"`
}

// ChapterBatchRequest 批量章节请求
type ChapterBatchRequest struct {
	UserID       uint
	ProjectID    uint
	Session      *model.Session
	SessionTitle string
	VolumeID     uint
	Items        []ChapterBatchItem
	Provider     string
	Path         string
	BodyTemplate string
	WriteBack    ChapterWriteBack
}

// ChapterBatchResult 批量章节结果
type ChapterBatchResult struct {
	Session   *model.Session
	Documents []*model.Document
}

// RunChapterGenerate 生成章节并写回文档
func (s *workflowService) RunChapterGenerate(req ChapterGenerateRequest) (*ChapterGenerateResult, error) {
	session, err := s.ensureSession(req.Session, req.SessionTitle, "chapter_generate", req.ProjectID, req.UserID)
	if err != nil {
		return nil, err
	}

	s.broadcastProgress(session.ID, 0, "生成开始")
	raw, content, err := callAI(s.aiConfigService, req.Provider, req.Path, req.Body)
	if err != nil {
		return nil, err
	}
	if content == "" {
		content = string(raw)
	}

	metadata := map[string]interface{}{
		"project_id":  req.ProjectID,
		"document_id": req.DocumentID,
		"volume_id":   req.VolumeID,
		"provider":    req.Provider,
		"path":        req.Path,
	}

	promptStep, err := s.appendStep(session.ID, "生成请求", req.Body, "chapter.generate.prompt", metadata)
	if err != nil {
		return nil, err
	}

	doc, err := s.writeBackGenerate(req, content)
	if err != nil {
		return nil, err
	}
	metadata["document_id"] = doc.ID

	resultStep, err := s.appendStep(session.ID, "生成结果", content, "chapter.generate.result", metadata)
	if err != nil {
		return nil, err
	}

	s.broadcastProgress(session.ID, 100, "生成完成")
	s.broadcastDone(session.ID, "chapter_generate", doc.ID)

	return &ChapterGenerateResult{
		Session:  session,
		Document: doc,
		Steps:    []*model.SessionStep{promptStep, resultStep},
		Content:  content,
		Raw:      raw,
	}, nil
}

// RunChapterAnalyze 分析章节并写回摘要
func (s *workflowService) RunChapterAnalyze(req ChapterAnalyzeRequest) (*ChapterAnalyzeResult, error) {
	session, err := s.ensureSession(req.Session, req.SessionTitle, "chapter_analyze", req.ProjectID, req.UserID)
	if err != nil {
		return nil, err
	}

	s.broadcastProgress(session.ID, 0, "分析开始")
	raw, content, err := callAI(s.aiConfigService, req.Provider, req.Path, req.Body)
	if err != nil {
		return nil, err
	}
	if content == "" {
		content = string(raw)
	}

	metadata := map[string]interface{}{
		"project_id":  req.ProjectID,
		"document_id": req.DocumentID,
		"provider":    req.Provider,
		"path":        req.Path,
	}
	_, err = s.appendStep(session.ID, "分析结果", content, "chapter.analyze.result", metadata)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.WriteBack.SetSummary {
		updates["summary"] = content
	}
	if req.WriteBack.SetStatus != "" {
		updates["status"] = req.WriteBack.SetStatus
	}
	if len(updates) > 0 {
		if _, err := s.documentService.Update(req.DocumentID, updates); err != nil {
			return nil, err
		}
	}

	doc, err := s.documentService.GetByID(req.DocumentID)
	if err != nil {
		return nil, err
	}

	s.broadcastProgress(session.ID, 100, "分析完成")
	s.broadcastDone(session.ID, "chapter_analyze", doc.ID)

	return &ChapterAnalyzeResult{
		Session:  session,
		Document: doc,
		Content:  content,
		Raw:      raw,
	}, nil
}

// RunChapterRewrite 重写章节并写回内容
func (s *workflowService) RunChapterRewrite(req ChapterRewriteRequest) (*ChapterRewriteResult, error) {
	session, err := s.ensureSession(req.Session, req.SessionTitle, "chapter_rewrite", req.ProjectID, req.UserID)
	if err != nil {
		return nil, err
	}

	doc, err := s.documentService.GetByID(req.DocumentID)
	if err != nil {
		return nil, err
	}

	s.broadcastProgress(session.ID, 0, "重写开始")
	raw, content, err := callAI(s.aiConfigService, req.Provider, req.Path, req.Body)
	if err != nil {
		return nil, err
	}
	if content == "" {
		content = string(raw)
	}

	metadata := map[string]interface{}{
		"project_id":   req.ProjectID,
		"document_id":  req.DocumentID,
		"provider":     req.Provider,
		"path":         req.Path,
		"rewrite_mode": req.RewriteMode,
		"prev_content": doc.Content,
	}
	_, err = s.appendStep(session.ID, "重写结果", content, "chapter.rewrite.result", metadata)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"content": content,
	}
	if req.WriteBack.SetStatus != "" {
		updates["status"] = req.WriteBack.SetStatus
	}
	if _, err := s.documentService.Update(req.DocumentID, updates); err != nil {
		return nil, err
	}

	updated, err := s.documentService.GetByID(req.DocumentID)
	if err != nil {
		return nil, err
	}

	s.broadcastProgress(session.ID, 100, "重写完成")
	s.broadcastDone(session.ID, "chapter_rewrite", updated.ID)

	return &ChapterRewriteResult{
		Session:  session,
		Document: updated,
		Content:  content,
		Raw:      raw,
	}, nil
}

// RunChapterBatch 批量生成章节
func (s *workflowService) RunChapterBatch(req ChapterBatchRequest) (*ChapterBatchResult, error) {
	session, err := s.ensureSession(req.Session, req.SessionTitle, "chapter_batch", req.ProjectID, req.UserID)
	if err != nil {
		return nil, err
	}

	documents := make([]*model.Document, 0, len(req.Items))
	for index, item := range req.Items {
		progress := int(float64(index) / float64(len(req.Items)) * 100)
		s.broadcastProgress(session.ID, progress, "批量生成中")

		body := buildBatchBody(req.BodyTemplate, item)
		metadata := map[string]interface{}{
			"project_id": req.ProjectID,
			"volume_id":  req.VolumeID,
			"provider":   req.Provider,
			"path":       req.Path,
			"title":      item.Title,
		}
		_, err := s.appendStep(session.ID, "批量生成开始", item.Outline, "chapter.batch.item.started", metadata)
		if err != nil {
			return nil, err
		}

		raw, content, err := callAI(s.aiConfigService, req.Provider, req.Path, body)
		if err != nil {
			return nil, err
		}
		if content == "" {
			content = string(raw)
		}

		orderIndex := item.OrderIndex
		if orderIndex <= 0 {
			orderIndex, err = s.documentService.GetNextOrderIndex(req.ProjectID, req.VolumeID)
			if err != nil {
				return nil, err
			}
		}

		doc, err := s.documentService.Create(req.ProjectID, item.Title, content, "", req.WriteBack.SetStatus, orderIndex, "", "", 0, "", "", "", "", "", req.VolumeID)
		if err != nil {
			return nil, err
		}

		if req.WriteBack.SetSummary {
			if _, err := s.documentService.Update(doc.ID, map[string]interface{}{"summary": content}); err != nil {
				return nil, err
			}
		}

		metadata["document_id"] = doc.ID
		_, err = s.appendStep(session.ID, "批量生成结果", content, "chapter.batch.item.result", metadata)
		if err != nil {
			return nil, err
		}

		documents = append(documents, doc)
	}

	s.broadcastProgress(session.ID, 100, "批量生成完成")
	s.broadcastDone(session.ID, "chapter_batch", 0)

	return &ChapterBatchResult{
		Session:   session,
		Documents: documents,
	}, nil
}


func (s *workflowService) broadcastStep(sessionID uint, step *model.SessionStep) {
	data := map[string]interface{}{
		"step_id":   step.ID,
		"title":     step.Title,
		"content":   step.Content,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	hub := sse.GetHub()
	hub.BroadcastToSession(fmt.Sprintf("%d", sessionID), sse.NewStepAppendedEvent(data))
}

func (s *workflowService) appendStep(sessionID uint, title, content, formatType string, metadata map[string]interface{}) (*model.SessionStep, error) {
	step := &model.SessionStep{
		Title:      title,
		Content:    content,
		FormatType: formatType,
		SessionID:  sessionID,
	}
	if metadata != nil {
		step.Metadata = encodeMetadata(metadata)
	}
	if err := s.sessionService.CreateStepAutoOrder(step); err != nil {
		return nil, err
	}
	s.broadcastStep(sessionID, step)
	return step, nil
}

func (s *workflowService) ensureSession(session *model.Session, title, mode string, projectID, userID uint) (*model.Session, error) {
	if session != nil {
		return session, nil
	}

	session = &model.Session{
		Title:     title,
		Mode:      mode,
		ProjectID: projectID,
		UserID:    userID,
	}
	if err := s.sessionService.CreateSession(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *workflowService) broadcastProgress(sessionID uint, progress int, message string) {
	data := map[string]interface{}{
		"progress":  progress,
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	hub := sse.GetHub()
	hub.BroadcastToSession(fmt.Sprintf("%d", sessionID), sse.NewProgressUpdatedEvent(data))
}

func (s *workflowService) broadcastDone(sessionID uint, mode string, documentID uint) {
	data := map[string]interface{}{
		"mode":       mode,
		"document_id": documentID,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	hub := sse.GetHub()
	hub.BroadcastToSession(fmt.Sprintf("%d", sessionID), sse.NewWorkflowDoneEvent(data))
}

func (s *workflowService) writeBackGenerate(req ChapterGenerateRequest, content string) (*model.Document, error) {
	if req.DocumentID > 0 {
		updates := map[string]interface{}{
			"content": content,
		}
		if req.WriteBack.SetStatus != "" {
			updates["status"] = req.WriteBack.SetStatus
		}
		if req.WriteBack.SetSummary {
			updates["summary"] = content
		}
		if _, err := s.documentService.Update(req.DocumentID, updates); err != nil {
			return nil, err
		}
		return s.documentService.GetByID(req.DocumentID)
	}

	orderIndex := req.OrderIndex
	if orderIndex <= 0 {
		var err error
		orderIndex, err = s.documentService.GetNextOrderIndex(req.ProjectID, req.VolumeID)
		if err != nil {
			return nil, err
		}
	}

	doc, err := s.documentService.Create(req.ProjectID, req.Title, content, "", req.WriteBack.SetStatus, orderIndex, "", "", 0, "", "", "", "", "", req.VolumeID)
	if err != nil {
		return nil, err
	}
	if req.WriteBack.SetSummary {
		if _, err := s.documentService.Update(doc.ID, map[string]interface{}{"summary": content}); err != nil {
			return nil, err
		}
		return s.documentService.GetByID(doc.ID)
	}
	return doc, nil
}

func buildBatchBody(template string, item ChapterBatchItem) string {
	body := strings.ReplaceAll(template, "{{title}}", item.Title)
	body = strings.ReplaceAll(body, "{{outline}}", item.Outline)
	return body
}

func encodeMetadata(metadata map[string]interface{}) datatypes.JSON {
	if metadata == nil {
		return nil
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return nil
	}
	return datatypes.JSON(raw)
}
