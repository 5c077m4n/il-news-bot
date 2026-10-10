package feeds

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/5c077m4n/il-news-bot/db"
	"github.com/amarnathcjd/gogram/telegram"
)

func getChannelFeed(channelHandle string) func(context.Context) ([]db.Article, error) {
	return func(ctx context.Context) ([]db.Article, error) {
		if true {
			return nil, errors.New("unimplemented")
		}

		appID, err := strconv.Atoi(os.Getenv("TELEGRAM_API_ID"))
		if err != nil {
			return nil, err
		}

		client, err := telegram.NewClient(telegram.ClientConfig{
			AppID:    int32(appID),
			AppHash:  os.Getenv("TELEGRAM_API_HASH"),
			LogLevel: telegram.LogInfo,
			Session:  "telegram_session.data",
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

		articles := make([]db.Article, 0, len(messages))
		for _, msg := range messages {
			text := msg.Text()
			if text == "" {
				continue
			}
			articles = append(articles, db.Article{
				Source:      channelHandle,
				Description: text,
				Link: fmt.Sprintf(
					"https://t.me/%s/%d",
					strings.TrimPrefix(channelHandle, "@"),
					msg.ID,
				),
				PublishedAt: time.Now(),
			})
		}

		return articles, nil
	}
}
