package grpc

import (
	"context"

	pb "github.com/JIeeiroSst/lib-gateway/cdn-service/gateway/cdn-service"

	"github.com/JIeeiroSst/cdn-service/internal/domain/model"
	"github.com/JIeeiroSst/cdn-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	pb.UnimplementedFileServiceServer
	usecase port.CDNUsecase
}

func NewHandler(usecase port.CDNUsecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (*pb.FileResponse, error) {
	lg := logger.WithContext(ctx)
	lg.Info("upload file", zap.String("filename", req.Filename), zap.String("file_type", req.FileType.String()))

	res, err := h.usecase.UploadFile(ctx, model.UploadFileInput{
		Content:  req.Content,
		Filename: req.Filename,
		MimeType: req.MimeType,
		FileType: fromPbFileType(req.FileType),
		Metadata: req.Metadata,
	})
	if err != nil {
		lg.Error("failed to upload file", zap.Error(err))
		return nil, err
	}

	return &pb.FileResponse{
		FileId:    res.FileID,
		Filename:  res.Filename,
		SizeBytes: res.SizeBytes,
		MimeType:  res.MimeType,
		FileType:  toPbFileType(res.FileType),
		Url:       res.URL,
		CreatedAt: timestamppb.New(res.CreatedAt),
		Metadata:  res.Metadata,
	}, nil
}

func (h *Handler) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.FileContentResponse, error) {
	lg := logger.WithContext(ctx)

	res, err := h.usecase.GetFile(ctx, req.FileId)
	if err != nil {
		lg.Error("failed to get file", zap.String("file_id", req.FileId), zap.Error(err))
		return nil, err
	}

	return &pb.FileContentResponse{
		Filename: res.Filename,
		Url:      res.URL,
	}, nil
}

func fromPbFileType(t pb.FileType) model.FileType {
	switch t {
	case pb.FileType_FILE_TYPE_IMAGE:
		return model.FileTypeImage
	case pb.FileType_FILE_TYPE_VIDEO:
		return model.FileTypeVideo
	default:
		return ""
	}
}

func toPbFileType(t model.FileType) pb.FileType {
	switch t {
	case model.FileTypeImage:
		return pb.FileType_FILE_TYPE_IMAGE
	case model.FileTypeVideo:
		return pb.FileType_FILE_TYPE_VIDEO
	default:
		return pb.FileType_FILE_TYPE_UNSPECIFIED
	}
}
