package usecase

import "github.com/JIeeiroSst/coupon-service/internal/repository"

type CouponRestrictionUsecase interface{}

type couponRestrictionUsecase struct {
	couponRestrictionRepo repository.CouponRestrictionRepository
	cacheUsecase          CacheUsecase
}

func NewCouponRestrictionUsecase(couponRestrictionRepo repository.CouponRestrictionRepository,
	cacheUsecase CacheUsecase) CouponRestrictionUsecase {
	return &couponRestrictionUsecase{
		couponRestrictionRepo: couponRestrictionRepo,
		cacheUsecase:          cacheUsecase,
	}
}
