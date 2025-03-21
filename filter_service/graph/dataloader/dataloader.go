package dataloader

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/filter-service/graph/model"
	"github.com/JIeeiroSst/filter-service/graph/storage"
	"github.com/graph-gophers/dataloader"
)

type ctxKey string

const (
	loadersKey = ctxKey("dataloaders")
)

// DataLoader offers data loaders scoped to a context
type DataLoader struct {
	userLoader *dataloader.Loader
}

// GetUser wraps the User dataloader for efficient retrieval by user ID
func (i *DataLoader) GetUser(ctx context.Context, userID string) (*model.User, error) {
	thunk := i.userLoader.Load(ctx, dataloader.StringKey(userID))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.(*model.User), nil
}

// NewDataLoader returns the instantiated Loaders struct for use in a request
func NewDataLoader(db storage.Storage) *DataLoader {
	// instantiate the user dataloader
	users := &userBatcher{db: db}
	// return the DataLoader
	return &DataLoader{
		userLoader: dataloader.NewBatchedLoader(users.get),
	}
}

// Middleware injects a DataLoader into the request context so it can be
// used later in the schema resolvers
func Middleware(loader *DataLoader, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCtx := context.WithValue(r.Context(), loadersKey, loader)
		r = r.WithContext(nextCtx)
		next.ServeHTTP(w, r)
	})
}

// For returns the dataloader for a given context
func For(ctx context.Context) *DataLoader {
	return ctx.Value(loadersKey).(*DataLoader)
}

// userBatcher wraps storage and provides a "get" method for the user dataloader
type userBatcher struct {
	db storage.Storage
}

// get implements the dataloader for finding many users by Id and returns
// them in the order requested
func (u *userBatcher) get(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	fmt.Printf("dataloader.userBatcher.get, users: [%s]\n", strings.Join(keys.Keys(), ","))
	// create a map for remembering the order of keys passed in
	keyOrder := make(map[string]int, len(keys))
	// collect the keys to search for
	var userIDs []string
	for ix, key := range keys {
		userIDs = append(userIDs, key.String())
		keyOrder[key.String()] = ix
	}
	// search for those users
	dbRecords, err := u.db.GetUsers(ctx, userIDs)
	// if DB error, return
	if err != nil {
		return []*dataloader.Result{{Data: nil, Error: err}}
	}
	// construct an output array of dataloader results
	results := make([]*dataloader.Result, len(keys))
	// enumerate records, put into output
	for _, record := range dbRecords {
		ix, ok := keyOrder[record.ID]
		// if found, remove from index lookup map so we know elements were found
		if ok {
			results[ix] = &dataloader.Result{Data: record, Error: nil}
			delete(keyOrder, record.ID)
		}
	}
	// fill array positions with errors where not found in DB
	for userID, ix := range keyOrder {
		err := fmt.Errorf("user not found %s", userID)
		results[ix] = &dataloader.Result{Data: nil, Error: err}
	}
	// return results
	return results
}
