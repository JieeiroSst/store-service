package http

import (
	"context"

	"github.com/JIeeiroSst/cdn-service/dto"
	"github.com/JIeeiroSst/cdn-service/internal/usecase"
	pb "github.com/JIeeiroSst/lib-gateway/cdn-service/gateway/cdn-service"
	"github.com/JIeeiroSst/utils/logger"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type HandlerV1 struct {
	usecase *usecase.Usecase
	pb.UnimplementedFileServiceServer
}

func NewHandlerV1(usecase *usecase.Usecase) *HandlerV1 {
	return &HandlerV1{
		usecase: usecase,
	}
}

func (h *HandlerV1) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (*pb.FileResponse, error) {
	lg := logger.WithContext(ctx)
	lg.Info("upload file", zap.String("filename", req.Filename), zap.String("file_type", req.FileType.String()))
	file := dto.UploadFileRequest{
		Filename: req.Filename,
		FileType: req.FileType,
		MimeType: req.MimeType,
		Metadata: req.Metadata,
		Content:  req.Content,
	}

	res, err := h.usecase.UploadFile(ctx, file)
	if err != nil {
		lg.Info("error get file: %v", zap.Error(err))
		return nil, err
	}

	return &pb.FileResponse{
		FileId:    res.FileID,
		Filename:  res.Filename,
		SizeBytes: res.SizeBytes,
		MimeType:  res.MimeType,
		FileType:  pb.FileType(pb.FileType_value[res.FileType]),
		Url:       res.Url,
		CreatedAt: timestamppb.New(res.CreatedAt),
		Metadata:  file.Metadata,
	}, err
}

func (h *HandlerV1) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.FileContentResponse, error) {
	lg := logger.WithContext(ctx)
	resp, err := h.usecase.CDNs.GetFile(ctx, req.FileId)
	if err != nil {
		lg.Info("error get file: %v", zap.Error(err))
		return nil, err
	}
	return &pb.FileContentResponse{
		Filename: resp.Filename,
		Url:      resp.Url,
	}, nil
}
