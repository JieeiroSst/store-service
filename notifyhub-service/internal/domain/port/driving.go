package port

import (
	"context"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
)

type ChannelUsecase interface {
	Create(ctx context.Context, c *model.Channel) (*model.Channel, error)
	Get(ctx context.Context, id string) (*model.Channel, error)
	List(ctx context.Context, f ChannelFilter) ([]*model.Channel, error)
	Update(ctx context.Context, id string, apply func(*model.Channel) error) (*model.Channel, error)
	Delete(ctx context.Context, id string) error
}

type DataSourceUsecase interface {
	Create(ctx context.Context, ds *model.DataSource) (*model.DataSource, error)
	Get(ctx context.Context, id string) (*model.DataSource, error)
	List(ctx context.Context) ([]*model.DataSource, error)
	Update(ctx context.Context, id string, apply func(*model.DataSource) error) (*model.DataSource, error)
	Delete(ctx context.Context, id string) error
}

type TemplateUsecase interface {
	Create(ctx context.Context, t *model.Template) (*model.Template, error)
	Get(ctx context.Context, id string) (*model.Template, error)
	List(ctx context.Context, channel string) ([]*model.Template, error)
	Update(ctx context.Context, id string, apply func(*model.Template) error) (*model.Template, error)
	Delete(ctx context.Context, id string) error
	WarmUp(ctx context.Context) (int, error)
}

type JobUsecase interface {
	Create(ctx context.Context, j *model.NotifyJob) (*model.NotifyJob, error)
	Get(ctx context.Context, id string) (*model.NotifyJob, error)
	List(ctx context.Context, f JobFilter) ([]*model.NotifyJob, int64, error)
	Update(ctx context.Context, id string, apply func(*model.NotifyJob) error) (*model.NotifyJob, error)
	Pause(ctx context.Context, id string) error
	Resume(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	Trigger(ctx context.Context, id string) error
	RestoreSchedules(ctx context.Context) error
	ScheduledJobs() []string
}

type HistoryUsecase interface {
	List(ctx context.Context, f HistoryFilter) ([]*model.NotifyHistory, int64, error)
}

type DispatchUsecase interface {
	Execute(ctx context.Context, t Task) error
}
