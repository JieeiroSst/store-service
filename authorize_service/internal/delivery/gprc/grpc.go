package gprc

import (
	"context"
	"errors"

	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/pb"
	"github.com/JieeiroSst/authorize-service/internal/usecase"
	"github.com/JieeiroSst/authorize-service/model"
)

type GRPCServer struct {
	usecase usecase.Casbins
	pb.UnimplementedAuthorizeServer
}

func (s *GRPCServer) NewGRPCServer(usecase usecase.Casbins) {
	s.usecase = usecase
}

func (s *GRPCServer) EnforceCasbin(ctx context.Context, req *pb.CasbinRuleRequest) (*pb.CasbinRuleReponse, error) {
	casbin := model.CasbinAuth{
		Sub: req.Sub,
		Obj: req.Obj,
		Act: req.Act,
	}

	err := s.usecase.EnforceCasbin(ctx, casbin)

	if errors.Is(err, common.FailedDB) {
		return &pb.CasbinRuleReponse{
			Message: common.FailedDB.Error(),
			Error:   false,
		}, err
	}

	if errors.Is(err, common.Failedenforcer) {
		return &pb.CasbinRuleReponse{
			Message: common.Failedenforcer.Error(),
			Error:   false,
		}, err
	}

	if errors.Is(err, common.NotAllow) {
		return &pb.CasbinRuleReponse{
			Message: common.NotAllow.Error(),
			Error:   false,
		}, err
	}

	return &pb.CasbinRuleReponse{
		Message: "THE CUSTOMER IS AUTHORIZED FOR THE CONTENT REQUESTED",
		Error:   true,
	}, nil
}
