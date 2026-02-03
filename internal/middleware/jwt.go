package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"novel-agent-os-backend/internal/config"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"
)

var (
	jwtSecret   string
	jwtExpire   time.Duration
	rateLimiter *RateLimiter
)

// JWTClaims JWT claims
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// 提取token
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Error(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Get().JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			response.Error(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

// JWTRequired 需要特定角色的JWT认证
func JWTRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先进行JWT认证
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Error(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Get().JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			response.Error(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// 检查角色权限
		if len(roles) > 0 {
			hasRole := false
			for _, role := range roles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}
			if !hasRole {
				response.Error(c, errors.ErrForbidden)
				c.Abort()
				return
			}
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

// GenerateToken 生成JWT token
func GenerateToken(userID uint, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Get().JWT.Secret))
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := c.GetTime("request_start")
		if start.IsZero() {
			start = c.GetTime("request_start")
		}

		c.Next()

		logger.Info("HTTP请求",
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.Int("status", c.Writer.Status()),
			logger.String("ip", c.ClientIP()),
		)
	}
}

// Recovery 错误恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("服务器内部错误",
					logger.Any("error", err),
					logger.String("path", c.Request.URL.Path),
				)
				response.Error(c, errors.ErrInternalServer)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// CORS 跨域处理中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
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
	return "req_" + randomString(16)
}

// randomString 生成随机字符串
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

// RateLimit 请求限流中间件（简化版）
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 简化实现，实际应该使用Redis等存储
		c.Next()
	}
}

// InitJWT 初始化JWT配置
func InitJWT(secret string, expireHour int) {
	jwtSecret = secret
	if expireHour <= 0 {
		expireHour = 24
	}
	jwtExpire = time.Duration(expireHour) * time.Hour
}

// InitRateLimiter 初始化限流器
func InitRateLimiter(redisAddr string, rps, burst int) {
	// 简化实现，实际应该使用Redis
	rateLimiter = &RateLimiter{
		rps:   rps,
		burst: burst,
	}
}

// RateLimiter 限流器
type RateLimiter struct {
	rps   int
	burst int
}
