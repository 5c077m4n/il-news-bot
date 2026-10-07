package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDatabase(t *testing.T) *Database {
	t.Helper()

	database, err := New(
		t.TempDir(),
		func(context.Context, string) ([]float32, error) {
			return []float32{1}, nil
		},
	)
	require.NoError(t, err, "new database")

	return database
}

func TestSaveArticles(t *testing.T) {
	database := newTestDatabase(t)

	ctx := context.Background()
	items := []Article{
		{
			Source:      "YNet",
			Title:       "T1",
			Description: "D1",
			Link:        "https://example.com/1",
		},
		{
			Source:      "JPost",
			Title:       "T2",
			Description: "D2",
		},
	}

	require.NoError(t, database.SaveArticles(ctx, items), "save")

	doc, err := database.articles.GetByID(ctx, "https://example.com/1")
	require.NoError(t, err, "get by link id")
	assert.Equal(t, "T1\nD1", doc.Content)
	assert.Equal(t, "YNet", doc.Metadata["source"])

	doc, err = database.articles.GetByID(ctx, items[1].id())
	require.NoError(t, err, "get by fallback id")
	assert.Equal(t, "T2\nD2", doc.Content)
}

func TestQueryArticles(t *testing.T) {
	database := newTestDatabase(t)

	ctx := context.Background()
	items := []Article{
		{
			Source:      "YNet",
			Title:       "T1",
			Description: "D1",
			Link:        "https://example.com/1",
		},
		{
			Source:      "JPost",
			Title:       "T2",
			Description: "D2",
			Link:        "https://example.com/2",
		},
		{
			Source:      "Haaretz",
			Title:       "T3",
			Description: "D3",
		},
	}

	require.NoError(t, database.SaveArticles(ctx, items), "save")

	articles, err := database.QueryArticles(
		ctx,
		"anything",
		10,
		map[string]string{"source": "YNet"},
	)
	require.NoError(t, err, "query")

	require.Len(t, articles, 1)
	assert.Equal(t, "T1", articles[0].Title)
	assert.Equal(t, "D1", articles[0].Description)
	assert.Equal(t, "https://example.com/1", articles[0].Link)
}

func TestDeleteOldArticles(t *testing.T) {
	database := newTestDatabase(t)

	ctx := context.Background()
	now := time.Now()
	items := []Article{
		{
			Source:      "YNet",
			Title:       "Old",
			Description: "Old",
			Link:        "https://example.com/old",
			PublishedAt: now.Add(-3 * 24 * time.Hour),
		},
		{
			Source:      "JPost",
			Title:       "New",
			Description: "New",
			Link:        "https://example.com/new",
			PublishedAt: now.Add(-time.Hour),
		},
	}

	require.NoError(t, database.SaveArticles(ctx, items), "save")

	removed, err := database.DeleteOldArticles(ctx, 48*time.Hour)
	require.NoError(t, err, "delete old")
	assert.Positive(t, removed, "removed articles count")

	_, err = database.articles.GetByID(ctx, "https://example.com/old")
	assert.Error(t, err, "old article was not removed")
	_, err = database.articles.GetByID(ctx, "https://example.com/new")
	assert.NoError(t, err, "recent article was removed")
}
