// Package feeds holds the data fetching functions
package feeds

import (
	"context"
	"log/slog"
	"time"

	"github.com/5c077m4n/il-news-bot/db"
)

const (
	pollInterval    = 15 * time.Minute
	cleanupInterval = time.Hour
	retention       = 24 * time.Hour
)

var allSources = map[string]func(context.Context) ([]db.Article, error){
	"Israel Hayom":           getRSSFeed("https://www.israelhayom.co.il/rss.xml"),
	"YNet":                   getRSSFeed("https://www.ynet.co.il/Integration/StoryRss2.xml"),
	"JPost":                  getRSSFeed("https://www.jpost.com/rss/rssfeedsfrontpage.aspx"),
	"Cyber Security News IL": getRSSFeed("https://rss.app/feeds/Ho4gIVhEXQwiIoOx.xml"),
	"Cyber Security News":    getChannelFeed("@CyberSecurityIL"),
	"Abu Ali Express":        getChannelFeed("@abualiexpress"),
	"Hacker News Feed":       getChannelFeed("@hacker_news_feed"),
	"Amit Segal":             getChannelFeed("@hacker_news_feed"),
	"Lobsters":               getChannelFeed("@lobste_rs"),
}

func refresh(ctx context.Context, database *db.Database) {
	start := time.Now()
	defer func() {
		slog.Info("fetching data sources done", "elapsed", time.Since(start))
	}()

	for name, getter := range allSources {
		articles, err := getter(ctx)
		if err != nil {
			slog.WarnContext(
				ctx,
				"could not refresh data source",
				slog.String("source", name),
				slog.Any("error", err),
			)
			continue
		}

		if err := database.SaveArticles(ctx, articles); err != nil {
			slog.WarnContext(
				ctx,
				"could not save articles",
				slog.String("source", name),
				slog.Any("error", err),
			)
		}
	}
}

func Cleanup(ctx context.Context, database *db.Database) {
	start := time.Now()
	defer func() {
		slog.Info("DB cleanup done", "elapsed", time.Since(start))
	}()

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			removed, err := database.DeleteOldArticles(ctx, retention)
			if err != nil {
				slog.WarnContext(
					ctx,
					"could not remove old articles",
					slog.Any("error", err),
				)
				continue
			}
			slog.InfoContext(
				ctx,
				"removed old articles",
				slog.Int("removed", removed),
			)
		}
	}
}

func Poll(ctx context.Context, database *db.Database) {
	refresh(ctx, database)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			refresh(ctx, database)
		}
	}
}
