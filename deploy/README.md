# Deploy BE (VPS + Cloudflare)

| Nhánh | API |
|-------|-----|
| `develop` | https://api-dev.viethung.uk |
| `pre` | https://api-staging.viethung.uk |

Hướng dẫn đầy đủ: [be/docs/deploy-vps.md](../be/docs/deploy-vps.md)

**CI:** `.github/workflows/deploy-api.yml` (root repo).

**Lần đầu trên VPS:**

```bash
docker network create cleverschool-edge
# dev
cp env.dev.example .env && nano .env
docker compose up -d --build && ./scripts/seed-admin.sh
# staging — thư mục khác, dùng env.staging.example
# proxy — xem deploy/proxy/
```
