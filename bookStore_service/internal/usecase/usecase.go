package usecase

type Usecase struct {
}

type Dependency struct {
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{}
}
