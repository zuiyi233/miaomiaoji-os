// 程序入口
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"novel-agent-os-backend/internal/config"
	"novel-agent-os-backend/internal/middleware"
	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/internal/router"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/pkg/database"
	"novel-agent-os-backend/pkg/logger"
)

func main() {
	configPath := flag.String("config", "", "config file path")
	configName := flag.String("config-name", "config", "config name without extension")
	flag.Parse()

	if err := config.Init(*configPath, *configName); err != nil {
		fmt.Printf("Failed to init config: %v\n", err)
		os.Exit(1)
	}

	cfg := config.Get()

	logger.Init(logger.Config{
		Level:      cfg.Logging.Level,
		Filename:   cfg.Logging.Filename,
		MaxSize:    cfg.Logging.MaxSize,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAge,
		Compress:   cfg.Logging.Compress,
		Console:    cfg.Logging.Console,
	})

	if err := database.Init(database.Config{
		Type:            cfg.Database.Type,
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Database,
		SSLMode:         cfg.Database.SSLMode,
		SQLitePath:      cfg.Database.SQLitePath,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		LogLevel:        cfg.Database.LogLevel,
	}); err != nil {
		logger.Error("Failed to init database", logger.Err(err))
		os.Exit(1)
	}

	middleware.InitJWT(cfg.JWT.Secret, cfg.JWT.ExpireHour)
	middleware.InitRateLimiter("", cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst)

	if err := autoMigrate(); err != nil {
		logger.Error("Failed to auto migrate", logger.Err(err))
		os.Exit(1)
	}

	if err := ensureDefaultAdmin(); err != nil {
		logger.Error("Failed to ensure default admin", logger.Err(err))
		os.Exit(1)
	}

	r := router.Setup()

	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		logger.Info("Starting server", logger.String("addr", addr))
		if err := r.Run(addr); err != nil {
			logger.Error("Server failed", logger.Err(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	if err := database.Close(); err != nil {
		logger.Error("Database close failed", logger.Err(err))
	}
	logger.Info("Server stopped")
}

func autoMigrate() error {
	return database.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.Volume{},
		&model.Document{},
		&model.DocumentEntityRef{},
		&model.Entity{},
		&model.EntityTag{},
		&model.EntityLink{},
		&model.Template{},
		&model.Plugin{},
		&model.PluginCapability{},
		&model.Session{},
		&model.SessionStep{},
		&model.Job{},
		&model.SettlementEntry{},
		&model.CorpusStory{},
		&model.File{},
		&model.RedemptionCode{},
		&model.RedemptionCodeUse{},
		&model.AIProviderConfig{},
	)
}

func ensureDefaultAdmin() error {
	userRepo := repository.NewUserRepository()
	userService := service.NewUserService(userRepo)
	return userService.EnsureDefaultAdmin()
}
