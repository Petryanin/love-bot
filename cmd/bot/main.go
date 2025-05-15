package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/Petryanin/love-bot/internal/bot"
	"github.com/Petryanin/love-bot/internal/clients"
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/scheduler"
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

		RelationshipService:    services.NewRelationshipService(cfg.DatingStartDate.In(cfg.DatingStartTZ)),
		ComplimentService:      services.NewComplimentService(),
		ImageComplimentService: services.NewImageComplimentService(clients.NewCatAASClient(cfg.CatAPIURL), cfg.FontPath),
		PlanService:            services.NewPlanService(cfg.DBPath, cfg.TgPartnerCharID),
		SessionManager:         services.NewSessionManager(),
		WeatherService:         services.NewWeatherService(clients.NewOpenWeatherMapClient(cfg.WeatherAPIURL, cfg.WeatherAPIKey), cfg.WeatherAPICity),
		DateTimeService:        services.NewDateTimeService(clients.NewDucklingClient(cfg.DucklingAPIURL, cfg.DucklingLocale, cfg.DucklingTZ)),
	}

	bot := bot.CreateBot(appCtx)

	scheduler.StartPlanScheduler(ctx, bot, appCtx, time.Minute)

	log.Print("starting bot...")
	bot.Start(ctx)
}
