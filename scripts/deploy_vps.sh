#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

: "${VPS_HOST:?Set VPS_HOST, e.g. 203.0.113.10}"
: "${VPS_USER:?Set VPS_USER, e.g. ubuntu}"
VPS_PATH="${VPS_PATH:-/opt/cleverschool-be}"

if ! command -v rsync >/dev/null 2>&1; then
  echo "[ERROR] rsync is required."
  exit 1
fi

if [ ! -f "$ROOT_DIR/.env" ]; then
  echo "[ERROR] Missing $ROOT_DIR/.env"
  exit 1
fi

echo "[1/4] Creating target directory on VPS..."
ssh "${VPS_USER}@${VPS_HOST}" "mkdir -p ${VPS_PATH}"

echo "[2/4] Syncing backend source to VPS..."
rsync -az --delete \
  --exclude '.git' \
  --exclude 'node_modules' \
  --exclude 'logs' \
  --exclude '.idea' \
  "$ROOT_DIR/" "${VPS_USER}@${VPS_HOST}:${VPS_PATH}/"

echo "[3/4] Running deployment routine on VPS..."
ssh "${VPS_USER}@${VPS_HOST}" "cd ${VPS_PATH} && chmod +x scripts/deploy_wsl.sh && ./scripts/deploy_wsl.sh"

echo "[4/4] Done. Verify service status on VPS with:"
echo "ssh ${VPS_USER}@${VPS_HOST} 'cd ${VPS_PATH} && docker compose -f docker-compose.wsl.yml ps'"
