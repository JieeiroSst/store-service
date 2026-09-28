package application

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
)

var t0 = time.Date(2026, 9, 28, 8, 0, 0, 0, time.UTC)

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

type store struct {
	mu            sync.Mutex
	devices       map[uint]*model.UserDevice
	notifications map[uint]model.Notification
	campaigns     map[uint]*model.Campaign
	recipients    []model.CampaignRecipient
	batches       map[uint]*model.CampaignBatch
	nextID        uint
}

func newStore() *store {
	return &store{
		devices: map[uint]*model.UserDevice{}, notifications: map[uint]model.Notification{},
		campaigns: map[uint]*model.Campaign{}, batches: map[uint]*model.CampaignBatch{},
	}
}

func (s *store) id() uint { s.nextID++; return s.nextID }

func (s *store) addDevice(userID uint, token string, active bool) *model.UserDevice {
	d := &model.UserDevice{ID: s.id(), UserID: userID, DeviceToken: token, IsActive: active}
	s.devices[d.ID] = d
	return d
}

type notificationRepo struct{ s *store }

func (r notificationRepo) Create(_ context.Context, n *model.Notification) error {
	n.ID = r.s.id()
	r.s.notifications[n.ID] = *n
	return nil
}
func (r notificationRepo) GetByID(_ context.Context, id uint) (*model.Notification, error) {
	n, ok := r.s.notifications[id]
	if !ok {
		return nil, common.ErrNotFound
	}
	return &n, nil
}
func (r notificationRepo) Update(_ context.Context, n *model.Notification) error {
	r.s.notifications[n.ID] = *n
	return nil
}
func (r notificationRepo) Delete(context.Context, uint) error { return nil }
func (r notificationRepo) List(context.Context) ([]model.Notification, error) {
	return nil, nil
}

type deviceRepo struct{ s *store }

func (r deviceRepo) Create(context.Context, *model.UserDevice) error { return nil }
func (r deviceRepo) GetByID(context.Context, uint) (*model.UserDevice, error) {
	return nil, common.ErrNotFound
}
func (r deviceRepo) ListActiveByUserID(_ context.Context, userID uint) ([]model.UserDevice, error) {
	var out []model.UserDevice
	for _, d := range r.s.devices {
		if d.UserID == userID && d.IsActive {
			out = append(out, *d)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (r deviceRepo) Update(context.Context, *model.UserDevice) error  { return nil }
func (r deviceRepo) Delete(context.Context, uint) error               { return nil }
func (r deviceRepo) List(context.Context) ([]model.UserDevice, error) { return nil, nil }
func (r deviceRepo) Register(_ context.Context, d *model.UserDevice, opts port.RegisterOptions) (*model.UserDevice, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var saved *model.UserDevice
	for _, x := range r.s.devices {
		if x.TokenHash != nil && *x.TokenHash == *d.TokenHash {
			x.UserID, x.DeviceID, x.DeviceType, x.IsActive, x.LastUsedAt = d.UserID, d.DeviceID, d.DeviceType, true, d.LastUsedAt
			saved = x
		}
	}
	if saved == nil {
		cp := *d
		cp.ID = r.s.id()
		r.s.devices[cp.ID] = &cp
		saved = &cp
	}
	for _, x := range r.s.devices {
		if x.ID == saved.ID {
			continue
		}
		if (opts.ReplaceSameInstall && x.DeviceID == saved.DeviceID) || (opts.SingleDevicePerUser && x.UserID == saved.UserID) {
			x.IsActive = false
		}
	}
	cp := *saved
	return &cp, nil
}

func (r deviceRepo) DeactivateByTokenHash(_ context.Context, hash string) (int64, error) {
	var n int64
	for _, x := range r.s.devices {
		if x.TokenHash != nil && *x.TokenHash == hash && x.IsActive {
			x.IsActive = false
			n++
		}
	}
	return n, nil
}

func (r deviceRepo) DeactivateByUserID(_ context.Context, userID uint) (int64, error) {
	var n int64
	for _, x := range r.s.devices {
		if x.UserID == userID && x.IsActive {
			x.IsActive = false
			n++
		}
	}
	return n, nil
}

func (r deviceRepo) DeactivateStale(_ context.Context, before time.Time) (int64, error) {
	var n int64
	for _, x := range r.s.devices {
		if x.IsActive && x.LastUsedAt.Before(before) {
			x.IsActive = false
			n++
		}
	}
	return n, nil
}

func (r deviceRepo) DeactivateByDeviceID(_ context.Context, userID uint, deviceID string) (int64, error) {
	var n int64
	for _, x := range r.s.devices {
		if x.UserID == userID && x.DeviceID == deviceID && x.IsActive {
			x.IsActive = false
			n++
		}
	}
	return n, nil
}

func (r deviceRepo) ListByUserID(_ context.Context, userID uint) ([]model.UserDevice, error) {
	var out []model.UserDevice
	for _, x := range r.s.devices {
		if x.UserID == userID {
			out = append(out, *x)
		}
	}
	return out, nil
}

func (r deviceRepo) DeactivateByIDs(_ context.Context, ids []uint) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, id := range ids {
		r.s.devices[id].IsActive = false
	}
	return nil
}

type contactRepo struct {
	byUser map[uint]*model.UserContact
}

func newContactRepo() *contactRepo { return &contactRepo{byUser: map[uint]*model.UserContact{}} }

func (r *contactRepo) Upsert(_ context.Context, c *model.UserContact) error {
	for id, x := range r.byUser {
		if id == c.UserID {
			continue
		}
		if c.Email != nil && x.Email != nil && *x.Email == *c.Email {
			x.Email = nil
		}
		if c.Phone != nil && x.Phone != nil && *x.Phone == *c.Phone {
			x.Phone = nil
		}
	}
	cur, ok := r.byUser[c.UserID]
	if !ok {
		cp := *c
		r.byUser[c.UserID] = &cp
		return nil
	}
	if c.Email != nil {
		cur.Email = c.Email
	}
	if c.Phone != nil {
		cur.Phone = c.Phone
	}
	return nil
}

func (r *contactRepo) Get(_ context.Context, id uint) (*model.UserContact, error) {
	c, ok := r.byUser[id]
	if !ok {
		return nil, common.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *contactRepo) UserIDByEmail(_ context.Context, email string) (uint, error) {
	for id, c := range r.byUser {
		if c.Email != nil && *c.Email == email {
			return id, nil
		}
	}
	return 0, common.ErrUnknownContact
}

func (r *contactRepo) UserIDByPhone(_ context.Context, phone string) (uint, error) {
	for id, c := range r.byUser {
		if c.Phone != nil && *c.Phone == phone {
			return id, nil
		}
	}
	return 0, common.ErrUnknownContact
}

type auditRepo struct {
	mu         sync.Mutex
	contents   []model.AuditContent
	deliveries []model.AuditDelivery
}

func (r *auditRepo) SaveContent(_ context.Context, c *model.AuditContent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, x := range r.contents {
		if x.Hash == c.Hash {
			c.ID = x.ID
			return nil
		}
	}
	c.ID = uint(len(r.contents) + 1)
	r.contents = append(r.contents, *c)
	return nil
}

func (r *auditRepo) SaveDeliveries(_ context.Context, ds []model.AuditDelivery) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, d := range ds {
		d.ID = uint(len(r.deliveries) + 1)
		r.deliveries = append(r.deliveries, d)
	}
	return nil
}

func (r *auditRepo) ListDeliveries(_ context.Context, f port.AuditFilter) ([]model.AuditDelivery, error) {
	var out []model.AuditDelivery
	for i := len(r.deliveries) - 1; i >= 0; i-- {
		d := r.deliveries[i]
		match := len(f.UserIDs) == 0 && len(f.Recipients) == 0
		for _, id := range f.UserIDs {
			match = match || d.UserID == id
		}
		for _, r := range f.Recipients {
			match = match || d.Recipient == r
		}
		if !match || (f.Channel != "" && d.Channel != f.Channel) || (f.Status != "" && d.Status != f.Status) ||
			(f.SourceType != "" && d.SourceType != f.SourceType) || (f.BeforeID != 0 && d.ID >= f.BeforeID) {
			continue
		}
		if len(out) < f.Limit {
			out = append(out, d)
		}
	}
	return out, nil
}

func (r *auditRepo) GetContents(_ context.Context, ids []uint) ([]model.AuditContent, error) {
	var out []model.AuditContent
	for _, id := range ids {
		for _, c := range r.contents {
			if c.ID == id {
				out = append(out, c)
			}
		}
	}
	return out, nil
}

func (r *auditRepo) DeleteBefore(_ context.Context, before time.Time, limit int) (int64, error) {
	var kept []model.AuditDelivery
	var n int64
	for _, d := range r.deliveries {
		if d.CreatedAt.Before(before) && int(n) < limit {
			n++
			continue
		}
		kept = append(kept, d)
	}
	r.deliveries = kept
	return n, nil
}

type fakeValidator struct{ invalid map[string]bool }

func (v fakeValidator) ValidateToken(_ context.Context, token string) error {
	if v.invalid[token] {
		return fmt.Errorf("%w: unregistered", common.ErrInvalidToken)
	}
	return nil
}

type campaignRepo struct{ s *store }

func (r campaignRepo) Create(_ context.Context, c *model.Campaign) error {
	c.ID = r.s.id()
	cp := *c
	r.s.campaigns[c.ID] = &cp
	return nil
}
func (r campaignRepo) AddRecipients(_ context.Context, rs []model.CampaignRecipient) (int64, error) {
	seen := map[string]bool{}
	for _, x := range r.s.recipients {
		seen[fmt.Sprint(x.CampaignID, x.RecipientKey)] = true
	}
	var n int64
	for _, x := range rs {
		k := fmt.Sprint(x.CampaignID, x.RecipientKey)
		if seen[k] {
			continue
		}
		seen[k] = true
		x.ID = r.s.id()
		r.s.recipients = append(r.s.recipients, x)
		n++
	}
	return n, nil
}
func (r campaignRepo) Activate(_ context.Context, id uint, recipients int) error {
	c := r.s.campaigns[id]
	if c.Status == model.CampaignDraft {
		c.Status, c.Recipients = model.CampaignPending, recipients
	}
	return nil
}
func (r campaignRepo) GetByID(_ context.Context, id uint) (*model.Campaign, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	c, ok := r.s.campaigns[id]
	if !ok {
		return nil, common.ErrNotFound
	}
	cp := *c
	return &cp, nil
}
func (r campaignRepo) List(context.Context, int, int) ([]model.Campaign, error) { return nil, nil }
func (r campaignRepo) SetStatus(_ context.Context, id uint, from []model.CampaignStatus, to model.CampaignStatus, fields map[string]any) (bool, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	c := r.s.campaigns[id]
	for _, f := range from {
		if c.Status == f {
			c.Status = to
			if v, ok := fields["planner_attempts"].(int); ok {
				c.PlannerAttempts = v
			}
			if v, ok := fields["last_error"].(string); ok {
				c.LastError = v
			}
			if v, ok := fields["planner_lease_until"].(time.Time); ok {
				c.PlannerLeaseUntil = &v
			}
			return true, nil
		}
	}
	return false, nil
}
func (r campaignRepo) Cancel(_ context.Context, id uint) (bool, error) {
	c := r.s.campaigns[id]
	if c.Status.Finished() {
		return false, nil
	}
	c.Status = model.CampaignCancelled
	return true, nil
}
func (r campaignRepo) ClaimForPlanning(_ context.Context, now, until time.Time) (*model.Campaign, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	ids := make([]uint, 0)
	for id := range r.s.campaigns {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		c := r.s.campaigns[id]
		if (c.Status == model.CampaignPending || c.Status == model.CampaignPlanning) && (c.PlannerLeaseUntil == nil || c.PlannerLeaseUntil.Before(now)) {
			c.Status, c.PlannerLeaseUntil = model.CampaignPlanning, &until
			cp := *c
			return &cp, nil
		}
	}
	return nil, nil
}
func (r campaignRepo) ReleasePlanning(_ context.Context, id uint) error {
	r.s.campaigns[id].PlannerLeaseUntil = nil
	return nil
}
func (r campaignRepo) CompleteIfDrained(_ context.Context, id uint, now time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	c := r.s.campaigns[id]
	if c.Status != model.CampaignSending {
		return nil
	}
	for _, b := range r.s.batches {
		if b.CampaignID == id && (b.State == model.BatchQueued || b.State == model.BatchSending) {
			return nil
		}
	}
	c.Status, c.CompletedAt = model.CampaignCompleted, &now
	return nil
}
func (r campaignRepo) Progress(_ context.Context, id uint) (model.CampaignProgress, error) {
	var p model.CampaignProgress
	for _, b := range r.s.batches {
		if b.CampaignID != id {
			continue
		}
		switch b.State {
		case model.BatchSent:
			p.BatchesSent++
		case model.BatchFailed:
			p.BatchesFailed++
		case model.BatchQueued:
			p.BatchesQueued++
		case model.BatchCancelled:
			p.BatchesCancelled++
		}
		p.Targets += b.Targets
		p.Sent += b.Sent
		p.Failed += b.Failed
		p.InvalidTokens += b.Invalid
	}
	return p, nil
}

type batchRepo struct{ s *store }

func (r batchRepo) CreateAndAdvance(_ context.Context, b *model.CampaignBatch, cursor uint) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	b.ID = r.s.id()
	cp := *b
	r.s.batches[b.ID] = &cp
	c := r.s.campaigns[b.CampaignID]
	c.Cursor, c.BatchesPlanned = cursor, c.BatchesPlanned+1
	return nil
}
func (r batchRepo) GetByID(_ context.Context, id uint) (*model.CampaignBatch, error) {
	b, ok := r.s.batches[id]
	if !ok {
		return nil, common.ErrNotFound
	}
	cp := *b
	return &cp, nil
}
func (r batchRepo) Claim(_ context.Context, id uint, now, until time.Time) (*model.CampaignBatch, bool, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	b, ok := r.s.batches[id]
	if !ok {
		return nil, false, common.ErrNotFound
	}
	if !(b.State == model.BatchQueued || (b.State == model.BatchSending && b.LeaseUntil.Before(now))) {
		return nil, false, nil
	}
	b.State, b.LeaseUntil, b.Attempts = model.BatchSending, &until, b.Attempts+1
	cp := *b
	return &cp, true, nil
}
func (r batchRepo) Finish(_ context.Context, b *model.CampaignBatch) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	x := r.s.batches[b.ID]
	x.State, x.Sent, x.Failed, x.Invalid, x.LastError, x.LeaseUntil = b.State, b.Sent, b.Failed, b.Invalid, b.LastError, nil
	return nil
}
func (r batchRepo) Requeue(_ context.Context, id uint, after time.Time, lastErr string) error {
	x := r.s.batches[id]
	x.State, x.LeaseUntil, x.RepublishAfter, x.LastError = model.BatchQueued, nil, after, lastErr
	return nil
}
func (r batchRepo) DueForRepublish(_ context.Context, now time.Time, limit int) ([]model.CampaignBatch, error) {
	var out []model.CampaignBatch
	for _, b := range r.s.batches {
		if (b.State == model.BatchQueued && b.RepublishAfter.Before(now)) || (b.State == model.BatchSending && b.LeaseUntil.Before(now)) {
			out = append(out, *b)
		}
	}
	return out, nil
}
func (r batchRepo) TouchRepublish(_ context.Context, id uint, after time.Time) error {
	r.s.batches[id].RepublishAfter = after
	return nil
}
func (r batchRepo) CancelPending(_ context.Context, campaignID uint) error {
	for _, b := range r.s.batches {
		if b.CampaignID == campaignID && b.State == model.BatchQueued {
			b.State = model.BatchCancelled
		}
	}
	return nil
}

type targetSource struct{ s *store }

func (t targetSource) candidates(c *model.Campaign) []model.BatchTarget {
	var out []model.BatchTarget
	switch c.Audience {
	case model.AudienceEmails:
		for _, r := range t.s.recipients {
			if r.CampaignID == c.ID {
				out = append(out, model.BatchTarget{ID: r.ID, Email: r.Email})
			}
		}
	default:
		users := map[uint]bool{}
		for _, r := range t.s.recipients {
			if r.CampaignID == c.ID {
				users[r.UserID] = true
			}
		}
		for _, d := range t.s.devices {
			if d.IsActive && (c.Audience == model.AudienceAllDevices || users[d.UserID]) {
				out = append(out, model.BatchTarget{ID: d.ID, UserID: d.UserID, DeviceType: d.DeviceType, Token: d.DeviceToken})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (t targetSource) NextIDs(_ context.Context, c *model.Campaign, after uint, limit int) ([]uint, error) {
	t.s.mu.Lock()
	defer t.s.mu.Unlock()
	var ids []uint
	for _, x := range t.candidates(c) {
		if x.ID > after && len(ids) < limit {
			ids = append(ids, x.ID)
		}
	}
	return ids, nil
}
func (t targetSource) Targets(_ context.Context, c *model.Campaign, after, until uint) ([]model.BatchTarget, error) {
	t.s.mu.Lock()
	defer t.s.mu.Unlock()
	var out []model.BatchTarget
	for _, x := range t.candidates(c) {
		if x.ID > after && x.ID <= until {
			out = append(out, x)
		}
	}
	return out, nil
}

type fakePublisher struct {
	mu        sync.Mutex
	published []uint
	err       error
}

func (p *fakePublisher) PublishBatch(_ context.Context, id uint) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil {
		return p.err
	}
	p.published = append(p.published, id)
	return nil
}
func (p *fakePublisher) Publish(context.Context, *model.Notification) error { return nil }

type fakePush struct {
	mu       sync.Mutex
	calls    [][]string
	outcome  func(token string) port.TokenResult
	err      error
	topics   []string
	topicErr error
}

func (p *fakePush) SendToTokens(_ context.Context, tokens []string, _, _ string, _ map[string]string) ([]port.TokenResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = append(p.calls, tokens)
	if p.err != nil {
		return nil, p.err
	}
	out := make([]port.TokenResult, len(tokens))
	for i, tk := range tokens {
		out[i] = port.TokenResult{Success: true}
		if p.outcome != nil {
			out[i] = p.outcome(tk)
		}
	}
	return out, nil
}
func (p *fakePush) SendToTopic(_ context.Context, topic, _, _ string, _ map[string]string) (string, error) {
	p.topics = append(p.topics, topic)
	return "msg-1", p.topicErr
}

type fakeEmail struct {
	batches  [][]port.EmailMessage
	keys     []string
	err      error
	lastTo   []string
	lastSubj string
	lastBody string
}

func (e *fakeEmail) Send(_ context.Context, to []string, subject, body string) error {
	e.lastTo, e.lastSubj, e.lastBody = to, subject, body
	return e.err
}

type fakeSlack struct{ title, text string }

func (s *fakeSlack) Send(_ context.Context, title, text string) error {
	s.title, s.text = title, text
	return nil
}

type fakeSlackTemplate struct{}

func (fakeSlackTemplate) Render(t string, d map[string]string) (string, string, error) {
	if t != "alert" {
		return "", "", fmt.Errorf("unknown template %q", t)
	}
	return "Alert", "service " + d["service"] + " is down", nil
}
func (e *fakeEmail) SendBatch(_ context.Context, m []port.EmailMessage, key string) error {
	e.batches = append(e.batches, m)
	e.keys = append(e.keys, key)
	return e.err
}
func (e *fakeEmail) MaxBatchSize() int { return 100 }

type fakeTemplate struct{}

func (fakeTemplate) Render(t string, d map[string]string) (string, string, error) {
	if t != "welcome" {
		return "", "", fmt.Errorf("unknown template %q", t)
	}
	return "Chào " + d["name"], "<p>xin chào</p>", nil
}

type fixture struct {
	s         *store
	contacts  *contactRepo
	audit     *auditRepo
	clock     *fakeClock
	push      *fakePush
	email     *fakeEmail
	slack     *fakeSlack
	pub       *fakePublisher
	notif     port.NotificationUsecase
	campaigns port.CampaignUsecase
	runner    port.CampaignRunner
	settings  CampaignSettings
}

func newFixture() *fixture {
	s := newStore()
	f := &fixture{s: s, contacts: newContactRepo(), audit: &auditRepo{}, clock: &fakeClock{now: t0}, push: &fakePush{}, email: &fakeEmail{}, slack: &fakeSlack{}, pub: &fakePublisher{}}
	f.settings = CampaignSettings{
		PushBatchSize: 500, EmailBatchSize: 100, MaxRecipients: 1_000_000, MaxAttempts: 3,
		PlannerLease: 2 * time.Minute, BatchLease: 5 * time.Minute, RepublishGrace: 10 * time.Minute,
		BatchesPerPlan: 100, RepublishPerTick: 500,
	}
	f.notif = NewNotificationService(notificationRepo{s}, deviceRepo{s}, f.contacts, deviceRepo{s}, f.pub, f.push, f.email, f.slack, fakeTemplate{}, fakeSlackTemplate{}, NewAuditor(f.audit, f.clock))
	f.campaigns = NewCampaignService(campaignRepo{s}, batchRepo{s}, fakeTemplate{}, f.settings)
	f.runner = NewCampaignRunner(campaignRepo{s}, batchRepo{s}, targetSource{s}, f.pub, f.push, f.email, fakeTemplate{}, deviceRepo{s}, f.clock, f.settings, NewAuditor(f.audit, f.clock))
	return f
}

func (f *fixture) devices(policy DevicePolicy, invalid ...string) port.UserDeviceUsecase {
	bad := map[string]bool{}
	for _, t := range invalid {
		bad[t] = true
	}
	return NewUserDeviceService(deviceRepo{f.s}, f.contacts, fakeValidator{invalid: bad}, policy)
}

func (f *fixture) drain(t interface{ Fatal(...any) }) {
	for {
		worked, err := f.runner.PlanNext(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if !worked {
			break
		}
	}
	for len(f.pub.published) > 0 {
		id := f.pub.published[0]
		f.pub.published = f.pub.published[1:]
		if _, err := f.runner.SendBatch(context.Background(), id); err != nil {
			t.Fatal(err)
		}
	}
}
