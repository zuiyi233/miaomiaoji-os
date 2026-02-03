package router

import (
	"github.com/gin-gonic/gin"
	"rulebacktest/internal/middleware"
	"rulebacktest/internal/wire"
)

// RouteRegister 路由注册函数类型
type RouteRegister func(rg *gin.RouterGroup, handlers *wire.Handlers)

// Setup 初始化并配置路由
func Setup(handlers *wire.Handlers, customRoutes ...RouteRegister) *gin.Engine {
	r := gin.New()

	registerGlobalMiddleware(r)
	registerHealthRoutes(r)
	registerAPIRoutes(r, handlers, customRoutes...)

	return r
}

// registerGlobalMiddleware 注册全局中间件
func registerGlobalMiddleware(r *gin.Engine) {
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
}

// registerHealthRoutes 注册健康检查路由
func registerHealthRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
}

// registerAPIRoutes 注册API路由
func registerAPIRoutes(r *gin.Engine, handlers *wire.Handlers, customRoutes ...RouteRegister) {
	v1 := r.Group("/api/v1")
	{
		registerPublicRoutes(v1, handlers)
		registerAuthenticatedRoutes(v1, handlers)

		for _, register := range customRoutes {
			register(v1, handlers)
		}
	}
}

// registerPublicRoutes 注册公开路由
func registerPublicRoutes(rg *gin.RouterGroup, handlers *wire.Handlers) {
	// 用户认证
	auth := rg.Group("/auth")
	{
		auth.POST("/register", handlers.UserHandler.Register)
		auth.POST("/login", handlers.UserHandler.Login)
	}

	// 商品分类（公开）
	categories := rg.Group("/categories")
	{
		categories.GET("", handlers.CategoryHandler.List)
		categories.GET("/:id", handlers.CategoryHandler.GetByID)
	}

	// 商品列表（公开）
	products := rg.Group("/products")
	{
		products.GET("", handlers.ProductHandler.List)
		products.GET("/:id", handlers.ProductHandler.GetByID)
		products.GET("/:id/reviews", handlers.ReviewHandler.ListByProduct)
		products.GET("/:id/rating", handlers.ReviewHandler.GetProductRating)
	}

	// 评论（公开查询）
	comments := rg.Group("/comments")
	{
		comments.GET("", handlers.CommentHandler.List)
		comments.GET("/count", handlers.CommentHandler.Count)
		comments.GET("/:id/replies", handlers.CommentHandler.ListReplies)
	}
}

// registerAuthenticatedRoutes 注册需要认证的路由
func registerAuthenticatedRoutes(rg *gin.RouterGroup, handlers *wire.Handlers) {
	authenticated := rg.Group("")
	authenticated.Use(middleware.Auth())

	// 用户信息
	user := authenticated.Group("/user")
	{
		user.GET("/profile", handlers.UserHandler.GetProfile)
		user.PUT("/profile", handlers.UserHandler.UpdateProfile)
	}

	// 分类管理（需要认证）
	categories := authenticated.Group("/categories")
	{
		categories.POST("", handlers.CategoryHandler.Create)
		categories.PUT("/:id", handlers.CategoryHandler.Update)
		categories.DELETE("/:id", handlers.CategoryHandler.Delete)
	}

	// 商品管理（需要认证）
	products := authenticated.Group("/products")
	{
		products.POST("", handlers.ProductHandler.Create)
		products.PUT("/:id", handlers.ProductHandler.Update)
		products.DELETE("/:id", handlers.ProductHandler.Delete)
	}

	// 购物车
	cart := authenticated.Group("/cart")
	{
		cart.GET("", handlers.CartHandler.List)
		cart.POST("", handlers.CartHandler.Add)
		cart.PUT("/:id", handlers.CartHandler.Update)
		cart.DELETE("/:id", handlers.CartHandler.Delete)
		cart.DELETE("", handlers.CartHandler.Clear)
	}

	// 订单
	orders := authenticated.Group("/orders")
	{
		orders.GET("", handlers.OrderHandler.List)
		orders.POST("", handlers.OrderHandler.Create)
		orders.GET("/:id", handlers.OrderHandler.GetByID)
		orders.POST("/:id/cancel", handlers.OrderHandler.Cancel)
		orders.POST("/:id/pay", handlers.OrderHandler.Pay)
		orders.POST("/:id/confirm", handlers.OrderHandler.ConfirmReceipt)
	}

	// 收货地址
	addresses := authenticated.Group("/addresses")
	{
		addresses.GET("", handlers.AddressHandler.List)
		addresses.POST("", handlers.AddressHandler.Create)
		addresses.GET("/default", handlers.AddressHandler.GetDefault)
		addresses.GET("/:id", handlers.AddressHandler.GetByID)
		addresses.PUT("/:id", handlers.AddressHandler.Update)
		addresses.DELETE("/:id", handlers.AddressHandler.Delete)
		addresses.POST("/:id/default", handlers.AddressHandler.SetDefault)
	}

	// 收藏
	favorites := authenticated.Group("/favorites")
	{
		favorites.GET("", handlers.FavoriteHandler.List)
		favorites.POST("", handlers.FavoriteHandler.Add)
		favorites.DELETE("/:product_id", handlers.FavoriteHandler.Remove)
		favorites.GET("/:product_id/check", handlers.FavoriteHandler.Check)
	}

	// 评价
	reviews := authenticated.Group("/reviews")
	{
		reviews.POST("", handlers.ReviewHandler.Create)
		reviews.GET("/user", handlers.ReviewHandler.ListByUser)
	}

	// 评论
	comments := authenticated.Group("/comments")
	{
		comments.POST("", handlers.CommentHandler.Create)
		comments.DELETE("/:id", handlers.CommentHandler.Delete)
	}

	// 管理员接口
	registerAdminRoutes(authenticated, handlers)
}

// registerAdminRoutes 注册管理员路由
func registerAdminRoutes(rg *gin.RouterGroup, handlers *wire.Handlers) {
	admin := rg.Group("/admin")
	admin.Use(middleware.RequireRole("admin"))
	{
		// 订单管理
		admin.GET("/orders", handlers.AdminHandler.ListOrders)
		admin.GET("/orders/:id", handlers.AdminHandler.GetOrder)
		admin.POST("/orders/:id/ship", handlers.AdminHandler.ShipOrder)
		admin.POST("/orders/:id/complete", handlers.AdminHandler.CompleteOrder)

		// 商品库存管理
		admin.PUT("/products/:id/stock", handlers.AdminHandler.UpdateProductStock)
	}
}

// RegisterAuthenticatedRoutes 返回一个带认证中间件的路由注册函数
func RegisterAuthenticatedRoutes(register RouteRegister) RouteRegister {
	return func(rg *gin.RouterGroup, handlers *wire.Handlers) {
		authenticated := rg.Group("")
		authenticated.Use(middleware.Auth())
		register(authenticated, handlers)
	}
}
