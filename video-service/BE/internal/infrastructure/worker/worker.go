package worker

import (
	"context"
	"log"

	"github.com/JIeeiroSst/video-service/config"
	"github.com/JIeeiroSst/video-service/internal/domain/port"
	"go.uber.org/fx"
)

func New(lc fx.Lifecycle, queue port.JobQueue, uc port.TranscodeUsecase, cfg *config.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				err := queue.Consume(ctx, func(ctx context.Context, id string) error {
					jctx, cancel := context.WithTimeout(ctx, cfg.Worker.JobTimeout)
					defer cancel()
					log.Printf("transcode %s: start", id)
					err := uc.Process(jctx, id)
					if err != nil {
						log.Printf("transcode %s: %v", id, err)
					} else {
						log.Printf("transcode %s: done", id)
					}
					return err
				})
				if err != nil {
					log.Printf("job consumer stopped: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			cancel()
			select {
			case <-done:
			case <-ctx.Done():
			}
			return nil
		},
	})
}
