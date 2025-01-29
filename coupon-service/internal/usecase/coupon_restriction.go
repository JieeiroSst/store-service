package usecase

import (
	"context"

	"github.com/JIeeiroSst/coupon-service/internal/dto"
	"github.com/JIeeiroSst/coupon-service/internal/model"
	"github.com/JIeeiroSst/coupon-service/internal/repository"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CouponRestrictionUsecase interface {
	GetCouponRestrictionByCouponID(ctx context.Context, couponID uint) (*dto.CouponRestriction, error)
	CreateCouponRestriction(ctx context.Context, restrictions *dto.CouponRestriction) (*dto.CouponRestriction, error)
	UpdateCouponRestriction(ctx context.Context, restrictions dto.CouponRestriction) (*dto.CouponRestriction, error)
	DeleteCouponRestriction(ctx context.Context, couponID int) error
	GetCouponRestrictionByCouponIDs(ctx context.Context) ([]*dto.CouponRestriction, error)
}

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

func (c *couponRestrictionUsecase) GetCouponRestrictionByCouponID(ctx context.Context, couponID uint) (*dto.CouponRestriction, error) {
	couponRestriction, err := c.couponRestrictionRepo.GetCouponRestriction(ctx, int(couponID))
	if err != nil {
		return nil, err
	}

	return &dto.CouponRestriction{
		Id:                 couponRestriction.Id,
		CouponId:           couponRestriction.CouponId,
		RestrictionType:    couponRestriction.RestrictionType,
		RestrictedEntityId: couponRestriction.RestrictedEntityId,
		IsExclude:          couponRestriction.IsExclude,
		CreatedAt:          timestamppb.New(couponRestriction.CreatedAt),
	}, nil
}

func (c *couponRestrictionUsecase) CreateCouponRestriction(ctx context.Context, restrictions *dto.CouponRestriction) (*dto.CouponRestriction, error) {
	couponRestrictions := &model.CouponRestriction{
		CouponId:           restrictions.CouponId,
		RestrictionType:    restrictions.RestrictionType,
		RestrictedEntityId: restrictions.RestrictedEntityId,
		IsExclude:          restrictions.IsExclude,
	}

	createdCouponRestriction, err := c.couponRestrictionRepo.CreateCouponRestriction(ctx, couponRestrictions)
	if err != nil {
		return nil, err
	}

	return &dto.CouponRestriction{
		Id:                 createdCouponRestriction.Id,
		CouponId:           createdCouponRestriction.CouponId,
		RestrictionType:    createdCouponRestriction.RestrictionType,
		RestrictedEntityId: createdCouponRestriction.RestrictedEntityId,
		IsExclude:          createdCouponRestriction.IsExclude,
		CreatedAt:          timestamppb.New(createdCouponRestriction.CreatedAt),
	}, nil
}

func (c *couponRestrictionUsecase) UpdateCouponRestriction(ctx context.Context, restrictions dto.CouponRestriction) (*dto.CouponRestriction, error) {
	couponRestrictions := &model.CouponRestriction{
		Id:                 restrictions.Id,
		CouponId:           restrictions.CouponId,
		RestrictionType:    restrictions.RestrictionType,
		RestrictedEntityId: restrictions.RestrictedEntityId,
		IsExclude:          restrictions.IsExclude,
	}

	updatedCouponRestriction, err := c.couponRestrictionRepo.UpdateCouponRestriction(ctx, couponRestrictions)
	if err != nil {
		return nil, err
	}

	return &dto.CouponRestriction{
		Id:                 updatedCouponRestriction.Id,
		CouponId:           updatedCouponRestriction.CouponId,
		RestrictionType:    updatedCouponRestriction.RestrictionType,
		RestrictedEntityId: updatedCouponRestriction.RestrictedEntityId,
		IsExclude:          updatedCouponRestriction.IsExclude,
		CreatedAt:          timestamppb.New(updatedCouponRestriction.CreatedAt),
	}, nil
}

func (c *couponRestrictionUsecase) DeleteCouponRestriction(ctx context.Context, couponID int) error {
	return c.couponRestrictionRepo.DeleteCouponRestriction(ctx, couponID)
}

func (c *couponRestrictionUsecase) GetCouponRestrictionByCouponIDs(ctx context.Context) ([]*dto.CouponRestriction, error) {
	couponRestrictions, err := c.couponRestrictionRepo.GetCouponRestrictions(ctx)
	if err != nil {
		return nil, err
	}

	var dtoCouponRestrictions []*dto.CouponRestriction
	for _, couponRestriction := range couponRestrictions {
		dtoCouponRestrictions = append(dtoCouponRestrictions, &dto.CouponRestriction{
			Id:                 couponRestriction.Id,
			CouponId:           couponRestriction.CouponId,
			RestrictionType:    couponRestriction.RestrictionType,
			RestrictedEntityId: couponRestriction.RestrictedEntityId,
			IsExclude:          couponRestriction.IsExclude,
			CreatedAt:          timestamppb.New(couponRestriction.CreatedAt),
		})
	}

	return dtoCouponRestrictions, nil
}
