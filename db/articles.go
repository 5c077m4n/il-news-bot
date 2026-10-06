package db

import (
	"context"
	"fmt"
	"hash/fnv"
	"runtime"
	"strings"
	"time"

	"github.com/philippgille/chromem-go"
)

type Lean string

const (
	LeanLeft    Lean = "left"
	LeanRight   Lean = "right"
	LeanNeutral Lean = "neutral"
)

type Article struct {
	Source      string    `json:"source"`
	Lean        Lean      `json:"lean"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	PublishedAt time.Time `json:"publishedAt"`
}

func (a Article) content() string {
	return strings.TrimSpace(a.Title + "\n" + a.Description)
}

func (a Article) id() string {
	if a.Link != "" {
		return a.Link
	}

	hash := fnv.New64a()
	for _, part := range []string{a.Source, a.Title, a.Description} {
		_, _ = fmt.Fprintln(hash, part)
	}
	return fmt.Sprintf("%s:%x", a.Source, hash.Sum64())
}

func (a Article) metadata() map[string]string {
	return map[string]string{
		"source":      a.Source,
		"lean":        string(a.Lean),
		"title":       a.Title,
		"description": a.Description,
		"link":        a.Link,
		"publishedAt": a.PublishedAt.Format(time.RFC3339),
	}
}

func from(result chromem.Result) Article {
	publishedAt, err := time.Parse(time.RFC3339, result.Metadata["publishedAt"])
	if err != nil {
		publishedAt = time.Time{}
	}

	return Article{
		Source:      result.Metadata["source"],
		Lean:        Lean(result.Metadata["lean"]),
		Title:       result.Metadata["title"],
		Description: result.Metadata["description"],
		Link:        result.Metadata["link"],
		PublishedAt: publishedAt,
	}
}

func QueryArticles(
	ctx context.Context,
	query string,
	limit int,
	where map[string]string,
) ([]Article, error) {
	instance, err := database()
	if err != nil {
		return nil, err
	}

	count := instance.articles.Count()
	if count == 0 {
		return nil, nil
	}

	results, err := instance.articles.Query(ctx, query, min(limit, count), where, nil)
	if err != nil {
		return nil, err
	}

	articles := make([]Article, 0, len(results))
	for _, result := range results {
		articles = append(articles, from(result))
	}

	return articles, nil
}

func SaveArticles(ctx context.Context, items []Article) error {
	if len(items) == 0 {
		return nil
	}

	instance, err := database()
	if err != nil {
		return err
	}

	documents := make([]chromem.Document, 0, len(items))
	for _, item := range items {
		content := item.content()
		if content == "" {
			continue
		}
		documents = append(documents, chromem.Document{
			ID:       item.id(),
			Metadata: item.metadata(),
			Content:  content,
		})
	}
	if len(documents) == 0 {
		return nil
	}

	return instance.articles.AddDocuments(ctx, documents, runtime.NumCPU())
}
