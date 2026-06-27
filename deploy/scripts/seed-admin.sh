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

ROOT="$(cd .. && pwd)"
NETWORK="$(docker network ls --format '{{.Name}}' | grep 'cleverschool.*internal' | head -1)"
if [ -z "$NETWORK" ]; then
  echo "❌ Chưa thấy network Docker của stack. Chạy: docker compose up -d (trong thư mục deploy) trước."
  exit 1
fi

if [ -z "${POSTGRES_PASSWORD:-}" ]; then
  echo "❌ POSTGRES_PASSWORD trống — kiểm tra file deploy/.env"
  exit 1
fi

GO_MOD_CACHE="${GO_MOD_CACHE:-cs-go-mod-cache}"
docker volume inspect "$GO_MOD_CACHE" >/dev/null 2>&1 || docker volume create "$GO_MOD_CACHE" >/dev/null

echo "Seeding roles & admin via network ${NETWORK}..."
echo "⏳ Lần đầu có thể mất 10–20 phút (tải Go modules + compile trên VPS). Đừng Ctrl+C."

docker run --rm \
  --network "${NETWORK}" \
  -v "${ROOT}":/app \
  -v "${GO_MOD_CACHE}":/go/pkg/mod \
  -w /app \
  -e GOPROXY=https://proxy.golang.org,direct \
  -e DB_MASTER_HOST=postgres \
  -e DB_MASTER_PORT=5432 \
  -e DB_MASTER_USER="${POSTGRES_USER:-lms_user}" \
  -e DB_MASTER_PASSWORD="${POSTGRES_PASSWORD}" \
  -e DB_MASTER_NAME="${POSTGRES_DB:-lms_db}" \
  -e REDIS_ENABLED="${REDIS_ENABLED:-true}" \
  -e REDIS_HOST="${REDIS_HOST:-redis}" \
  -e REDIS_PORT="${REDIS_PORT:-6379}" \
  golang:1.24-alpine \
  sh -c '
    set -e
    apk add --no-cache git ca-certificates >/dev/null
    if [ ! -x /app/deploy/bin/seeder ]; then
      echo "→ Building seeder binary (one-time)..."
      CGO_ENABLED=0 go build -o /app/deploy/bin/seeder ./database/seeder
    else
      echo "→ Using cached seeder binary"
    fi
    /app/deploy/bin/seeder -model Role
  '

echo "Done. User admin đã seed — đổi mật khẩu mặc định ngay (xem docs/getting-started.md)."
