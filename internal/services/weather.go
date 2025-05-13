package services

import (
	"fmt"
	"math"
	"strings"

	clients "github.com/Petryanin/love-bot/internal/clients"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type WeatherService struct {
	Client *clients.OpenWeatherMapClient
	City   string
}

func NewWeatherService(client *clients.OpenWeatherMapClient, city string) *WeatherService {
	return &WeatherService{
		Client: client,
		City:   city,
	}
}

func (ws *WeatherService) TodaySummary() (string, error) {
	data, err := ws.Client.CurrentWeather(ws.City)
	if err != nil {
		return "", err
	}

	// Берём первый элемент массива погодных условий
	w := data.Weather[0]
	emoji := weatherEmoji(w.Icon, w.Description)
	temp := int(math.Round(data.Main.Temp))
	feels := int(math.Round(data.Main.FeelsLike))

	summary := fmt.Sprintf(
		"Сейчас в городе %s:\n\n"+
			"%s %s\n"+
			"🌡️ Температура: %d°C (ощущается как %d°C)\n"+
			"💨 Ветер: %.1f м/с\n"+
			"💧 Влажность: %d%%",
		data.Name,
		emoji, cases.Title(language.Russian).String(w.Description),
		temp,
		feels,
		data.Wind.Speed,
		data.Main.Humidity,
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
