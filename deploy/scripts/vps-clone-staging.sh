#!/usr/bin/env bash
# Clone CleverSchool-be (staging) đúng cấu trúc cho Docker build — chạy trên VPS một lần.
#
# Cách 1 — Repo public:
#   bash vps-clone-staging.sh
#
# Cách 2 — Repo private (dùng GitHub PAT, không phải mật khẩu đăng nhập):
#   export GITHUB_TOKEN=ghp_xxxxxxxx
#   bash vps-clone-staging.sh

set -euo pipefail

REPO_URL="https://github.com/nviethung04/CleverSchool-be.git"
BRANCH="staging"
TARGET="/opt/cleverschool-staging"
ENV_BACKUP="/tmp/cleverschool-staging.env.bak"

if [ -f "${TARGET}/deploy/.env" ]; then
  cp "${TARGET}/deploy/.env" "${ENV_BACKUP}"
  echo "Đã backup .env → ${ENV_BACKUP}"
fi

CLONE_URL="${REPO_URL}"
if [ -n "${GITHUB_TOKEN:-}" ]; then
  CLONE_URL="https://${GITHUB_TOKEN}@github.com/nviethung04/CleverSchool-be.git"
fi

rm -rf /tmp/cleverschool-staging-clone
git clone -b "${BRANCH}" "${CLONE_URL}" /tmp/cleverschool-staging-clone

mkdir -p "${TARGET}"
# Ghi đè bằng full repo (Dockerfile ở root, deploy/ bên trong)
rsync -a --delete /tmp/cleverschool-staging-clone/ "${TARGET}/"

if [ -f "${ENV_BACKUP}" ]; then
  cp "${ENV_BACKUP}" "${TARGET}/deploy/.env"
  echo "Đã khôi phục .env cũ"
fi

if [ ! -f "${TARGET}/deploy/.env" ]; then
  cp "${TARGET}/deploy/env.staging.viethung.template" "${TARGET}/deploy/.env"
  echo "Tạo .env từ env.staging.viethung.template — hãy nano sửa mật khẩu/URL"
fi

docker network create cleverschool-edge 2>/dev/null || true

echo ""
echo "Xong. Tiếp theo:"
echo "  cd ${TARGET}/deploy"
echo "  docker compose up -d --build"
echo "  ./scripts/seed-admin.sh"
