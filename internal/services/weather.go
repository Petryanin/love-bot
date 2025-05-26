package services

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/Petryanin/love-bot/internal/clients"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type WeatherService struct {
	client clients.WeatherFetcher
	city   string
}

func NewWeatherService(client clients.WeatherFetcher, city string) *WeatherService {
	return &WeatherService{
		client: client,
	}
}

func (ws *WeatherService) TodaySummary(ctx context.Context, city string) (string, error) {
	weather, err := ws.client.Fetch(ctx, city)
	if err != nil {
		return "", err
	}

	emoji := weatherEmoji(weather.Icon, weather.Description)
	temp := int(math.Round(weather.Temp))
	feels := int(math.Round(weather.FeelsLike))

	summary := fmt.Sprintf(
		"Сейчас в городе %s:\n\n"+
			"%s %s\n"+
			"🌡️ Температура: %d°C (ощущается как %d°C)\n"+
			"💨 Ветер: %.1f м/с\n"+
			"💧 Влажность: %d%%",
		weather.City,
		emoji, cases.Title(language.Russian).String(weather.Description),
		temp,
		feels,
		weather.WindSpeed,
		weather.Humidity,
	)
	return summary, nil
}

// weatherEmoji возвращает эмодзи по коду и/или описанию
func weatherEmoji(icon, desc string) string {
	switch {
	case strings.HasPrefix(icon, "01"):
		return "☀️" // ясно
	case strings.HasPrefix(icon, "02"):
		return "🌤️" // малооблачно
	case strings.HasPrefix(icon, "03"), strings.HasPrefix(icon, "04"):
		return "☁️" // облачно
	case strings.HasPrefix(icon, "09"):
		return "🌧️" // морось
	case strings.HasPrefix(icon, "10"):
		return "🌦️" // дождь
	case strings.HasPrefix(icon, "11"):
		return "⛈️" // гроза
	case strings.HasPrefix(icon, "13"):
		return "❄️" // снег
	case strings.HasPrefix(icon, "50"):
		return "🌫️" // туман
	default:
		// как запасной вариант, можно смотреть по описанию
		d := strings.ToLower(desc)
		switch {
		case strings.Contains(d, "rain"):
			return "🌧️"
		case strings.Contains(d, "cloud"):
			return "☁️"
		case strings.Contains(d, "snow"):
			return "❄️"
		case strings.Contains(d, "clear"):
			return "☀️"
		default:
			return "🌈" // универсальный
		}
	}
}
