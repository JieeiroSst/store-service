package usecase

import (
	"github.com/JieeiroSst/authorize-service/internal/repository"
	"github.com/JieeiroSst/authorize-service/pkg/cache"
	"github.com/JieeiroSst/authorize-service/pkg/otp"
	"github.com/casbin/casbin/v2/persist"
)

type Usecase struct {
	Casbins
	Otps
}

type Dependency struct {
	Repos       *repository.Repositories
	Adapter     persist.Adapter
	OTP         otp.OTP
	CacheHelper cache.CacheHelper
}

func NewUsecase(deps Dependency) *Usecase {
	casbinUsecase := NewCasbinUsecase(deps.Repos.Casbins, deps.Adapter, deps.CacheHelper)
	otpUsecase := NewOTPUsecase(deps.OTP, deps.CacheHelper)

	return &Usecase{
		Casbins: casbinUsecase,
		Otps:    otpUsecase,
	}
}
