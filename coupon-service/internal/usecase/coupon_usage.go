package usecase

import (
	"context"

	"github.com/JIeeiroSst/coupon-service/internal/dto"
	"github.com/JIeeiroSst/coupon-service/internal/model"
	"github.com/JIeeiroSst/coupon-service/internal/repository"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CouponUsageUsecase interface {
	GetCouponUsageByCouponID(ctx context.Context, couponID uint) (*dto.CouponUsage, error)
	CreateCouponUsage(ctx context.Context, usage *dto.CouponUsage) (*dto.CouponUsage, error)
	UpdateCouponUsage(ctx context.Context, usage dto.CouponUsage) (*dto.CouponUsage, error)
	DeleteCouponUsage(ctx context.Context, couponID int) error
	GetCouponUsageByCouponIDs(ctx context.Context) ([]*dto.CouponUsage, error)
}

type couponUsageUsecase struct {
	couponUsageRepo repository.CouponUsageRepository
	cacheUsecase    CacheUsecase
}

func (c *couponUsageUsecase) GetCouponUsageByCouponIDs(ctx context.Context) ([]*dto.CouponUsage, error) {
	couponUsages, err := c.couponUsageRepo.GetCouponUsages(ctx)
	if err != nil {
		return nil, err
	}

	var dtoCouponUsages []*dto.CouponUsage
	for _, couponUsage := range couponUsages {
		dtoCouponUsages = append(dtoCouponUsages, &dto.CouponUsage{
			Id:             couponUsage.Id,
			CouponId:       couponUsage.CouponId,
			UserId:         couponUsage.UserId,
			UsedAt:         timestamppb.New(couponUsage.UsedAt),
			DiscountAmount: couponUsage.DiscountAmount,
			OrderId:        couponUsage.OrderId,
		})
	}

	return dtoCouponUsages, nil
}

func NewCouponUsageUsecase(couponUsageRepo repository.CouponUsageRepository, cacheUsecase CacheUsecase) CouponUsageUsecase {
	return &couponUsageUsecase{
		couponUsageRepo: couponUsageRepo,
		cacheUsecase:    cacheUsecase,
	}
}

func (c *couponUsageUsecase) GetCouponUsageByCouponID(ctx context.Context, couponID uint) (*dto.CouponUsage, error) {
	couponUsage, err := c.couponUsageRepo.GetCouponUsage(ctx, int(couponID))
	if err != nil {
		return nil, err
	}

	return &dto.CouponUsage{
		Id:             couponUsage.Id,
		CouponId:       couponUsage.CouponId,
		UserId:         couponUsage.UserId,
		UsedAt:         timestamppb.New(couponUsage.UsedAt),
		DiscountAmount: couponUsage.DiscountAmount,
		OrderId:        couponUsage.OrderId,
	}, nil
}

func (c *couponUsageUsecase) CreateCouponUsage(ctx context.Context, usage *dto.CouponUsage) (*dto.CouponUsage, error) {
	couponUsage := &model.CouponUsage{
		CouponId:       usage.CouponId,
		UserId:         usage.UserId,
		UsedAt:         usage.UsedAt.AsTime(),
		DiscountAmount: usage.DiscountAmount,
		OrderId:        usage.OrderId,
	}

	if err := c.couponUsageRepo.CreateCouponUsage(ctx, *couponUsage); err != nil {
		return nil, err
	}

	return &dto.CouponUsage{
		Id:             couponUsage.Id,
		CouponId:       couponUsage.CouponId,
		UserId:         couponUsage.UserId,
		UsedAt:         timestamppb.New(couponUsage.UsedAt),
		DiscountAmount: couponUsage.DiscountAmount,
		OrderId:        couponUsage.OrderId,
	}, nil
}

func (c *couponUsageUsecase) UpdateCouponUsage(ctx context.Context, usage dto.CouponUsage) (*dto.CouponUsage, error) {
	couponUsage := &model.CouponUsage{
		Id:             usage.Id,
		CouponId:       usage.CouponId,
		UserId:         usage.UserId,
		UsedAt:         usage.UsedAt.AsTime(),
		DiscountAmount: usage.DiscountAmount,
		OrderId:        usage.OrderId,
	}

	if err := c.couponUsageRepo.UpdateCouponUsage(ctx, *couponUsage); err != nil {
		return nil, err
	}

	return &dto.CouponUsage{
		Id:             couponUsage.Id,
		CouponId:       couponUsage.CouponId,
		UserId:         couponUsage.UserId,
		UsedAt:         timestamppb.New(couponUsage.UsedAt),
		DiscountAmount: couponUsage.DiscountAmount,
		OrderId:        couponUsage.OrderId,
	}, nil
}

func (c *couponUsageUsecase) DeleteCouponUsage(ctx context.Context, couponID int) error {
	return c.couponUsageRepo.DeleteCouponUsage(ctx, couponID)
}
