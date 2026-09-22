package facebio

import (
	"context"
	"errors"
	"sync"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/model"
)

var errNoFramesAnalyzed = errors.New("facebio: no frame produced a usable face template")

func (a *analyzer) analyzeFramesConcurrently(ctx context.Context, frames [][]byte) (*model.FaceTemplate, error) {
	type frameResult struct {
		tmpl *model.FaceTemplate
		err  error
	}

	jobs := make(chan []byte, len(frames))
	for _, f := range frames {
		jobs <- f
	}
	close(jobs)

	results := make(chan frameResult, len(frames))
	var wg sync.WaitGroup
	for i := 0; i < numWorkers(len(frames)); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for frame := range jobs {
				tmpl, err := a.Analyze(ctx, frame)
				results <- frameResult{tmpl: tmpl, err: err}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	var best *model.FaceTemplate
	var lastErr error
	for res := range results {
		if res.err != nil {
			lastErr = res.err
			continue
		}
		if best == nil || res.tmpl.QualityScore > best.QualityScore {
			best = res.tmpl
		}
	}
	if best == nil {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, errNoFramesAnalyzed
	}
	return best, nil
}

func numWorkers(n int) int {
	const maxWorkers = 4
	if n <= 0 {
		return 1
	}
	if n < maxWorkers {
		return n
	}
	return maxWorkers
}
