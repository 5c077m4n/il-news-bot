package agents

import (
	"context"
	"log/slog"
	"time"

	"github.com/5c077m4n/il-news-bot/db"
)

func GetNews(database *db.Database, prompt string) (*AnchorResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	language, err := sanitizePrompt(ctx, prompt)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "detected prompt language", slog.String("language", language))

	response, err := anchor(ctx, database, prompt)
	if err != nil {
		return nil, err
	}

	return factChecker(ctx, language, response)
}
