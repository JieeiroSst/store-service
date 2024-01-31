package usecase

type Outbound interface {
}

type OutboundUsecase struct {
}

func NewOutboundUsecase() *OutboundUsecase {
	return &OutboundUsecase{}
}
