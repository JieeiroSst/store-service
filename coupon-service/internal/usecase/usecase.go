package usecase

import "github.com/JIeeiroSst/coupon-service/internal/repository"

type Usecase struct {
	CouponRestrictionUsecase
	CouponUsecase
	CouponUsageUsecase
	UserCouponUsecase
}

type Dependency struct {
	Repos *repository.Repositories
	Cache CacheUsecase
}

func NewUsecase(dep *Dependency) *Usecase {
	return &Usecase{
		CouponRestrictionUsecase: NewCouponRestrictionUsecase(dep.Repos.CouponRepository, dep.Cache),
		CouponUsecase:            NewCouponUsecase(dep.Repos.CouponRepository, dep.Cache),
		CouponUsageUsecase:       NewCouponUsageUsecase(dep.Repos.CouponRepository, dep.Cache),
		UserCouponUsecase:        NewUserCouponUsecase(dep.Repos.UserCouponRepository, dep.Cache),
	}
}
