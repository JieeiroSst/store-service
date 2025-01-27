package usecase

import "github.com/JIeeiroSst/coupon-service/internal/repository"

type UserCouponUsecase interface {
}

type userCouponUsecase struct {
	userCouponRepo repository.UserCouponRepository
	cacheUsecase   CacheUsecase
}

func NewUserCouponUsecase(userCouponRepo repository.UserCouponRepository, cacheUsecase CacheUsecase) UserCouponUsecase {
	return &userCouponUsecase{
		userCouponRepo: userCouponRepo,
		cacheUsecase:   cacheUsecase,
	}
}
