// internal/app/app.go
package app

import (
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/db"
	"github.com/Petryanin/love-bot/internal/services"
)

type AppContext struct {
	Cfg *config.Config

	RelationshipService    *services.RelationshipService
	ComplimentService      *services.ComplimentService
	ImageComplimentService *services.ImageComplimentService
	SessionManager         *services.SessionManager
	WeatherService         *services.WeatherService
	DateTimeService        *services.DateTimeService
	MagicBallService       *services.MagicBallService
	GeoService             *services.GeoService

	PlanService db.Planner
	UserService db.UserManager
}
