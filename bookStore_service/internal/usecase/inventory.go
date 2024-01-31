package usecase

type Inventories interface {
}

type InventoriesUsecase struct {
}

func NewInventoriesUsecase() *InventoriesUsecase {
	return &InventoriesUsecase{}
}
