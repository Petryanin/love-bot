-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS plan (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id INTEGER NOT NULL,
    description TEXT NOT NULL,
    event_time DATETIME NOT NULL,
    remind_time DATETIME NOT NULL,
    reminded BOOLEAN NOT NULL DEFAULT FALSE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS plan;
-- +goose StatementEnd
