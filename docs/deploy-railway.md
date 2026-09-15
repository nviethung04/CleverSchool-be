# Deploy BE lên Railway (không VPS)

Frontend giữ trên **Vercel** (`fe/README_DEPLOY_VERCEL.md`). File này chỉ lo **API + Postgres + Redis** trên Railway.

Railway **không chạy** `docker-compose.yml`. Mỗi service Compose thành một service Railway. Dùng Postgres/Redis **managed**, không deploy image DB tự host.

Nguồn:

- [Deploy a Docker Compose App](https://docs.railway.com/guides/docker-compose)
- [Dockerfiles](https://docs.railway.com/builds/dockerfiles)
- [PostgreSQL](https://docs.railway.com/databases/postgresql)
- [Redis](https://docs.railway.com/databases/redis)
- [Variables](https://docs.railway.com/guides/variables)
- [Monorepo / Root Directory](https://docs.railway.com/guides/deploying-a-monorepo)
- [Volumes](https://docs.railway.com/guides/volumes)

---

## 1. Ba service trên một project

| Compose (VPS) | Railway |
|---------------|---------|
| `api` (Dockerfile) | Service **api** — GitHub + `be/Dockerfile` |
| `postgres` image | Database → **PostgreSQL** (managed) |
| `redis` image | Database → **Redis** (managed) |
| `ports: 8080` | Settings → Networking → **Generate Domain** |
| volumes upload | Volume mount `/data` (entrypoint map sang `public` + `scorm-packages`) |
| `depends_on` | Không có — app retry kết nối lúc start |

---

## 2. Làm một lần trên dashboard

1. [railway.com](https://railway.com) → **New Project** → **Empty project**.
2. **+ New** → **Database** → PostgreSQL. Giữ tên service `Postgres` (hoặc sửa reference cho khớp).
3. **+ New** → **Database** → Redis. Giữ tên `Redis`.
4. **+ New** → **GitHub Repo** → chọn repo chứa backend.
5. Service API — **Settings → Build**:

   | Repo | Root Directory | Watch Paths |
   |------|----------------|-------------|
   | Monorepo `CleverSchool/` | `/be` | `/be/**` |
   | Repo riêng `CleverSchool-be` | `/` (trống) | để trống |

   Railway tự nhận `Dockerfile` trong root directory ([docs](https://docs.railway.com/builds/dockerfiles)).

6. **Settings → Networking** → **Generate Domain**. Ghi lại URL dạng `https://xxx.up.railway.app`.
7. **Settings → Volumes** → Add Volume, mount path **`/data`**.
8. **Settings → Healthcheck Path** = `/health`. Timeout dài (migrate chạy trước khi listen): biến `RAILWAY_HEALTHCHECK_TIMEOUT_SEC=300`.
9. **Variables** → Raw Editor → dán `be/deploy/env.railway.example`, sửa:

   - `FRONTEND_URL` / `ALLOW_ORIGINS` = URL Vercel
   - `JWT_SECRET` = chuỗi mạnh
   - Tên reference `${{Postgres.DATABASE_URL}}` / `${{Redis.REDIS_URL}}` khớp tên service trên canvas

10. Deploy. Log API phải có `Connected to master database` và `Server starting on :$PORT`.

`PORT` do Railway inject — không set tay trừ khi cần cố định.

---

## 3. Sau deploy lần đầu

Trong tab **api** → terminal / one-off command (hoặc Railway CLI):

```bash
./myapp migrate:version
./myapp refresh-permissions
./seeder -model Role
```

Tạo admin: xem `be/deploy/scripts/seed-admin.sh` (chạy tương đương trong container).

Kiểm tra:

```bash
curl https://YOUR-SERVICE.up.railway.app/health
```

Phải trả `{"status":"ok"}`.

---

## 4. Nối frontend Vercel

Trên Vercel:

- `NEXT_PUBLIC_API_BASE_URL=https://YOUR-SERVICE.up.railway.app/api`

Trên Railway API:

- `ALLOW_ORIGINS` chứa origin Vercel (kể cả preview nếu cần)
- Redeploy FE sau khi đổi env

---

## 5. Cập nhật sau này

Push lên branch Railway đang theo dõi → build Dockerfile → `docker-entrypoint.sh` chạy `./myapp migrate` rồi start.

Rollback: Deployments → chọn bản cũ → Redeploy.

---

## 6. Domain riêng (tuỳ chọn)

Settings → Networking → Custom Domain (ví dụ `api.yourdomain.com`). Cập nhật `API_DOMAIN` / `APP_FULL_URL` và `NEXT_PUBLIC_API_BASE_URL`.

---

## 7. Lỗi thường gặp

| Triệu chứng | Cách xử lý |
|-------------|------------|
| Build không thấy Dockerfile | Root Directory sai (`/be` trên monorepo) |
| `Failed to connect to master database` | Reference `DATABASE_URL` sai tên service; xem Variables đã resolve chưa |
| Redis fail, app vẫn lên | Redis optional lúc start; set `REDIS_URL=${{Redis.REDIS_URL}}` |
| CORS | Thêm origin Vercel vào `ALLOW_ORIGINS`, redeploy API |
| Upload mất sau deploy | Chưa gắn Volume `/data` |
| Healthcheck fail lần đầu | Tăng `RAILWAY_HEALTHCHECK_TIMEOUT_SEC` (migrate + buf build không nằm ở health; migrate nằm ở start) |
| SSL Postgres | Dùng `DATABASE_URL` Railway nguyên bản (private). Không thêm `sslmode=disable` lên URL managed trừ khi log yêu cầu |

---

## 8. VPS cũ

`deploy-vps.md` + GitHub Actions SSH vẫn dùng được. Railway là đường **không VPS**. Hai môi trường không chia sẻ DB trừ khi trỏ cùng `DATABASE_URL`.
