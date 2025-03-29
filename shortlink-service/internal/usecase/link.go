package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/shortlink-service/dto"
	"github.com/JIeeiroSst/shortlink-service/internal/repository"
	"github.com/JIeeiroSst/shortlink-service/model"
	"github.com/JIeeiroSst/utils/cache/expire"
	"github.com/rs/xid"
)

type Links interface {
	RedirectLink(ctx context.Context, shortCode string) (string, error)
	CreateLink(ctx context.Context, link *dto.Link) (string, error)
	GetLinkByID(ctx context.Context, id string) (*dto.Link, error)
	DeleteLink(ctx context.Context, id string) error
	GetLinks(ctx context.Context, page, pageSize int) ([]dto.Link, int64, error)
}

type LinkUsecase struct {
	Repo   *repository.Repositories
	Expire expire.CacheHelper
	Domain string
}

func NewLinkUsecase(repo *repository.Repositories, expire expire.CacheHelper, domain string) *LinkUsecase {
	return &LinkUsecase{
		Repo:   repo,
		Expire: expire,
		Domain: domain,
	}
}

func (u *LinkUsecase) RedirectLink(ctx context.Context, shortCode string) (string, error) {
	link, err := u.Repo.Links.GetLinkByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}
	if link == nil {
		return "", nil
	}

	err = u.Repo.Links.CreateOrUpdateLinkClick(ctx, link.ID)
	if err != nil {
		return "", err
	}

	return link.OriginalURL, nil
}

func (u *LinkUsecase) CreateLink(ctx context.Context, req *dto.Link) (string, error) {
	if req == nil {
		return "", nil
	}
	shortcode := u.createShortCode()
	shortlink := fmt.Sprintf("%s/%s", u.Domain, shortcode)
	link := model.Link{
		ID:          xid.New().String(),
		OriginalURL: req.OriginalURL,
		ShortCode:   shortcode,
		UserID:      req.UserID,
		ExpiredAt:   time.Now().Add(time.Hour * 24 * 7),
		Status:      1,
		CreatedAt:   time.Now(),
		Shortlink:   shortlink,
	}
	err := u.Repo.Links.CreateLink(ctx, &link)
	if err != nil {
		return "", err
	}
	return shortlink, nil
}

func (u *LinkUsecase) GetLinkByID(ctx context.Context, id string) (*dto.Link, error) {
	link, err := u.Repo.Links.GetLinkByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, nil
	}

	return &dto.Link{
		ID:          link.ID,
		OriginalURL: link.OriginalURL,
		ShortCode:   link.ShortCode,
		UserID:      link.UserID,
		ExpiredAt:   link.ExpiredAt,
		CreatedAt:   link.CreatedAt,
		Shortlink:   link.Shortlink,
		TotalClicks: link.TotalClicks,
		Status:      dto.LinkStatus(link.Status),
	}, nil
}

func (u *LinkUsecase) DeleteLink(ctx context.Context, id string) error {
	err := u.Repo.Links.DeleteLink(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (u *LinkUsecase) createShortCode() string {
	return xid.New().String()[:8]
}

func (u *LinkUsecase) GetLinks(ctx context.Context, page, pageSize int) ([]dto.Link, int64, error) {
	links, total, err := u.Repo.Links.GetLinks(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var result []dto.Link
	for _, link := range links {
		result = append(result, dto.Link{
			ID:          link.ID,
			OriginalURL: link.OriginalURL,
			ShortCode:   link.ShortCode,
			UserID:      link.UserID,
			ExpiredAt:   link.ExpiredAt,
			CreatedAt:   link.CreatedAt,
			Shortlink:   link.Shortlink,
			TotalClicks: link.TotalClicks,
			Status:      dto.LinkStatus(link.Status),
		})
	}

	return result, total, nil
}
