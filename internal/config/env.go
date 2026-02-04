package config

import "os"

// GetEnvFirst 获取环境变量值
func GetEnvFirst(keys ...string) string {
	for _, key := range keys {
		value := os.Getenv(key)
		if value != "" {
			return value
		}
	}
	return ""
}
