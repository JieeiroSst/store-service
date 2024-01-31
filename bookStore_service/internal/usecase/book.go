package usecase

type Books interface {
}

type BookUsecase struct {
}

func NewBookUsecase() *BookUsecase {
	return &BookUsecase{}
}
