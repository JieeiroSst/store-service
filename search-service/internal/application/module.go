package application

import (
	"context"

	"github.com/JIeeiroSst/search-service/config"
	"github.com/JIeeiroSst/search-service/internal/domain"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) (domain.SensitivePolicy, error) {
		return domain.NewSensitivePolicy(cfg.Elasticsearch.SensitivePattern)
	}),
	fx.Provide(NewDocumentService),
	fx.Provide(NewSchemaService),
	fx.Invoke(applySchemaOnStart),
)

func applySchemaOnStart(lc fx.Lifecycle, cfg *config.Config, schema *SchemaService) {
	if !cfg.Elasticsearch.Schema.ShouldApplyOnStartup() {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				schema.ApplyWithRetry(ctx)
			}()
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			select {
			case <-done:
			case <-stopCtx.Done():
			}
			return nil
		},
	})
}
