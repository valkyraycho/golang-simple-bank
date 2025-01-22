#!/bin/sh

set -e

echo "running db migration"
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

exec "$@"
