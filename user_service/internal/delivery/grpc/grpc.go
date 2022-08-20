package grpc

import (
	"context"

	"github.com/JIeeiroSst/user-service/internal/pb"
	"github.com/JIeeiroSst/user-service/internal/usecase"
)

type GRPCServer struct {
	usecase usecase.Usecase
}

func (s *GRPCServer) NewGRPCServer(usecase usecase.Usecase) {
	s.usecase = usecase
}

func (s *GRPCServer) Authentication(ctx context.Context, req *pb.AuthenticationRequest) (*pb.AuthenticationReponse, error) {
	token := req.Token
	username := req.Username
	if err := s.usecase.Users.Authentication(token, username); err != nil {
		return &pb.AuthenticationReponse{
			Message: err.Error(),
			Code: "-1",
		} ,err 
	}
	return &pb.AuthenticationReponse{
		Code: "1",
		Message: "user authorized success",
	}, nil
}
