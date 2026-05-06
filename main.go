package main

import (
	"github.com/5c077m4n/il-news-bot/telegram"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	if err := telegram.Run(); err != nil {
		panic(err)
	}
}
