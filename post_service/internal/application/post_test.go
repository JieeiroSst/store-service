package application

import (
	"context"
	"strings"
	"testing"

	"github.com/JIeeiroSst/post-service/internal/domain/port"
	"github.com/JIeeiroSst/post-service/model"
)

func TestCreatePostWithoutFileLeavesMediaIDNil(t *testing.T) {
	posts := newFakePostRepository()
	svc := NewPostService(posts, newFakeMediaRepository(), &fakeObjectStorage{}, &fakeIDGenerator{})

	post, err := svc.CreatePost(context.Background(), model.CreatePostInput{AuthorID: "author-1", Name: "hello"}, nil)
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if post.MediaID != nil {
		t.Errorf("MediaID = %q, want nil for a text-only post (must be SQL NULL, not \"\", or the fk_news_media constraint rejects it)", *post.MediaID)
	}
}

func TestSafeExt(t *testing.T) {
	cases := map[string]string{
		"cover.PNG":        ".png",
		"../../etc/passwd": "",
		"a.html\x00.png":   ".png",
		"noext":            "",
		"x.tar.gz":         ".gz",
		"evil.p/h":         "",
	}
	for in, want := range cases {
		if got := safeExt(in); got != want {
			t.Errorf("safeExt(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCreatePostWithFileLinksTheUploadedMedia(t *testing.T) {
	posts := newFakePostRepository()
	media := newFakeMediaRepository()
	storage := &fakeObjectStorage{}
	svc := NewPostService(posts, media, storage, &fakeIDGenerator{})

	file := &port.UploadFileInput{FileName: "cover.png"}
	post, err := svc.CreatePost(context.Background(), model.CreatePostInput{AuthorID: "author-1", Name: "hello"}, file)
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if storage.uploadCalls != 1 {
		t.Fatalf("uploadCalls = %d, want 1", storage.uploadCalls)
	}
	if post.MediaID == nil {
		t.Fatal("MediaID is nil, want the uploaded file's media ID")
	}
	stored, ok := media.media[*post.MediaID]
	if !ok {
		t.Fatalf("no media row saved under post.MediaID = %q", *post.MediaID)
	}
	if want := *post.MediaID + ".png"; storage.lastKey != want {
		t.Errorf("object key = %q, want server-generated %q", storage.lastKey, want)
	}
	if !strings.HasSuffix(stored.URL, storage.lastKey) {
		t.Errorf("stored media URL = %q, want it to reference %q", stored.URL, storage.lastKey)
	}
	if stored.Description != "cover.png" {
		t.Errorf("media description = %q, want original file name", stored.Description)
	}
}

func TestCreatePostLinksToCategory(t *testing.T) {
	posts := newFakePostRepository()
	svc := NewPostService(posts, newFakeMediaRepository(), &fakeObjectStorage{}, &fakeIDGenerator{})

	post, err := svc.CreatePost(context.Background(), model.CreatePostInput{AuthorID: "author-1", Name: "hello", CategoryID: "cat-1"}, nil)
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if got := posts.categoryOf[post.ID]; got != "cat-1" {
		t.Errorf("category linked = %q, want cat-1", got)
	}
}

func TestUpdatePostRejectsNonAuthor(t *testing.T) {
	posts := newFakePostRepository()
	svc := NewPostService(posts, newFakeMediaRepository(), &fakeObjectStorage{}, &fakeIDGenerator{})

	post, err := svc.CreatePost(context.Background(), model.CreatePostInput{AuthorID: "author-1", Name: "hello"}, nil)
	if err != nil {
		t.Fatalf("setup CreatePost() error = %v", err)
	}

	err = svc.UpdatePost(context.Background(), post.ID, "someone-else", model.UpdatePostInput{Name: "edited"})
	if err != port.ErrForbidden {
		t.Fatalf("UpdatePost() error = %v, want ErrForbidden", err)
	}
}

func TestUpdatePostAllowsAuthor(t *testing.T) {
	posts := newFakePostRepository()
	svc := NewPostService(posts, newFakeMediaRepository(), &fakeObjectStorage{}, &fakeIDGenerator{})

	post, err := svc.CreatePost(context.Background(), model.CreatePostInput{AuthorID: "author-1", Name: "hello"}, nil)
	if err != nil {
		t.Fatalf("setup CreatePost() error = %v", err)
	}

	if err := svc.UpdatePost(context.Background(), post.ID, "author-1", model.UpdatePostInput{Name: "edited"}); err != nil {
		t.Fatalf("UpdatePost() error = %v", err)
	}
	updated, _ := svc.GetPost(context.Background(), post.ID)
	if updated.Name != "edited" {
		t.Errorf("Name = %q, want edited", updated.Name)
	}
}
