// internal/services/plan.go
package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Petryanin/love-bot/internal/config"
	_ "github.com/mattn/go-sqlite3"
)

type Plan struct {
	ID          int64
	ChatID      int64
	Description string
	EventTime   time.Time
	RemindTime  time.Time
	Reminded    bool
}

type PlanService struct {
	db              *sql.DB
	PartnersChatIDs []int64
}

func NewPlanService(dbPath string, partnersChatIDs []int64) *PlanService {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("failed to open a database")
	}
	// создаём таблицу, если нет
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS plan (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chat_id INTEGER NOT NULL,
			description TEXT NOT NULL,
			event_time DATETIME NOT NULL,
			remind_time DATETIME NOT NULL,
			reminded BOOLEAN NOT NULL DEFAULT FALSE
		)`,
	)
	if err != nil {
		log.Fatal("failed to create table")
	}
	return &PlanService{db: db, PartnersChatIDs: partnersChatIDs}
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

func (s *PlanService) List(
	chatID int64,
	pageNumber int,
	cfg *config.Config,
) (plans []Plan, hasPrev, hasNext bool, err error) {
	offset := pageNumber * config.NavPageSize
	limit := config.NavPageSize + 1

	rows, err := s.db.Query(`
        SELECT id, description, event_time, remind_time
        FROM plan
        WHERE chat_id = ?
        ORDER BY id ASC
        LIMIT ? OFFSET ?`,
		chatID, limit, offset,
	)
	if err != nil {
		log.Print("error executing query: %w", err)
		return nil, false, false, err
	}
	defer rows.Close()

	var result []Plan
	for rows.Next() {
		var p Plan
		p.ChatID = chatID

		err := rows.Scan(
			&p.ID,
			&p.Description,
			&p.EventTime,
			&p.RemindTime,
		)
		if err != nil {
			log.Print("error scanning row: %w", err)
			return nil, false, false, err
		}

		p.EventTime = p.EventTime.In(cfg.DefaultTZ)
		p.RemindTime = p.RemindTime.In(cfg.DefaultTZ)
		result = append(result, p)
	}

	if len(result) > config.NavPageSize {
		hasNext = true
		result = result[:config.NavPageSize]
	}

	hasPrev = pageNumber > 0

	return result, hasPrev, hasNext, nil
}

func (s *PlanService) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM plan WHERE id = ?`, id)
	return err
}

func (s *PlanService) GetDueAndMark(now time.Time) ([]Plan, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
        SELECT id, chat_id, description, event_time, remind_time
        FROM plan
        WHERE remind_time <= ? AND reminded = 0
	`, now.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var duePlans []Plan
	for rows.Next() {
		var p Plan
		if err := rows.Scan(
			&p.ID,
			&p.ChatID,
			&p.Description,
			&p.EventTime,
			&p.RemindTime,
		); err != nil {
			return nil, err
		}
		duePlans = append(duePlans, p)
	}

	if len(duePlans) == 0 {
		return duePlans, tx.Commit()
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(duePlans)), ",")
	args := make([]any, len(duePlans)+1)
	for i, p := range duePlans {
		args[i] = p.ID
	}

	_, err = tx.Exec(fmt.Sprintf(`
        UPDATE plan
        SET reminded = TRUE
        WHERE id IN (%s)
    `, placeholders), args...)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return duePlans, nil
}

func (s *PlanService) Schedule(id int64, t time.Time) error {
	_, err := s.db.Exec(`
		UPDATE plan SET remind_time = ?, reminded = FALSE WHERE id = ?`, t, id,
	)
	return err
}
