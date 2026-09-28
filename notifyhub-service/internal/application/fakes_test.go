package application

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
)

type fakeChannels struct{ items map[string]*model.Channel }

func (f *fakeChannels) Create(_ context.Context, c *model.Channel) error {
	f.items[c.ID] = c
	return nil
}
func (f *fakeChannels) Get(_ context.Context, id string) (*model.Channel, error) {
	c, ok := f.items[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	cp := *c
	return &cp, nil
}
func (f *fakeChannels) List(context.Context, port.ChannelFilter) ([]*model.Channel, error) {
	return nil, nil
}
func (f *fakeChannels) Update(_ context.Context, c *model.Channel) error {
	f.items[c.ID] = c
	return nil
}
func (f *fakeChannels) Delete(_ context.Context, id string) error {
	delete(f.items, id)
	return nil
}

type fakeTemplates struct{ items map[string]*model.Template }

func (f *fakeTemplates) Create(_ context.Context, t *model.Template) error {
	f.items[t.ID] = t
	return nil
}
func (f *fakeTemplates) Get(_ context.Context, id string) (*model.Template, error) {
	t, ok := f.items[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	cp := *t
	return &cp, nil
}
func (f *fakeTemplates) ListActive(context.Context, string) ([]*model.Template, error) {
	return nil, nil
}
func (f *fakeTemplates) Update(_ context.Context, t *model.Template) error {
	f.items[t.ID] = t
	return nil
}

type fakeSources struct{ items map[string]*model.DataSource }

func (f *fakeSources) Create(_ context.Context, ds *model.DataSource) error {
	f.items[ds.ID] = ds
	return nil
}
func (f *fakeSources) Get(_ context.Context, id string) (*model.DataSource, error) {
	ds, ok := f.items[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	return ds, nil
}
func (f *fakeSources) ListActive(context.Context) ([]*model.DataSource, error) { return nil, nil }
func (f *fakeSources) Update(_ context.Context, ds *model.DataSource) error {
	f.items[ds.ID] = ds
	return nil
}

type fakeJobs struct {
	mu       sync.Mutex
	items    map[string]*model.NotifyJob
	channels *fakeChannels
	tmpls    *fakeTemplates
}

func (f *fakeJobs) Create(_ context.Context, j *model.NotifyJob) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *j
	f.items[j.ID] = &cp
	return nil
}
func (f *fakeJobs) Get(_ context.Context, id string) (*model.NotifyJob, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	j, ok := f.items[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	cp := *j
	if c, ok := f.channels.items[j.ChannelID]; ok {
		cp.Channel = c
	}
	if t, ok := f.tmpls.items[j.TemplateID]; ok {
		cp.Template = t
	}
	return &cp, nil
}
func (f *fakeJobs) ListActive(context.Context) ([]*model.NotifyJob, error) { return nil, nil }
func (f *fakeJobs) List(context.Context, port.JobFilter) ([]*model.NotifyJob, int64, error) {
	return nil, 0, nil
}
func (f *fakeJobs) Update(_ context.Context, j *model.NotifyJob) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *j
	f.items[j.ID] = &cp
	return nil
}
func (f *fakeJobs) UpdateStatus(_ context.Context, id string, s model.JobStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	j, ok := f.items[id]
	if !ok {
		return port.ErrNotFound
	}
	j.Status = s
	return nil
}
func (f *fakeJobs) RecordRun(_ context.Context, id string, _ *time.Time, failed bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	j := f.items[id]
	now := time.Now()
	j.LastRunAt = &now
	j.RunCount++
	if failed {
		j.FailCount++
	}
	return nil
}
func (f *fakeJobs) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.items[id]; !ok {
		return port.ErrNotFound
	}
	delete(f.items, id)
	return nil
}
func (f *fakeJobs) CountByChannel(_ context.Context, channelID string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var n int64
	for _, j := range f.items {
		if j.ChannelID == channelID {
			n++
		}
	}
	return n, nil
}

type fakeHistory struct {
	mu      sync.Mutex
	entries map[string]*model.NotifyHistory
}

func (f *fakeHistory) Create(_ context.Context, h *model.NotifyHistory) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries[h.ID] = h
	return nil
}
func (f *fakeHistory) UpdateStatus(_ context.Context, id string, s model.NotifyStatus, retries int, errMsg string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	h := f.entries[id]
	h.Status, h.RetryCount, h.Error = s, retries, errMsg
	return nil
}
func (f *fakeHistory) List(context.Context, port.HistoryFilter) ([]*model.NotifyHistory, int64, error) {
	return nil, 0, nil
}

// fakeSender fails the first failTimes sends for each recipient.
type fakeSender struct {
	mu        sync.Mutex
	failTimes int
	calls     map[string]int
	sent      []port.Message
}

func (f *fakeSender) Type() model.ChannelType { return model.ChannelSMS }
func (f *fakeSender) Send(_ context.Context, m port.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls[m.Recipient]++
	if f.calls[m.Recipient] <= f.failTimes {
		return errors.New("provider down")
	}
	f.sent = append(f.sent, m)
	return nil
}

type fakeRegistry struct{ s port.Sender }

func (r fakeRegistry) Get(t model.ChannelType) (port.Sender, error) {
	if r.s == nil || r.s.Type() != t {
		return nil, errors.New("not configured")
	}
	return r.s, nil
}

type fakeFetcher struct {
	data interface{}
	err  error
}

func (f fakeFetcher) Fetch(context.Context, *model.DataSource) (interface{}, error) {
	return f.data, f.err
}

// fakeRenderer "renders" by returning the raw subject/body; Compile fails
// when the body is "{{bad".
type fakeRenderer struct{ evicted []string }

func (r *fakeRenderer) Compile(t *model.Template) error {
	if t.Body == "{{bad" {
		return errors.New("parse error")
	}
	return nil
}
func (r *fakeRenderer) Render(t *model.Template, data map[string]interface{}) (string, string, error) {
	if err := r.Compile(t); err != nil {
		return "", "", err
	}
	return t.Subject, t.Body, nil
}
func (r *fakeRenderer) Evict(id string) { r.evicted = append(r.evicted, id) }

type fakeScheduler struct {
	mu         sync.Mutex
	registered map[string]bool
	failWith   error
}

func (s *fakeScheduler) Register(j *model.NotifyJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failWith != nil {
		return s.failWith
	}
	s.registered[j.ID] = true
	return nil
}
func (s *fakeScheduler) Unregister(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.registered, id)
}
func (s *fakeScheduler) Scheduled() []string          { return nil }
func (s *fakeScheduler) NextRun(string) *time.Time    { return nil }
func (s *fakeScheduler) isRegistered(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.registered[id]
}

type fakeQueue struct{ tasks []port.Task }

func (q *fakeQueue) Enqueue(t port.Task) error {
	q.tasks = append(q.tasks, t)
	return nil
}

type nopMetrics struct{}

func (nopMetrics) JobExecuted(model.ScheduleType, string)           {}
func (nopMetrics) NotificationSent(model.ChannelType, string)       {}
func (nopMetrics) SendRetried(model.ChannelType)                    {}
func (nopMetrics) FetchObserved(string, string, time.Duration)      {}

type env struct {
	channels  *fakeChannels
	templates *fakeTemplates
	sources   *fakeSources
	jobs      *fakeJobs
	history   *fakeHistory
	sender    *fakeSender
	renderer  *fakeRenderer
	scheduler *fakeScheduler
	queue     *fakeQueue
}

func newEnv() *env {
	ch := &fakeChannels{items: map[string]*model.Channel{
		"ch-sms": {ID: "ch-sms", Name: "sms", Type: model.ChannelSMS, IsActive: true},
	}}
	tp := &fakeTemplates{items: map[string]*model.Template{
		"tpl-sms": {ID: "tpl-sms", Name: "hello", Channel: model.ChannelSMS, Body: "Hi {{.name}}", IsActive: true},
	}}
	return &env{
		channels:  ch,
		templates: tp,
		sources:   &fakeSources{items: map[string]*model.DataSource{}},
		jobs:      &fakeJobs{items: map[string]*model.NotifyJob{}, channels: ch, tmpls: tp},
		history:   &fakeHistory{entries: map[string]*model.NotifyHistory{}},
		sender:    &fakeSender{calls: map[string]int{}},
		renderer:  &fakeRenderer{},
		scheduler: &fakeScheduler{registered: map[string]bool{}},
		queue:     &fakeQueue{},
	}
}
