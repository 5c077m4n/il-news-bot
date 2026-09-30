package db

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/goccy/go-json"
)

var db = sync.OnceValues(func() (*pebble.DB, error) {
	db, err := pebble.Open("pebble_data", &pebble.Options{})
	if err != nil {
		return nil, err
	}

	return db, nil
})

type entry[T any] struct {
	CreatedAt time.Time `json:"createdAt"`
	Value     T         `json:"value"`
}

func Get[T any](ctx context.Context, key []byte) (*T, time.Time, error) {
	var zero time.Time

	db, err := db()
	if err != nil {
		return nil, zero, err
	}

	value, closer, err := db.Get(key)
	if err != nil {
		return nil, zero, err
	}
	defer func() {
		if err := closer.Close(); err != nil {
			slog.ErrorContext(
				ctx,
				"could not close PebbleDB instance",
				slog.String("error", err.Error()),
			)
		}
	}()

	var record entry[T]
	if err := json.UnmarshalContext(ctx, value, &record); err != nil {
		return nil, zero, err
	}

	return &record.Value, record.CreatedAt, nil
}

func Set[T any](ctx context.Context, key []byte, value T) error {
	db, err := db()
	if err != nil {
		return err
	}

	encoded, err := json.MarshalContext(ctx, entry[T]{
		CreatedAt: time.Now(),
		Value:     value,
	})
	if err != nil {
		return err
	}

	return db.Set(key, encoded, pebble.Sync)
}
