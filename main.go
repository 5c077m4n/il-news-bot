package main

import (
	"context"

	"github.com/5c077m4n/il-news-bot/db"
	"github.com/5c077m4n/il-news-bot/feeds"
	"github.com/5c077m4n/il-news-bot/telegram"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	database, err := db.New(db.DefaultDirectory, nil)
	if err != nil {
		panic(err)
	}

	go feeds.Poll(context.Background(), database)
	go database.PollCleanup(context.Background())

	if err := telegram.Run(database); err != nil {
		panic(err)
	}
}
