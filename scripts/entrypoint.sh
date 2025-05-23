#!/bin/sh
set -euo pipefail

./scripts/migrate.sh

exec /usr/local/bin/app
