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

var (
	GetIsrealHayom = getRSSFeed(
		"Israel Hayom",
		db.LeanRight,
		"https://www.israelhayom.co.il/rss.xml",
	)
	GetYNet = getRSSFeed(
		"YNet",
		db.LeanLeft,
		"https://www.ynet.co.il/Integration/StoryRss2.xml",
	)
	GetJPost = getRSSFeed(
		"JPost",
		db.LeanRight,
		"https://www.jpost.com/rss/rssfeedsfrontpage.aspx",
	)
	GetMakorRishon = getRSSFeed(
		"Makor Rishon",
		db.LeanRight,
		"https://www.makorrishon.co.il/feed/",
	)
	GetCyberNews = getRSSFeed(
		"Cyber News",
		db.LeanNeutral,
		"https://rss.app/feeds/Ho4glVhEXQwiloOx.xml",
	)
	GetAbuAliExpress = getChannelFeed("Abu Ali Express", db.LeanRight, "@abualiexpress")
)

var allSources = map[string]func(context.Context) ([]db.Article, error){
	"Israel Hayom":    GetIsrealHayom,
	"YNet":            GetYNet,
	"JPost":           GetJPost,
	"Makor Rishon":    GetMakorRishon,
	"Cyber News":      GetCyberNews,
	"Abu Ali Express": GetAbuAliExpress,
}

func refresh(ctx context.Context) {
	start := time.Now()
	defer func() {
		slog.Info("fetching data sources done", "elapsed", time.Since(start))
	}()

	for name, getter := range allSources {
		if _, err := getter(ctx); err != nil {
			slog.WarnContext(
				ctx,
				"could not refresh data source",
				slog.String("source", name),
				slog.Any("error", err),
			)
		}
	}
}

func Cleanup(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			removed, err := db.DeleteOldArticles(ctx, retention)
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

func Poll(ctx context.Context) {
	refresh(ctx)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			refresh(ctx)
		}
	}
}
