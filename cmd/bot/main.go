package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/Petryanin/love-bot/internal/bot"
	"github.com/Petryanin/love-bot/internal/clients"
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/db"
	"github.com/Petryanin/love-bot/internal/scheduler"
	"github.com/Petryanin/love-bot/internal/services"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatal("main: failed to load configuration: %w", err)
	}

	database, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		log.Fatal("main: failed to open db: %w", err)
	}

	appCtx := &app.AppContext{
		Cfg: cfg,

		RelationshipService:    services.NewRelationshipService(cfg.DatingStartDate.In(cfg.DatingStartTZ)),
		ComplimentService:      services.NewComplimentService(cfg.ComplimentsFilePath),
		ImageComplimentService: services.NewImageComplimentService(clients.NewCatAASClient(cfg.CatAPIURL), cfg.FontPath),
		SessionManager:         services.NewSessionManager(),
		WeatherService:         services.NewWeatherService(clients.NewOpenWeatherMapClient(cfg.WeatherAPIURL, cfg.WeatherAPIKey, cfg.WeatherAPILang), cfg.WeatherAPICity),
		DateTimeService:        services.NewDateTimeService(clients.NewDucklingClient(cfg.DucklingAPIURL, cfg.DucklingLocale)),
		MagicBallService:       services.NewMagicBallService(cfg.MagicBallImagesPath),
		GeoService:             services.NewGeoService(clients.NewGeoNamesClient(cfg.GeoNamesAPIURL, cfg.GeoNamesAPIUsername, cfg.GeoNamesAPILang)),

		PlanService: db.NewPlanService(database),
		UserService: db.NewUserManager(database),
	}

	bot := bot.CreateBot(appCtx)

	log.Print("starting plan scheduler...")
	scheduler.StartPlanScheduler(ctx, bot, appCtx, time.Second*5)

	log.Print("starting bot...")
	bot.Start(ctx)
}
