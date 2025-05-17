package services

import (
	"fmt"
	"strings"
	"time"
)

// RelationshipService хранит точку старта отношений
type RelationshipService struct {
	startDating time.Time
}

// NewRelationshipService создаёт сервис с указанной датой начала
func NewRelationshipService(startDating time.Time) *RelationshipService {
	return &RelationshipService{startDating: startDating}
}

func (rs *RelationshipService) Duration() string {
	now := time.Now()
	// высчитываем годы, месяцы, дни
	y1, m1, d1 := diffYMD(rs.startDating, now)

	together := buildDurationString("*Вы вместе уже:*", y1, m1, d1)

	nextAnn := rs.nextRoundDate(now)
	// разница между сейчас и nextAnn
	y2, m2, d2 := diffYMD(now, nextAnn)
	until := buildDurationString("*До следующей круглой даты осталось:*", y2, m2, d2)

	return together + "\n" + until + " ❤️"
}

// buildDurationString собирает фразу типа "префикс X лет Y месяцев Z дней",
// пропуская те части, которые равны нулю
func buildDurationString(prefix string, years, months, days int) string {
	parts := []string{}

	if years > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", years, Pluralize(years, "год", "года", "лет")))
	}
	if months > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", months, Pluralize(months, "месяц", "месяца", "месяцев")))
	}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", days, Pluralize(days, "день", "дня", "дней")))
	}
	// Если ничего не добавилось — хотя бы вывести "0 дней"
	if len(parts) == 0 {
		parts = append(parts, "всего ничего\\.\\.\\.")
	}

	return prefix + " " + strings.Join(parts, " ")
}

// nextRoundDate возвращает ближайшую дату с днём = 27:
// если сегодня до 27-го включительно, берёт в этом месяце,
// иначе — 27-го следующего месяца.
func (rs *RelationshipService) nextRoundDate(now time.Time) time.Time {
	year, month, day := now.Date()
	loc := now.Location()

	// если сегодня до 27-го (неважно включительно или строго, тут включительно)
	if day <= 27 {
		return time.Date(year, month, 27, 0, 0, 0, 0, loc)
	}
	// иначе — следующий месяц
	nextMonth := month + 1
	nextYear := year
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}
	return time.Date(nextYear, nextMonth, 27, 0, 0, 0, 0, loc)
}

// diffYMD возвращает разницу в годах, месяцах и днях между a и b (b ≥ a)
func diffYMD(a, b time.Time) (years, months, days int) {
	// начнём с простого год/месяц/день
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()

	years = by - ay
	months = int(bm - am)
	days = bd - ad

	// если дни стали отрицательными — «занимаем» месяц
	if days < 0 {
		// сколько дней в предыдущем месяце b?
		prevMonth := b.AddDate(0, -1, 0)
		daysInPrevMonth := daysIn(prevMonth)
		days += daysInPrevMonth
		months--
	}

	// если месяцы стали отрицательными — «занимаем» год
	if months < 0 {
		months += 12
		years--
	}
	return
}

// daysIn возвращает количество дней в месяце t
func daysIn(t time.Time) int {
	y, m, _ := t.Date()
	// первый день следующего месяца
	firstNext := time.Date(y, m+1, 1, 0, 0, 0, 0, t.Location())
	return int(firstNext.Add(-time.Nanosecond).Day())
}
