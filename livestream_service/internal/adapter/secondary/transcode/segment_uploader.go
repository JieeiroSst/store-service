package transcode

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

type segmentUploader struct {
	storage port.ObjectStorage
}

func newSegmentUploader(storage port.ObjectStorage) *segmentUploader {
	return &segmentUploader{storage: storage}
}

func (u *segmentUploader) watch(ctx context.Context, streamKey, dir string, lastActivity *atomic.Int64) (func(), error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := watcher.Add(dir); err != nil {
		_ = watcher.Close()
		return nil, err
	}

	watchCtx, cancel := context.WithCancel(ctx)
	go func() {
		defer watcher.Close()
		for {
			select {
			case <-watchCtx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
					continue
				}
				u.uploadFile(watchCtx, streamKey, event.Name, lastActivity)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.WithContext(watchCtx).Warn("segment watcher error",
					zap.String("streamKey", streamKey), zap.Error(err))
			}
		}
	}()

	return cancel, nil
}

func (u *segmentUploader) uploadFile(ctx context.Context, streamKey, path string, lastActivity *atomic.Int64) {
	name := filepath.Base(path)
	isPlaylist := strings.HasSuffix(name, ".m3u8")
	isSegment := strings.HasSuffix(name, ".ts")
	if !isPlaylist && !isSegment {
		return
	}

	info, err := os.Stat(path)
	if err != nil || info.Size() == 0 {
		return // still being written, or already rotated away by ffmpeg
	}

	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	contentType := "video/mp2t"
	cacheControl := "public, max-age=31536000, immutable"
	if isPlaylist {
		contentType = "application/vnd.apple.mpegurl"
		cacheControl = "public, max-age=1"
	}

	key := streamKey + "/" + name
	if err := u.storage.PutObject(ctx, key, f, info.Size(), contentType, cacheControl); err != nil {
		logger.WithContext(ctx).Warn("upload hls object failed",
			zap.String("streamKey", streamKey), zap.String("key", key), zap.Error(err))
		return
	}
	lastActivity.Store(time.Now().UnixNano())

	if isSegment {
		_ = os.Remove(path)
	}
}
