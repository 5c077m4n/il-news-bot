package db

import (
	"context"
	"time"
)

func DeleteOldArticles(ctx context.Context, maxAge time.Duration) (int, error) {
	instance, err := database()
	if err != nil {
		return 0, err
	}

	count := instance.articles.Count()
	if count == 0 {
		return 0, nil
	}

	results, err := instance.articles.QueryEmbedding(ctx, []float32{1}, count, nil, nil)
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

	if err := instance.articles.Delete(ctx, nil, nil, ids...); err != nil {
		return 0, err
	}

	return len(ids), nil
}
