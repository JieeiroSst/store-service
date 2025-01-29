package usecase

import (
	"github.com/JIeeiroSst/coupon-service/internal/repository"
)

type Usecase struct {
	CouponRestrictionUsecase
	CouponUsecase
	CouponUsageUsecase
	UserCouponUsecase
}

type Dependency struct {
	Repos    *repository.Repositories
	RedisURl []string
}

func NewUsecase(dep *Dependency) *Usecase {
	cache := NewCacheUsecase(dep.RedisURl)
	return &Usecase{
		CouponRestrictionUsecase: NewCouponRestrictionUsecase(dep.Repos.CouponRestrictionRepository, cache),
		CouponUsecase:            NewCouponUsecase(dep.Repos.CouponRepository, cache),
		CouponUsageUsecase:       NewCouponUsageUsecase(dep.Repos.CouponUsageRepository, cache),
		UserCouponUsecase:        NewUserCouponUsecase(dep.Repos.UserCouponRepository, cache),
	}
}
