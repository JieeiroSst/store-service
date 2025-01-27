package usecase

import "github.com/JIeeiroSst/coupon-service/internal/repository"

type CouponUsecase interface{}

type couponUsecase struct {
	couponRepo   repository.CouponRepository
	cacheUsecase CacheUsecase
}

func NewCouponUsecase(couponRepo repository.CouponRepository, cacheUsecase CacheUsecase) CouponUsecase {
	return &couponUsecase{
		couponRepo:   couponRepo,
		cacheUsecase: cacheUsecase,
	}
}
