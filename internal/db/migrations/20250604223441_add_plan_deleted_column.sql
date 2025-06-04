-- +goose Up
-- +goose StatementBegin
ALTER TABLE plan ADD COLUMN deleted NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE plan DROP COLUMN deleted;
-- +goose StatementEnd
