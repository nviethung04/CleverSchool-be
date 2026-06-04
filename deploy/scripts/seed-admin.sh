#!/usr/bin/env sh
# Chạy một lần sau deploy lần đầu — tạo admin + permissions
set -e
cd "$(dirname "$0")/.."
if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  . ./.env
  set +a
fi
COMPOSE="docker compose"
NETWORK="${COMPOSE_PROJECT_NAME:-cleverschool}_internal"

echo "Seeding roles & admin (go run seeder)..."
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
  sh -c "go run ./database/seeder -model Role"

echo "Done. Login: admin / admin123 — đổi mật khẩu ngay sau khi đăng nhập."
