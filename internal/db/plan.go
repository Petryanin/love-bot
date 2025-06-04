package db

import (
	"context"
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

type Planner interface {
	Add(ctx context.Context, p *Plan) error
	GetByID(ctx context.Context, id int64, cfg *config.Config) (*Plan, error)
	List(ctx context.Context, pageNumber int) (plans []Plan, hasPrev, hasNext bool, err error)
	Delete(ctx context.Context, id int64) error
	GetDueAndMark(ctx context.Context, now time.Time) ([]Plan, error)
	Schedule(ctx context.Context, id int64, t time.Time) error
	DeleteExpired(ctx context.Context, retention time.Duration) (int64, error)
}

type PlanService struct {
	db *sql.DB
}

var _ Planner = (*PlanService)(nil)

func NewPlanService(db *sql.DB) *PlanService {
	return &PlanService{db: db}
}

func (s *PlanService) Add(ctx context.Context, p *Plan) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO plan (chat_id, description, event_time, remind_time)
        VALUES(?, ?, ?, ?)`,
		p.ChatID, p.Description, p.EventTime.UTC(), p.RemindTime.UTC(),
	)
	return err
}

func (s *PlanService) GetByID(ctx context.Context, id int64, cfg *config.Config) (*Plan, error) {
	// подготавливаем SQL
	row := s.db.QueryRowContext(ctx, `
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

	return &p, nil
}

func (s *PlanService) List(ctx context.Context, pageNumber int) (plans []Plan, hasPrev, hasNext bool, err error) {
	offset := pageNumber * config.NavPageSize
	limit := config.NavPageSize + 1

	rows, err := s.db.QueryContext(ctx, `
        SELECT id, description, event_time, remind_time
        FROM plan
		WHERE deleted IS FALSE
        ORDER BY id ASC
        LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		log.Print("error executing query: %w", err)
		return nil, false, false, err
	}
	defer rows.Close()

	var result []Plan
	for rows.Next() {
		var p Plan

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
		result = append(result, p)
	}

	if len(result) > config.NavPageSize {
		hasNext = true
		result = result[:config.NavPageSize]
	}

	hasPrev = pageNumber > 0

	return result, hasPrev, hasNext, nil
}

func (s *PlanService) Delete(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE plan
		SET deleted = TRUE
		WHERE id = ?
	`, id)
	return err
}

func (s *PlanService) GetDueAndMark(ctx context.Context, now time.Time) ([]Plan, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
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

	_, err = tx.ExecContext(ctx, fmt.Sprintf(`
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

func (s *PlanService) Schedule(ctx context.Context, id int64, t time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE plan SET remind_time = ?, reminded = FALSE WHERE id = ?`, t, id,
	)
	return err
}

func (s *PlanService) DeleteExpired(ctx context.Context, retention time.Duration) (int64, error) {
	threshold := time.Now().UTC().Add(-retention)

	res, err := s.db.ExecContext(ctx, `
        UPDATE plan
		SET deleted = TRUE
    	WHERE event_time <= ? AND deleted is FALSE
    `, threshold)
	if err != nil {
		return 0, fmt.Errorf("db: exec error: %w", err)
	}
	return res.RowsAffected()
}
