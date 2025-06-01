package bot

import (
	"log"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/handlers"
	"github.com/go-telegram/bot"
)

func CreateBot(app *app.App) *bot.Bot {
	log.Print("creating bot...")
	b, err := bot.New(app.Cfg.TgToken)
	if err != nil {
		log.Fatal(err)
	}

	registerHandlers(app, b)

	return b
}

func registerHandlers(app *app.App, b *bot.Bot) {
	log.Print("registering handlers...")
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.StartCmd,
		bot.MatchTypeCommand,
		handlers.StartHandler(app),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.HelpCmd,
		bot.MatchTypeCommand,
		handlers.HelpHandler(app),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.WeatherBtn,
		bot.MatchTypeExact,
		handlers.WeatherHandler(app),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.TogetherTimeBtn,
		bot.MatchTypeExact,
		handlers.TogetherTimeHandler(app),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.ComplimentBtn,
		bot.MatchTypeExact,
		handlers.ComplimentImageHandler(app),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.MagicBallBtn,
		bot.MatchTypeExact,
		handlers.MagicBallHandler(app),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"",
		bot.MatchTypePrefix,
		handlers.StateRootHandler(app),
	)
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"plan:",
		bot.MatchTypePrefix,
		handlers.PlansDetailsHandler(app),
	)
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"plan_delete:",
		bot.MatchTypePrefix,
		handlers.PlansDeleteHandler(app),
	)
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"plans",
		bot.MatchTypePrefix,
		handlers.PlansListHandler(app),
	)
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"remind:change:",
		bot.MatchTypePrefix,
		handlers.PlansChangeRemindTimeHandler(app),
	)
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"remind:",
		bot.MatchTypePrefix,
		handlers.PlansRemindHandler(nil, app),
	)
}
