package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"go.uber.org/zap"
)

var ctx = context.Background()

func (e *env) jobService() port.JobUsecase {
	return NewJobService(e.jobs, e.channels, e.templates, e.sources, e.scheduler, e.queue, zap.NewNop())
}

func (e *env) dispatcher(retryMax int) port.DispatchUsecase {
	return NewDispatcher(e.jobs, e.history, fakeRegistry{s: e.sender}, fakeFetcher{}, e.renderer,
		e.scheduler, nopMetrics{}, Settings{RetryMax: retryMax, RetryDelay: time.Millisecond}, zap.NewNop())
}

func intervalJob() *model.NotifyJob {
	return &model.NotifyJob{
		Name:         "reminder",
		ChannelID:    "ch-sms",
		TemplateID:   "tpl-sms",
		ScheduleType: model.ScheduleInterval,
		IntervalSec:  60,
		Recipients:   model.StringSlice{"+84900000001", "+84900000002"},
	}
}

func jsonPatch(body string) func(v any) error {
	return func(v any) error { return json.Unmarshal([]byte(body), v) }
}

func TestCreateJobRegistersSchedule(t *testing.T) {
	e := newEnv()
	j, err := e.jobService().Create(ctx, intervalJob())
	if err != nil {
		t.Fatal(err)
	}
	if j.Status != model.JobStatusActive || !e.scheduler.isRegistered(j.ID) {
		t.Fatalf("job not active/registered: status=%s", j.Status)
	}
}

func TestCreateJobRejectsUnknownReferences(t *testing.T) {
	e := newEnv()
	j := intervalJob()
	j.TemplateID = "missing"
	if _, err := e.jobService().Create(ctx, j); !errors.Is(err, port.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestCreateJobRollsBackWhenScheduleInvalid(t *testing.T) {
	e := newEnv()
	e.scheduler.failWith = errors.New("bad cron")
	if _, err := e.jobService().Create(ctx, intervalJob()); !errors.Is(err, port.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
	if len(e.jobs.items) != 0 {
		t.Fatal("job persisted although it cannot be scheduled")
	}
}

func TestUpdateJobKeepsBookkeepingFields(t *testing.T) {
	e := newEnv()
	svc := e.jobService()
	j, _ := svc.Create(ctx, intervalJob())
	_ = e.jobs.UpdateStatus(ctx, j.ID, model.JobStatusPaused)
	e.jobs.items[j.ID].RunCount, e.jobs.items[j.ID].FailCount = 7, 2

	updated, err := svc.Update(ctx, j.ID, func(nj *model.NotifyJob) error {
		return jsonPatch(`{"name":"renamed","status":"active","run_count":0}`)(nj)
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "renamed" || len(updated.Recipients) != 2 {
		t.Fatalf("patch not applied onto existing job: %+v", updated)
	}
	if updated.Status != model.JobStatusPaused || updated.RunCount != 7 || updated.FailCount != 2 {
		t.Fatalf("bookkeeping overwritten: status=%s run=%d fail=%d", updated.Status, updated.RunCount, updated.FailCount)
	}
}

func TestDeleteChannelInUseConflicts(t *testing.T) {
	e := newEnv()
	if _, err := e.jobService().Create(ctx, intervalJob()); err != nil {
		t.Fatal(err)
	}
	err := NewChannelService(e.channels, e.jobs).Delete(ctx, "ch-sms")
	if !errors.Is(err, port.ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}
}

func TestTemplateCreateRejectsInvalidSyntax(t *testing.T) {
	e := newEnv()
	svc := NewTemplateService(e.templates, e.renderer, zap.NewNop())
	_, err := svc.Create(ctx, &model.Template{Name: "x", Channel: model.ChannelSMS, Body: "{{bad"})
	if !errors.Is(err, port.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestTemplateUpdateKeepsActiveFlag(t *testing.T) {
	e := newEnv()
	svc := NewTemplateService(e.templates, e.renderer, zap.NewNop())
	updated, err := svc.Update(ctx, "tpl-sms", func(t *model.Template) error {
		return jsonPatch(`{"body":"Hello {{.name}}"}`)(t)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !updated.IsActive || updated.Name != "hello" || updated.Body != "Hello {{.name}}" {
		t.Fatalf("unexpected template after update: %+v", updated)
	}
}

func TestDispatchSendsToAllRecipientsWithRetry(t *testing.T) {
	e := newEnv()
	e.sender.failTimes = 1 // every recipient fails once, then succeeds
	j, _ := e.jobService().Create(ctx, intervalJob())

	if err := e.dispatcher(2).Execute(ctx, port.Task{JobID: j.ID}); err != nil {
		t.Fatal(err)
	}
	if len(e.sender.sent) != 2 {
		t.Fatalf("want 2 sent, got %d", len(e.sender.sent))
	}
	for _, h := range e.history.entries {
		if h.Status != model.NotifyStatusSent || h.RetryCount != 1 {
			t.Fatalf("history not updated: %+v", h)
		}
	}
	if got := e.jobs.items[j.ID]; got.RunCount != 1 || got.FailCount != 0 {
		t.Fatalf("run not recorded: run=%d fail=%d", got.RunCount, got.FailCount)
	}
}

func TestDispatchRecordsFailure(t *testing.T) {
	e := newEnv()
	e.sender.failTimes = 10
	j, _ := e.jobService().Create(ctx, intervalJob())

	if err := e.dispatcher(1).Execute(ctx, port.Task{JobID: j.ID}); err == nil {
		t.Fatal("want error when all sends fail")
	}
	if got := e.jobs.items[j.ID]; got.RunCount != 1 || got.FailCount != 1 {
		t.Fatalf("failure not recorded: run=%d fail=%d", got.RunCount, got.FailCount)
	}
}

func TestDispatchCompletesAfterMaxRuns(t *testing.T) {
	e := newEnv()
	job := intervalJob()
	job.MaxRuns = 1
	j, _ := e.jobService().Create(ctx, job)
	d := e.dispatcher(0)

	if err := d.Execute(ctx, port.Task{JobID: j.ID}); err != nil {
		t.Fatal(err)
	}
	if got := e.jobs.items[j.ID]; got.Status != model.JobStatusCompleted || e.scheduler.isRegistered(j.ID) {
		t.Fatalf("job not completed after max_runs: status=%s", got.Status)
	}
	// A late scheduled firing must not send again.
	sent := len(e.sender.sent)
	_ = d.Execute(ctx, port.Task{JobID: j.ID})
	if len(e.sender.sent) != sent {
		t.Fatal("completed job was executed again")
	}
}

func TestDispatchSkipsPausedJobUnlessManual(t *testing.T) {
	e := newEnv()
	svc := e.jobService()
	j, _ := svc.Create(ctx, intervalJob())
	_ = svc.Pause(ctx, j.ID)
	d := e.dispatcher(0)

	_ = d.Execute(ctx, port.Task{JobID: j.ID})
	if len(e.sender.sent) != 0 {
		t.Fatal("paused job ran on schedule")
	}
	if err := d.Execute(ctx, port.Task{JobID: j.ID, Manual: true}); err != nil {
		t.Fatal(err)
	}
	if len(e.sender.sent) != 2 {
		t.Fatal("manual trigger of paused job did not run")
	}
}

func TestDispatchOnceJobCompletes(t *testing.T) {
	e := newEnv()
	job := intervalJob()
	job.ScheduleType = model.ScheduleOnce
	runAt := time.Now().Add(time.Hour)
	job.RunAt = &runAt
	j, _ := e.jobService().Create(ctx, job)

	if err := e.dispatcher(0).Execute(ctx, port.Task{JobID: j.ID}); err != nil {
		t.Fatal(err)
	}
	if got := e.jobs.items[j.ID]; got.Status != model.JobStatusCompleted {
		t.Fatalf("once job status = %s, want completed", got.Status)
	}
}

func TestDispatchFailsGracefullyWithoutChannel(t *testing.T) {
	e := newEnv()
	j, _ := e.jobService().Create(ctx, intervalJob())
	delete(e.channels.items, "ch-sms")

	if err := e.dispatcher(0).Execute(ctx, port.Task{JobID: j.ID}); err == nil {
		t.Fatal("want error for job whose channel is gone")
	}
	if got := e.jobs.items[j.ID]; got.FailCount != 1 {
		t.Fatalf("fail_count = %d, want 1", got.FailCount)
	}
}

func TestDispatchMergesFetchedData(t *testing.T) {
	d := &dispatcher{fetcher: fakeFetcher{data: map[string]interface{}{"name": "fetched"}}, metrics: nopMetrics{}}
	job := &model.NotifyJob{
		StaticPayload: model.JSONMap{"name": "static", "env": "prod"},
		DataSource:    &model.DataSource{Name: "api", IsActive: true},
	}
	data := d.templateData(ctx, job, zap.NewNop())
	if data["name"] != "fetched" || data["env"] != "prod" {
		t.Fatalf("unexpected data: %v", data)
	}
}
