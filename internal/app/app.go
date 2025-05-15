// internal/app/app.go
package app

import (
	"github.com/Petryanin/love-bot/internal/config"
    "github.com/Petryanin/love-bot/internal/clients"
    "github.com/Petryanin/love-bot/internal/services"
)

type AppContext struct {
    Cfg              *config.Config

    WeatherClient    *clients.OpenWeatherMapClient
    DucklingClient   *clients.DucklingClient
    CatClient        *clients.CatAASClient

    RelationshipService       *services.RelationshipService
    ComplimentService    *services.ComplimentService
    ImageComplimentService  *services.ImageComplimentService
    PlanService      *services.PlanService
    SessionManager   *services.SessionManager
    WeatherService   *services.WeatherService
}
