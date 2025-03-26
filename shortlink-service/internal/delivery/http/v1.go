package http

import (
	"context"

	grpc "github.com/JIeeiroSst/lib-gateway/shortlink-service/gateway/shortlink-service"
	"github.com/JIeeiroSst/shortlink-service/dto"
	"github.com/JIeeiroSst/shortlink-service/internal/usecase"
	"google.golang.org/protobuf/types/known/emptypb"
)

type HandlerV1 struct {
	usecase *usecase.Usecase
	grpc.UnimplementedShortlinkServiceServer
}

func NewHandlerV1(usecase *usecase.Usecase) *HandlerV1 {
	return &HandlerV1{
		usecase: usecase,
	}
}

func (h *HandlerV1) CreateLink(ctx context.Context, in *grpc.CreateLinkRequest) (*grpc.Link, error) {
	shortlink, err := h.usecase.Links.CreateLink(ctx, &dto.Link{
		OriginalURL: in.OriginalUrl,
	})
	if err != nil {
		return nil, err
	}
	return &grpc.Link{
		ShortCode: shortlink,
	}, nil
}
func (h *HandlerV1) GetLink(ctx context.Context, in *grpc.GetLinkRequest) (*grpc.Link, error) {

	return nil, nil
}

func (h *HandlerV1) ListLinks(ctx context.Context, in *grpc.ListLinksRequest) (*grpc.ListLinksResponse, error) {
	return nil, nil
}

func (h *HandlerV1) DeleteLink(ctx context.Context, in *grpc.GetLinkRequest) (*emptypb.Empty, error) {
	if err := h.usecase.Links.DeleteLink(ctx, in.ShortCode); err != nil {
		return nil, err
	}
	return nil, nil
}

func (h *HandlerV1) GetLinkStats(ctx context.Context, in *grpc.LinkStatsRequest) (*grpc.LinkStatsResponse, error) {

	return &grpc.LinkStatsResponse{}, nil
}
