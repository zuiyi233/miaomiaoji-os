package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/errors"
	"novel-agent-os-backend/pkg/response"
)

// RequireAIAccess 校验AI权限
func RequireAIAccess(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserIDFromContext(c)
		if userID == 0 {
			response.Fail(c, errors.CodeUnauthorized, "未登录")
			c.Abort()
			return
		}

		user, err := userService.GetUserByID(userID)
		if err != nil {
			response.Fail(c, errors.CodeNotFound, "用户不存在")
			c.Abort()
			return
		}

		if user.Role != "admin" && (user.AIAccessUntil == nil || user.AIAccessUntil.Before(time.Now())) {
			response.Fail(c, errors.CodeForbidden, "AI权限已过期")
			c.Abort()
			return
		}

		c.Next()
	}
}
