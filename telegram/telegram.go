// Package telegram that holds the telegram init logic
package telegram

import (
	"log/slog"

	"github.com/5c077m4n/il-news-bot/agents"
	"github.com/amarnathcjd/gogram/telegram"
)

func Run() error {
	client, err := getTelegramBotClient()
	if err != nil {
		return err
	}

	client.On(
		telegram.OnMessage,
		func(message *telegram.NewMessage) error {
			if message.Text() == "/start" {
				articles, err := agents.GetNewsAritcles("Please get me the lastest news")
				if err != nil {
					if _, err := message.Reply(
						"Sorry, couldn't fetch your news just now...",
					); err != nil {
						slog.Error("could not send message", slog.String("error", err.Error()))
					}
					return err
				}

				resp := agents.AnchorResponse{List: articles}
				if _, err := message.Reply(resp.String()); err != nil {
					return err
				}
			}
			return nil
		},
	)
	if _, err := client.BotsSetBotCommands(
		&telegram.BotCommandScopeDefault{},
		"en",
		[]*telegram.BotCommand{
			{Command: "start", Description: "Get the latest news"},
		},
	); err != nil {
		slog.Warn("could not register bot commands", "error", err)
	}

	client.Idle()
	return nil
}
