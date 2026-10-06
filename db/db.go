package db

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/philippgille/chromem-go"
)

const DefaultDirectory = "chromem_data"
const openRouterBaseURL = "https://openrouter.ai/api/v1"
const embeddingModel = "openai/text-embedding-3-small"

type Database struct {
	articles *chromem.Collection
}

func defaultEmbeddingFunc() chromem.EmbeddingFunc {
	return chromem.NewEmbeddingFuncOpenAICompat(
		openRouterBaseURL,
		os.Getenv("OPENROUTER_API_KEY"),
		embeddingModel,
		new(true),
	)
}

func (d *Database) QueryArticles(
	ctx context.Context,
	query string,
	limit int,
	where map[string]string,
) ([]Article, error) {
	count := d.articles.Count()
	if count == 0 {
		return nil, nil
	}

	results, err := d.articles.Query(ctx, query, min(limit, count), where, nil)
	if err != nil {
		return nil, err
	}

	articles := make([]Article, 0, len(results))
	for _, result := range results {
		articles = append(articles, from(result))
	}

	return articles, nil
}

func (d *Database) SaveArticles(ctx context.Context, items []Article) error {
	if len(items) == 0 {
		return nil
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

	return d.articles.AddDocuments(ctx, documents, runtime.NumCPU())
}

func (d *Database) DeleteOldArticles(ctx context.Context, maxAge time.Duration) (int, error) {
	count := d.articles.Count()
	if count == 0 {
		return 0, nil
	}

	results, err := d.articles.QueryEmbedding(ctx, []float32{1}, count, nil, nil)
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-maxAge)
	ids := make([]string, 0, len(results))
	for _, result := range results {
		if from(result).PublishedAt.Before(cutoff) {
			ids = append(ids, result.ID)
		}
	}
	if len(ids) == 0 {
		return 0, nil
	}

	if err := d.articles.Delete(ctx, nil, nil, ids...); err != nil {
		return 0, err
	}

	return len(ids), nil
}

func New(directory string, embeddingFunc chromem.EmbeddingFunc) (*Database, error) {
	instance, err := chromem.NewPersistentDB(directory, false)
	if err != nil {
		return nil, err
	}

	if embeddingFunc == nil {
		embeddingFunc = defaultEmbeddingFunc()
	}
	articles, err := instance.GetOrCreateCollection("articles", nil, embeddingFunc)
	if err != nil {
		return nil, err
	}

	return &Database{articles: articles}, nil
}
