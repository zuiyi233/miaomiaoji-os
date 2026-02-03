package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"rulebacktest/internal/service"
	"rulebacktest/pkg/errors"
	"rulebacktest/pkg/logger"
	"rulebacktest/pkg/response"
)

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		if query != "" {
			path = path + "?" + query
		}

		logger.Info("HTTP请求",
			logger.String("method", method),
			logger.String("path", path),
			logger.String("ip", clientIP),
			logger.Int("status", statusCode),
			logger.Float64("latency_ms", float64(latency.Milliseconds())),
		)
	}
}

// Recovery 错误恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("服务器内部错误",
					logger.Field("error", err),
					logger.String("path", c.Request.URL.Path),
					logger.String("method", c.Request.Method),
				)
				response.InternalServerError(c, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}

// CORS 跨域处理中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Auth JWT认证中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			response.Unauthorized(c, "未提供认证Token")
			c.Abort()
			return
		}

		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		userService := service.GetUserService()
		claims, err := userService.ParseToken(token)
		if err != nil {
			appErr := errors.GetAppError(err)
			if appErr != nil {
				response.Fail(c, appErr.Code, appErr.Message)
			} else {
				response.Unauthorized(c, "Token无效或已过期")
			}
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// RequireRole 角色权限中间件，验证用户是否具有指定角色
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			response.Forbidden(c, "无法获取用户角色")
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			response.Forbidden(c, "用户角色格式错误")
			c.Abort()
			return
		}

		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "无权限访问该资源")
		c.Abort()
	}
}

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return uuid.New().String()
}

// rateLimiter 内存限流器
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// newRateLimiter 创建限流器
func newRateLimiter(limit int, windowSeconds int) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   time.Duration(windowSeconds) * time.Second,
	}
}

// allow 检查是否允许请求
func (r *rateLimiter) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-r.window)

	// 清理过期记录
	if times, exists := r.requests[key]; exists {
		var valid []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		r.requests[key] = valid
	}

	// 检查是否超出限制
	if len(r.requests[key]) >= r.limit {
		return false
	}

	// 记录本次请求
	r.requests[key] = append(r.requests[key], now)
	return true
}

// RateLimit 请求限流中间件
// limit: 时间窗口内允许的最大请求数
// windowSeconds: 时间窗口(秒)
func RateLimit(limit int, windowSeconds int) gin.HandlerFunc {
	limiter := newRateLimiter(limit, windowSeconds)

	return func(c *gin.Context) {
		key := c.ClientIP()

		if !limiter.allow(key) {
			logger.Warn("请求被限流",
				logger.String("ip", key),
				logger.Int("limit", limit),
				logger.Int("window", windowSeconds),
			)
			c.Header("Retry-After", "1")
			response.Fail(c, errors.CodeRateLimited, "请求过于频繁，请稍后重试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// Timeout 请求超时中间件
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			logger.Warn("请求超时",
				logger.String("path", c.Request.URL.Path),
				logger.String("method", c.Request.Method),
			)
			response.Fail(c, errors.CodeTimeout, "请求超时")
			c.Abort()
		}
	}
}

// ErrorHandler 全局错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			appErr := errors.GetAppError(err)
			if appErr != nil {
				response.Fail(c, appErr.Code, appErr.Message)
				return
			}

			logger.Error("未处理的错误", logger.Err(err))
			response.InternalServerError(c, "服务器内部错误")
		}
	}
}
