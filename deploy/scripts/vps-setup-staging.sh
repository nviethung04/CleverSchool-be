#!/usr/bin/env bash
# Setup staging từ đầu sau khi chạy vps-reset.sh
# Cần: GITHUB_TOKEN (nếu repo private) hoặc repo public.
#
#   export GITHUB_TOKEN=ghp_xxx   # tùy chọn
#   bash vps-setup-staging.sh

set -euo pipefail

REPO="https://github.com/nviethung04/CleverSchool-be.git"
BRANCH="staging"
TARGET="/opt/cleverschool-staging"
PROXY="/opt/cleverschool-proxy"

CLONE_URL="$REPO"
if [ -n "${GITHUB_TOKEN:-}" ]; then
  CLONE_URL="https://${GITHUB_TOKEN}@github.com/nviethung04/CleverSchool-be.git"
fi

apt-get update -qq
apt-get install -y -qq git rsync

docker network create cleverschool-edge 2>/dev/null || true

rm -rf "${TARGET}"
git clone -b "${BRANCH}" "${CLONE_URL}" "${TARGET}"

cd "${TARGET}/deploy"
if [ ! -f .env ]; then
  cp env.staging.viethung.template .env
  echo "Đã tạo .env từ env.staging.viethung.template — chạy: nano .env (FRONTEND_URL, ALLOW_ORIGINS)"
fi

docker compose up -d --build
docker compose logs api --tail 40

chmod +x scripts/seed-admin.sh
./scripts/seed-admin.sh

mkdir -p "${PROXY}"
cp -r "${TARGET}/deploy/proxy/"* "${PROXY}/"
cd "${PROXY}"
docker compose up -d

echo ""
echo "✅ Staging stack + Caddy."
echo "Kiểm tra: curl login — dùng mật khẩu admin sau khi seed (không ghi password vào repo)."
