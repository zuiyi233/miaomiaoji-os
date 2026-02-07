// Viper配置管理
package config

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"novel-agent-os-backend/pkg/database"
	"novel-agent-os-backend/pkg/logger"

	"github.com/spf13/viper"
)

var cfg *Config

type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Server    ServerConfig    `mapstructure:"server"`
	Database  database.Config `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Logging   logger.Config   `mapstructure:"logging"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	AI        AIConfig        `mapstructure:"ai"`
}

type AppConfig struct {
	Env   string `mapstructure:"env"`
	Name  string `mapstructure:"name"`
	Debug bool   `mapstructure:"debug"`
}

type ServerConfig struct {
	Port           int `mapstructure:"port"`
	ReadTimeout    int `mapstructure:"read_timeout"`
	WriteTimeout   int `mapstructure:"write_timeout"`
	MaxHeaderBytes int `mapstructure:"max_header_bytes"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireHour int    `mapstructure:"expire_hour"`
	Issuer     string `mapstructure:"issuer"`
}

type RateLimitConfig struct {
	RequestsPerSecond int `mapstructure:"requests_per_second"`
	Burst             int `mapstructure:"burst"`
}

type AIConfig struct {
	DefaultProvider      string `mapstructure:"default_provider"`
	ProvidersPath        string `mapstructure:"providers_path"`
	AllowInsecureHTTP    bool   `mapstructure:"allow_insecure_http"`
	ModelsCacheTTL       int    `mapstructure:"models_cache_ttl"`
	UseStaleCacheOnError bool   `mapstructure:"use_stale_cache_on_error"`
}

var cfgMu sync.RWMutex

func Init(configPath string, configName string) error {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName(configName)
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.novel-agent-os")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logger.Error("读取配置文件失败", logger.Err(err))
			return fmt.Errorf("failed to read config file: %w", err)
		}
		logger.Warn("配置文件未找到，使用默认配置")
	}

	v.SetEnvPrefix("NOVEL_AGENT_OS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	loaded := &Config{}
	if err := v.Unmarshal(loaded); err != nil {
		logger.Error("解析配置文件失败", logger.Err(err))
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if loaded.App.Env == "" {
		loaded.App.Env = "development"
	}
	if loaded.Server.Port == 0 {
		loaded.Server.Port = 8080
	}
	if loaded.Database.Type == "" {
		loaded.Database.Type = "sqlite"
	}
	if loaded.Database.SQLitePath == "" {
		loaded.Database.SQLitePath = "./data.db"
	}
	if loaded.JWT.Secret == "" {
		loaded.JWT.Secret = "your-secret-key-change-in-production"
	}
	if loaded.JWT.ExpireHour == 0 {
		loaded.JWT.ExpireHour = 24
	}
	if loaded.RateLimit.RequestsPerSecond == 0 {
		loaded.RateLimit.RequestsPerSecond = 100
	}
	if loaded.RateLimit.Burst == 0 {
		loaded.RateLimit.Burst = 200
	}
	if loaded.AI.ProvidersPath == "" {
		loaded.AI.ProvidersPath = "./configs/providers"
	}
	if loaded.AI.ModelsCacheTTL == 0 {
		loaded.AI.ModelsCacheTTL = 3600
	}

	cfgMu.Lock()
	cfg = loaded
	cfgMu.Unlock()

	return nil
}

func Get() *Config {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if cfg == nil {
		panic("config not initialized, call config.Init() first")
	}
	return cfg
}

// Reload 重新加载配置
func Reload() error {
	return Init("", "config")
}

// NowUTCString 返回UTC时间字符串
func NowUTCString() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func GetDBConfig() database.Config {
	return Get().Database
}

func GetJWTConfig() JWTConfig {
	return Get().JWT
}
