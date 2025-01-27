package usecase

import "github.com/JIeeiroSst/coupon-service/internal/repository"

type CouponUsageUsecase interface{}

type couponUsageUsecase struct {
	couponRepo   repository.CouponRepository
	cacheUsecase CacheUsecase
}

func NewCouponUsageUsecase(couponRepo repository.CouponRepository, cacheUsecase CacheUsecase) CouponUsageUsecase {
	return &couponUsageUsecase{
		couponRepo:   couponRepo,
		cacheUsecase: cacheUsecase,
	}
}
