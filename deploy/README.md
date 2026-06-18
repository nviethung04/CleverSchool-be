# Deploy BE (VPS + Cloudflare)

| Nhánh | API |
|-------|-----|
| `develop` | https://api-dev.viethung.uk |
| `pre` | https://api-staging.viethung.uk |

Hướng dẫn đầy đủ: [be/docs/deploy-vps.md](../be/docs/deploy-vps.md)

**Sau mỗi release (tránh lệch code):** [be/docs/vps-release-checklist.md](../be/docs/vps-release-checklist.md)

**CI:** `.github/workflows/deploy-api.yml` (root repo).

**Lần đầu trên VPS:**

```bash
docker network create cleverschool-edge
# dev
cp env.dev.example .env && nano .env
docker compose up -d --build && ./scripts/seed-admin.sh
./scripts/post-deploy-check.sh   # kiểm tra migration + permissions
# staging — thư mục khác, dùng env.staging.example
# proxy — xem deploy/proxy/
```
