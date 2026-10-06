package db

import (
	"context"
	"testing"
	"time"
)

func newTestDatabase(t *testing.T) *Database {
	t.Helper()

	database, err := New(
		t.TempDir(),
		func(context.Context, string) ([]float32, error) {
			return []float32{1}, nil
		},
	)
	if err != nil {
		t.Fatalf("new database: %v", err)
	}

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

	if err := database.SaveArticles(ctx, items); err != nil {
		t.Fatalf("save: %v", err)
	}

	doc, err := database.articles.GetByID(ctx, "https://example.com/1")
	if err != nil {
		t.Fatalf("get by link id: %v", err)
	}
	if doc.Content != "T1\nD1" {
		t.Errorf("got content %q, want %q", doc.Content, "T1\nD1")
	}
	if doc.Metadata["source"] != "YNet" || doc.Metadata["lean"] != "left" {
		t.Errorf("got metadata %v, want source YNet and lean left", doc.Metadata)
	}

	doc, err = database.articles.GetByID(ctx, items[1].id())
	if err != nil {
		t.Fatalf("get by fallback id: %v", err)
	}
	if doc.Content != "T2\nD2" {
		t.Errorf("got content %q, want %q", doc.Content, "T2\nD2")
	}
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

	if err := database.SaveArticles(ctx, items); err != nil {
		t.Fatalf("save: %v", err)
	}

	articles, err := database.QueryArticles(
		ctx,
		"anything",
		10,
		map[string]string{"lean": string(leftLean)},
	)
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	if len(articles) != 2 {
		t.Fatalf("got %d articles, want 2", len(articles))
	}
	var ynet *Article
	for i, article := range articles {
		if article.Lean != leftLean {
			t.Errorf("got lean %q, want %q", article.Lean, leftLean)
		}
		if article.Source == "YNet" {
			ynet = &articles[i]
		}
	}
	if ynet == nil {
		t.Fatal("YNet article not found in results")
	}
	if ynet.Title != "T1" || ynet.Description != "D1" ||
		ynet.Link != "https://example.com/1" {
		t.Errorf("got reconstructed article %+v, want the YNet item", *ynet)
	}
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

	if err := database.SaveArticles(ctx, items); err != nil {
		t.Fatalf("save: %v", err)
	}

	removed, err := database.DeleteOldArticles(ctx, 48*time.Hour)
	if err != nil {
		t.Fatalf("delete old: %v", err)
	}
	if removed == 0 {
		t.Fatalf("got %d removed articles, want at least 1", removed)
	}

	if _, err := database.articles.GetByID(ctx, "https://example.com/old"); err == nil {
		t.Errorf("old article was not removed")
	}
	if _, err := database.articles.GetByID(ctx, "https://example.com/new"); err != nil {
		t.Errorf("recent article was removed: %v", err)
	}
}
