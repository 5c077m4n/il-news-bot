package agents

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

func GetNews(prompt string) (*AnchorResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	safePrompt, err := sanitizePrompt(ctx, prompt)
	if err != nil {
		return nil, err
	}

	leftAnchorRespChan := make(chan *AnchorResponse)
	rightAnchorRespChan := make(chan *AnchorResponse)

	wg := sync.WaitGroup{}
	wg.Go(func() {
		response, err := lefty(ctx, safePrompt)
		if err != nil {
			slog.WarnContext(
				ctx,
				"couldn't fetch articles",
				slog.String("side", "left"),
				slog.String("error", err.Error()),
			)
			return
		}
		leftAnchorRespChan <- response
		close(leftAnchorRespChan)
	})
	wg.Go(func() {
		resposne, err := righty(ctx, safePrompt)
		if err != nil {
			slog.WarnContext(
				ctx,
				"couldn't fetch articles",
				slog.String("side", "right"),
				slog.String("error", err.Error()),
			)
			return
		}
		rightAnchorRespChan <- resposne
		close(rightAnchorRespChan)
	})
	wg.Wait()

	accu, err := accumilator(ctx, <-leftAnchorRespChan, <-rightAnchorRespChan)
	if err != nil {
		return nil, err
	}

	return accu, nil
}
