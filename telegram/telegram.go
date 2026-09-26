// Package telegram that holds the telegram init logic
package telegram

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/5c077m4n/il-news-bot/agents"
	"github.com/amarnathcjd/gogram/telegram"
)

func handleNewsRequest(message *telegram.NewMessage, prompt string) error {
	if _, err := message.Reply(
		"Fetching news...",
	); err != nil {
		slog.Error("could not send message", slog.String("error", err.Error()))
	}

	response, err := agents.GetNews(prompt)
	if err != nil {
		if _, err := message.Reply(
			"Sorry, couldn't fetch your news just now...",
		); err != nil {
			slog.Error("could not send message", slog.String("error", err.Error()))
		}
		return err
	}
	slog.Info("fetched news successfully")

	if message.Sender != nil {
		slog.Info(
			"sending resposne to user",
			slog.String("username", message.Sender.Username),
		)
	}
	if _, err := message.Reply(response.String()); err != nil {
		return err
	}
	return nil
}

func handleMessage(message *telegram.NewMessage) error {
	command, args, _ := strings.Cut(strings.TrimSpace(message.Text()), " ")
	args = strings.TrimSpace(args)

	switch command {
	case "/start":
		return handleNewsRequest(message, "Please get me the lastest news")
	case "/subject":
		if args == "" {
			if _, err := message.Reply(
				"Please provide a subject, e.g. `/subject politics`",
			); err != nil {
				slog.Error("could not send message", slog.String("error", err.Error()))
			}
			return nil
		}
		return handleNewsRequest(
			message,
			fmt.Sprintf("Please get me the lastest news about %s", args),
		)
	}
	return nil
}

func Run() error {
	client, err := getBotClient()
	if err != nil {
		return err
	}

	client.On(telegram.OnMessage, handleMessage)
	if _, err := client.BotsSetBotCommands(
		&telegram.BotCommandScopeDefault{},
		"en",
		[]*telegram.BotCommand{
			{Command: "start", Description: "Get the latest news"},
			{Command: "subject", Description: "Get the latest news about a subject"},
		},
	); err != nil {
		slog.Warn("could not register bot commands", "error", err)
	}

	client.Idle()
	return nil
}
