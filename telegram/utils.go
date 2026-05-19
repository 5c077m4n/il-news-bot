package telegram

import (
	"os"
	"strconv"

	"github.com/amarnathcjd/gogram/telegram"
)

func getBotClient() (*telegram.Client, error) {
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

	if err := client.LoginBot(os.Getenv("TELEGRAM_BOT_TOKEN")); err != nil {
		return nil, err
	}

	return client, nil
}
