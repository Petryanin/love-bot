package bot

import (
	"log"

	"github.com/Petryanin/love-bot/internal/clients"
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/handlers"
	"github.com/Petryanin/love-bot/internal/services"
	"github.com/go-telegram/bot"
)

func CreateBot(cfg *config.Config) *bot.Bot {
	log.Print("creating bot...")
	b, err := bot.New(cfg.TgToken)
	if err != nil {
		log.Fatal(err)
	}

	registerHandlers(cfg, b)

	return b
}

func registerHandlers(cfg *config.Config, b *bot.Bot) {
	log.Print("initializing services...")
	weatherService := services.NewWeatherService(
		clients.NewOpenWeatherMapClient(cfg.WeatherAPIURL, cfg.WeatherAPIKey),
		cfg.WeatherAPICity,
	)
	relationshipService := services.NewRelationshipService(
		cfg.DatingStartDate.In(cfg.DatingStartTZ),
	)
	complimentService := services.NewComplimentService()
	imgService := services.NewImageComplimentService(
		clients.NewCatAASClient(cfg.CatAPIURL),
		cfg.FontPath,
	)

	planService, err := services.NewPlanService(cfg.DBPath, 369618248)
	if err != nil {
		panic(err)
	}
	sessionManager := services.NewSessionManager()

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
		handlers.WeatherHandler(weatherService),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.TogetherTimeButton,
		bot.MatchTypeExact,
		handlers.TogetherTimeHandler(relationshipService),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		config.ComplimentButton,
		bot.MatchTypeExact,
		handlers.ComplimentImageHandler(imgService, complimentService),
	)
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"",
		bot.MatchTypePrefix,
		handlers.PlansHandler(cfg, sessionManager, planService),
	)
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"plan:",
		bot.MatchTypePrefix,
		handlers.PlansDetailsHandler(cfg, planService),
	)
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"plan_delete:",
		bot.MatchTypePrefix,
		handlers.PlansDeleteHandler(cfg, planService),
	)
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"plan_list",
		bot.MatchTypeExact,
		handlers.PlansListHandler(cfg, sessionManager, planService),
	)
}
