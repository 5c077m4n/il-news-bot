package feeds

import (
	"context"
	"log/slog"
	"time"

	"github.com/5c077m4n/il-news-bot/db"
	"github.com/mmcdole/gofeed"
)

func getRSSFeed(
	source string,
	lean db.Lean,
	url string,
) func(*db.Database, context.Context) ([]db.Article, error) {
	return func(database *db.Database, ctx context.Context) ([]db.Article, error) {
		parserCtx, parserCancel := context.WithTimeout(ctx, 10*time.Second)
		defer parserCancel()

		feedParser := gofeed.NewParser()
		feed, err := feedParser.ParseURLWithContext(url, parserCtx)
		if err != nil {
			return nil, err
		}

		articles := make([]db.Article, 0, len(feed.Items))
		for _, item := range feed.Items {
			description := item.Description
			if description == "" {
				description = item.Content
			}
			articles = append(articles, db.Article{
				Source:      source,
				Lean:        lean,
				Title:       item.Title,
				Description: description,
				Link:        item.Link,
				PublishedAt: publishedAt(item),
			})
		}

		if err := database.SaveArticles(ctx, articles); err != nil {
			slog.WarnContext(
				ctx,
				"could not save articles",
				slog.String("source", source),
				slog.Any("error", err),
			)
		}

		return articles, nil
	}
}

func publishedAt(item *gofeed.Item) time.Time {
	if item.PublishedParsed != nil {
		return *item.PublishedParsed
	}
	if item.UpdatedParsed != nil {
		return *item.UpdatedParsed
	}
	return time.Time{}
}
