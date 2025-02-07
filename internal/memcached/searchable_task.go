package memcached

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
)

// SearchableTask ...
type SearchableTask struct {
	client *memcache.Client
	orig   SearchableTaskStore
}

type SearchableTaskStore interface {
	Delete(ctx context.Context, id string) error
	Index(ctx context.Context, task internal.Task) error
	Search(ctx context.Context, args internal.SearchParams) (internal.SearchResults, error)
}

// NewSearchableTask instantiates the Task repository.
func NewSearchableTask(client *memcache.Client, orig SearchableTaskStore) *SearchableTask {
	return &SearchableTask{
		client: client,
		orig:   orig,
	}
}

// Index ...
func (t *SearchableTask) Index(ctx context.Context, task internal.Task) error {
	defer newOTELSpan(ctx, "SearchableTask.Index").End()

	//-

	if err := t.orig.Index(ctx, task); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "orig.Index")
	}

	return nil
}

// Delete ...
func (t *SearchableTask) Delete(ctx context.Context, id string) error {
	defer newOTELSpan(ctx, "SearchableTask.Delete").End()

	//-

	if err := t.orig.Delete(ctx, id); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "orig.Delete")
	}

	return nil
}

// Search ...
func (t *SearchableTask) Search(ctx context.Context, args internal.SearchParams) (internal.SearchResults, error) {
	defer newOTELSpan(ctx, "SearchableTask.Search").End()

	//-

	key := newSearchableKey(args)

	var res internal.SearchResults

	if err := getTask(ctx, t.client, key, &res); err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			res, err := t.orig.Search(ctx, args)
			if err != nil {
				return internal.SearchResults{}, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "orig.Search")
			}

			setTask(ctx, t.client, key, &res, 25*time.Second)

			return res, nil
		}

		return internal.SearchResults{}, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "get")
	}

	return res, nil
}

func newSearchableKey(args internal.SearchParams) string {
	var (
		description string
		priority    int8
		isDone      bool
	)

	if args.Description != nil {
		description = *args.Description
	}

	if args.Priority != nil {
		priority = int8(*args.Priority)
	}

	if args.IsDone != nil {
		isDone = *args.IsDone
	}

	return fmt.Sprintf("%s_%d_%t_%d_%d", description, priority, isDone, args.From, args.Size)
}
