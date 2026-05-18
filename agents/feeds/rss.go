package feeds

import (
	"context"
	"time"

	"github.com/mmcdole/gofeed"
)

func getRSSFeed(url string) func(context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		parserCtx, parserCancel := context.WithTimeout(ctx, 10*time.Second)
		defer parserCancel()

		fp := gofeed.NewParser()
		feed, err := fp.ParseURLWithContext(url, parserCtx)
		if err != nil {
			return "", err
		}

		return feed.String(), nil
	}
}
