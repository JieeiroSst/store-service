package repository

import (
	"context"
	"strings"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"gorm.io/gorm"
)

type eventRepository struct{ db *gorm.DB }

func NewEventRepository(db *gorm.DB) *eventRepository { return &eventRepository{db: db} }

func (r *eventRepository) Create(ctx context.Context, e *model.Event) error {
	return mapWriteErr(conn(ctx, r.db).Create(e).Error)
}

func (r *eventRepository) GetByID(ctx context.Context, id int64) (*model.Event, error) {
	var e model.Event
	if err := conn(ctx, r.db).First(&e, id).Error; err != nil {
		return nil, notFound(err)
	}
	return &e, nil
}

func (r *eventRepository) GetBySlug(ctx context.Context, slug string) (*model.Event, error) {
	var e model.Event
	if err := conn(ctx, r.db).Where("slug = ?", slug).First(&e).Error; err != nil {
		return nil, notFound(err)
	}
	return &e, nil
}

func (r *eventRepository) GetByIDs(ctx context.Context, ids []int64) ([]model.Event, error) {
	var items []model.Event
	err := conn(ctx, r.db).Where("id IN ?", ids).Order("volume DESC").Find(&items).Error
	return items, err
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func (r *eventRepository) List(ctx context.Context, f port.EventFilter) (*port.Page[model.Event], error) {
	sort := f.Sort
	if sort != "newest" && sort != "ending" {
		sort = "volume"
	}
	cur, err := decodeCursor(f.Cursor, sort)
	if err != nil {
		return nil, err
	}

	q := conn(ctx, r.db).Model(&model.Event{})
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Category != "" {
		q = q.Where("category = ?", strings.ToLower(f.Category))
	}
	if f.Tag != "" {
		q = q.Where("tags LIKE ?", `%"`+escapeLike(strings.ToLower(f.Tag))+`"%`)
	}
	if f.Featured {
		q = q.Where("featured = 1")
	}
	if f.Query != "" {
		like := "%" + escapeLike(f.Query) + "%"
		q = q.Where("title LIKE ? OR EXISTS (SELECT 1 FROM markets m WHERE m.event_id = events.id AND m.question LIKE ?)", like, like)
	}

	switch sort {
	case "newest":
		if cur != nil {
			q = q.Where("id < ?", cur.ID)
		}
		q = q.Order("id DESC")
	case "ending":
		q = q.Where("status = 'open' AND end_date > ?", time.Now())
		if cur != nil {
			at := time.Unix(0, cur.T)
			q = q.Where("end_date > ? OR (end_date = ? AND id > ?)", at, at, cur.ID)
		}
		q = q.Order("end_date ASC, id ASC")
	default:
		if cur != nil {
			q = q.Where("volume < ? OR (volume = ? AND id < ?)", cur.N, cur.N, cur.ID)
		}
		q = q.Order("volume DESC, id DESC")
	}

	var rows []model.Event
	if err := q.Limit(f.Limit + 1).Find(&rows).Error; err != nil {
		return nil, err
	}
	return paginate(rows, f.Limit, func(e model.Event) cursor {
		return cursor{S: sort, ID: e.ID, N: e.Volume, T: e.EndDate.UnixNano()}
	}), nil
}

func (r *eventRepository) Related(ctx context.Context, id int64, category string, limit int) ([]model.Event, error) {
	q := conn(ctx, r.db).Where("id <> ? AND status = 'open'", id)
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var items []model.Event
	err := q.Order("volume DESC").Limit(limit).Find(&items).Error
	return items, err
}

func (r *eventRepository) Categories(ctx context.Context) ([]port.CategoryCount, error) {
	var items []port.CategoryCount
	err := conn(ctx, r.db).Model(&model.Event{}).
		Select("category, COUNT(*) AS events").
		Where("status = 'open' AND category <> ''").
		Group("category").Order("events DESC").Scan(&items).Error
	return items, err
}

func (r *eventRepository) AddVolume(ctx context.Context, id, delta int64) error {
	return conn(ctx, r.db).Model(&model.Event{}).Where("id = ?", id).
		UpdateColumn("volume", gorm.Expr("volume + ?", delta)).Error
}

func (r *eventRepository) Save(ctx context.Context, e *model.Event) error {
	return conn(ctx, r.db).Save(e).Error
}
