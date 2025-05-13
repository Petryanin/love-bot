// internal/services/plan.go
package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Petryanin/love-bot/internal/config"
	_ "github.com/mattn/go-sqlite3"
)

type Plan struct {
	ID          int
	ChatID      int64
	Description string
	EventTime   time.Time
	RemindTime  time.Time
}

type PlanService struct {
	db              *sql.DB
	PartnersChatIDs int64
}

func NewPlanService(dbPath string, partnerChatID int64) (*PlanService, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	// создаём таблицу, если нет
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS plan (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chat_id INTEGER NOT NULL,
			description TEXT NOT NULL,
			event_time DATETIME NOT NULL,
			remind_time DATETIME NOT NULL
		)`,
	)
	if err != nil {
		return nil, err
	}
	return &PlanService{db: db, PartnersChatIDs: partnerChatID}, nil
}

func (s *PlanService) Add(p *Plan) error {
	_, err := s.db.Exec(
		`INSERT INTO plan (chat_id, description, event_time, remind_time)
        VALUES(?, ?, ?, ?)`,
		p.ChatID, p.Description, p.EventTime.UTC(), p.RemindTime.UTC(),
	)
	return err
}

func (s *PlanService) GetByID(id int64, cfg *config.Config) (*Plan, error) {
	// подготавливаем SQL
	row := s.db.QueryRow(`
        SELECT id, chat_id, description, event_time, remind_time
        FROM plan
        WHERE id = ?
    `, id)

	// временные контейнеры для строковых дат
	var (
		p         Plan
		eventStr  string
		remindStr string
	)
	// сканируем
	err := row.Scan(
		&p.ID,
		&p.ChatID,
		&p.Description,
		&eventStr,
		&remindStr,
	)
	if err != nil {
		return nil, err
	}

	p.EventTime, err = time.Parse(time.RFC3339, eventStr)
	if err != nil {
		return nil, fmt.Errorf("invalid event_time in db: %w", err)
	}
	p.RemindTime, err = time.Parse(time.RFC3339, remindStr)
	if err != nil {
		return nil, fmt.Errorf("invalid remind_time in db: %w", err)
	}
	p.EventTime = p.EventTime.In(cfg.DefaultTZ)
	p.RemindTime = p.RemindTime.In(cfg.DefaultTZ)

	return &p, nil
}

func (s *PlanService) List(chatID int64, cfg *config.Config) ([]Plan, error) {
	rows, err := s.db.Query(
		`SELECT
			id,
			description,
			event_time,
			remind_time
        FROM
			plan
		WHERE
			chat_id = ?
		ORDER BY id ASC`,
		chatID,
	)
	if err != nil {
		log.Print("error executing query: %w", err)
		return nil, err
	}
	defer rows.Close()
	var res []Plan
	for rows.Next() {
		var p Plan
		p.ChatID = chatID
		if err := rows.Scan(&p.ID, &p.Description, &p.EventTime, &p.RemindTime); err != nil {
			log.Print("error iterating over rows: %w", err)
			return nil, err
		}
		p.EventTime = p.EventTime.In(cfg.DefaultTZ)
		p.RemindTime = p.RemindTime.In(cfg.DefaultTZ)
		res = append(res, p)
	}
	return res, nil
}

func (s *PlanService) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM plan WHERE id = ?`, id)
	return err
}
