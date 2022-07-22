package usecase

import (
	"github.com/JieeiroSst/authorize-service/internal/repository"
	"github.com/JieeiroSst/authorize-service/pkg/otp"
	"github.com/JieeiroSst/authorize-service/pkg/snowflake"
	"github.com/casbin/casbin/v2/persist"
)

type Usecase struct {
	Casbins 
	Otps
}

type Dependency struct {
	Repos     *repository.Repositories
	Snowflake snowflake.SnowflakeData
	Adapter   persist.Adapter
	OTP       otp.OTP
}

func NewUsecase(deps Dependency) *Usecase {
	casbinUsecase := NewCasbinUsecase(deps.Repos.Casbins, deps.Snowflake, deps.Adapter)
	otpUsecase := NewOTPUsecase(deps.OTP)

	return &Usecase{
		Casbins: casbinUsecase,
		Otps: otpUsecase,
	}
}
