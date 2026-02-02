// Viper配置管理
package config

import (
	"fmt"
	"strings"

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

	cfg = &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		logger.Error("解析配置文件失败", logger.Err(err))
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.App.Env == "" {
		cfg.App.Env = "development"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Database.Type == "" {
		cfg.Database.Type = "sqlite"
	}
	if cfg.Database.SQLitePath == "" {
		cfg.Database.SQLitePath = "./data.db"
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "your-secret-key-change-in-production"
	}
	if cfg.JWT.ExpireHour == 0 {
		cfg.JWT.ExpireHour = 24
	}
	if cfg.RateLimit.RequestsPerSecond == 0 {
		cfg.RateLimit.RequestsPerSecond = 100
	}
	if cfg.RateLimit.Burst == 0 {
		cfg.RateLimit.Burst = 200
	}

	return nil
}

func Get() *Config {
	if cfg == nil {
		panic("config not initialized, call config.Init() first")
	}
	return cfg
}

func GetDBConfig() database.Config {
	return Get().Database
}

func GetJWTConfig() JWTConfig {
	return Get().JWT
}
