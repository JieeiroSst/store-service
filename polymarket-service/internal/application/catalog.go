package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

type catalogService struct {
	events  port.EventRepository
	markets port.MarketRepository
	tx      port.TxManager
	opts    Options
	now     func() time.Time
}

func NewCatalogService(events port.EventRepository, markets port.MarketRepository, tx port.TxManager, opts Options) *catalogService {
	return &catalogService{events: events, markets: markets, tx: tx, opts: opts.withDefaults(), now: time.Now}
}

func slugify(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			dash = false
		case !dash && b.Len() > 0:
			b.WriteByte('-')
			dash = true
		}
	}
	slug := strings.Trim(b.String(), "-")
	if len(slug) > 120 {
		slug = strings.Trim(slug[:120], "-")
	}
	return slug
}

func (s *catalogService) CreateEvent(ctx context.Context, in port.CreateEventInput) (*model.Event, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || len(in.Markets) == 0 || !in.EndDate.After(s.now()) {
		return nil, fmt.Errorf("%w: title, at least one market and a future end_date are required", port.ErrInvalidInput)
	}
	slug := in.Slug
	if slug == "" {
		slug = slugify(title)
	}
	if slug == "" {
		return nil, fmt.Errorf("%w: cannot derive a slug from the title", port.ErrInvalidInput)
	}

	event := &model.Event{
		Slug: slug, Title: title, Description: in.Description, Category: strings.ToLower(strings.TrimSpace(in.Category)),
		Tags: normalizeTags(in.Tags), ImageURL: in.ImageURL, Featured: in.Featured, NegRisk: in.NegRisk,
		Status: model.EventOpen, EndDate: in.EndDate,
	}
	markets := make([]model.Market, 0, len(in.Markets))
	seen := map[string]bool{}
	for i, mi := range in.Markets {
		question := strings.TrimSpace(mi.Question)
		if question == "" {
			return nil, fmt.Errorf("%w: market %d has no question", port.ErrInvalidInput, i)
		}
		end := mi.EndTime
		if end.IsZero() {
			end = in.EndDate
		}
		if !end.After(s.now()) {
			return nil, fmt.Errorf("%w: market %d must end in the future", port.ErrInvalidInput, i)
		}
		value := mi.ShareValue
		if value == 0 {
			value = s.opts.ShareValue
		}
		if value < 2 {
			return nil, fmt.Errorf("%w: share_value must be at least 2", port.ErrInvalidInput)
		}
		minSize := mi.MinOrderSize
		if minSize < 1 {
			minSize = s.opts.MinOrderSize
		}
		if mi.RewardPool < 0 || mi.RewardMaxSpread < 0 || mi.RewardMaxSpread >= value || (mi.RewardPool > 0 && mi.RewardMaxSpread == 0) {
			return nil, fmt.Errorf("%w: market %d has invalid liquidity reward settings", port.ErrInvalidInput, i)
		}
		if event.NegRisk && len(markets) > 0 && value != markets[0].ShareValue {
			return nil, fmt.Errorf("%w: markets of a neg-risk event must share one share_value", port.ErrInvalidInput)
		}
		mslug := mi.Slug
		if mslug == "" {
			mslug = slugify(question)
			if len(in.Markets) > 1 && mi.GroupItemTitle != "" {
				mslug = slug + "-" + slugify(mi.GroupItemTitle)
			}
		}
		if seen[mslug] {
			mslug += "-" + strconv.Itoa(i+1)
		}
		seen[mslug] = true
		markets = append(markets, model.Market{
			Slug: mslug, Question: question, GroupItemTitle: mi.GroupItemTitle, Description: mi.Description,
			ShareValue: value, MinOrderSize: minSize, EndTime: end, Status: model.MarketOpen,
			RewardPool: mi.RewardPool, RewardMaxSpread: mi.RewardMaxSpread, RewardMinSize: mi.RewardMinSize,
		})
	}

	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		event.ID = 0
		if err := s.events.Create(ctx, event); err != nil {
			return err
		}
		event.Markets = nil
		for i := range markets {
			markets[i].ID = 0
			markets[i].EventID = event.ID
			if err := s.markets.Create(ctx, &markets[i]); err != nil {
				return err
			}
		}
		event.Markets = markets
		return nil
	})
	if err != nil {
		return nil, err
	}
	return event, nil
}

func normalizeTags(tags []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, t := range tags {
		t = strings.ToLower(strings.TrimSpace(t))
		if t != "" && !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}

func (s *catalogService) loadEvent(ctx context.Context, idOrSlug string) (*model.Event, error) {
	if id, err := strconv.ParseInt(idOrSlug, 10, 64); err == nil {
		return s.events.GetByID(ctx, id)
	}
	return s.events.GetBySlug(ctx, idOrSlug)
}

func (s *catalogService) GetEvent(ctx context.Context, idOrSlug string) (*model.Event, error) {
	event, err := s.loadEvent(ctx, idOrSlug)
	if err != nil {
		return nil, err
	}
	events := []model.Event{*event}
	if err := s.attachMarkets(ctx, events); err != nil {
		return nil, err
	}
	return &events[0], nil
}

func (s *catalogService) ListEvents(ctx context.Context, f port.EventFilter) (*port.Page[model.Event], error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	page, err := s.events.List(ctx, f)
	if err != nil {
		return nil, err
	}
	return page, s.attachMarkets(ctx, page.Items)
}

func (s *catalogService) RelatedEvents(ctx context.Context, idOrSlug string, limit int) ([]model.Event, error) {
	event, err := s.loadEvent(ctx, idOrSlug)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	items, err := s.events.Related(ctx, event.ID, event.Category, limit)
	if err != nil {
		return nil, err
	}
	return items, s.attachMarkets(ctx, items)
}

func (s *catalogService) Categories(ctx context.Context) ([]port.CategoryCount, error) {
	return s.events.Categories(ctx)
}

func (s *catalogService) GetMarket(ctx context.Context, idOrSlug string) (*model.Market, error) {
	if id, err := strconv.ParseInt(idOrSlug, 10, 64); err == nil {
		return s.markets.GetByID(ctx, id)
	}
	return s.markets.GetBySlug(ctx, idOrSlug)
}

func (s *catalogService) SearchMarkets(ctx context.Context, f port.MarketFilter) (*port.MarketList, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	f.Offset = max(f.Offset, 0)
	items, total, err := s.markets.Search(ctx, f)
	if err != nil {
		return nil, err
	}
	return &port.MarketList{Items: items, Total: total}, nil
}

func (s *catalogService) attachMarkets(ctx context.Context, events []model.Event) error {
	if len(events) == 0 {
		return nil
	}
	ids := make([]int64, len(events))
	for i := range events {
		ids[i] = events[i].ID
	}
	markets, err := s.markets.ListByEvents(ctx, ids)
	if err != nil {
		return err
	}
	byEvent := map[int64][]model.Market{}
	for _, m := range markets {
		byEvent[m.EventID] = append(byEvent[m.EventID], m)
	}
	for i := range events {
		events[i].Markets = byEvent[events[i].ID]
	}
	return nil
}
