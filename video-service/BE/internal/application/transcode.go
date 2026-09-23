package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/JIeeiroSst/video-service/internal/domain/model"
	"github.com/JIeeiroSst/video-service/internal/domain/port"
	"golang.org/x/sync/errgroup"
)

type TranscodeService struct {
	repo       port.VideoRepository
	storage    port.VideoStorage
	transcoder port.Transcoder
}

func NewTranscodeService(repo port.VideoRepository, storage port.VideoStorage, t port.Transcoder) port.TranscodeUsecase {
	return &TranscodeService{repo: repo, storage: storage, transcoder: t}
}

func (s *TranscodeService) Process(ctx context.Context, id string) error {
	v, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, port.ErrNotFound) {
			return nil
		}
		return err
	}
	if v.Status == model.StatusReady {
		return nil
	}

	dir, err := os.MkdirTemp("", "transcode-"+id+"-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	src := filepath.Join(dir, "source")
	if err := s.download(ctx, v.ObjectKey(), src); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	out := filepath.Join(dir, "out")
	res, err := s.transcoder.Transcode(ctx, src, out)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		log.Printf("transcode %s failed: %v", id, err)
		v.Status = model.StatusFailed
		return s.repo.Save(ctx, *v)
	}

	if err := s.upload(ctx, *v, out); err != nil {
		return fmt.Errorf("upload renditions: %w", err)
	}

	v.Status = model.StatusReady
	v.Duration, v.Width, v.Height = res.Duration, res.Width, res.Height
	return s.repo.Save(ctx, *v)
}

func (s *TranscodeService) download(ctx context.Context, key, dst string) error {
	rc, err := s.storage.Open(ctx, key)
	if err != nil {
		return err
	}
	defer rc.Close()
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, rc)
	return err
}

func (s *TranscodeService) upload(ctx context.Context, v model.Video, out string) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(8)

	err := filepath.WalkDir(out, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(out, path)
		rel = filepath.ToSlash(rel)
		key := v.HLSPrefix() + rel
		if rel == "thumbnail.jpg" {
			key = v.ThumbnailKey()
		}
		g.Go(func() error {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			ct := mime.TypeByExtension(filepath.Ext(path))
			switch {
			case strings.HasSuffix(path, ".m3u8"):
				ct = "application/vnd.apple.mpegurl"
			case strings.HasSuffix(path, ".ts"):
				ct = "video/mp2t"
			}
			_, err = s.storage.Put(ctx, key, f, fileSize(f), ct)
			return err
		})
		return nil
	})
	if werr := g.Wait(); werr != nil {
		return werr
	}
	return err
}

func fileSize(f *os.File) int64 {
	st, err := f.Stat()
	if err != nil {
		return -1
	}
	return st.Size()
}
