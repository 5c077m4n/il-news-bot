package agents

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

func GetNewsAritcles(prompt string) ([]*NewsItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	safePrompt, err := sanitizePrompt(ctx, prompt)
	if err != nil {
		return nil, err
	}

	leftArticlesChan := make(chan []*NewsItem)
	rightArticlesChan := make(chan []*NewsItem)

	wg := sync.WaitGroup{}
	wg.Go(func() {
		newsArticles, err := lefty(ctx, safePrompt)
		if err != nil {
			slog.WarnContext(
				ctx,
				"couldn't fetch articles",
				slog.String("side", "left"),
				slog.String("error", err.Error()),
			)
			return
		}
		leftArticlesChan <- newsArticles
		close(leftArticlesChan)
	})
	wg.Go(func() {
		newsArticles, err := righty(ctx, safePrompt)
		if err != nil {
			slog.WarnContext(
				ctx,
				"couldn't fetch articles",
				slog.String("side", "right"),
				slog.String("error", err.Error()),
			)
			return
		}
		rightArticlesChan <- newsArticles
		close(leftArticlesChan)
	})
	wg.Wait()

	allArticles, err := accumilator(ctx, <-leftArticlesChan, <-rightArticlesChan)
	if err != nil {
		return nil, err
	}

	return allArticles, nil
}
