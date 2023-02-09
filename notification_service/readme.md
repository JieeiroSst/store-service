mockey-gen:
	mockgen -destination=mocks/mock_repository.go -package=mocks github.com/yanoandri/simple-goorm/repository IPaymentRepository
	mockgen -destination=mocks/mock_usecase.go -package=mocks github.com/yanoandri/simple-goorm/usecase IPaymentUsecase