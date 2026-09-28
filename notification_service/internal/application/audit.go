package application

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
)

const (
	auditDefaultLimit = 50
	auditMaxLimit     = 500
	auditPurgeChunk   = 5000
)

type trail struct {
	content    model.AuditContent
	base       model.AuditDelivery
	deliveries []model.AuditDelivery
}

func newTrail(sourceType string, sourceID, batchID uint, attempt int, channel, requestedBy string) *trail {
	if strings.TrimSpace(requestedBy) == "" {
		requestedBy = model.DefaultRequester
	}
	return &trail{
		content: model.AuditContent{Channel: channel},
		base: model.AuditDelivery{
			SourceType: sourceType, SourceID: sourceID, BatchID: batchID,
			Attempt: attempt, Channel: channel, RequestedBy: requestedBy,
		},
	}
}

func (t *trail) setContent(title, body string, data map[string]string, templateType string) {
	t.content.Title, t.content.Body, t.content.Data, t.content.TemplateType = title, body, data, templateType
}

func (t *trail) add(d model.AuditDelivery) {
	full := t.base
	full.UserID, full.Recipient, full.DeviceID, full.DeviceType = d.UserID, d.Recipient, d.DeviceID, d.DeviceType
	full.TokenHash, full.Status, full.MessageID, full.Error = d.TokenHash, d.Status, d.MessageID, d.Error
	t.deliveries = append(t.deliveries, full)
}

func (t *trail) failAll(err error) {
	for i := range t.deliveries {
		if t.deliveries[i].Status == "" {
			t.deliveries[i].Status, t.deliveries[i].Error = model.AuditFailed, err.Error()
		}
	}
}

type auditor struct {
	repo  port.AuditRepository
	clock port.Clock
}

func (a *auditor) record(ctx context.Context, t *trail) {
	if a == nil || a.repo == nil || len(t.deliveries) == 0 {
		return
	}
	ctx = context.WithoutCancel(ctx)
	now := a.clock.Now()
	t.content.CreatedAt = now
	t.content.ComputeHash()
	if err := a.repo.SaveContent(ctx, &t.content); err != nil {
		log.Printf("audit: save content: %v", err)
		return
	}
	for i := range t.deliveries {
		t.deliveries[i].ContentID = t.content.ID
		t.deliveries[i].CreatedAt = now
		if len(t.deliveries[i].Error) > 1000 {
			t.deliveries[i].Error = t.deliveries[i].Error[:1000]
		}
	}
	if err := a.repo.SaveDeliveries(ctx, t.deliveries); err != nil {
		log.Printf("audit: save %d deliveries for %s %d: %v", len(t.deliveries), t.base.SourceType, t.base.SourceID, err)
	}
}

type auditService struct {
	repo      port.AuditRepository
	contacts  port.ContactRepository
	clock     port.Clock
	retention time.Duration
}

func NewAuditService(repo port.AuditRepository, contacts port.ContactRepository, clock port.Clock, cfg *config.Config) port.AuditUsecase {
	s := &auditService{repo: repo, contacts: contacts, clock: clock}
	if cfg.Audit.RetentionDays > 0 {
		s.retention = time.Duration(cfg.Audit.RetentionDays) * 24 * time.Hour
	}
	return s
}

func NewAuditor(repo port.AuditRepository, clock port.Clock) *auditor {
	return &auditor{repo: repo, clock: clock}
}

func (s *auditService) ListDeliveries(ctx context.Context, q port.AuditQuery) (*port.AuditPage, error) {
	f := port.AuditFilter{
		Channel: q.Channel, Status: q.Status, SourceType: q.SourceType, SourceID: q.SourceID,
		RequestedBy: q.RequestedBy, From: q.From, To: q.To, BeforeID: q.BeforeID, Limit: q.Limit,
	}
	if f.Limit <= 0 {
		f.Limit = auditDefaultLimit
	}
	f.Limit = min(f.Limit, auditMaxLimit)

	identified := q.UserID != 0 || strings.TrimSpace(q.Email) != "" || strings.TrimSpace(q.Phone) != ""
	if q.UserID != 0 {
		f.UserIDs = append(f.UserIDs, q.UserID)
	}
	if email := model.NormalizeEmail(q.Email); email != "" {
		f.Recipients = append(f.Recipients, email)
		if id, err := s.contacts.UserIDByEmail(ctx, email); err == nil {
			f.UserIDs = append(f.UserIDs, id)
		}
	}
	if strings.TrimSpace(q.Phone) != "" {
		phone, ok := model.NormalizePhone(q.Phone)
		if !ok {
			return nil, invalid("phone %q is not valid", q.Phone)
		}
		f.Recipients = append(f.Recipients, phone)
		if id, err := s.contacts.UserIDByPhone(ctx, phone); err == nil {
			f.UserIDs = append(f.UserIDs, id)
		}
	}
	for _, id := range f.UserIDs {
		if c, err := s.contacts.Get(ctx, id); err == nil && c.Email != nil {
			f.Recipients = append(f.Recipients, *c.Email)
		}
	}
	if identified && len(f.UserIDs) == 0 && len(f.Recipients) == 0 {
		return &port.AuditPage{Items: []model.AuditDelivery{}, Contents: map[uint]model.AuditContent{}}, nil
	}

	items, err := s.repo.ListDeliveries(ctx, f)
	if err != nil {
		return nil, err
	}
	page := &port.AuditPage{Items: items, Contents: map[uint]model.AuditContent{}}
	if items == nil {
		page.Items = []model.AuditDelivery{}
	}
	var ids []uint
	seen := map[uint]bool{}
	for _, d := range items {
		if !seen[d.ContentID] {
			seen[d.ContentID] = true
			ids = append(ids, d.ContentID)
		}
	}
	contents, err := s.repo.GetContents(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, c := range contents {
		page.Contents[c.ID] = c
	}
	if len(items) == f.Limit {
		page.NextBeforeID = items[len(items)-1].ID
	}
	return page, nil
}

func (s *auditService) GetContent(ctx context.Context, id uint) (*model.AuditContent, error) {
	contents, err := s.repo.GetContents(ctx, []uint{id})
	if err != nil {
		return nil, err
	}
	if len(contents) == 0 {
		return nil, common.ErrNotFound
	}
	return &contents[0], nil
}

func (s *auditService) PurgeExpired(ctx context.Context) (int64, error) {
	if s.retention <= 0 {
		return 0, nil
	}
	before := s.clock.Now().Add(-s.retention)
	var total int64
	for {
		n, err := s.repo.DeleteBefore(ctx, before, auditPurgeChunk)
		total += n
		if err != nil || n < auditPurgeChunk {
			return total, err
		}
	}
}
