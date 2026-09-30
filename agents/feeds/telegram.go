package feeds

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/5c077m4n/il-news-bot/db"
	"github.com/amarnathcjd/gogram/telegram"
)

func getChannelFeed(channelHandle string) func(context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		key := fmt.Appendf(nil, "telegram:%s", channelHandle)
		if cached, createdAt, err := db.Get[[]string](ctx, key); err == nil {
			if time.Since(createdAt) < time.Hour {
				return strings.Join(*cached, "\n"), nil
			}
		}

		appID, err := strconv.Atoi(os.Getenv("TELEGRAM_API_ID"))
		if err != nil {
			return "", err
		}

		client, err := telegram.NewClient(telegram.ClientConfig{
			AppID:    int32(appID),
			AppHash:  os.Getenv("TELEGRAM_API_HASH"),
			LogLevel: telegram.LogInfo,
			Session:  "telegram_session.data",
		})
		if err != nil {
			return "", err
		}

		if _, err := client.Login(os.Getenv("TELEGRAM_PHONE_NUMBER")); err != nil {
			return "", err
		}

		messages, err := client.GetMessages(
			channelHandle,
			&telegram.SearchOption{Context: ctx, Limit: 70},
		)
		if err != nil {
			return "", err
		}

		results := make([]string, 0, len(messages))
		for _, msg := range messages {
			results = append(results, msg.Text())
		}

		if err := db.Set(ctx, key, results); err != nil {
			slog.WarnContext(
				ctx,
				"could not set value in cache",
				slog.String("key", string(key)),
				slog.String("error", err.Error()),
			)
		}

		return strings.Join(results, "\n"), nil
	}
}
