package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/Petryanin/love-bot/internal/bot"
	"github.com/Petryanin/love-bot/internal/config"
)

func main() {
	cfg, err := config.Load(".env")
    if err != nil {
        log.Fatal(err)
    }

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	bot := bot.CreateBot(cfg)

	log.Print("starting bot...")
	bot.Start(ctx)
}
