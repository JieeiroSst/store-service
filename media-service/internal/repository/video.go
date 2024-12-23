package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/JIeeiroSst/media-service/model"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/pagination"
	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"
)

type Video interface {
	UploadVideo(ctx context.Context, video model.Video, tag model.Tag) error
	PaginateVideo(ctx context.Context, p pagination.Pagination) (*pagination.Pagination, error)
	InsertOrUpdateTag(ctx context.Context, tag model.Tag) error
	InsertOrUpdateVideo(ctx context.Context, video model.Video) error
	SearchVideo(ctx context.Context, query string, page int, size int) (*model.SearchVideo, error)
}

type VideoRepository struct {
	db      *gorm.DB
	elastic *elastic.Client
}

func NewVideoRepository(db *gorm.DB, elastic *elastic.Client) *VideoRepository {
	return &VideoRepository{
		db:      db,
		elastic: elastic,
	}
}

func (r *VideoRepository) UploadVideo(ctx context.Context, video model.Video, tag model.Tag) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	if err := tx.Create(&video).Error; err != nil {
		logger.Error(ctx, "CreateVideo err %v", err)
		return err
	}

	if err := tx.Create(&tag).Error; err != nil {
		logger.Error(ctx, "CreateTag err %v", err)
		return err
	}

	return nil
}

func (r *VideoRepository) PaginateVideo(ctx context.Context, param pagination.Pagination) (*pagination.Pagination, error) {
	var videos []model.Video

	r.db.Scopes(pagination.Paginate(videos, &param, r.db, "Tag")).Find(&videos)
	param.Rows = videos

	return &param, nil

}

func (r *VideoRepository) InsertOrUpdateTag(ctx context.Context, tag model.Tag) error {
	index := "tags"
	id := fmt.Sprintf("%d", tag.TagID)

	doc, err := json.Marshal(tag)
	if err != nil {
		return err
	}

	exists, err := r.elastic.Exists().Index(index).Id(id).Do(ctx)
	if err != nil {
		return err
	}

	if exists {
		_, err = r.elastic.Update().Index(index).Id(id).Doc(string(doc)).Do(ctx)
	} else {
		_, err = r.elastic.Index().Index(index).Id(id).BodyJson(tag).Do(ctx)
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *VideoRepository) InsertOrUpdateVideo(ctx context.Context, video model.Video) error {
	index := "videos"
	id := fmt.Sprintf("%d", video.VideoID)

	doc, err := json.Marshal(video)
	if err != nil {
		return err
	}

	exists, err := r.elastic.Exists().Index(index).Id(id).Do(ctx)
	if err != nil {
		return err
	}

	if exists {
		_, err = r.elastic.Update().Index(index).Id(id).Doc(string(doc)).Do(ctx)
	} else {
		_, err = r.elastic.Index().Index(index).Id(id).BodyJson(video).Do(ctx)
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *VideoRepository) SearchVideo(ctx context.Context, query string, page int, size int) (*model.SearchVideo, error) {
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 10
	}
	index := "videos"
	searchQuery := elastic.NewBoolQuery().Must(
		elastic.NewMatchQuery("description", query),
	)
	searchSource := elastic.NewSearchSource().
		Query(searchQuery).
		From((page - 1) * size).
		Size(size)
	searchResult, err := r.elastic.Search().
		Index(index).
		SearchSource(searchSource).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	total := searchResult.TotalHits()
	totalPages := int(math.Ceil(float64(total) / float64(size)))

	var videos []model.Video
	for _, hit := range searchResult.Hits.Hits {
		var video model.Video
		data, _ := hit.Source.MarshalJSON()
		err := json.Unmarshal(data, &video)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	return &model.SearchVideo{
		Videos: videos,
		Total:  total,
		Page:   page,
		Size:   size,
		Pages:  totalPages,
	}, nil
}
