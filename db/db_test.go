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
			Lean:        LeanLeft,
			Title:       "T1",
			Description: "D1",
			Link:        "https://example.com/1",
		},
		{
			Source:      "JPost",
			Lean:        LeanRight,
			Title:       "T2",
			Description: "D2",
		},
	}

	require.NoError(t, database.SaveArticles(ctx, items), "save")

	doc, err := database.articles.GetByID(ctx, "https://example.com/1")
	require.NoError(t, err, "get by link id")
	assert.Equal(t, "T1\nD1", doc.Content)
	assert.Equal(t, "YNet", doc.Metadata["source"])
	assert.Equal(t, "left", doc.Metadata["lean"])

	doc, err = database.articles.GetByID(ctx, items[1].id())
	require.NoError(t, err, "get by fallback id")
	assert.Equal(t, "T2\nD2", doc.Content)
}

func TestQueryArticles(t *testing.T) {
	database := newTestDatabase(t)

	ctx := context.Background()
	leftLean := Lean("test-left")
	items := []Article{
		{
			Source:      "YNet",
			Lean:        leftLean,
			Title:       "T1",
			Description: "D1",
			Link:        "https://example.com/1",
		},
		{
			Source:      "JPost",
			Lean:        Lean("test-right"),
			Title:       "T2",
			Description: "D2",
			Link:        "https://example.com/2",
		},
		{
			Source:      "Haaretz",
			Lean:        leftLean,
			Title:       "T3",
			Description: "D3",
		},
	}

	require.NoError(t, database.SaveArticles(ctx, items), "save")

	articles, err := database.QueryArticles(
		ctx,
		"anything",
		10,
		map[string]string{"lean": string(leftLean)},
	)
	require.NoError(t, err, "query")

	require.Len(t, articles, 2)
	var ynet *Article
	for i, article := range articles {
		assert.Equal(t, leftLean, article.Lean)
		if article.Source == "YNet" {
			ynet = &articles[i]
		}
	}
	require.NotNil(t, ynet, "YNet article not found in results")
	assert.Equal(t, "T1", ynet.Title)
	assert.Equal(t, "D1", ynet.Description)
	assert.Equal(t, "https://example.com/1", ynet.Link)
}

func TestDeleteOldArticles(t *testing.T) {
	database := newTestDatabase(t)

	ctx := context.Background()
	now := time.Now()
	items := []Article{
		{
			Source:      "YNet",
			Lean:        LeanNeutral,
			Title:       "Old",
			Description: "Old",
			Link:        "https://example.com/old",
			PublishedAt: now.Add(-3 * 24 * time.Hour),
		},
		{
			Source:      "JPost",
			Lean:        LeanNeutral,
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
