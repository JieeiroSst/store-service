package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type socialRepository struct{ db *gorm.DB }

func NewSocialRepository(db *gorm.DB) *socialRepository { return &socialRepository{db: db} }

func (r *socialRepository) GetProfile(ctx context.Context, userID string) (*model.Profile, error) {
	var p model.Profile
	if err := conn(ctx, r.db).Where("user_id = ?", userID).First(&p).Error; err != nil {
		return nil, notFound(err)
	}
	return &p, nil
}

func (r *socialRepository) GetProfiles(ctx context.Context, userIDs []string) (map[string]model.Profile, error) {
	out := map[string]model.Profile{}
	if len(userIDs) == 0 {
		return out, nil
	}
	var items []model.Profile
	if err := conn(ctx, r.db).Where("user_id IN ?", userIDs).Find(&items).Error; err != nil {
		return nil, err
	}
	for _, p := range items {
		out[p.UserID] = p
	}
	return out, nil
}

func (r *socialRepository) SaveProfile(ctx context.Context, p *model.Profile) error {
	c := conn(ctx, r.db)
	res := c.Model(&model.Profile{}).Where("user_id = ?", p.UserID).
		Updates(map[string]any{"username": p.Username, "bio": p.Bio, "avatar_url": p.AvatarURL})
	if res.Error != nil {
		return mapWriteErr(res.Error)
	}
	if res.RowsAffected > 0 {
		return nil
	}
	var exists int64
	if err := c.Model(&model.Profile{}).Where("user_id = ?", p.UserID).Count(&exists).Error; err != nil || exists > 0 {
		return err
	}
	return mapWriteErr(c.Create(p).Error)
}

func (r *socialRepository) CreateComment(ctx context.Context, c *model.Comment) error {
	return conn(ctx, r.db).Create(c).Error
}

func (r *socialRepository) GetComment(ctx context.Context, id int64) (*model.Comment, error) {
	var c model.Comment
	if err := conn(ctx, r.db).First(&c, id).Error; err != nil {
		return nil, notFound(err)
	}
	return &c, nil
}

func (r *socialRepository) DeleteComment(ctx context.Context, id int64) error {
	return conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		ids := []int64{id}
		var replies []int64
		if err := tx.Model(&model.Comment{}).Where("parent_id = ?", id).Pluck("id", &replies).Error; err != nil {
			return err
		}
		ids = append(ids, replies...)
		if err := tx.Where("comment_id IN ?", ids).Delete(&model.CommentLike{}).Error; err != nil {
			return err
		}
		return tx.Where("id IN ?", ids).Delete(&model.Comment{}).Error
	})
}

const likesExpr = "(SELECT COUNT(*) FROM comment_likes l WHERE l.comment_id = comments.id)"

func (r *socialRepository) ListComments(ctx context.Context, eventID int64, sort string, limit int, cursorStr string) (*port.Page[model.Comment], error) {
	if sort != "likes" {
		sort = "newest"
	}
	cur, err := decodeCursor(cursorStr, "comments:"+sort)
	if err != nil {
		return nil, err
	}
	q := conn(ctx, r.db).Model(&model.Comment{}).
		Select("comments.*, "+likesExpr+" AS likes").Where("event_id = ?", eventID)
	if sort == "likes" {
		if cur != nil {
			q = q.Where(likesExpr+" < ? OR ("+likesExpr+" = ? AND comments.id < ?)", cur.N, cur.N, cur.ID)
		}
		q = q.Order("likes DESC, comments.id DESC")
	} else {
		if cur != nil {
			q = q.Where("comments.id < ?", cur.ID)
		}
		q = q.Order("comments.id DESC")
	}
	var rows []model.Comment
	if err := q.Limit(limit + 1).Find(&rows).Error; err != nil {
		return nil, err
	}
	return paginate(rows, limit, func(c model.Comment) cursor {
		return cursor{S: "comments:" + sort, ID: c.ID, N: c.Likes}
	}), nil
}

func (r *socialRepository) AddLike(ctx context.Context, commentID int64, userID string) error {
	return conn(ctx, r.db).Clauses(clause.OnConflict{DoNothing: true}).
		Create(&model.CommentLike{CommentID: commentID, UserID: userID, CreatedAt: time.Now()}).Error
}

func (r *socialRepository) RemoveLike(ctx context.Context, commentID int64, userID string) error {
	return conn(ctx, r.db).Where("comment_id = ? AND user_id = ?", commentID, userID).Delete(&model.CommentLike{}).Error
}

func (r *socialRepository) AddBookmark(ctx context.Context, userID string, eventID int64) error {
	return conn(ctx, r.db).Clauses(clause.OnConflict{DoNothing: true}).
		Create(&model.Bookmark{UserID: userID, EventID: eventID, CreatedAt: time.Now()}).Error
}

func (r *socialRepository) RemoveBookmark(ctx context.Context, userID string, eventID int64) error {
	return conn(ctx, r.db).Where("user_id = ? AND event_id = ?", userID, eventID).Delete(&model.Bookmark{}).Error
}

func (r *socialRepository) ListBookmarks(ctx context.Context, userID string) ([]int64, error) {
	var ids []int64
	err := conn(ctx, r.db).Model(&model.Bookmark{}).Where("user_id = ?", userID).Order("created_at DESC").Pluck("event_id", &ids).Error
	return ids, err
}

type statsRepository struct{ db *gorm.DB }

func NewStatsRepository(db *gorm.DB) *statsRepository { return &statsRepository{db: db} }

func (r *statsRepository) Leaderboard(ctx context.Context, metric string, since time.Time, limit int) ([]port.LeaderboardRow, error) {
	var rows []port.LeaderboardRow
	c := conn(ctx, r.db)
	if since.IsZero() {
		since = time.Unix(0, 0)
	}
	if metric == "volume" {
		err := c.Raw(`
			SELECT user_id, SUM(v) AS volume FROM (
				SELECT taker_user_id AS user_id, yes_price * size AS v, created_at FROM trades
				UNION ALL
				SELECT maker_user_id AS user_id, yes_price * size AS v, created_at FROM trades
			) t WHERE created_at >= ? GROUP BY user_id ORDER BY volume DESC, user_id LIMIT ?`, since, limit).Scan(&rows).Error
		return rows, err
	}
	err := c.Model(&model.PnLEntry{}).
		Select("user_id, SUM(amount) AS profit").
		Where("created_at >= ?", since).
		Group("user_id").Order("profit DESC, user_id").Limit(limit).Scan(&rows).Error
	return rows, err
}
