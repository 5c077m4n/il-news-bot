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

var allSources = map[string]func(*db.Database, context.Context) ([]db.Article, error){
	"Israel Hayom":     getRSSFeed("Israel Hayom", "https://www.israelhayom.co.il/rss.xml"),
	"YNet":             getRSSFeed("YNet", "https://www.ynet.co.il/Integration/StoryRss2.xml"),
	"JPost":            getRSSFeed("JPost", "https://www.jpost.com/rss/rssfeedsfrontpage.aspx"),
	"Cyber News":       getChannelFeed("Cyber News", "@CyberSecurityIL"),
	"Abu Ali Express":  getChannelFeed("Abu Ali Express", "@abualiexpress"),
	"Hacker News Feed": getChannelFeed("Hacker News Feed", "@hacker_news_feed"),
	"Amit Segal":       getChannelFeed("Amit Segal", "@hacker_news_feed"),
	"Lobsters":         getChannelFeed("Lobsters", "@lobste_rs"),
}

func refresh(ctx context.Context, database *db.Database) {
	start := time.Now()
	defer func() {
		slog.Info("fetching data sources done", "elapsed", time.Since(start))
	}()

	for name, getter := range allSources {
		if _, err := getter(database, ctx); err != nil {
			slog.WarnContext(
				ctx,
				"could not refresh data source",
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
