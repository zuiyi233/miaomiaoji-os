package wire

import (
	"gorm.io/gorm"
	"rulebacktest/internal/handler"
	"rulebacktest/internal/repository"
	"rulebacktest/internal/service"
)

// ProvideBaseRepository 提供BaseRepository实例
func ProvideBaseRepository(db *gorm.DB) *repository.BaseRepository {
	return repository.NewBaseRepository(db)
}

// ProvideUserRepository 提供UserRepository实例
func ProvideUserRepository(base *repository.BaseRepository) *repository.UserRepository {
	return repository.NewUserRepository(base)
}

// ProvideCategoryRepository 提供CategoryRepository实例
func ProvideCategoryRepository(base *repository.BaseRepository) *repository.CategoryRepository {
	return repository.NewCategoryRepository(base)
}

// ProvideProductRepository 提供ProductRepository实例
func ProvideProductRepository(base *repository.BaseRepository) *repository.ProductRepository {
	return repository.NewProductRepository(base)
}

// ProvideCartRepository 提供CartRepository实例
func ProvideCartRepository(base *repository.BaseRepository) *repository.CartRepository {
	return repository.NewCartRepository(base)
}

// ProvideOrderRepository 提供OrderRepository实例
func ProvideOrderRepository(base *repository.BaseRepository) *repository.OrderRepository {
	return repository.NewOrderRepository(base)
}

// ProvideAddressRepository 提供AddressRepository实例
func ProvideAddressRepository(base *repository.BaseRepository) *repository.AddressRepository {
	return repository.NewAddressRepository(base)
}

// ProvideFavoriteRepository 提供FavoriteRepository实例
func ProvideFavoriteRepository(base *repository.BaseRepository) *repository.FavoriteRepository {
	return repository.NewFavoriteRepository(base)
}

// ProvideReviewRepository 提供ReviewRepository实例
func ProvideReviewRepository(base *repository.BaseRepository) *repository.ReviewRepository {
	return repository.NewReviewRepository(base)
}

// ProvideCommentRepository 提供CommentRepository实例
func ProvideCommentRepository(base *repository.BaseRepository) *repository.CommentRepository {
	return repository.NewCommentRepository(base)
}

// ProvideUserService 提供UserService实例
func ProvideUserService(repo *repository.UserRepository) *service.UserService {
	return service.NewUserService(repo)
}

// ProvideCategoryService 提供CategoryService实例
func ProvideCategoryService(repo *repository.CategoryRepository) *service.CategoryService {
	return service.NewCategoryService(repo)
}

// ProvideProductService 提供ProductService实例
func ProvideProductService(repo *repository.ProductRepository, categoryRepo *repository.CategoryRepository) *service.ProductService {
	return service.NewProductService(repo, categoryRepo)
}

// ProvideCartService 提供CartService实例
func ProvideCartService(repo *repository.CartRepository, productRepo *repository.ProductRepository) *service.CartService {
	return service.NewCartService(repo, productRepo)
}

// ProvideOrderService 提供OrderService实例
func ProvideOrderService(repo *repository.OrderRepository, cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *service.OrderService {
	return service.NewOrderService(repo, cartRepo, productRepo)
}

// ProvideAddressService 提供AddressService实例
func ProvideAddressService(repo *repository.AddressRepository) *service.AddressService {
	return service.NewAddressService(repo)
}

// ProvideFavoriteService 提供FavoriteService实例
func ProvideFavoriteService(repo *repository.FavoriteRepository, productRepo *repository.ProductRepository) *service.FavoriteService {
	return service.NewFavoriteService(repo, productRepo)
}

// ProvideReviewService 提供ReviewService实例
func ProvideReviewService(repo *repository.ReviewRepository, orderRepo *repository.OrderRepository) *service.ReviewService {
	return service.NewReviewService(repo, orderRepo)
}

// ProvideCommentService 提供CommentService实例
func ProvideCommentService(repo *repository.CommentRepository, productRepo *repository.ProductRepository, reviewRepo *repository.ReviewRepository) *service.CommentService {
	return service.NewCommentService(repo, productRepo, reviewRepo)
}

// ProvideUserHandler 提供UserHandler实例
func ProvideUserHandler(svc *service.UserService) *handler.UserHandler {
	return handler.NewUserHandler(svc)
}

// ProvideCategoryHandler 提供CategoryHandler实例
func ProvideCategoryHandler(svc *service.CategoryService) *handler.CategoryHandler {
	return handler.NewCategoryHandler(svc)
}

// ProvideProductHandler 提供ProductHandler实例
func ProvideProductHandler(svc *service.ProductService) *handler.ProductHandler {
	return handler.NewProductHandler(svc)
}

// ProvideCartHandler 提供CartHandler实例
func ProvideCartHandler(svc *service.CartService) *handler.CartHandler {
	return handler.NewCartHandler(svc)
}

// ProvideOrderHandler 提供OrderHandler实例
func ProvideOrderHandler(svc *service.OrderService) *handler.OrderHandler {
	return handler.NewOrderHandler(svc)
}

// ProvideAddressHandler 提供AddressHandler实例
func ProvideAddressHandler(svc *service.AddressService) *handler.AddressHandler {
	return handler.NewAddressHandler(svc)
}

// ProvideAdminHandler 提供AdminHandler实例
func ProvideAdminHandler(orderSvc *service.OrderService, productSvc *service.ProductService, userSvc *service.UserService) *handler.AdminHandler {
	return handler.NewAdminHandler(orderSvc, productSvc, userSvc)
}

// ProvideFavoriteHandler 提供FavoriteHandler实例
func ProvideFavoriteHandler(svc *service.FavoriteService) *handler.FavoriteHandler {
	return handler.NewFavoriteHandler(svc)
}

// ProvideReviewHandler 提供ReviewHandler实例
func ProvideReviewHandler(svc *service.ReviewService) *handler.ReviewHandler {
	return handler.NewReviewHandler(svc)
}

// ProvideCommentHandler 提供CommentHandler实例
func ProvideCommentHandler(svc *service.CommentService) *handler.CommentHandler {
	return handler.NewCommentHandler(svc)
}

// Handlers 包含所有Handler实例
type Handlers struct {
	UserHandler     *handler.UserHandler
	CategoryHandler *handler.CategoryHandler
	ProductHandler  *handler.ProductHandler
	CartHandler     *handler.CartHandler
	OrderHandler    *handler.OrderHandler
	AddressHandler  *handler.AddressHandler
	AdminHandler    *handler.AdminHandler
	FavoriteHandler *handler.FavoriteHandler
	ReviewHandler   *handler.ReviewHandler
	CommentHandler  *handler.CommentHandler
}

// ProvideHandlers 提供所有Handler实例
func ProvideHandlers(
	userHandler *handler.UserHandler,
	categoryHandler *handler.CategoryHandler,
	productHandler *handler.ProductHandler,
	cartHandler *handler.CartHandler,
	orderHandler *handler.OrderHandler,
	addressHandler *handler.AddressHandler,
	adminHandler *handler.AdminHandler,
	favoriteHandler *handler.FavoriteHandler,
	reviewHandler *handler.ReviewHandler,
	commentHandler *handler.CommentHandler,
) *Handlers {
	return &Handlers{
		UserHandler:     userHandler,
		CategoryHandler: categoryHandler,
		ProductHandler:  productHandler,
		CartHandler:     cartHandler,
		OrderHandler:    orderHandler,
		AddressHandler:  addressHandler,
		AdminHandler:    adminHandler,
		FavoriteHandler: favoriteHandler,
		ReviewHandler:   reviewHandler,
		CommentHandler:  commentHandler,
	}
}
