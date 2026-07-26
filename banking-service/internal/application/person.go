package application

import (
	"context"

	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
)

type personService struct {
	repo port.PersonRepository
}

func NewPersonService(repo port.PersonRepository) port.PersonUsecase {
	return &personService{repo: repo}
}

func (s *personService) CreatePerson(ctx context.Context, person *model.Person) (*model.Person, error) {
	if err := s.repo.Create(ctx, person); err != nil {
		return nil, err
	}
	return person, nil
}

func (s *personService) GetPerson(ctx context.Context, id int) (*model.Person, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *personService) ListPersons(ctx context.Context) ([]model.Person, error) {
	return s.repo.List(ctx)
}

func (s *personService) UpdatePerson(ctx context.Context, person *model.Person) (*model.Person, error) {
	if _, err := s.repo.GetByID(ctx, person.PersonID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, person); err != nil {
		return nil, err
	}
	return person, nil
}

func (s *personService) DeletePerson(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
