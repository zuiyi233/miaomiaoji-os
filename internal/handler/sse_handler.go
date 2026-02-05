package handler

import (
	"fmt"
	"net/http"

	"novel-agent-os-backend/pkg/response"
	"novel-agent-os-backend/pkg/sse"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SSEHandler SSE 流处理器
type SSEHandler struct{}

func NewSSEHandler() *SSEHandler {
	return &SSEHandler{}
}

// Stream 建立 SSE 流连接
func (h *SSEHandler) Stream(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	clientID := uuid.NewString()
	hub := sse.GetHub()
	client := hub.AddClient(clientID, sessionID)
	defer hub.RemoveClient(clientID)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	writer := c.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	// 发送一条注释行，帮助客户端尽快进入 connected
	_, _ = fmt.Fprint(writer, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case evt, ok := <-client.Channel:
			if !ok {
				return
			}
			_, _ = fmt.Fprint(writer, evt.ToSSEFormat())
			flusher.Flush()
		}
	}
}

// BroadcastTestEvent 测试广播事件
func (h *SSEHandler) BroadcastTestEvent(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		response.SuccessWithData(c, gin.H{"ok": true})
		return
	}

	hub := sse.GetHub()
	hub.BroadcastToSession(sessionID, sse.NewStepAppendedEvent(gin.H{
		"step_id":   0,
		"title":     "test",
		"content":   "test event",
		"timestamp": "",
	}))
	response.SuccessWithData(c, gin.H{"ok": true})
}
