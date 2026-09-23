package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

var usernameRE = regexp.MustCompile(`^[A-Za-z0-9_]{3,30}$`)

const maxCommentLen = 2000

type socialService struct {
	social    port.SocialRepository
	events    port.EventRepository
	markets   port.MarketRepository
	positions port.PositionRepository
	stats     port.StatsRepository
	now       func() time.Time
}

func NewSocialService(
	social port.SocialRepository,
	events port.EventRepository,
	markets port.MarketRepository,
	positions port.PositionRepository,
	stats port.StatsRepository,
) *socialService {
	return &socialService{social: social, events: events, markets: markets, positions: positions, stats: stats, now: time.Now}
}

func (s *socialService) GetProfile(ctx context.Context, userID string) (*model.Profile, error) {
	if userID == "" {
		return nil, port.ErrInvalidUser
	}
	p, err := s.social.GetProfile(ctx, userID)
	if errors.Is(err, port.ErrNotFound) {
		return &model.Profile{UserID: userID}, nil
	}
	return p, err
}

func (s *socialService) UpdateProfile(ctx context.Context, in port.UpdateProfileInput) (*model.Profile, error) {
	if in.UserID == "" {
		return nil, port.ErrInvalidUser
	}
	if in.Username != "" && !usernameRE.MatchString(in.Username) {
		return nil, fmt.Errorf("%w: username must be 3-30 letters, digits or underscores", port.ErrInvalidInput)
	}
	if len(in.Bio) > 500 {
		return nil, fmt.Errorf("%w: bio is limited to 500 characters", port.ErrInvalidInput)
	}
	p, err := s.social.GetProfile(ctx, in.UserID)
	if errors.Is(err, port.ErrNotFound) {
		p = &model.Profile{UserID: in.UserID}
	} else if err != nil {
		return nil, err
	}
	if in.Username != "" {
		p.Username = in.Username
	}
	if p.Username == "" {
		return nil, fmt.Errorf("%w: a username is required to create a profile", port.ErrInvalidInput)
	}
	p.Bio, p.AvatarURL = in.Bio, in.AvatarURL
	if err := s.social.SaveProfile(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *socialService) AddComment(ctx context.Context, eventID int64, userID, body string, parentID int64) (*model.Comment, error) {
	body = strings.TrimSpace(body)
	if userID == "" {
		return nil, port.ErrInvalidUser
	}
	if body == "" || len([]rune(body)) > maxCommentLen {
		return nil, fmt.Errorf("%w: comment must be 1-%d characters", port.ErrInvalidInput, maxCommentLen)
	}
	if _, err := s.events.GetByID(ctx, eventID); err != nil {
		return nil, err
	}
	if parentID != 0 {
		parent, err := s.social.GetComment(ctx, parentID)
		if err != nil {
			return nil, err
		}
		if parent.EventID != eventID {
			return nil, fmt.Errorf("%w: parent comment belongs to another event", port.ErrInvalidInput)
		}
	}
	c := &model.Comment{EventID: eventID, ParentID: parentID, UserID: userID, Body: body}
	if err := s.social.CreateComment(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *socialService) ListComments(ctx context.Context, eventID int64, sort string, limit int, cursor string) (*port.Page[model.Comment], error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	return s.social.ListComments(ctx, eventID, sort, limit, cursor)
}

func (s *socialService) DeleteComment(ctx context.Context, commentID int64, userID string) error {
	c, err := s.social.GetComment(ctx, commentID)
	if err != nil {
		return err
	}
	if c.UserID != userID {
		return port.ErrForbidden
	}
	return s.social.DeleteComment(ctx, commentID)
}

func (s *socialService) LikeComment(ctx context.Context, commentID int64, userID string) error {
	if userID == "" {
		return port.ErrInvalidUser
	}
	if _, err := s.social.GetComment(ctx, commentID); err != nil {
		return err
	}
	return s.social.AddLike(ctx, commentID, userID)
}

func (s *socialService) UnlikeComment(ctx context.Context, commentID int64, userID string) error {
	if userID == "" {
		return port.ErrInvalidUser
	}
	return s.social.RemoveLike(ctx, commentID, userID)
}

func (s *socialService) Watch(ctx context.Context, userID string, eventID int64) error {
	if userID == "" {
		return port.ErrInvalidUser
	}
	if _, err := s.events.GetByID(ctx, eventID); err != nil {
		return err
	}
	return s.social.AddBookmark(ctx, userID, eventID)
}

func (s *socialService) Unwatch(ctx context.Context, userID string, eventID int64) error {
	if userID == "" {
		return port.ErrInvalidUser
	}
	return s.social.RemoveBookmark(ctx, userID, eventID)
}

func (s *socialService) Watchlist(ctx context.Context, userID string) ([]model.Event, error) {
	if userID == "" {
		return nil, port.ErrInvalidUser
	}
	ids, err := s.social.ListBookmarks(ctx, userID)
	if err != nil || len(ids) == 0 {
		return []model.Event{}, err
	}
	return s.events.GetByIDs(ctx, ids)
}

func (s *socialService) Leaderboard(ctx context.Context, metric, window string, limit int) ([]port.LeaderboardEntry, error) {
	if metric == "" {
		metric = "profit"
	}
	if metric != "profit" && metric != "volume" {
		return nil, fmt.Errorf("%w: metric must be profit or volume", port.ErrInvalidInput)
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var since time.Time
	switch window {
	case "day":
		since = s.now().Add(-24 * time.Hour)
	case "week":
		since = s.now().Add(-7 * 24 * time.Hour)
	case "month":
		since = s.now().Add(-30 * 24 * time.Hour)
	case "", "all":
	default:
		return nil, fmt.Errorf("%w: window must be day, week, month or all", port.ErrInvalidInput)
	}

	rows, err := s.stats.Leaderboard(ctx, metric, since, limit)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(rows))
	for i, r := range rows {
		ids[i] = r.UserID
	}
	profiles, err := s.social.GetProfiles(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]port.LeaderboardEntry, len(rows))
	for i, r := range rows {
		out[i] = port.LeaderboardEntry{Rank: i + 1, UserID: r.UserID, Username: profiles[r.UserID].Username, Profit: r.Profit, Volume: r.Volume}
	}
	return out, nil
}

func (s *socialService) TopHolders(ctx context.Context, marketID int64, outcome model.Outcome, limit int) ([]port.HolderEntry, error) {
	if !outcome.Valid() {
		return nil, port.ErrInvalidOutcome
	}
	if _, err := s.markets.GetByID(ctx, marketID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	holders, err := s.positions.TopHolders(ctx, marketID, outcome, limit)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(holders))
	for i, h := range holders {
		ids[i] = h.UserID
	}
	profiles, err := s.social.GetProfiles(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]port.HolderEntry, len(holders))
	for i, h := range holders {
		out[i] = port.HolderEntry{UserID: h.UserID, Username: profiles[h.UserID].Username, Shares: h.Shares}
	}
	return out, nil
}
