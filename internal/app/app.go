// internal/app/app.go
package app

import (
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/db"
	"github.com/Petryanin/love-bot/internal/services"
)

type App struct {
	Cfg *config.Config

	Relationship    *services.RelationshipService
	Compliment      *services.ComplimentService
	ImageCompliment *services.ImageComplimentService
	Session         *services.SessionManager
	Weather         *services.WeatherService
	DateTime        *services.DateTimeService
	MagicBall       *services.MagicBallService
	Geo             *services.GeoService

	Plan db.Planner
	User db.UserManager
}
