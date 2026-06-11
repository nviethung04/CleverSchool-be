#!/usr/bin/env sh
# Seed dữ liệu demo MVP (trường, khóa, GV/HS, exam, homework) — chạy sau seed-admin.sh
set -e
cd "$(dirname "$0")/.."
if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  . ./.env
  set +a
fi
NETWORK="$(docker network ls --format '{{.Name}}' | grep 'cleverschool.*internal' | head -1)"
if [ -z "$NETWORK" ]; then
  echo "❌ Chạy docker compose up -d trước."
  exit 1
fi

echo "Seeding demo MVP data..."
docker run --rm \
  --network "${NETWORK}" \
  -v "$(cd .. && pwd)":/app \
  -w /app \
  -e DB_MASTER_HOST=postgres \
  -e DB_MASTER_PORT=5432 \
  -e DB_MASTER_USER="${POSTGRES_USER:-lms_user}" \
  -e DB_MASTER_PASSWORD="${POSTGRES_PASSWORD}" \
  -e DB_MASTER_NAME="${POSTGRES_DB:-lms_db}" \
  golang:1.24-alpine \
  sh -c "go run ./database/seeder -model Demo"
