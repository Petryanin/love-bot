package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Petryanin/love-bot/internal/clients"
)

type DateTimeService struct {
	client clients.DucklingParser
	months map[time.Month]string
	days   map[time.Weekday]string
}

func NewDateTimeService(client clients.DucklingParser) *DateTimeService {
	return &DateTimeService{
		client: client,
		months: map[time.Month]string{
			time.January:   "января",
			time.February:  "февраля",
			time.March:     "марта",
			time.April:     "апреля",
			time.May:       "мая",
			time.June:      "июня",
			time.July:      "июля",
			time.August:    "августа",
			time.September: "сентября",
			time.October:   "октября",
			time.November:  "ноября",
			time.December:  "декабря",
		},
		days: map[time.Weekday]string{
			time.Monday:    "Пн",
			time.Tuesday:   "Вт",
			time.Wednesday: "Ср",
			time.Thursday:  "Чт",
			time.Friday:    "Пт",
			time.Saturday:  "Сб",
			time.Sunday:    "Вс",
		},
	}
}

func (s *DateTimeService) Parse(
	ctx context.Context,
	text string,
	ref time.Time,
	tz string,
) (string, time.Time, error) {
	DTItems, err := s.client.Parse(ctx, text, ref, tz)
	if err != nil {
		return text, time.Time{}, err
	}

	for _, it := range DTItems {
		if it.Dim == "time" {
			if value, ok := it.Value.Value.(string); ok {
				dt, err := time.Parse(time.RFC3339Nano, value)
				if err != nil {
					log.Print("dt service: wrong datetime format: %w", err)
					return text, time.Time{}, err
				}
				return it.Body, dt, nil
			}
		}
	}

	return text, time.Time{}, fmt.Errorf("dt service: failed to parse datetime %q", text)
}

// FormatRu возвращает строку вида:
//   - если год = текущий:
//     "27 мая (Вт), 12:00"
//   - если год ≠ текущий:
//     "27 мая (Ср) 2026 в 12:00"
func (s *DateTimeService) FormatRu(t time.Time) string {
	day := t.Day()
	monthName := s.months[t.Month()]
	weekday := s.days[t.Weekday()]
	timePart := t.Format("15:04")
	currentYear := time.Now().Year()

	if t.Year() == currentYear {
		return fmt.Sprintf("%d %s (%s), %s", day, monthName, weekday, timePart)
	}
	return fmt.Sprintf("%d %s (%s) %d в %s", day, monthName, weekday, t.Year(), timePart)
}
