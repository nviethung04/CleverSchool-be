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
# Network do docker compose tạo (tên project trong docker-compose.yml: cleverschool)
NETWORK="$(docker network ls --format '{{.Name}}' | grep 'cleverschool.*internal' | head -1)"
if [ -z "$NETWORK" ]; then
  echo "❌ Chưa thấy network Docker của stack. Chạy: docker compose up -d (trong thư mục deploy) trước."
  exit 1
fi

echo "Seeding roles & admin (go run seeder) via network ${NETWORK}..."
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
