package usecase

import (
	"context"

	"github.com/JIeeiroSst/car-rental-service/model"
)

type LocationPage struct {
	Locations     []model.Location
	Total         int64
	NextPageToken string
}

func (u *Usecase) ListLocations(ctx context.Context, city, state, country string, page Page) (*LocationPage, error) {
	ls, total, err := u.repos.Locations.List(ctx, city, state, country, page.Offset, page.Limit)
	if err != nil {
		return nil, err
	}
	return &LocationPage{Locations: ls, Total: total, NextPageToken: page.NextToken(len(ls), total)}, nil
}
