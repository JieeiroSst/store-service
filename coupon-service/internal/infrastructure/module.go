package infrastructure

import (
	cronadapter "github.com/JIeeiroSst/coupon-service/internal/adapter/primary/cron"
	httpadapter "github.com/JIeeiroSst/coupon-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/coupon-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/coupon-service/internal/application"
	"github.com/JIeeiroSst/coupon-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/coupon-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB

	repository.Module, // port.CouponRepository, port.CouponRestrictionRepository, port.CouponUsageRepository, port.UserCouponRepository

	application.Module, // port.CouponUsecase, port.CouponRestrictionUsecase, port.CouponUsageUsecase, port.UserCouponUsecase

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
	fx.Invoke(cronadapter.New),
)
