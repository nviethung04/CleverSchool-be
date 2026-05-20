#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.wsl.yml"
ENV_FILE="$ROOT_DIR/.env.wsl"
COMPOSE_CMD=()

compose() {
  "${COMPOSE_CMD[@]}" --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

if ! command -v docker >/dev/null 2>&1; then
  echo "[ERROR] docker is not installed in WSL."
  exit 1
fi

if docker compose version >/dev/null 2>&1; then
  COMPOSE_CMD=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE_CMD=(docker-compose)
else
  echo "[ERROR] docker compose is not available (tried 'docker compose' and 'docker-compose')."
  exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
  ENV_FILE="$ROOT_DIR/.env"
fi

if [ ! -f "$ENV_FILE" ]; then
  echo "[ERROR] Missing .env.wsl and .env"
  echo "Copy .env.wsl.example to .env.wsl and fill required values first."
  exit 1
fi

echo "[1/5] Starting Postgres and Redis..."
compose up -d postgres redis

echo "[2/5] Building app image..."
compose build app

echo "[3/5] Running database migrations..."
compose run --rm app ./myapp migrate

echo "[4/5] Starting backend app..."
compose up -d app

PORT_VALUE="$(grep -E '^PORT=' "$ENV_FILE" | head -n1 | cut -d'=' -f2-)"
if [ -z "$PORT_VALUE" ]; then
  PORT_VALUE="8080"
fi

HEALTH_URL="http://127.0.0.1:${PORT_VALUE}/ws/health"
echo "[5/5] Smoke test: $HEALTH_URL"

for _ in $(seq 1 30); do
  if curl -fsS "$HEALTH_URL" >/dev/null 2>&1; then
    echo "[OK] Backend is healthy on $HEALTH_URL"
    compose ps
    exit 0
  fi
  sleep 2
done

echo "[ERROR] Health check failed. Showing app logs:"
compose logs --tail=120 app
exit 1
