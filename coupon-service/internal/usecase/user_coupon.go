package usecase

import (
	"context"

	"github.com/JIeeiroSst/coupon-service/internal/dto"
	"github.com/JIeeiroSst/coupon-service/internal/model"
	"github.com/JIeeiroSst/coupon-service/internal/repository"
	"github.com/JIeeiroSst/utils/geared_id"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserCouponUsecase interface {
	GetUserCouponByUserID(ctx context.Context, userID uint) ([]dto.UserCoupon, error)
	GetUserCouponByCouponID(ctx context.Context, couponID uint) ([]dto.UserCoupon, error)
	CreateUserCoupon(ctx context.Context, userCoupon *dto.UserCoupon) (*dto.UserCoupon, error)
	UpdateUserCoupon(ctx context.Context, userCoupon dto.UserCoupon) (*dto.UserCoupon, error)
	DeleteUserCoupon(ctx context.Context, userCouponID int) error
	GetUserCoupons(ctx context.Context) ([]dto.UserCoupon, error)
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

func (u *userCouponUsecase) GetUserCouponByUserID(ctx context.Context, userID uint) ([]dto.UserCoupon, error) {
	userCoupons, err := u.userCouponRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var dtoUserCoupons []dto.UserCoupon
	for _, userCoupon := range userCoupons {
		dtoUserCoupons = append(dtoUserCoupons, dto.UserCoupon{
			Id:         userCoupon.Id,
			UserId:     userCoupon.UserId,
			CouponId:   userCoupon.CouponId,
			IsUsed:     userCoupon.IsUsed,
			UsedAt:     timestamppb.New(userCoupon.UsedAt),
			AssignedAt: timestamppb.New(userCoupon.AssignedAt),
		})
	}

	return dtoUserCoupons, nil
}

func (u *userCouponUsecase) GetUserCouponByCouponID(ctx context.Context, couponID uint) ([]dto.UserCoupon, error) {
	userCoupons, err := u.userCouponRepo.GetByCouponID(ctx, couponID)
	if err != nil {
		return nil, err
	}

	var dtoUserCoupons []dto.UserCoupon
	for _, userCoupon := range userCoupons {
		dtoUserCoupons = append(dtoUserCoupons, dto.UserCoupon{
			Id:         userCoupon.Id,
			UserId:     userCoupon.UserId,
			CouponId:   userCoupon.CouponId,
			IsUsed:     userCoupon.IsUsed,
			UsedAt:     timestamppb.New(userCoupon.UsedAt),
			AssignedAt: timestamppb.New(userCoupon.AssignedAt),
		})
	}

	return dtoUserCoupons, nil
}

func (u *userCouponUsecase) CreateUserCoupon(ctx context.Context, userCoupon *dto.UserCoupon) (*dto.UserCoupon, error) {
	userCoupons := &model.UserCoupon{
		Id:         int64(geared_id.GearedIntID()),
		UserId:     userCoupon.UserId,
		CouponId:   userCoupon.CouponId,
		IsUsed:     userCoupon.IsUsed,
		UsedAt:     userCoupon.UsedAt.AsTime(),
		AssignedAt: userCoupon.AssignedAt.AsTime(),
	}

	createdUserCoupon, err := u.userCouponRepo.Create(ctx, userCoupons)
	if err != nil {
		return nil, err
	}

	return &dto.UserCoupon{
		Id:         createdUserCoupon.Id,
		UserId:     createdUserCoupon.UserId,
		CouponId:   createdUserCoupon.CouponId,
		IsUsed:     createdUserCoupon.IsUsed,
		UsedAt:     timestamppb.New(createdUserCoupon.UsedAt),
		AssignedAt: timestamppb.New(createdUserCoupon.AssignedAt),
	}, nil
}

func (u *userCouponUsecase) UpdateUserCoupon(ctx context.Context, userCoupon dto.UserCoupon) (*dto.UserCoupon, error) {
	userCoupons := &model.UserCoupon{
		Id:         userCoupon.Id,
		UserId:     userCoupon.UserId,
		CouponId:   userCoupon.CouponId,
		IsUsed:     userCoupon.IsUsed,
		UsedAt:     userCoupon.UsedAt.AsTime(),
		AssignedAt: userCoupon.AssignedAt.AsTime(),
	}

	updatedUserCoupon, err := u.userCouponRepo.Update(ctx, userCoupons)
	if err != nil {
		return nil, err
	}

	return &dto.UserCoupon{
		Id:         updatedUserCoupon.Id,
		UserId:     updatedUserCoupon.UserId,
		CouponId:   updatedUserCoupon.CouponId,
		IsUsed:     updatedUserCoupon.IsUsed,
		UsedAt:     timestamppb.New(updatedUserCoupon.UsedAt),
		AssignedAt: timestamppb.New(updatedUserCoupon.AssignedAt),
	}, nil
}

func (u *userCouponUsecase) DeleteUserCoupon(ctx context.Context, userCouponID int) error {
	userCoupon := &model.UserCoupon{
		Id: int64(userCouponID),
	}

	return u.userCouponRepo.Delete(ctx, userCoupon)
}

func (u *userCouponUsecase) GetUserCoupons(ctx context.Context) ([]dto.UserCoupon, error) {
	userCoupons, err := u.userCouponRepo.GetUserCoupons(ctx)
	if err != nil {
		return nil, err
	}

	var dtoUserCoupons []dto.UserCoupon
	for _, userCoupon := range userCoupons {
		dtoUserCoupons = append(dtoUserCoupons, dto.UserCoupon{
			Id:         userCoupon.Id,
			UserId:     userCoupon.UserId,
			CouponId:   userCoupon.CouponId,
			IsUsed:     userCoupon.IsUsed,
			UsedAt:     timestamppb.New(userCoupon.UsedAt),
			AssignedAt: timestamppb.New(userCoupon.AssignedAt),
		})
	}

	return dtoUserCoupons, nil
}
