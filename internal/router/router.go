// 路由配置
package router

import (
	"novel-agent-os-backend/internal/config"
	"novel-agent-os-backend/internal/handler"
	"novel-agent-os-backend/internal/middleware"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/internal/service"
	"novel-agent-os-backend/internal/storage"
	"novel-agent-os-backend/pkg/logger"
	"novel-agent-os-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// Setup 设置路由
func Setup() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit())

	// 健康检查
	r.GET("/healthz", HealthCheck)
	r.GET("/ready", ReadinessCheck)

	// 初始化依赖
	db := repository.GetDB()

	userRepo := repository.NewUserRepository()
	userService := service.NewUserService(userRepo)
	authHandler := handler.NewAuthHandler(userService)
	userHandler := handler.NewUserHandler(userService)

	projectRepo := repository.NewProjectRepository()
	volumeRepo := repository.NewVolumeRepository()
	documentRepo := repository.NewDocumentRepository()
	entityRepo := repository.NewEntityRepository()
	templateRepo := repository.NewTemplateRepository()
	projectService := service.NewProjectService(projectRepo, volumeRepo, documentRepo, entityRepo, templateRepo)

	volumeService := service.NewVolumeService(volumeRepo, projectRepo)
	volumeHandler := handler.NewVolumeHandler(volumeService)

	documentService := service.NewDocumentService(documentRepo, projectRepo, volumeRepo)
	documentHandler := handler.NewDocumentHandler(documentService)

	entityService := service.NewEntityService(entityRepo, projectRepo)
	entityHandler := handler.NewEntityHandler(entityService)

	templateService := service.NewTemplateService(templateRepo, projectRepo)
	templateHandler := handler.NewTemplateHandler(templateService)

	redemptionRepo := repository.NewRedemptionCodeRepository()
	redemptionService := service.NewRedemptionCodeService(redemptionRepo)
	redemptionHandler := handler.NewRedemptionCodeHandler(redemptionService, userService)

	aiConfigRepo := repository.NewAIConfigRepository()
	aiConfigService := service.NewAIConfigService(aiConfigRepo)
	aiModelService := service.NewAIModelService(aiConfigRepo)
	aiConfigHandler := handler.NewAIConfigHandler(aiConfigService)
	aiModelHandler := handler.NewAIModelHandler(aiModelService)
	aiProxyHandler := handler.NewAIProxyHandler(aiConfigService)
	aiProxyStreamHandler := handler.NewAIProxyStreamHandler(aiConfigService)

	pluginRepo := repository.NewPluginRepository(db)
	pluginService := service.NewPluginService(pluginRepo)

	// Session 依赖
	sessionRepo := repository.NewSessionRepository(db)
	sessionService := service.NewSessionService(sessionRepo)
	sessionHandler := handler.NewSessionHandler(sessionService)

	// Job 依赖
	jobRepo := repository.NewJobRepository(db)
	jobService := service.NewJobService(jobRepo, sessionRepo, pluginService, sessionService)
	jobHandler := handler.NewJobHandler(jobService)

	pluginHandler := handler.NewPluginHandler(pluginService, jobService)

	// Settlement 依赖
	settlementRepo := repository.NewSettlementRepository(db)
	settlementService := service.NewSettlementService(settlementRepo)
	settlementHandler := handler.NewSettlementHandler(settlementService)

	// File 依赖
	fileRepo := repository.NewFileRepository(db)
	localStorage, err := storage.NewLocalStorage("./uploads")
	if err != nil {
		logger.Error("初始化本地存储失败", logger.Err(err))
	}
	fileService := service.NewFileService(fileRepo, localStorage)
	fileHandler := handler.NewFileHandler(fileService)

	projectHandler := handler.NewProjectHandler(projectService, fileService)

	// Corpus 依赖
	corpusRepo := repository.NewCorpusRepository(db)
	corpusService := service.NewCorpusService(corpusRepo)
	corpusHandler := handler.NewCorpusHandler(corpusService)

	// Formatting & Quality 依赖
	formattingService := service.NewFormattingService(config.Config{})
	qualityGateService := service.NewQualityGateService(config.Config{})
	formattingHandler := handler.NewFormattingHandler(formattingService)
	qualityHandler := handler.NewQualityHandler(qualityGateService)

	// SSE 依赖
	sseHandler := handler.NewSSEHandler()

	// API v1
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", HealthCheck)

		// 认证路由（无需认证）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/refresh", middleware.JWTAuth(), authHandler.Refresh)
		}

		// 用户路由
		users := v1.Group("/users")
		{
			// 需要登录的路由
			users.GET("/profile", middleware.JWTAuth(), userHandler.GetProfile)
			users.PUT("/profile", middleware.JWTAuth(), userHandler.UpdateProfile)
			users.PUT("/password", middleware.JWTAuth(), userHandler.ChangePassword)
			users.POST("/check-in", middleware.JWTAuth(), userHandler.CheckIn)
			users.GET("/points", middleware.JWTAuth(), userHandler.GetPoints)
			users.POST("/heartbeat", middleware.JWTAuth(), userHandler.Heartbeat)

			// 管理员路由
			users.GET("/", middleware.JWTRequired("admin"), userHandler.ListUsers)
			users.PUT("/:id/status", middleware.JWTRequired("admin"), userHandler.UpdateUserStatus)
		}

		// 兑换码路由
		codes := v1.Group("/codes")
		{
			codes.POST("/redeem", middleware.JWTAuth(), redemptionHandler.Redeem)
			codes.GET("", middleware.JWTRequired("admin"), redemptionHandler.ListCodes)
			codes.POST("/generate", middleware.JWTRequired("admin"), redemptionHandler.GenerateCodes)
			codes.PUT("/batch", middleware.JWTRequired("admin"), redemptionHandler.BatchUpdateCodes)
			codes.GET("/export", middleware.JWTRequired("admin"), redemptionHandler.ExportCodes)
		}

		// AI 供应商与模型
		ai := v1.Group("/ai")
		{
			ai.GET("/models", middleware.JWTAuth(), aiModelHandler.ListModels)
			ai.PUT("/providers", middleware.JWTRequired("admin"), aiConfigHandler.UpdateProvider)
			ai.GET("/providers", middleware.JWTRequired("admin"), aiConfigHandler.GetProvider)
			ai.POST("/providers/test", middleware.JWTRequired("admin"), aiConfigHandler.TestProvider)
			ai.POST("/proxy", middleware.JWTAuth(), handler.RequireAIAccess(userService), aiProxyHandler.Proxy)
			ai.POST("/proxy/stream", middleware.JWTAuth(), handler.RequireAIAccess(userService), aiProxyStreamHandler.ProxyStream)
		}

		// 项目路由
		projects := v1.Group("/projects")
		{
			projects.GET("", middleware.JWTAuth(), projectHandler.List)
			projects.GET("/", middleware.JWTAuth(), projectHandler.List)
			projects.POST("/", middleware.JWTAuth(), projectHandler.Create)
			projects.POST("/snapshot", middleware.JWTAuth(), projectHandler.UpsertSnapshot)
			projects.POST("/backup", middleware.JWTAuth(), projectHandler.BackupSnapshot)
			projects.GET("/:project_id/backup/latest", middleware.JWTAuth(), projectHandler.GetLatestBackup)
			projects.GET("/:project_id", middleware.JWTAuth(), projectHandler.GetByID)
			projects.PUT("/:project_id", middleware.JWTAuth(), projectHandler.Update)
			projects.DELETE("/:project_id", middleware.JWTAuth(), projectHandler.Delete)
			projects.GET("/:project_id/export", middleware.JWTAuth(), projectHandler.Export)

			// 项目下的卷路由
			projects.GET("/:project_id/volumes", middleware.JWTAuth(), volumeHandler.List)
			projects.POST("/:project_id/volumes", middleware.JWTAuth(), volumeHandler.Create)
			projects.POST("/:project_id/volumes/reorder", middleware.JWTAuth(), volumeHandler.Reorder)

			// 项目下的文档路由
			projects.GET("/:project_id/documents", middleware.JWTAuth(), documentHandler.ListByProject)
			projects.POST("/:project_id/documents", middleware.JWTAuth(), documentHandler.Create)

			// 项目下的实体路由
			projects.GET("/:project_id/entities", middleware.JWTAuth(), entityHandler.List)
			projects.POST("/:project_id/entities", middleware.JWTAuth(), entityHandler.Create)

			// 项目下的模板路由
			projects.GET("/:project_id/templates", middleware.JWTAuth(), templateHandler.ListByProject)
			projects.POST("/:project_id/templates", middleware.JWTAuth(), templateHandler.Create)
		}

		// 卷路由
		volumes := v1.Group("/volumes")
		{
			volumes.GET("/:volume_id", middleware.JWTAuth(), volumeHandler.GetByID)
			volumes.PUT("/:volume_id", middleware.JWTAuth(), volumeHandler.Update)
			volumes.DELETE("/:volume_id", middleware.JWTAuth(), volumeHandler.Delete)

			// 卷下的文档路由
			volumes.GET("/:volume_id/documents", middleware.JWTAuth(), documentHandler.ListByVolume)
		}

		// 文档路由
		documents := v1.Group("/documents")
		{
			documents.GET("/:id", middleware.JWTAuth(), documentHandler.GetByID)
			documents.PUT("/:id", middleware.JWTAuth(), documentHandler.Update)
			documents.DELETE("/:id", middleware.JWTAuth(), documentHandler.Delete)
			documents.POST("/:id/bookmarks", middleware.JWTAuth(), documentHandler.AddBookmark)
			documents.DELETE("/:id/bookmarks/:index", middleware.JWTAuth(), documentHandler.RemoveBookmark)
			documents.POST("/:id/entities", middleware.JWTAuth(), documentHandler.LinkEntity)
			documents.DELETE("/:id/entities/:entity_id", middleware.JWTAuth(), documentHandler.UnlinkEntity)
		}

		// 实体路由
		entities := v1.Group("/entities")
		{
			entities.GET("/:id", middleware.JWTAuth(), entityHandler.GetByID)
			entities.PUT("/:id", middleware.JWTAuth(), entityHandler.Update)
			entities.DELETE("/:id", middleware.JWTAuth(), entityHandler.Delete)
			entities.POST("/:id/tags", middleware.JWTAuth(), entityHandler.AddTag)
			entities.DELETE("/:id/tags/:tag", middleware.JWTAuth(), entityHandler.RemoveTag)
			entities.POST("/:id/links", middleware.JWTAuth(), entityHandler.CreateLink)
			entities.DELETE("/:id/links/:target_id", middleware.JWTAuth(), entityHandler.DeleteLink)
		}

		// 模板路由
		templates := v1.Group("/templates")
		{
			templates.GET("/system", middleware.JWTAuth(), templateHandler.ListSystem)
			templates.GET("/:id", middleware.JWTAuth(), templateHandler.GetByID)
			templates.PUT("/:id", middleware.JWTAuth(), templateHandler.Update)
			templates.DELETE("/:id", middleware.JWTAuth(), templateHandler.Delete)
		}

		// 插件路由
		plugins := v1.Group("/plugins")
		{
			plugins.POST("", middleware.JWTAuth(), pluginHandler.CreatePlugin)
			plugins.GET("", middleware.JWTAuth(), pluginHandler.ListPlugins)
			plugins.GET("/:plugin_id", middleware.JWTAuth(), pluginHandler.GetPlugin)
			plugins.PUT("/:plugin_id", middleware.JWTAuth(), pluginHandler.UpdatePlugin)
			plugins.DELETE("/:plugin_id", middleware.JWTAuth(), pluginHandler.DeletePlugin)
			plugins.PUT("/:plugin_id/enable", middleware.JWTAuth(), pluginHandler.EnablePlugin)
			plugins.PUT("/:plugin_id/disable", middleware.JWTAuth(), pluginHandler.DisablePlugin)
			plugins.POST("/:plugin_id/ping", middleware.JWTAuth(), pluginHandler.PingPlugin)

			plugins.GET("/:plugin_id/capabilities", middleware.JWTAuth(), pluginHandler.GetCapabilities)
			plugins.POST("/:plugin_id/capabilities", middleware.JWTAuth(), pluginHandler.AddCapability)
			plugins.DELETE("/capabilities/:id", middleware.JWTAuth(), pluginHandler.RemoveCapability)

			plugins.POST("/:plugin_id/invoke", middleware.JWTAuth(), pluginHandler.InvokePlugin)
			plugins.POST("/:plugin_id/invoke-async", middleware.JWTAuth(), pluginHandler.InvokePluginAsync)
		}

		// 任务路由
		jobs := v1.Group("/jobs")
		{
			jobs.GET("/:job_uuid", middleware.JWTAuth(), jobHandler.GetJob)
			jobs.POST("/:job_uuid/cancel", middleware.JWTAuth(), jobHandler.CancelJob)
		}

		// 会话路由
		sessions := v1.Group("/sessions")
		{
			sessions.POST("", middleware.JWTAuth(), sessionHandler.CreateSession)
			sessions.GET("", middleware.JWTAuth(), sessionHandler.ListSessions)
			sessions.GET("/projects/:project_id", middleware.JWTAuth(), sessionHandler.ListSessionsByProject)
			sessions.GET("/:session_id", middleware.JWTAuth(), sessionHandler.GetSession)
			sessions.PUT("/:session_id", middleware.JWTAuth(), sessionHandler.UpdateSession)
			sessions.DELETE("/:session_id", middleware.JWTAuth(), sessionHandler.DeleteSession)

			// SessionStep 路由
			sessions.POST("/:session_id/steps", middleware.JWTAuth(), sessionHandler.CreateStep)
			sessions.GET("/:session_id/steps", middleware.JWTAuth(), sessionHandler.ListSteps)
			sessions.GET("/steps/:id", middleware.JWTAuth(), sessionHandler.GetStep)
			sessions.PUT("/steps/:id", middleware.JWTAuth(), sessionHandler.UpdateStep)
			sessions.DELETE("/steps/:id", middleware.JWTAuth(), sessionHandler.DeleteStep)
		}

		// SSE 路由
		sse := v1.Group("/sse")
		{
			sse.GET("/stream", middleware.JWTAuth(), sseHandler.Stream)
			sse.POST("/test", middleware.JWTAuth(), sseHandler.BroadcastTestEvent)
		}

		// 结算路由
		settlements := v1.Group("/settlements")
		{
			settlements.POST("", middleware.JWTAuth(), settlementHandler.CreateEntry)
			settlements.GET("", middleware.JWTAuth(), settlementHandler.ListEntries)
			settlements.GET("/filter", middleware.JWTAuth(), settlementHandler.FilterEntries)
			settlements.GET("/total-points", middleware.JWTAuth(), settlementHandler.GetTotalPoints)
			settlements.GET("/:id", middleware.JWTAuth(), settlementHandler.GetEntry)
			settlements.PUT("/:id", middleware.JWTAuth(), settlementHandler.UpdateEntry)
			settlements.DELETE("/:id", middleware.JWTAuth(), settlementHandler.DeleteEntry)
		}

		// 语料库路由
		corpus := v1.Group("/corpus")
		{
			corpus.POST("", middleware.JWTAuth(), corpusHandler.CreateStory)
			corpus.GET("", middleware.JWTAuth(), corpusHandler.ListStories)
			corpus.GET("/search", middleware.JWTAuth(), corpusHandler.SearchStories)
			corpus.GET("/genre", middleware.JWTAuth(), corpusHandler.ListStoriesByGenre)
			corpus.GET("/:id", middleware.JWTAuth(), corpusHandler.GetStory)
			corpus.PUT("/:id", middleware.JWTAuth(), corpusHandler.UpdateStory)
			corpus.DELETE("/:id", middleware.JWTAuth(), corpusHandler.DeleteStory)
		}

		// 文件路由
		files := v1.Group("/files")
		{
			files.POST("", middleware.JWTAuth(), fileHandler.CreateFile)
			files.GET("", middleware.JWTAuth(), fileHandler.ListFiles)
			files.GET("/project/:project_id", middleware.JWTAuth(), fileHandler.ListFilesByProject)
			files.GET("/:id", middleware.JWTAuth(), fileHandler.GetFile)
			files.GET("/:id/download", middleware.JWTAuth(), fileHandler.DownloadFile)
			files.PUT("/:id", middleware.JWTAuth(), fileHandler.UpdateFile)
			files.DELETE("/:id", middleware.JWTAuth(), fileHandler.DeleteFile)
		}

		// 排版路由
		formatting := v1.Group("/formatting")
		{
			formatting.POST("/format", middleware.JWTAuth(), formattingHandler.FormatText)
			formatting.GET("/styles", middleware.JWTAuth(), formattingHandler.GetAvailableStyles)
		}

		// 质量门禁路由
		quality := v1.Group("/quality")
		{
			quality.POST("/check", middleware.JWTAuth(), qualityHandler.CheckQuality)
			quality.GET("/thresholds", middleware.JWTAuth(), qualityHandler.GetThresholds)
		}
	}

	return r
}

// GetEngine 获取引擎
func GetEngine() *gin.Engine {
	return Setup()
}

// RegisterRoutes 注册路由
func RegisterRoutes(r *gin.Engine) {
	// Additional routes can be registered here
}

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	response.SuccessWithData(c, gin.H{
		"status":  "ok",
		"message": "Service is healthy",
	})
}

// ReadinessCheck 就绪检查
func ReadinessCheck(c *gin.Context) {
	response.SuccessWithData(c, gin.H{
		"status":   "ready",
		"database": "connected",
	})
}
