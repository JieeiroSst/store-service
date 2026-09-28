package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
)

type ChannelFilter struct {
	Type   string
	Active *bool
}

type JobFilter struct {
	Status   string
	Page     int
	PageSize int
}

type HistoryFilter struct {
	JobID    string
	Status   string
	Page     int
	PageSize int
}

type ChannelRepository interface {
	Create(ctx context.Context, c *model.Channel) error
	Get(ctx context.Context, id string) (*model.Channel, error)
	List(ctx context.Context, f ChannelFilter) ([]*model.Channel, error)
	Update(ctx context.Context, c *model.Channel) error
	Delete(ctx context.Context, id string) error
}

type DataSourceRepository interface {
	Create(ctx context.Context, ds *model.DataSource) error
	Get(ctx context.Context, id string) (*model.DataSource, error)
	ListActive(ctx context.Context) ([]*model.DataSource, error)
	Update(ctx context.Context, ds *model.DataSource) error
}

type TemplateRepository interface {
	Create(ctx context.Context, t *model.Template) error
	Get(ctx context.Context, id string) (*model.Template, error)
	ListActive(ctx context.Context, channel string) ([]*model.Template, error)
	Update(ctx context.Context, t *model.Template) error
}

type JobRepository interface {
	Create(ctx context.Context, j *model.NotifyJob) error
	Get(ctx context.Context, id string) (*model.NotifyJob, error)
	ListActive(ctx context.Context) ([]*model.NotifyJob, error)
	List(ctx context.Context, f JobFilter) ([]*model.NotifyJob, int64, error)
	Update(ctx context.Context, j *model.NotifyJob) error
	UpdateStatus(ctx context.Context, id string, status model.JobStatus) error
	RecordRun(ctx context.Context, id string, next *time.Time, failed bool) error
	Delete(ctx context.Context, id string) error
	CountByChannel(ctx context.Context, channelID string) (int64, error)
}

type HistoryRepository interface {
	Create(ctx context.Context, h *model.NotifyHistory) error
	UpdateStatus(ctx context.Context, id string, status model.NotifyStatus, retries int, errMsg string) error
	List(ctx context.Context, f HistoryFilter) ([]*model.NotifyHistory, int64, error)
}

type Message struct {
	Recipient string
	Subject   string
	Body      string
	Data      map[string]string
}

type Sender interface {
	Type() model.ChannelType
	Send(ctx context.Context, msg Message) error
}

type SenderRegistry interface {
	Get(t model.ChannelType) (Sender, error)
}

type DataFetcher interface {
	Fetch(ctx context.Context, ds *model.DataSource) (interface{}, error)
}

type TemplateRenderer interface {
	Compile(t *model.Template) error
	Render(t *model.Template, data map[string]interface{}) (subject, body string, err error)
	Evict(templateID string)
}

type JobScheduler interface {
	Register(j *model.NotifyJob) error
	Unregister(jobID string)
	Scheduled() []string
	NextRun(jobID string) *time.Time
}

type Task struct {
	JobID  string
	Manual bool
}

type TaskQueue interface {
	Enqueue(t Task) error
}

type Metrics interface {
	JobExecuted(scheduleType model.ScheduleType, status string)
	NotificationSent(channel model.ChannelType, status string)
	SendRetried(channel model.ChannelType)
	FetchObserved(source, status string, d time.Duration)
}
