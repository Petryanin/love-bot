// internal/app/app.go
package app

import (
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/services"
)

type AppContext struct {
	Cfg *config.Config

	RelationshipService    *services.RelationshipService
	ComplimentService      *services.ComplimentService
	ImageComplimentService *services.ImageComplimentService
	PlanService            services.Planner
	SessionManager         *services.SessionManager
	WeatherService         *services.WeatherService
	DateTimeService        *services.DateTimeService
	MagicBallService       *services.MagicBallService
}
