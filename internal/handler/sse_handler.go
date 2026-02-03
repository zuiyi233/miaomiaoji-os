package handler

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"
	"novel-agent-os-backend/pkg/sse"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SSEHandler struct{}

func NewSSEHandler() *SSEHandler {
	return &SSEHandler{}
}

func (h *SSEHandler) Stream(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		response.Fail(c, errors.CodeInvalidParams, "session_id is required")
		return
	}

	clientID := uuid.New().String()
	hub := sse.GetHub()
	client := hub.AddClient(clientID, sessionID)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ctx := c.Request.Context()

	go func() {
		<-ctx.Done()
		client.Close()
	}()

	c.Stream(func(w io.Writer) bool {
		select {
		case event := <-client.Channel:
			_, err := fmt.Fprint(w, event.ToSSEFormat())
			if err != nil {
				logger.Error("Failed to write SSE event", logger.Err(err), logger.String("client_id", clientID))
				return false
			}
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			return true
		case <-ctx.Done():
			return false
		}
	})
}

func (h *SSEHandler) BroadcastTestEvent(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		response.Fail(c, errors.CodeInvalidParams, "session_id is required")
		return
	}

	eventType := c.Query("type")
	var event sse.Event

	switch eventType {
	case "step":
		event = sse.NewStepAppendedEvent(map[string]interface{}{
			"step_id":   1,
			"title":     "Test Step",
			"content":   "This is a test step content",
			"timestamp": time.Now(),
		})
	case "quality":
		event = sse.NewQualityCheckedEvent(map[string]interface{}{
			"step_id":   1,
			"passed":    true,
			"score":     95,
			"issues":    []string{},
			"timestamp": time.Now(),
		})
	case "export":
		event = sse.NewExportReadyEvent(map[string]interface{}{
			"export_id": "export-123",
			"format":    "pdf",
			"file_url":  "/exports/export-123.pdf",
			"timestamp": time.Now(),
		})
	default:
		event = sse.NewErrorEvent("Unknown event type", nil)
	}

	hub := sse.GetHub()
	hub.BroadcastToSession(sessionID, event)

	response.SuccessWithData(c, gin.H{
		"message":    "Event broadcasted",
		"session_id": sessionID,
		"event_type": event.Type,
	})
}
