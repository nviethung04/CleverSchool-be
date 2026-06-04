#!/usr/bin/env bash
# Chạy trên VPS Ubuntu/Debian lần đầu (root hoặc sudo)
set -euo pipefail

apt-get update
apt-get install -y ca-certificates curl git

if ! command -v docker >/dev/null 2>&1; then
  curl -fsSL https://get.docker.com | sh
fi

if ! docker compose version >/dev/null 2>&1; then
  apt-get install -y docker-compose-plugin || true
fi

DEPLOY_DIR="${DEPLOY_DIR:-/opt/cleverschool}"
mkdir -p "$DEPLOY_DIR"

echo "Bootstrap xong. Tiếp theo:"
echo "  1. Clone repo vào $DEPLOY_DIR (hoặc CI rsync deploy/)"
echo "  2. cp deploy/.env.example deploy/.env && chỉnh secret + domain"
echo "  3. cd deploy && cp .env.example .env && chỉnh API_DOMAIN + ALLOW_ORIGINS (Vercel)"
echo "  4. docker compose up -d && ./scripts/seed-admin.sh"
