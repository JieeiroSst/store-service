package application

import (
	"context"
	"strconv"

	"github.com/JIeeiroSst/post-service/internal/domain/port"
	"github.com/JIeeiroSst/post-service/model"
)

type fakePostRepository struct {
	posts      map[string]*model.Post
	categoryOf map[string]string // postID -> categoryID, for assertions
}

func newFakePostRepository() *fakePostRepository {
	return &fakePostRepository{posts: make(map[string]*model.Post), categoryOf: make(map[string]string)}
}

func (f *fakePostRepository) Create(ctx context.Context, post *model.Post, categoryID string) error {
	f.posts[post.ID] = post
	if categoryID != "" {
		f.categoryOf[post.ID] = categoryID
	}
	return nil
}

func (f *fakePostRepository) GetByID(ctx context.Context, id string) (*model.Post, error) {
	post, ok := f.posts[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	got := *post
	return &got, nil
}

func (f *fakePostRepository) List(ctx context.Context, cursor string, limit int) ([]model.Post, string, error) {
	var result []model.Post
	for _, p := range f.posts {
		result = append(result, *p)
	}
	return result, "", nil
}

func (f *fakePostRepository) Update(ctx context.Context, id string, post *model.Post) error {
	if _, ok := f.posts[id]; !ok {
		return port.ErrNotFound
	}
	f.posts[id] = post
	return nil
}

type fakeMediaRepository struct {
	media map[string]*model.Media
}

func newFakeMediaRepository() *fakeMediaRepository {
	return &fakeMediaRepository{media: make(map[string]*model.Media)}
}

func (f *fakeMediaRepository) Create(ctx context.Context, media *model.Media) error {
	f.media[media.ID] = media
	return nil
}

type fakeObjectStorage struct {
	uploadCalls int
	err         error
}

func (f *fakeObjectStorage) UploadFile(ctx context.Context, input port.UploadFileInput) (*port.UploadFileResult, error) {
	f.uploadCalls++
	if f.err != nil {
		return nil, f.err
	}
	return &port.UploadFileResult{URL: "https://cdn.example.com/" + input.FileName}, nil
}

func (f *fakeObjectStorage) RemoveObject(ctx context.Context, fileName string) error {
	return nil
}

type fakeIDGenerator struct {
	next int
}

func (f *fakeIDGenerator) NewID() string {
	f.next++
	return "id-" + strconv.Itoa(f.next)
}

type fakeCategoryRepository struct {
	categories map[string]*model.Category
}

func newFakeCategoryRepository() *fakeCategoryRepository {
	return &fakeCategoryRepository{categories: make(map[string]*model.Category)}
}

func (f *fakeCategoryRepository) Create(ctx context.Context, category *model.Category) error {
	f.categories[category.ID] = category
	return nil
}

func (f *fakeCategoryRepository) Update(ctx context.Context, id string, category *model.Category) error {
	existing, ok := f.categories[id]
	if !ok {
		return port.ErrNotFound
	}
	existing.Name = category.Name
	existing.Description = category.Description
	return nil
}

func (f *fakeCategoryRepository) Delete(ctx context.Context, id string) error {
	if _, ok := f.categories[id]; !ok {
		return port.ErrNotFound
	}
	delete(f.categories, id)
	return nil
}

func (f *fakeCategoryRepository) GetByID(ctx context.Context, id string) (*model.Category, error) {
	category, ok := f.categories[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	got := *category
	return &got, nil
}

func (f *fakeCategoryRepository) List(ctx context.Context, cursor string, limit int) ([]model.Category, string, error) {
	var result []model.Category
	for _, c := range f.categories {
		result = append(result, *c)
	}
	return result, "", nil
}
