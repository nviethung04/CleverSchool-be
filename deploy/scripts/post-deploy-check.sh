#!/usr/bin/env sh
# Kiểm tra sau deploy VPS — migration, permission cache, API smoke.
# Chạy trong thư mục deploy/ (có .env và docker-compose.yml).
set -e

cd "$(dirname "$0")/.."

if [ ! -f .env ]; then
  echo "❌ Thiếu deploy/.env"
  exit 1
fi

# shellcheck disable=SC1091
. ./.env

STACK="${STACK_NAME:?STACK_NAME chưa set trong .env}"
API_CONTAINER="${STACK}-api"
EXPECTED_FILE="$(dirname "$0")/EXPECTED_MIGRATION_VERSION"
EXPECTED="$(tr -d ' \r\n' < "$EXPECTED_FILE" 2>/dev/null || echo "")"

if [ -z "$EXPECTED" ]; then
  echo "⚠️  Không đọc được EXPECTED_MIGRATION_VERSION — bỏ qua so sánh version."
fi

if ! docker ps --format '{{.Names}}' | grep -qx "$API_CONTAINER"; then
  echo "❌ Container $API_CONTAINER chưa chạy. Chạy: docker compose up -d"
  exit 1
fi

echo "=== Migration version ==="
VERSION_OUT="$(docker exec "$API_CONTAINER" ./myapp migrate:version 2>&1 || true)"
echo "$VERSION_OUT"
CURRENT="$(echo "$VERSION_OUT" | grep -oE '[0-9]+' | tail -1)"

if [ -n "$EXPECTED" ] && [ -n "$CURRENT" ]; then
  if [ "$CURRENT" != "$EXPECTED" ]; then
    echo "❌ Migration lệch: DB=$CURRENT, code mong đợi=$EXPECTED"
    echo "   → Xem be/docs/vps-release-checklist.md"
    exit 1
  fi
  echo "✅ Migration khớp version $EXPECTED"
fi

echo "=== Refresh permission cache (Redis) ==="
docker exec "$API_CONTAINER" ./myapp refresh-permissions || {
  echo "⚠️  refresh-permissions lỗi (Redis?) — kiểm tra REDIS_* trong .env"
}

if [ -n "${API_DOMAIN:-}" ]; then
  echo "=== Smoke: login endpoint ==="
  CODE="$(curl -s -o /dev/null -w '%{http_code}' "${API_DOMAIN}/api/login" -X POST \
    -H 'Content-Type: application/json' \
    -d '{"username":"admin","password":"invalid"}' || echo "000")"
  if [ "$CODE" = "401" ] || [ "$CODE" = "400" ] || [ "$CODE" = "200" ]; then
    echo "✅ API phản hồi HTTP $CODE (endpoint sống)"
  else
    echo "⚠️  API trả HTTP $CODE — kiểm tra Caddy/proxy và container api"
  fi
fi

echo "=== Done ==="
