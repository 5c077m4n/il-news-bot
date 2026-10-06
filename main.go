package main

import (
	"context"

	"github.com/5c077m4n/il-news-bot/agents/feeds"
	"github.com/5c077m4n/il-news-bot/db"
	"github.com/5c077m4n/il-news-bot/telegram"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	database, err := db.New(db.DefaultDirectory, nil)
	if err != nil {
		panic(err)
	}

	go feeds.Poll(context.Background(), database)
	go feeds.Cleanup(context.Background(), database)

	if err := telegram.Run(database); err != nil {
		panic(err)
	}
}
