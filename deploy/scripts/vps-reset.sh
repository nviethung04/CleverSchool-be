#!/usr/bin/env bash
# Dọn sạch CleverSchool trên VPS (container, volume, thư mục /opt).
# Chạy trên VPS: bash vps-reset.sh
# Giữ Docker đã cài — chỉ xóa stack CleverSchool.

set -euo pipefail

echo "=============================================="
echo " RESET CleverSchool trên VPS"
echo " - Dừng & xóa container + volume Postgres/Redis/Caddy"
echo " - Xóa /opt/cleverschool-dev, staging, proxy"
echo " - KHÔNG gỡ Docker, KHÔNG đổi DNS/SSH"
echo "=============================================="
read -r -p "Gõ yes để tiếp tục: " confirm
if [ "$confirm" != "yes" ]; then
  echo "Đã hủy."
  exit 0
fi

down_compose() {
  local dir="$1"
  if [ -f "${dir}/docker-compose.yml" ]; then
    echo "→ docker compose down -v (${dir})"
    (cd "${dir}" && docker compose down -v --remove-orphans) || true
  fi
}

down_compose /opt/cleverschool-dev/deploy
down_compose /opt/cleverschool-staging/deploy
down_compose /opt/cleverschool-proxy

docker rm -f cs-caddy \
  csstaging-api csstaging-postgres csstaging-redis \
  csdev-api csdev-postgres csdev-redis 2>/dev/null || true

for vol in cleverschool_postgres_data cleverschool_redis_data \
  cleverschool-proxy_caddy_data cleverschool-proxy_caddy_config; do
  docker volume rm "$vol" 2>/dev/null || true
done

docker network rm cleverschool-edge 2>/dev/null || true

rm -rf /opt/cleverschool-dev /opt/cleverschool-staging /opt/cleverschool-proxy
rm -rf /tmp/CleverSchool-be /tmp/cleverschool-staging-clone

echo ""
echo "✅ Đã dọn sạch."
echo "Tiếp theo (staging trước):"
echo "  docker network create cleverschool-edge"
echo "  git clone -b staging ... /opt/cleverschool-staging"
echo "  cd /opt/cleverschool-staging/deploy && cp env.staging.viethung.template .env && nano .env"
echo "  docker compose up -d --build && ./scripts/seed-admin.sh"
echo "  cp -r deploy/proxy/* /opt/cleverschool-proxy/ && cd /opt/cleverschool-proxy && docker compose up -d"
echo "Xem: be/docs/deploy-vps.md — Phần Z"
