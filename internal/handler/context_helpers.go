package handler

import "github.com/gin-gonic/gin"

// getUserIDFromContext 从上下文获取用户ID
func getUserIDFromContext(c *gin.Context) uint {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	uid, ok := userID.(uint)
	if !ok {
		return 0
	}
	return uid
}
