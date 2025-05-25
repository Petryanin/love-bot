package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	ChatID    int64
	Name      string
	City      string
	TZ        *time.Location
	PartnerID int64
}

type UserFull struct {
	User
	PartnerName string
}

type UserManager interface {
	GetByID(ctx context.Context, chatID int64, byPartner bool) (*UserFull, error)
	Upsert(ctx context.Context, user *User) error
	UpdateGeo(ctx context.Context, chatID int64, city, tz string) error
	UpdatePartner(ctx context.Context, chatID int64, partnerName string) error
}

type userManager struct {
	db *sql.DB
}

var _ UserManager = (*userManager)(nil)

func NewUserManager(db *sql.DB) *userManager {
	return &userManager{db: db}
}

func (um *userManager) GetByID(ctx context.Context, id int64, byPartner bool) (*UserFull, error) {
	var where string
	if byPartner {
		where = "u.partner_id = ?"
	} else {
		where = "u.chat_id = ?"
	}

	query := fmt.Sprintf(`
        SELECT
            u.chat_id, u.username, u.city, u.tz,
            u.partner_id, p.username
        FROM user AS u
        JOIN user AS p ON p.chat_id = u.partner_id
        WHERE %s
    `, where)

	row := um.db.QueryRowContext(ctx, query, id)

	var (
		user   UserFull
		tzName string
	)

	if err := row.Scan(
		&user.ChatID,
		&user.Name,
		&user.City,
		&tzName,
		&user.PartnerID,
		&user.PartnerName,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("db: user not found (byPartner=%v, id=%d)", byPartner, id)
		}
		return nil, fmt.Errorf("db: query scan error: %w", err)
	}

	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("db: failed to load user location: %w", err)
	}
	user.TZ = loc

	return &user, nil
}

func (um *userManager) Upsert(ctx context.Context, user *User) error {
	_, err := um.db.ExecContext(ctx, `
        INSERT INTO user(chat_id, username, city, tz, partner_id)
        VALUES(?,?,?,?,?)
        ON CONFLICT(chat_id) DO UPDATE SET
          username=excluded.username,
          city=excluded.city,
		  tz=excluded.tz,
          partner_id=excluded.partner_id
    `, user.ChatID, user.City, user.TZ.String(), user.PartnerID)
	return err
}

func (um *userManager) UpdateGeo(ctx context.Context, chatID int64, city, tz string) error {
	res, err := um.db.ExecContext(ctx, `
		UPDATE user
    	SET city = ?, tz = ?
        WHERE chat_id = ?
		`, city, tz, chatID,
	)
	if err != nil {
		return fmt.Errorf("db: exec update: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db: checking affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("db: user %d not found", chatID)
	}
	return nil
}

func (um *userManager) UpdatePartner(ctx context.Context, chatID int64, partnerName string) error {
	var partnerID int64
	err := um.db.QueryRowContext(ctx, `
		SELECT chat_id
        FROM user
        WHERE username = ?
		`, partnerName,
	).Scan(&partnerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("db: partner %q not found", partnerName)
		}
		return fmt.Errorf("db: query partner id: %w", err)
	}

	res, err := um.db.ExecContext(ctx, `
		UPDATE user
		SET partner_id = ?
        WHERE chat_id = ?
		`, partnerID, chatID,
	)
	if err != nil {
		return fmt.Errorf("db: exec update: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db: checking affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("db: user %d not found", chatID)
	}
	return nil
}
