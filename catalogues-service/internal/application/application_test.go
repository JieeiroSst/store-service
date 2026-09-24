package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
)

type fakeCategoryRepo struct {
	byID map[int64]model.Category
	next int64
}

func newFakeCategoryRepo() *fakeCategoryRepo {
	return &fakeCategoryRepo{byID: map[int64]model.Category{}}
}

func (f *fakeCategoryRepo) Create(_ context.Context, c *model.Category) error {
	f.next++
	c.ID = f.next
	f.byID[c.ID] = *c
	return nil
}
func (f *fakeCategoryRepo) Get(_ context.Context, id int64) (*model.Category, error) {
	c, ok := f.byID[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return &c, nil
}
func (f *fakeCategoryRepo) List(context.Context) ([]model.Category, error) { return nil, nil }
func (f *fakeCategoryRepo) Update(_ context.Context, c *model.Category) error {
	f.byID[c.ID] = *c
	return nil
}
func (f *fakeCategoryRepo) Delete(context.Context, int64) error { return nil }

func TestSlugify(t *testing.T) {
	for in, want := range map[string]string{
		"Go in Action": "go-in-action", "  A -- B!! ": "a-b", "***": "", "Sách 2e": "s-ch-2e",
	} {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCategoryCreate(t *testing.T) {
	ctx := context.Background()
	svc := NewCategoryService(newFakeCategoryRepo())

	if _, err := svc.Create(ctx, model.Category{}); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("empty name: got %v, want ErrInvalid", err)
	}

	root, err := svc.Create(ctx, model.Category{Name: "Books", IsPublic: false})
	if err != nil || root.Slug != "books" {
		t.Fatalf("root = %+v, %v", root, err)
	}

	child, err := svc.Create(ctx, model.Category{Name: "Fiction", IsPublic: true, ParentID: &root.ID})
	if err != nil {
		t.Fatal(err)
	}
	if child.AncestorsArePublic {
		t.Error("child of a private category must not have public ancestors")
	}

	missing := int64(99)
	if _, err := svc.Create(ctx, model.Category{Name: "X", ParentID: &missing}); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("missing parent: got %v, want ErrInvalid", err)
	}
}

func TestCategoryUpdateRejectsSelfParent(t *testing.T) {
	ctx := context.Background()
	svc := NewCategoryService(newFakeCategoryRepo())
	c, _ := svc.Create(ctx, model.Category{Name: "A"})
	if _, err := svc.Update(ctx, model.Category{ID: c.ID, Name: "A", ParentID: &c.ID}); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("got %v, want ErrInvalid", err)
	}
	if _, err := svc.Update(ctx, model.Category{ID: 42, Name: "A"}); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestPrepareProduct(t *testing.T) {
	one, upc, blank := int64(1), "  123 ", "  "
	rating := 6.0

	cases := []struct {
		name    string
		in      model.Product
		invalid bool
	}{
		{"defaults to standalone", model.Product{Title: "T"}, false},
		{"missing title", model.Product{}, true},
		{"unknown structure", model.Product{Title: "T", Structure: "weird"}, true},
		{"child needs parent", model.Product{Title: "T", Structure: model.StructureChild}, true},
		{"standalone with parent", model.Product{Title: "T", ParentID: &one}, true},
		{"child with parent", model.Product{Title: "T", Structure: model.StructureChild, ParentID: &one}, false},
		{"self parent", model.Product{ID: 1, Title: "T", Structure: model.StructureChild, ParentID: &one}, true},
		{"rating out of range", model.Product{Title: "T", Rating: &rating}, true},
		{"upc trimmed", model.Product{Title: "T", UPC: &upc}, false},
		{"blank upc becomes nil", model.Product{Title: "T", UPC: &blank}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := prepareProduct(&tc.in)
			if tc.invalid != errors.Is(err, model.ErrInvalid) {
				t.Fatalf("err = %v, invalid want %v", err, tc.invalid)
			}
		})
	}

	p := model.Product{Title: "Go in Action", UPC: &upc, CategoryIDs: []int64{1, 1, 2}}
	if err := prepareProduct(&p); err != nil {
		t.Fatal(err)
	}
	if p.Slug != "go-in-action" || p.Structure != model.StructureStandalone || *p.UPC != "123" || len(p.CategoryIDs) != 2 {
		t.Errorf("normalised product = %+v", p)
	}
}

func TestSetRecommendationsValidation(t *testing.T) {
	svc := &productService{}
	ctx := context.Background()
	if _, err := svc.SetRecommendations(ctx, 1, []model.Recommendation{{RecommendedID: 1}}); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("self recommendation: got %v", err)
	}
	dup := []model.Recommendation{{RecommendedID: 2}, {RecommendedID: 2}}
	if _, err := svc.SetRecommendations(ctx, 1, dup); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("duplicate recommendation: got %v", err)
	}
}

func TestPrepareOption(t *testing.T) {
	o := model.Option{Name: "Gift wrap"}
	if err := prepareOption(&o); err != nil || o.Code != "gift-wrap" || o.Type != model.OptionText {
		t.Errorf("defaults: %+v, %v", o, err)
	}
	bad := model.Option{Name: "x", Type: "blob"}
	if err := prepareOption(&bad); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("bad type: got %v", err)
	}
}
