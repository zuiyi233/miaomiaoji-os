//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet 所有Provider的集合
var ProviderSet = wire.NewSet(
	ProvideBaseRepository,
	ProvideUserRepository,
	ProvideCategoryRepository,
	ProvideProductRepository,
	ProvideCartRepository,
	ProvideOrderRepository,
	ProvideAddressRepository,
	ProvideFavoriteRepository,
	ProvideReviewRepository,
	ProvideCommentRepository,
	ProvideUserService,
	ProvideCategoryService,
	ProvideProductService,
	ProvideCartService,
	ProvideOrderService,
	ProvideAddressService,
	ProvideFavoriteService,
	ProvideReviewService,
	ProvideCommentService,
	ProvideUserHandler,
	ProvideCategoryHandler,
	ProvideProductHandler,
	ProvideCartHandler,
	ProvideOrderHandler,
	ProvideAddressHandler,
	ProvideAdminHandler,
	ProvideFavoriteHandler,
	ProvideReviewHandler,
	ProvideCommentHandler,
	ProvideHandlers,
)

// InitializeHandlers 初始化所有Handler
func InitializeHandlers(db *gorm.DB) (*Handlers, error) {
	wire.Build(ProviderSet)
	return nil, nil
}
