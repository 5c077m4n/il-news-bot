package db

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

func TestUnmarshalContextTime(t *testing.T) {
	ctx := context.Background()

	encoded, err := json.MarshalContext(ctx, entry[[]string]{
		CreatedAt: time.Now(),
		Value:     []string{"a", "b"},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var record entry[[]string]
	if err := json.UnmarshalContext(ctx, encoded, &record); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if record.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
	if !reflect.DeepEqual(record.Value, []string{"a", "b"}) {
		t.Errorf("got value %v, want [a b]", record.Value)
	}
}
