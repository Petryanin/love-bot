#!/bin/sh
set -euo pipefail

DB_DRIVER="${DB_DRIVER:-sqlite3}"

DSN="${DB_PATH:?"Environment variable DB_PATH must be set"}"

MIGRATIONS_DIR="${MIGRATIONS_DIR:-/app/migrations}"

GOOSE_BIN="${GOOSE_BIN:-goose}"

echo "Goose: running migrations"
echo "  driver: $DB_DRIVER"
echo "  dsn: $DSN"
echo "  migrations dir: $MIGRATIONS_DIR"
echo

$GOOSE_BIN -dir $MIGRATIONS_DIR $DB_DRIVER $DSN up
