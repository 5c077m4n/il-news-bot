package feeds

import (
	"context"
	"time"

	"github.com/5c077m4n/il-news-bot/db"
	"github.com/mmcdole/gofeed"
)

func publishedAt(item *gofeed.Item) time.Time {
	if item.PublishedParsed != nil {
		return *item.PublishedParsed
	}
	if item.UpdatedParsed != nil {
		return *item.UpdatedParsed
	}
	return time.Time{}
}

func getRSSFeed(url string) func(context.Context) ([]db.Article, error) {
	return func(ctx context.Context) ([]db.Article, error) {
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
				Source:      url,
				Title:       item.Title,
				Description: description,
				Link:        item.Link,
				PublishedAt: publishedAt(item),
			})
		}

		return articles, nil
	}
}
