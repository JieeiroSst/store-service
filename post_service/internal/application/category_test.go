package application

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/post-service/internal/domain/port"
	"github.com/JIeeiroSst/post-service/model"
)

func TestCreateCategoryThenGetByID(t *testing.T) {
	svc := NewCategoryService(newFakeCategoryRepository(), &fakeIDGenerator{})

	created, err := svc.CreateCategory(context.Background(), model.CreateCategoryInput{Name: "tech"})
	if err != nil {
		t.Fatalf("CreateCategory() error = %v", err)
	}

	got, err := svc.GetCategory(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetCategory() error = %v", err)
	}
	if got.Name != "tech" {
		t.Errorf("Name = %q, want tech", got.Name)
	}
}

func TestDeleteCategoryThenGetReturnsNotFound(t *testing.T) {
	svc := NewCategoryService(newFakeCategoryRepository(), &fakeIDGenerator{})
	created, err := svc.CreateCategory(context.Background(), model.CreateCategoryInput{Name: "tech"})
	if err != nil {
		t.Fatalf("setup CreateCategory() error = %v", err)
	}

	if err := svc.DeleteCategory(context.Background(), created.ID); err != nil {
		t.Fatalf("DeleteCategory() error = %v", err)
	}
	if _, err := svc.GetCategory(context.Background(), created.ID); err != port.ErrNotFound {
		t.Fatalf("GetCategory() after delete error = %v, want ErrNotFound", err)
	}
}
