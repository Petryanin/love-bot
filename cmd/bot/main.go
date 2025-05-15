package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/Petryanin/love-bot/internal/bot"
	"github.com/Petryanin/love-bot/internal/clients"
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/services"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	appCtx := &app.AppContext{
		Cfg: cfg,

		WeatherClient:  clients.NewOpenWeatherMapClient(cfg.WeatherAPIURL, cfg.WeatherAPIKey),
		DucklingClient: clients.NewDucklingClient(cfg.DucklingAPIURL, cfg.DucklingLocale, cfg.DucklingTZ),
		CatClient:      clients.NewCatAASClient(cfg.CatAPIURL),

		RelationshipService:    services.NewRelationshipService(cfg.DatingStartDate.In(cfg.DatingStartTZ)),
		ComplimentService:      services.NewComplimentService(),
		ImageComplimentService: services.NewImageComplimentService(clients.NewCatAASClient(cfg.CatAPIURL), cfg.FontPath),
		PlanService:            services.NewPlanService(cfg.DBPath, cfg.TgPartnerCharID),
		SessionManager:         services.NewSessionManager(),
		WeatherService:         services.NewWeatherService(clients.NewOpenWeatherMapClient(cfg.WeatherAPIURL, cfg.WeatherAPIKey), cfg.WeatherAPICity),
	}

	bot := bot.CreateBot(appCtx)

	log.Print("starting bot...")
	bot.Start(ctx)
}
