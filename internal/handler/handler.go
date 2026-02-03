// Package handler HTTP处理器层
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// parseUintParam 解析URL参数为uint
func parseUintParam(c *gin.Context, param string) (uint, error) {
	idStr := c.Param(param)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// parseUintQuery 解析查询参数为uint
func parseUintQuery(c *gin.Context, query string, defaultValue uint) uint {
	valueStr := c.DefaultQuery(query, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}
	return uint(value)
}

// parseIntQuery 解析查询参数为int
func parseIntQuery(c *gin.Context, query string, defaultValue int) int {
	valueStr := c.DefaultQuery(query, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
