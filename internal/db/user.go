package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
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
	Get(ctx context.Context, opts ...UserOption) (*UserFull, error)
	Upsert(ctx context.Context, chatID int64, name, city, tz string, partnerID int64) error
	UpdateGeo(ctx context.Context, chatID int64, city, tz string) error
	UpdatePartner(ctx context.Context, chatID int64, partnerName string) error
	TZ(ctx context.Context, defaultTZ *time.Location, opts ...UserOption) (tz *time.Location)
}

type userManager struct {
	db *sql.DB
}

var _ UserManager = (*userManager)(nil)

func NewUserManager(db *sql.DB) *userManager {
	return &userManager{db: db}
}

type UserOption func(*userQueryOptions)

type userQueryOptions struct {
	byChatID       bool
	byPartnerID    bool
	byUsername     bool
	chatID         int64
	partnerID      int64
	username       string
	includePartner bool
}

func WithChatID(id int64) UserOption {
	return func(opts *userQueryOptions) {
		opts.byChatID = true
		opts.chatID = id
	}
}

func WithPartnerID(id int64) UserOption {
	return func(opts *userQueryOptions) {
		opts.byPartnerID = true
		opts.partnerID = id
	}
}

func WithUsername(name string) UserOption {
	return func(opts *userQueryOptions) {
		opts.byUsername = true
		opts.username = name
	}
}

func WithPartnerInfo() UserOption {
	return func(opts *userQueryOptions) {
		opts.includePartner = true
	}
}

func (um *userManager) Get(
	ctx context.Context,
	opts ...UserOption,
) (*UserFull, error) {
	options := userQueryOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	switch {
	case options.byChatID && options.chatID == 0,
		options.byPartnerID && options.partnerID == 0,
		options.byUsername && options.username == "":
		return nil, fmt.Errorf("db: invalid search criteria")
	}

	var (
		where  string
		arg    interface{}
		fields = []string{"u.chat_id", "u.username", "u.city", "u.tz", "u.partner_id"}
		joins  []string
	)

	switch {
	case options.byChatID:
		where = "u.chat_id = ?"
		arg = options.chatID
	case options.byPartnerID:
		where = "u.partner_id = ?"
		arg = options.partnerID
	case options.byUsername:
		where = "u.username = ?"
		arg = options.username
	default:
		return nil, fmt.Errorf("db: no search criteria provided")
	}

	if options.includePartner {
		fields = append(fields, "p.username AS partner_name")
		joins = append(joins, "JOIN user AS p ON p.chat_id = u.partner_id")
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM user AS u
		%s
		WHERE %s
	`,
		strings.Join(fields, ", "),
		strings.Join(joins, " "),
		where,
	)

	row := um.db.QueryRowContext(ctx, query, arg)

	var (
		chatID      int64
		username    string
		city        string
		tzName      string
		partnerID   int64
		partnerName sql.NullString
	)

	dest := []interface{}{
		&chatID,
		&username,
		&city,
		&tzName,
		&partnerID,
	}

	if options.includePartner {
		dest = append(dest, &partnerName)
	}

	if err := row.Scan(dest...); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("db: user not found")
		}
		return nil, fmt.Errorf("db: scan error: %w", err)
	}

	user := &UserFull{
		User: User{
			ChatID:    chatID,
			Name:      username,
			City:      city,
			PartnerID: partnerID,
		},
	}

	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("db: invalid timezone %q: %w", tzName, err)
	}
	user.TZ = loc

	if options.includePartner {
		user.PartnerName = partnerName.String
	}

	return user, nil
}

func (um *userManager) Upsert(ctx context.Context, chatID int64, name, city, tz string, partnerID int64) error {
	_, err := um.db.ExecContext(ctx, `
        INSERT INTO user(chat_id, username, city, tz, partner_id)
        VALUES(?,?,?,?,?)
        ON CONFLICT(chat_id) DO UPDATE SET
          username=excluded.username,
          city=excluded.city,
		  tz=excluded.tz,
          partner_id=excluded.partner_id
    `, chatID, name, city, tz, partnerID)
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

func (um *userManager) TZ(ctx context.Context, defaultTZ *time.Location, opts ...UserOption) (tz *time.Location) {
	user, err := um.Get(ctx, opts...)
	if err != nil {
		log.Print("db: failed to get tz for user: %w", err)
		tz = defaultTZ
	} else {
		tz = user.TZ
	}
	return tz
}
