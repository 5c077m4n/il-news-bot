package db

import (
	"context"
	"testing"
)

func TestSaveArticles(t *testing.T) {
	directory = t.TempDir()
	embeddingFunc = func(context.Context, string) ([]float32, error) {
		return []float32{1}, nil
	}

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

	if err := SaveArticles(ctx, items); err != nil {
		t.Fatalf("save: %v", err)
	}

	instance, err := database()
	if err != nil {
		t.Fatalf("database: %v", err)
	}

	doc, err := instance.articles.GetByID(ctx, "https://example.com/1")
	if err != nil {
		t.Fatalf("get by link id: %v", err)
	}
	if doc.Content != "T1\nD1" {
		t.Errorf("got content %q, want %q", doc.Content, "T1\nD1")
	}
	if doc.Metadata["source"] != "YNet" || doc.Metadata["lean"] != "left" {
		t.Errorf("got metadata %v, want source YNet and lean left", doc.Metadata)
	}

	doc, err = instance.articles.GetByID(ctx, items[1].id())
	if err != nil {
		t.Fatalf("get by fallback id: %v", err)
	}
	if doc.Content != "T2\nD2" {
		t.Errorf("got content %q, want %q", doc.Content, "T2\nD2")
	}
}

func TestQueryArticles(t *testing.T) {
	directory = t.TempDir()
	embeddingFunc = func(context.Context, string) ([]float32, error) {
		return []float32{1}, nil
	}

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

	if err := SaveArticles(ctx, items); err != nil {
		t.Fatalf("save: %v", err)
	}

	articles, err := QueryArticles(
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
