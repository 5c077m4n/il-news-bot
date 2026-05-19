package feeds

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/cockroachdb/pebble"
	"github.com/goccy/go-json"
)

type cachedMessages struct {
	Timestamp time.Time `json:"timestamp"`
	Messages  []string  `json:"messages"`
}

func getChannelFeed(channelHandle string) func(context.Context) ([]string, error) {
	return func(ctx context.Context) ([]string, error) {
		db, err := DB()
		if err != nil {
			return nil, err
		}

		key := fmt.Appendf(nil, "telegram:%s", channelHandle)
		if value, closer, err := db.Get(key); err == nil {
			defer func() {
				if err := closer.Close(); err != nil {
					slog.ErrorContext(
						ctx,
						"could not close PebbleDB instance",
						slog.String("error", err.Error()),
					)
				}
			}()

			var cached cachedMessages
			if err := json.UnmarshalContext(ctx, value, &cached); err == nil {
				if time.Since(cached.Timestamp) < time.Hour {
					return cached.Messages, nil
				}
			}
		}

		appID, err := strconv.Atoi(os.Getenv("TELEGRAM_API_ID"))
		if err != nil {
			return nil, err
		}

		client, err := telegram.NewClient(telegram.ClientConfig{
			AppID:    int32(appID),
			AppHash:  os.Getenv("TELEGRAM_API_HASH"),
			LogLevel: telegram.LogInfo,
			Session:  "bot_session.data",
		})
		if err != nil {
			return nil, err
		}

		if _, err := client.Login(os.Getenv("TELEGRAM_PHONE_NUMBER")); err != nil {
			return nil, err
		}

		messages, err := client.GetMessages(
			channelHandle,
			&telegram.SearchOption{Context: ctx, Limit: 70},
		)
		if err != nil {
			return nil, err
		}

		results := make([]string, 0, len(messages))
		for _, msg := range messages {
			results = append(results, msg.Text())
		}

		cache := cachedMessages{Timestamp: time.Now(), Messages: results}
		cacheBytes, err := json.MarshalContext(ctx, cache)
		if err != nil {
			slog.WarnContext(ctx, "could not update cache", slog.String("error", err.Error()))
		} else {
			if err := db.Set(key, cacheBytes, pebble.Sync); err != nil {
				slog.WarnContext(
					ctx,
					"could not set value in cache",
					slog.String("key", string(key)),
				)
			}
		}

		return results, nil
	}
}
