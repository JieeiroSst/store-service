package usecase

import (
	"context"

	"github.com/JIeeiroSst/coupon-service/internal/dto"
	"github.com/JIeeiroSst/coupon-service/internal/model"
	"github.com/JIeeiroSst/coupon-service/internal/repository"
	"github.com/JIeeiroSst/utils/geared_id"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CouponUsecase interface {
	GetCouponByID(ctx context.Context, couponID uint) (*dto.Coupon, error)
	CreateCoupon(ctx context.Context, coupon *dto.Coupon) (*dto.Coupon, error)
	UpdateCoupon(ctx context.Context, coupon dto.Coupon) (*dto.Coupon, error)
	DeleteCoupon(ctx context.Context, couponID int) error
	GetCoupons(ctx context.Context) ([]*dto.Coupon, error)
}

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

func (c *couponUsecase) GetCouponByID(ctx context.Context, couponID uint) (*dto.Coupon, error) {
	coupon, err := c.couponRepo.GetCoupon(ctx, int(couponID))
	if err != nil {
		return nil, err
	}

	return &dto.Coupon{
		Id:                coupon.Id,
		Code:              coupon.Code,
		DiscountValue:     coupon.DiscountValue,
		MinimumPurchase:   coupon.MinimumPurchase,
		MaxDiscountAmount: coupon.MaxDiscountAmount,
		Description:       coupon.Description,
		StartDate:         timestamppb.New(coupon.StartDate),
		EndDate:           timestamppb.New(coupon.EndDate),
		IsActive:          coupon.IsActive,
		MaxUses:           coupon.MaxUses,
		CurrentUses:       coupon.CurrentUses,
		CreatedAt:         timestamppb.New(coupon.CreatedAt),
		UpdatedAt:         timestamppb.New(coupon.UpdatedAt),
	}, nil
}

func (c *couponUsecase) CreateCoupon(ctx context.Context, coupon *dto.Coupon) (*dto.Coupon, error) {
	coupons := &model.Coupon{
		Id:                int64(geared_id.GearedIntID()),
		Code:              coupon.Code,
		DiscountValue:     coupon.DiscountValue,
		MinimumPurchase:   coupon.MinimumPurchase,
		MaxDiscountAmount: coupon.MaxDiscountAmount,
		Description:       coupon.Description,
		StartDate:         coupon.StartDate.AsTime(),
		EndDate:           coupon.EndDate.AsTime(),
		IsActive:          coupon.IsActive,
		MaxUses:           coupon.MaxUses,
		CurrentUses:       coupon.CurrentUses,
	}

	createdCoupon, err := c.couponRepo.CreateCoupon(ctx, coupons)
	if err != nil {
		return nil, err
	}

	return &dto.Coupon{
		Id:                createdCoupon.Id,
		Code:              createdCoupon.Code,
		DiscountValue:     createdCoupon.DiscountValue,
		MinimumPurchase:   createdCoupon.MinimumPurchase,
		MaxDiscountAmount: createdCoupon.MaxDiscountAmount,
		Description:       createdCoupon.Description,
		StartDate:         timestamppb.New(createdCoupon.StartDate),
		EndDate:           timestamppb.New(createdCoupon.EndDate),
		IsActive:          createdCoupon.IsActive,
		MaxUses:           createdCoupon.MaxUses,
		CurrentUses:       createdCoupon.CurrentUses,
		CreatedAt:         timestamppb.New(createdCoupon.CreatedAt),
		UpdatedAt:         timestamppb.New(createdCoupon.UpdatedAt),
	}, nil
}

func (c *couponUsecase) UpdateCoupon(ctx context.Context, coupon dto.Coupon) (*dto.Coupon, error) {
	coupons := &model.Coupon{
		Id:                coupon.Id,
		Code:              coupon.Code,
		DiscountValue:     coupon.DiscountValue,
		MinimumPurchase:   coupon.MinimumPurchase,
		MaxDiscountAmount: coupon.MaxDiscountAmount,
		Description:       coupon.Description,
		StartDate:         coupon.StartDate.AsTime(),
		EndDate:           coupon.EndDate.AsTime(),
		IsActive:          coupon.IsActive,
		MaxUses:           coupon.MaxUses,
		CurrentUses:       coupon.CurrentUses,
	}

	updatedCoupon, err := c.couponRepo.UpdateCoupon(ctx, coupons)
	if err != nil {
		return nil, err
	}

	return &dto.Coupon{
		Id:                updatedCoupon.Id,
		Code:              updatedCoupon.Code,
		DiscountValue:     updatedCoupon.DiscountValue,
		MinimumPurchase:   updatedCoupon.MinimumPurchase,
		MaxDiscountAmount: updatedCoupon.MaxDiscountAmount,
		Description:       updatedCoupon.Description,
		StartDate:         timestamppb.New(updatedCoupon.StartDate),
		EndDate:           timestamppb.New(updatedCoupon.EndDate),
		IsActive:          updatedCoupon.IsActive,
		MaxUses:           updatedCoupon.MaxUses,
		CurrentUses:       updatedCoupon.CurrentUses,
		CreatedAt:         timestamppb.New(updatedCoupon.CreatedAt),
		UpdatedAt:         timestamppb.New(updatedCoupon.UpdatedAt),
	}, nil
}

func (c *couponUsecase) DeleteCoupon(ctx context.Context, couponID int) error {
	return c.couponRepo.DeleteCoupon(ctx, couponID)
}

func (c *couponUsecase) GetCoupons(ctx context.Context) ([]*dto.Coupon, error) {
	coupons, err := c.couponRepo.GetCoupons(ctx)
	if err != nil {
		return nil, err
	}

	var couponsDTO []*dto.Coupon
	for _, coupon := range coupons {
		couponsDTO = append(couponsDTO, &dto.Coupon{
			Id:                coupon.Id,
			Code:              coupon.Code,
			DiscountValue:     coupon.DiscountValue,
			MinimumPurchase:   coupon.MinimumPurchase,
			MaxDiscountAmount: coupon.MaxDiscountAmount,
			Description:       coupon.Description,
			StartDate:         timestamppb.New(coupon.StartDate),
			EndDate:           timestamppb.New(coupon.EndDate),
			IsActive:          coupon.IsActive,
			MaxUses:           coupon.MaxUses,
			CurrentUses:       coupon.CurrentUses,
			CreatedAt:         timestamppb.New(coupon.CreatedAt),
			UpdatedAt:         timestamppb.New(coupon.UpdatedAt),
		})
	}

	return couponsDTO, nil
}
