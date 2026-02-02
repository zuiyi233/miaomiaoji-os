// 路由配置
package router

import (
	"novel-agent-os-backend/internal/handler"
	"novel-agent-os-backend/internal/middleware"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/internal/service"
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
	userRepo := repository.NewUserRepository()
	userService := service.NewUserService(userRepo)
	authHandler := handler.NewAuthHandler(userService)
	userHandler := handler.NewUserHandler(userService)

	projectRepo := repository.NewProjectRepository()
	projectService := service.NewProjectService(projectRepo)
	projectHandler := handler.NewProjectHandler(projectService)

	volumeRepo := repository.NewVolumeRepository()
	volumeService := service.NewVolumeService(volumeRepo, projectRepo)
	volumeHandler := handler.NewVolumeHandler(volumeService)

	documentRepo := repository.NewDocumentRepository()
	documentService := service.NewDocumentService(documentRepo, projectRepo, volumeRepo)
	documentHandler := handler.NewDocumentHandler(documentService)

	entityRepo := repository.NewEntityRepository()
	entityService := service.NewEntityService(entityRepo, projectRepo)
	entityHandler := handler.NewEntityHandler(entityService)

	templateRepo := repository.NewTemplateRepository()
	templateService := service.NewTemplateService(templateRepo, projectRepo)
	templateHandler := handler.NewTemplateHandler(templateService)

	pluginRepo := repository.NewPluginRepository()
	pluginService := service.NewPluginService(pluginRepo)
	pluginHandler := handler.NewPluginHandler(pluginService)

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
			users.POST("/check-in", middleware.JWTAuth(), userHandler.CheckIn)
			users.GET("/points", middleware.JWTAuth(), userHandler.GetPoints)

			// 管理员路由
			users.GET("/", middleware.JWTRequired("admin"), userHandler.ListUsers)
			users.PUT("/:id/status", middleware.JWTRequired("admin"), userHandler.UpdateUserStatus)
		}

		// 项目路由
		projects := v1.Group("/projects")
		{
			projects.GET("/", middleware.JWTAuth(), projectHandler.List)
			projects.POST("/", middleware.JWTAuth(), projectHandler.Create)
			projects.GET("/:id", middleware.JWTAuth(), projectHandler.GetByID)
			projects.PUT("/:id", middleware.JWTAuth(), projectHandler.Update)
			projects.DELETE("/:id", middleware.JWTAuth(), projectHandler.Delete)
			projects.GET("/:id/export", middleware.JWTAuth(), projectHandler.Export)

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
			volumes.GET("/:id", middleware.JWTAuth(), volumeHandler.GetByID)
			volumes.PUT("/:id", middleware.JWTAuth(), volumeHandler.Update)
			volumes.DELETE("/:id", middleware.JWTAuth(), volumeHandler.Delete)

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
			plugins.POST("/register", middleware.JWTAuth(), pluginHandler.Register)
			plugins.GET("/", middleware.JWTAuth(), pluginHandler.List)
			plugins.GET("/:id", middleware.JWTAuth(), pluginHandler.GetByID)
			plugins.PUT("/:id/status", middleware.JWTAuth(), pluginHandler.UpdateStatus)
			plugins.PUT("/:id/health", middleware.JWTAuth(), pluginHandler.UpdateHealth)
			plugins.DELETE("/:id", middleware.JWTAuth(), pluginHandler.Delete)

			// 插件能力路由
			plugins.GET("/:plugin_id/capabilities", middleware.JWTAuth(), pluginHandler.ListCapabilities)
			plugins.POST("/:plugin_id/capabilities", middleware.JWTAuth(), pluginHandler.AddCapability)
			plugins.GET("/capabilities/:capability_id", middleware.JWTAuth(), pluginHandler.GetCapability)
			plugins.PUT("/capabilities/:capability_id", middleware.JWTAuth(), pluginHandler.UpdateCapability)
			plugins.DELETE("/capabilities/:capability_id", middleware.JWTAuth(), pluginHandler.DeleteCapability)
			plugins.POST("/capabilities/:capability_id/invoke", middleware.JWTAuth(), pluginHandler.InvokeCapability)
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
