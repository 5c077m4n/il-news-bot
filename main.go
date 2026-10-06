package main

import (
	"context"

	"github.com/5c077m4n/il-news-bot/agents/feeds"
	"github.com/5c077m4n/il-news-bot/telegram"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	go feeds.Poll(context.Background())

	if err := telegram.Run(); err != nil {
		panic(err)
	}
}
