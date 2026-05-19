package feeds

import (
	"os"
	"strconv"

	"github.com/amarnathcjd/gogram/telegram"
)

func getClient() (*telegram.Client, error) {
	appID, err := strconv.Atoi(os.Getenv("TELEGRAM_API_ID"))
	if err != nil {
		return nil, err
	}

	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:    int32(appID),
		AppHash:  os.Getenv("TELEGRAM_API_HASH"),
		LogLevel: telegram.LogInfo,
		Session:  "session.data",
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

func FetchChannelMessages(channelHandle string) ([]string, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}
	if _, err := client.Login(os.Getenv("TELEGRAM_PHONE_NUMBER")); err != nil {
		return nil, err
	}

	messages, err := client.GetMessages(channelHandle, &telegram.SearchOption{Limit: 70})
	if err != nil {
		return nil, err
	}

	results := make([]string, 0, len(messages))
	for _, msg := range messages {
		results = append(results, msg.Text())
	}
	return results, nil
}
