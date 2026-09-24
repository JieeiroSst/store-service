package metadata

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/upload-service/internal/domain/model"
	"github.com/JIeeiroSst/upload-service/internal/domain/port"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

const collection = "files"

type Repo struct{ files *mongo.Collection }

func New(db *mongo.Database) *Repo { return &Repo{files: db.Collection(collection)} }

func (r *Repo) EnsureIndexes(ctx context.Context) error {
	_, err := r.files.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "receiver_id", Value: 1}, {Key: "created_at", Value: -1}},
	})
	return err
}

func (r *Repo) Insert(ctx context.Context, f *model.File) error {
	_, err := r.files.InsertOne(ctx, f)
	return err
}

func (r *Repo) Get(ctx context.Context, id string) (*model.File, error) {
	var f model.File
	err := r.files.FindOne(ctx, bson.M{"_id": id}).Decode(&f)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *Repo) List(ctx context.Context, receiverID string, limit, offset int) ([]model.File, int64, error) {
	filter := bson.M{"receiver_id": receiverID}
	total, err := r.files.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	cur, err := r.files.Find(ctx, filter, options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: 1}}).
		SetSkip(int64(offset)).SetLimit(int64(limit)))
	if err != nil {
		return nil, 0, err
	}
	out := []model.File{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *Repo) ReplaceContent(ctx context.Context, f *model.File) error {
	res, err := r.files.UpdateByID(ctx, f.ID, bson.M{"$set": bson.M{
		"object_key": f.ObjectKey, "file_name": f.FileName, "content_type": f.ContentType,
		"size": f.Size, "sha256": f.SHA256, "updated_at": f.UpdatedAt,
	}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *Repo) Delete(ctx context.Context, id string) error {
	res, err := r.files.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return model.ErrNotFound
	}
	return nil
}

var Module = fx.Options(
	fx.Provide(New, func(r *Repo) port.MetadataRepository { return r }),
	fx.Invoke(func(lc fx.Lifecycle, r *Repo) {
		lc.Append(fx.Hook{OnStart: r.EnsureIndexes})
	}),
)
