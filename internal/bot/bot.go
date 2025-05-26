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
		config.StartCmd,
		bot.MatchTypeCommand,
		handlers.StartHandler(appCtx),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.HelpCmd,
		bot.MatchTypeCommand,
		handlers.HelpHandler,
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.WeatherBtn,
		bot.MatchTypeExact,
		handlers.WeatherHandler(appCtx),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.TogetherTimeBtn,
		bot.MatchTypeExact,
		handlers.TogetherTimeHandler(appCtx),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.ComplimentBtn,
		bot.MatchTypeExact,
		handlers.ComplimentImageHandler(appCtx),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.MagicBallBtn,
		bot.MatchTypeExact,
		handlers.MagicBallHandler(appCtx),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"",
		bot.MatchTypePrefix,
		handlers.StateRootHandler(appCtx),
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
		"plans",
		bot.MatchTypePrefix,
		handlers.PlansListHandler(appCtx),
	)
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"remind:change:",
		bot.MatchTypePrefix,
		handlers.PlansChangeRemindTimeHandler(appCtx),
	)
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"remind:",
		bot.MatchTypePrefix,
		handlers.PlansRemindHandler(nil, appCtx),
	)
}
