package agents

import (
	"context"
	"log/slog"
	"time"

	"github.com/5c077m4n/il-news-bot/db"
	"golang.org/x/sync/errgroup"
)

func GetNews(database *db.Database, prompt string) (*AnchorResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	language, err := sanitizePrompt(ctx, prompt)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "detected prompt language", slog.String("language", language))

	var leftResponse, rightResponse *AnchorResponse

	errGroup, errGroupCtx := errgroup.WithContext(ctx)
	errGroup.Go(func() error {
		resp, err := lefty(errGroupCtx, database, prompt)
		if err != nil {
			return err
		}

		leftResponse = resp
		return nil
	})
	errGroup.Go(func() error {
		resp, err := righty(errGroupCtx, database, prompt)
		if err != nil {
			return err
		}

		rightResponse = resp
		return nil
	})

	if err := errGroup.Wait(); err != nil {
		slog.WarnContext(
			ctx,
			"failed to fetch articles in parallel",
			slog.String("error", err.Error()),
		)
	}

	accu, err := accumilator(ctx, language, leftResponse, rightResponse)
	if err != nil {
		return nil, err
	}

	return accu, nil
}
