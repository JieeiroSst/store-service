package transcode

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JIeeiroSst/livestream-service/config"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/JIeeiroSst/livestream-service/internal/infrastructure/metrics"
	"github.com/JIeeiroSst/utils/logger"
	"go.uber.org/zap"
)

type job struct {
	streamKey string

	mu                sync.Mutex
	cmd               *exec.Cmd
	cancel            context.CancelFunc
	stopping          bool
	restartTimestamps []time.Time
	stopUploader      func()

	lastActivity atomic.Int64
}

type ffmpegRunner struct {
	mu   sync.Mutex
	jobs map[string]*job

	cfg      *config.Config
	uploader *segmentUploader
}

func NewFFmpegRunner(cfg *config.Config, storage port.ObjectStorage) port.TranscodeRunner {
	return &ffmpegRunner{
		jobs:     make(map[string]*job),
		cfg:      cfg,
		uploader: newSegmentUploader(storage),
	}
}

func (r *ffmpegRunner) Start(ctx context.Context, streamKey, rtmpInput string) error {
	r.mu.Lock()
	if _, exists := r.jobs[streamKey]; exists {
		r.mu.Unlock()
		return nil // already running - treat as idempotent (duplicate webhook)
	}
	j := &job{streamKey: streamKey}
	j.lastActivity.Store(time.Now().UnixNano())
	r.jobs[streamKey] = j
	r.mu.Unlock()

	go r.runLoop(j, rtmpInput)
	return nil
}

func (r *ffmpegRunner) runLoop(j *job, rtmpInput string) {
	outDir := filepath.Join(r.cfg.Transcode.HLSDir, j.streamKey)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		logger.WithContext(context.Background()).Error("create hls output dir failed",
			zap.String("streamKey", j.streamKey), zap.Error(err))
		r.cleanup(j)
		return
	}

	stopUpload, err := r.uploader.watch(context.Background(), j.streamKey, outDir, &j.lastActivity)
	if err != nil {
		logger.WithContext(context.Background()).Error("start segment uploader failed",
			zap.String("streamKey", j.streamKey), zap.Error(err))
	}
	j.mu.Lock()
	j.stopUploader = stopUpload
	j.mu.Unlock()

	for {
		if j.isStopping() {
			break
		}

		cmdCtx, cancel := context.WithCancel(context.Background())
		args := buildArgs(rtmpInput, outDir, r.cfg.Transcode)
		cmd := exec.CommandContext(cmdCtx, r.cfg.Transcode.FFmpegPath, args...)
		stderr, _ := cmd.StderrPipe()

		j.mu.Lock()
		j.cmd = cmd
		j.cancel = cancel
		j.mu.Unlock()

		startErr := cmd.Start()
		if startErr != nil {
			logger.WithContext(context.Background()).Error("ffmpeg start failed",
				zap.String("streamKey", j.streamKey), zap.Error(startErr))
			cancel()
			if !r.scheduleRestart(j) {
				break
			}
			continue
		}

		go streamLogs(j.streamKey, stderr)
		waitErr := cmd.Wait()
		cancel()

		if j.isStopping() {
			break
		}

		logger.WithContext(context.Background()).Warn("ffmpeg exited unexpectedly",
			zap.String("streamKey", j.streamKey), zap.Error(waitErr))
		if !r.scheduleRestart(j) {
			logger.WithContext(context.Background()).Error("ffmpeg giving up after repeated crashes",
				zap.String("streamKey", j.streamKey))
			break
		}
		time.Sleep(r.cfg.Transcode.RestartDelayDuration())
	}

	r.cleanup(j)
}

// scheduleRestart records a restart attempt and reports whether another
// attempt is allowed within the configured backoff window.
func (r *ffmpegRunner) scheduleRestart(j *job) bool {
	j.mu.Lock()
	defer j.mu.Unlock()

	now := time.Now()
	window := r.cfg.Transcode.RestartWindowDuration()
	kept := j.restartTimestamps[:0]
	for _, t := range j.restartTimestamps {
		if now.Sub(t) < window {
			kept = append(kept, t)
		}
	}
	j.restartTimestamps = kept

	if len(j.restartTimestamps) >= r.cfg.Transcode.MaxRestartsPerWindow {
		return false
	}
	j.restartTimestamps = append(j.restartTimestamps, now)
	metrics.FFmpegRestartsTotal.Inc()
	return true
}

func (r *ffmpegRunner) cleanup(j *job) {
	j.mu.Lock()
	stopUploader := j.stopUploader
	j.mu.Unlock()
	if stopUploader != nil {
		stopUploader()
	}

	r.mu.Lock()
	delete(r.jobs, j.streamKey)
	r.mu.Unlock()

	_ = os.RemoveAll(filepath.Join(r.cfg.Transcode.HLSDir, j.streamKey))
}

func (r *ffmpegRunner) Stop(ctx context.Context, streamKey string) error {
	j, ok := r.getJob(streamKey)
	if !ok {
		return nil
	}

	j.mu.Lock()
	j.stopping = true
	cmd := j.cmd
	cancel := j.cancel
	j.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		// Process may have already exited; nothing left to signal.
		return nil
	}
	go func() {
		time.Sleep(10 * time.Second)
		if cancel != nil {
			cancel() // force-kill if it hasn't exited gracefully from SIGINT
		}
	}()
	return nil
}

func (r *ffmpegRunner) Restart(ctx context.Context, streamKey string) error {
	j, ok := r.getJob(streamKey)
	if !ok {
		return fmt.Errorf("no active job for stream key %q", streamKey)
	}
	j.mu.Lock()
	cancel := j.cancel
	j.mu.Unlock()
	if cancel != nil {
		cancel() // runLoop observes the exit and respawns since stopping is still false
	}
	return nil
}

func (r *ffmpegRunner) IsRunning(streamKey string) bool {
	j, ok := r.getJob(streamKey)
	if !ok {
		return false
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.cmd != nil && !j.stopping
}

func (r *ffmpegRunner) ActiveStreamKeys() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	keys := make([]string, 0, len(r.jobs))
	for k := range r.jobs {
		keys = append(keys, k)
	}
	return keys
}

func (r *ffmpegRunner) StaleSince(threshold time.Duration) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	var stale []string
	for k, j := range r.jobs {
		last := j.lastActivity.Load()
		if last == 0 {
			continue
		}
		if now.Sub(time.Unix(0, last)) > threshold {
			stale = append(stale, k)
		}
	}
	return stale
}

func (r *ffmpegRunner) getJob(streamKey string) (*job, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.jobs[streamKey]
	return j, ok
}

func (j *job) isStopping() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.stopping
}

func streamLogs(streamKey string, stderr io.Reader) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		logger.WithContext(context.Background()).Info("ffmpeg",
			zap.String("streamKey", streamKey), zap.String("line", scanner.Text()))
	}
}
