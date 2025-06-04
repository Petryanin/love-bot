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

	app := &app.App{
		Cfg: cfg,

		Relationship:    services.NewRelationshipService(cfg.DatingStartDate.In(cfg.DatingStartTZ)),
		Compliment:      services.NewComplimentService(cfg.ComplimentsFilePath),
		ImageCompliment: services.NewImageComplimentService(clients.NewCatAASClient(cfg.CatAPIURL), cfg.FontPath),
		Session:         services.NewSessionManager(),
		Weather:         services.NewWeatherService(clients.NewOpenWeatherMapClient(cfg.WeatherAPIURL, cfg.WeatherAPIKey, cfg.WeatherAPILang), cfg.WeatherAPICity),
		DateTime:        services.NewDateTimeService(clients.NewDucklingClient(cfg.DucklingAPIURL, cfg.DucklingLocale)),
		MagicBall:       services.NewMagicBallService(cfg.MagicBallImagesPath),
		Geo:             services.NewGeoService(clients.NewGeoNamesClient(cfg.GeoNamesAPIURL, cfg.GeoNamesAPIUsername, cfg.GeoNamesAPILang)),

		Plan: db.NewPlanService(database),
		User: db.NewUserManager(database),
	}

	bot := bot.CreateBot(app)

	log.Print("starting schedulers...")
	scheduler.StartPlanScheduler(ctx, bot, app, time.Second*5)
	scheduler.StartCatScheduler(ctx, bot, app, time.Minute)
	scheduler.StartCleanupScheduler(ctx, app, time.Hour*2, time.Hour*24)

	log.Print("starting bot...")
	bot.Start(ctx)
}
