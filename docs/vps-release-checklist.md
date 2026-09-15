# Checklist đồng bộ release — tránh lệch code / DB / FE

Mục tiêu: **máy local, GitHub, Railway (hoặc VPS) và Vercel** cùng một phiên bản tính năng. Mỗi lần merge/push có thay đổi BE hoặc FE, làm theo checklist này.

Cập nhật: **2026-09-15**  
Migration mới nhất: **53** (file `be/deploy/EXPECTED_MIGRATION_VERSION`)

Tài liệu liên quan: [deploy-railway.md](./deploy-railway.md), [deploy-vps.md](./deploy-vps.md), [features.md](./features.md), [getting-started.md](./getting-started.md)

---

## 1. Ba repo / ba luồng deploy (đừng nhầm)

| Thành phần | Repo / nơi chạy | Cách lên môi trường |
|------------|-----------------|---------------------|
| **Backend (Railway)** | `be/` monorepo hoặc repo BE | Push branch Railway theo dõi → build Dockerfile → migrate + start. [deploy-railway.md](./deploy-railway.md) |
| **Backend (VPS)** | `CleverSchool-be` (hoặc `be/` trong monorepo) | Push `develop` / `staging` → GitHub Actions → VPS Docker |
| **Frontend** | `fe/` (Vercel) | Push branch → Vercel auto build |
| **Monorepo local** | `CleverSchool/` (cha) | Dev local; CI VPS không chạy nếu chỉ push folder cha. Railway monorepo: Root Directory `/be` |

**Tránh lệch:** Sau khi sửa `be/` trên máy local, phải **push lên repo mà CI đang theo dõi** (xem `deploy-vps.md` §B5). Chỉ commit trong folder cha không kích hoạt deploy VPS.

---

## 2. Quy trình release chuẩn (mỗi lần có tính năng mới)

### Bước A — Trên máy dev (trước khi push)

1. **Cập nhật docs** (bắt buộc): `features.md`, `CHANGELOG.md`, `tables-reference.md`, `ui-design.md` / `AGENTS.md` nếu đổi UX.
2. **Tăng** `be/deploy/EXPECTED_MIGRATION_VERSION` nếu có migration mới.
3. Chạy local:
   ```bash
   cd be
   go run . migrate
   go run . migrate:version    # phải = EXPECTED_MIGRATION_VERSION
   go run . refresh-permissions
   go run ./database/seeder -model Role   # chỉ khi thêm permission trong config/permission.go (không có migration seed)
   ```
4. Smoke local: login, API mới (assessments, settings, course-schedule, …).
5. FE: `pnpm build` trong `fe/`.

### Bước B — Push backend → Railway (không VPS)

Push nhánh Railway đang theo dõi. Root Directory `/be` nếu monorepo. Kiểm tra Deploy Logs + `GET /health`. Chi tiết: [deploy-railway.md](./deploy-railway.md).

Hoặc **Bước B (VPS)** bên dưới.

### Bước B (VPS) — Push backend → VPS

```bash
# Trong repo BE (nhánh develop hoặc staging)
git add be/ deploy/ be/docs/
git commit -m "feat: ..."
git push origin develop    # → api-dev
# hoặc
git push origin staging    # → api-staging
```

CI (`.github/workflows/deploy-api.yml`):

1. Build image GHCR (có **embed migration** trong binary).
2. SCP `be/deploy/` lên VPS.
3. `docker compose pull api && docker compose up -d`.
4. Container start → `docker-entrypoint.sh` chạy `./myapp migrate` tự động.
5. Chạy `post-deploy-check.sh` (nếu đã bật trong workflow).

### Bước C — Sau khi Actions xanh (trên VPS hoặc để CI chạy)

SSH vào VPS (nếu cần kiểm tra tay):

```bash
cd /opt/cleverschool-staging/deploy   # hoặc cleverschool-dev
chmod +x scripts/post-deploy-check.sh
./scripts/post-deploy-check.sh
```

Script kiểm tra:

- Container `api` đang chạy
- `migrate:version` = `EXPECTED_MIGRATION_VERSION`
- `refresh-permissions` (xóa cache Redis)
- `curl` endpoint login (nếu có `API_DOMAIN` trong `.env`)

### Bước D — Push frontend → Vercel

1. Push `fe/` lên branch Vercel đang theo dõi.
2. **Redeploy** nếu chỉ đổi env (`NEXT_PUBLIC_API_BASE_URL`).
3. Đảm bảo URL API khớp môi trường:

| Môi trường FE | `NEXT_PUBLIC_API_BASE_URL` |
|---------------|----------------------------|
| Dev / Preview | `https://api-dev.viethung.uk/api` |
| Staging / Prod | `https://api-staging.viethung.uk/api` |

4. `ALLOW_ORIGINS` trên VPS `.env` phải chứa URL Vercel (CORS).

---

## 3. Migration 0039–0044 — cần có trên VPS

Sau deploy image mới, migration chạy **tự động** khi container `api` khởi động (`RUN_MIGRATIONS=true`).

| Ver | Nội dung | Nếu thiếu sẽ lỗi |
|-----|----------|------------------|
| 39 | Seed grades 1–12 | Tạo lớp: FK `grade_id` |
| 40–41 | Bảng + quyền `assessments` | POST assessments **404** / **403** |
| 42 | `user_courses.main_teacher` | GET course users **500** |
| 43–44 | Bảng + quyền `settings` | GET settings **404** / **403** |

**Permission:** Migration 0041/0044 seed quyền vào DB. Nếu vẫn 403 sau migrate:

```bash
docker exec csstaging-api ./myapp refresh-permissions
# Lần đầu hoặc sau khi đổi config/permission.go nhiều:
cd /opt/cleverschool-staging/deploy && ./scripts/seed-admin.sh
```

(`csstaging-api` / `csdev-api` — tên theo `STACK_NAME` trong `.env`.)

---

## 4. Đồng bộ FE–BE (tránh “code mới FE, API cũ”)

| Triệu chứng | Nguyên nhân | Cách xử lý |
|-------------|-------------|------------|
| 404 API mới | VPS chưa deploy image mới | Push BE + Actions xanh + `post-deploy-check` |
| 403 permission | Cache Redis / thiếu seed | `refresh-permissions` hoặc `seed-admin.sh` |
| UI lỗi, API OK | Vercel chưa build FE mới | Push `fe/` hoặc Redeploy Vercel |
| CORS | `ALLOW_ORIGINS` thiếu URL FE | Sửa `.env` VPS → `docker compose restart api` |
| Migration lệch | Deploy cũ / dirty DB | `migrate:version`, log api, xem §5 |

---

## 5. Xử lý migration lỗi / lệch version

```bash
docker compose logs api --tail 100
docker exec csstaging-api ./myapp migrate:version
```

| Tình huống | Hành động |
|------------|-----------|
| Version < 44, không dirty | `docker compose restart api` (chạy lại migrate) hoặc `docker exec ... ./myapp migrate` |
| `dirty=true` | **Backup DB trước** → `migrate:force N` → `migrate` (xem [getting-started.md](./getting-started.md)) |
| Version đúng nhưng thiếu cột | Image cũ — pull image mới: `docker compose pull api && docker compose up -d` |

Backup trước khi sửa tay: [deploy-vps.md](./deploy-vps.md) Phần E.

---

## 6. Checklist nhanh (in / tick)

**Trước merge**

- [ ] Docs cập nhật (`features.md`, `CHANGELOG`, …)
- [ ] `EXPECTED_MIGRATION_VERSION` đúng
- [ ] Local migrate + smoke OK

**Sau push BE**

- [ ] GitHub Actions deploy **xanh**
- [ ] `post-deploy-check.sh` pass
- [ ] `migrate:version` = 53 (hoặc version hiện tại trong `EXPECTED_MIGRATION_VERSION`)

**Sau push FE**

- [ ] Vercel build xanh
- [ ] `NEXT_PUBLIC_API_BASE_URL` đúng môi trường
- [ ] Test login + 1 flow vừa sửa (vd. sửa chương trình, assessments)

**Định kỳ**

- [ ] Backup Postgres (Phần E `deploy-vps.md`)
- [ ] Đổi mật khẩu admin nếu dùng seed mặc định

---

## 7. Lệnh tham chiếu (copy-paste VPS)

Thay `csstaging` / đường dẫn theo môi trường:

```bash
export DEPLOY=/opt/cleverschool-staging/deploy
cd "$DEPLOY"

docker compose ps
docker compose logs api --tail 50
docker exec csstaging-api ./myapp migrate:version
docker exec csstaging-api ./myapp refresh-permissions
./scripts/post-deploy-check.sh

# Pull image mới thủ công (khi CI lỗi):
# export API_IMAGE=ghcr.io/<owner>/cleverschool-api:staging-latest
# docker compose pull api && API_IMAGE=$API_IMAGE docker compose up -d
```

---

## 8. Khi thêm migration mới (agent / dev)

1. Tạo `NNNN_*.up.sql` / `.down.sql`
2. Cập nhật `tables-reference.md`, `CHANGELOG.md`, `features.md`
3. **Tăng** `be/deploy/EXPECTED_MIGRATION_VERSION`
4. Ghi thêm dòng vào bảng §3 file này (hoặc mục tương ứng trong `features.md` §8)
5. Push BE → verify `post-deploy-check.sh` trên VPS

Quy tắc Cursor: `.cursor/rules/document-changes.mdc`
