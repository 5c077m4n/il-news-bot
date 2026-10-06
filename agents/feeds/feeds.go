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
	retention       = 48 * time.Hour
)

var allSources = map[string]func(*db.Database, context.Context) ([]db.Article, error){
	"Israel Hayom": getRSSFeed(
		"Israel Hayom",
		db.LeanRight,
		"https://www.israelhayom.co.il/rss.xml",
	),
	"YNet": getRSSFeed(
		"YNet",
		db.LeanLeft,
		"https://www.ynet.co.il/Integration/StoryRss2.xml",
	),
	"JPost": getRSSFeed(
		"JPost",
		db.LeanRight,
		"https://www.jpost.com/rss/rssfeedsfrontpage.aspx",
	),
	"Makor Rishon": getRSSFeed(
		"Makor Rishon",
		db.LeanRight,
		"https://www.makorrishon.co.il/feed/",
	),
	"Cyber News": getRSSFeed(
		"Cyber News",
		db.LeanNeutral,
		"https://rss.app/feeds/Ho4glVhEXQwiloOx.xml",
	),
	"Abu Ali Express": getChannelFeed("Abu Ali Express", db.LeanRight, "@abualiexpress"),
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
