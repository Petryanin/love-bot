-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user (
    chat_id INTEGER PRIMARY KEY,
    username TEXT UNIQUE,
    city TEXT NOT NULL,
    tz TEXT NOT NULL,
    partner_id INTEGER
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user;
-- +goose StatementEnd
