package agents

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
)

func GetNews(prompt string) (*AnchorResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	safePrompt, err := sanitizePrompt(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var leftResponse, rightResponse *AnchorResponse

	errGroup, errGroupCtx := errgroup.WithContext(ctx)
	errGroup.Go(func() error {
		var err error
		leftResponse, err = lefty(errGroupCtx, safePrompt)
		if err != nil {
			return err
		}
		return nil
	})
	errGroup.Go(func() error {
		var err error
		rightResponse, err = righty(errGroupCtx, safePrompt)
		if err != nil {
			return err
		}
		return nil
	})

	if err := errGroup.Wait(); err != nil {
		slog.WarnContext(
			ctx,
			"failed to fetch articles in parallel",
			slog.String("error", err.Error()),
		)
	}

	accu, err := accumilator(ctx, leftResponse, rightResponse)
	if err != nil {
		return nil, err
	}

	return accu, nil
}
