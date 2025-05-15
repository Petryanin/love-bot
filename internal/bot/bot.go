package bot

import (
	"log"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/handlers"
	"github.com/go-telegram/bot"
)

func CreateBot(appCtx *app.AppContext) *bot.Bot {
	log.Print("creating bot...")
	b, err := bot.New(appCtx.Cfg.TgToken)
	if err != nil {
		log.Fatal(err)
	}

	registerHandlers(appCtx, b)

	return b
}

func registerHandlers(appCtx *app.AppContext, b *bot.Bot) {
	log.Print("registering handlers...")
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.StartCommand,
		bot.MatchTypeCommand,
		handlers.StartHandler,
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.HelpCommand,
		bot.MatchTypeCommand,
		handlers.HelpHandler,
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.WeatherButton,
		bot.MatchTypeExact,
		handlers.WeatherHandler(appCtx),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.TogetherTimeButton,
		bot.MatchTypeExact,
		handlers.TogetherTimeHandler(appCtx),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.ComplimentButton,
		bot.MatchTypeExact,
		handlers.ComplimentImageHandler(appCtx),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"",
		bot.MatchTypePrefix,
		handlers.PlansHandler(appCtx),
	)
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"plan:",
		bot.MatchTypePrefix,
		handlers.PlansDetailsHandler(appCtx),
	)
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"plan_delete:",
		bot.MatchTypePrefix,
		handlers.PlansDeleteHandler(appCtx),
	)
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"plan_list",
		bot.MatchTypeExact,
		handlers.PlansListHandler(appCtx),
	)
}
