# Hướng dẫn deploy BE lên VPS (chi tiết từng bước)

Tài liệu này mô tả **việc anh cần làm** để:

1. Chạy **backend + PostgreSQL + Redis** trên VPS (Docker).
2. **Hai môi trường tách nhau** — dev và staging, **mỗi bên một database riêng**.
3. Domain Cloudflare: `api-dev.viethung.uk`, `api-staging.viethung.uk`.
4. **Tự động CI/CD**: push code lên GitHub → build Docker → deploy VPS (không K8s).

Frontend giữ trên **Vercel**; file này chỉ lo **BE + DB**.

**VPS IP (của anh):** `160.250.4.181`

| Nhánh Git | Môi trường | API | Thư mục VPS | Database |
|-----------|------------|-----|-------------|----------|
| `develop` | Dev | https://api-dev.viethung.uk | `/opt/cleverschool-dev` | `lms_db_dev` |
| `pre` | Staging | https://api-staging.viethung.uk | `/opt/cleverschool-staging` | `lms_db_staging` |

Workflow CI/CD: `.github/workflows/deploy-api.yml` (thư mục **gốc** repo CleverSchool).

---

## Phần A — Làm một lần (setup ban đầu)

### A1. Cloudflare DNS

1. Đăng nhập [Cloudflare](https://dash.cloudflare.com) → chọn zone **viethung.uk**.
2. Vào **DNS → Records → Add record**.
3. Tạo **2 bản ghi A** (không đảo ô Name và IP):

**Bản ghi 1 — Dev**

| Ô trên form | Điền |
|-------------|------|
| Type | A |
| Name | `api-dev` |
| IPv4 address | `160.250.4.181` |
| Proxy | Proxied (đám mây **cam**) |

**Bản ghi 2 — Staging**

| Ô trên form | Điền |
|-------------|------|
| Type | A |
| Name | `api-staging` |
| IPv4 address | `160.250.4.181` |
| Proxy | Proxied |

4. Vào **SSL/TLS → Overview**:
   - Lúc mới setup: chọn **Full**.
   - Sau khi Caddy trên VPS cấp cert xong: đổi **Full (strict)**.

5. Đợi 1–5 phút cho DNS cập nhật.

---

### A2. SSH vào VPS

Trên máy Windows (PowerShell):

```powershell
ssh root@160.250.4.181
# hoặc: ssh ubuntu@160.250.4.181  (tùy user VPS)
```

Nếu chưa có key SSH cho deploy, tạo trên **máy Windows** (PowerShell), **không** tạo trên VPS nếu mục đích là GitHub Actions SSH vào VPS:

```powershell
ssh-keygen -t ed25519 -C "cleverschool-deploy"
```

Khi `ssh-keygen` hỏi, điền như sau:

| Câu hỏi | Điền gì |
|---------|---------|
| `Enter file in which to save the key (...)` | **Enter** (giữ mặc định, vd. `C:\Users\hng\.ssh\id_ed25519`) |
| `Enter passphrase` | **Enter** (để trống — CI/CD dùng key không passphrase) |
| `Enter same passphrase again` | **Enter** |

Sau đó:

1. **Public key** (`.pub`) → copy vào VPS `~/.ssh/authorized_keys` (xem bên dưới).
2. **Private key** (không có đuôi `.pub`) → dán vào GitHub Secret `VPS_SSH_KEY`.

Copy public key lên VPS (chạy trên **Windows**, thay user VPS nếu không dùng `root`):

```powershell
type $env:USERPROFILE\.ssh\id_ed25519.pub | ssh root@160.250.4.181 "mkdir -p ~/.ssh && chmod 700 ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
```

> Nếu anh đã chạy `ssh-keygen` **trên VPS** (`root@viethung`): key nằm ở `/root/.ssh/id_ed25519` — dùng để SSH **từ VPS đi nơi khác**, không phải để GitHub vào VPS. Nên tạo lại key trên **máy anh** như trên.

Các bước sau **chạy trên VPS** (Linux).

---

### A3. Cài Docker

```bash
curl -fsSL https://get.docker.com | sh
apt update && apt install -y git docker-compose-plugin

docker --version
docker compose version
```

---

### A4. Tạo thư mục và network Docker

```bash
docker network create cleverschool-edge

sudo mkdir -p /opt/cleverschool-dev/deploy
sudo mkdir -p /opt/cleverschool-staging/deploy
sudo mkdir -p /opt/cleverschool-proxy

sudo chown -R $USER:$USER /opt/cleverschool-dev /opt/cleverschool-staging /opt/cleverschool-proxy
```

**Giải thích DB riêng:** Mỗi thư mục có `docker-compose` riêng → container Postgres riêng → volume `postgres_data` riêng → **dev và staging không dùng chung database**.

---

### A5. Đưa file deploy lên VPS

**Cách 1 — Clone repo (khuyến nghị lần đầu):**

```bash
cd /tmp
git clone https://github.com/<tên-org-hoặc-user>/CleverSchool.git
cp -r CleverSchool/deploy/* /opt/cleverschool-dev/deploy/
cp -r CleverSchool/deploy/* /opt/cleverschool-staging/deploy/
cp -r CleverSchool/deploy/proxy/* /opt/cleverschool-proxy/
```

**Cách 2 — Để CI copy sau:** vẫn cần tạo `.env` thủ công trước lần deploy đầu (xem A6).

---

### A6. Tạo file `.env` (bắt buộc, không commit Git)

**Dev — dùng file mẫu sẵn (viethung.uk):**

Trong repo có `deploy/env.dev.viethung.template` (đã điền domain + mật khẩu dev mẫu). Trên VPS:

```bash
cp /opt/cleverschool-dev/deploy/env.dev.viethung.template /opt/cleverschool-dev/deploy/.env
nano /opt/cleverschool-dev/deploy/.env   # sửa FRONTEND_URL = URL Vercel thật của anh
```

Nội dung dev (copy-paste nếu cần):

```env
STACK_NAME=csdev
API_DOMAIN=https://api-dev.viethung.uk
FRONTEND_URL=https://viethung.uk
ALLOW_ORIGINS=https://viethung.uk,http://localhost:3000
POSTGRES_DB=lms_db_dev
POSTGRES_USER=lms_user
POSTGRES_PASSWORD=admin123
REDIS_PASSWORD=admin123
JWT_SECRET=CsDev_viethung_uk_JWT_2026_min32chars_xY3z
RUN_MIGRATIONS=true
ENABLE_SWAGGER=true
APP_DEBUG=false
```

> **Đổi** `FRONTEND_URL` và `ALLOW_ORIGINS` cho khớp project Vercel. **Đổi** mật khẩu nếu repo public.

Hoặc từ example cơ bản:

```bash
cp /opt/cleverschool-dev/deploy/env.dev.example /opt/cleverschool-dev/deploy/.env
nano /opt/cleverschool-dev/deploy/.env
```

**Staging — dùng file mẫu sẵn (viethung.uk):**

```bash
cp /opt/cleverschool-staging/deploy/env.staging.viethung.template /opt/cleverschool-staging/deploy/.env
nano /opt/cleverschool-staging/deploy/.env
```

Nội dung staging (copy-paste nếu cần):

```env
STACK_NAME=csstaging
API_DOMAIN=https://api-staging.viethung.uk
FRONTEND_URL=https://staging.viethung.uk
ALLOW_ORIGINS=https://staging.viethung.uk,http://localhost:3000
POSTGRES_DB=lms_db_staging
POSTGRES_USER=lms_user
POSTGRES_PASSWORD=admin123
REDIS_PASSWORD=admin123
JWT_SECRET=CsStg_viethung_uk_JWT_2026_min32chars_bW7q
RUN_MIGRATIONS=true
ENABLE_SWAGGER=false
APP_DEBUG=false
REDIS_ENABLED=true
REDIS_DB=0
```

> Mật khẩu **khác hẳn dev** — hai DB độc lập. Thêm URL Vercel staging vào `ALLOW_ORIGINS` nếu FE chạy domain khác.

Hoặc:

```bash
cp /opt/cleverschool-staging/deploy/env.staging.example /opt/cleverschool-staging/deploy/.env
nano /opt/cleverschool-staging/deploy/.env
```

**Phải sửa** (mỗi file khác nhau):

| Biến | Dev (ví dụ) | Staging (ví dụ) |
|------|-------------|-----------------|
| `STACK_NAME` | `csdev` | `csstaging` |
| `API_DOMAIN` | `https://api-dev.viethung.uk` | `https://api-staging.viethung.uk` |
| `FRONTEND_URL` | URL Vercel dev | URL Vercel staging |
| `ALLOW_ORIGINS` | URL Vercel dev (có thể nhiều URL, cách nhau `,`) | URL Vercel staging |
| `POSTGRES_PASSWORD` | Mật khẩu mạnh **khác staging** | Mật khẩu mạnh **khác dev** |
| `REDIS_PASSWORD` | Mật khẩu Redis dev | Mật khẩu Redis staging |
| `JWT_SECRET` | Chuỗi bí mật ≥ 32 ký tự | Chuỗi khác dev |

`POSTGRES_DB` mặc định: `lms_db_dev` / `lms_db_staging` — **hai DB độc lập**.

---

### A7. Chạy stack lần đầu (build image trên VPS)

```bash
cd /opt/cleverschool-dev/deploy
docker compose up -d --build

cd /opt/cleverschool-staging/deploy
docker compose up -d --build
```

Kiểm tra container:

```bash
docker ps
# Phải thấy: csdev-api, csdev-postgres, csdev-redis
#           csstaging-api, csstaging-postgres, csstaging-redis
```

Container `api` **tự chạy migration** khi khởi động (`RUN_MIGRATIONS=true`).

Xem log migration:

```bash
cd /opt/cleverschool-dev/deploy
docker compose logs api --tail 80
```

---

### A8. Tạo tài khoản admin (mỗi môi trường một lần)

```bash
chmod +x /opt/cleverschool-dev/deploy/scripts/seed-admin.sh
chmod +x /opt/cleverschool-staging/deploy/scripts/seed-admin.sh

cd /opt/cleverschool-dev/deploy && ./scripts/seed-admin.sh
cd /opt/cleverschool-staging/deploy && ./scripts/seed-admin.sh
```

Đăng nhập mặc định: `admin` / `admin123` — **đổi mật khẩu ngay** sau khi vào được.

---

### A9. Bật HTTPS (Caddy — một lần cho cả hai domain)

```bash
cd /opt/cleverschool-proxy
docker compose up -d
docker compose logs -f caddy
```

Chờ thấy log cấp certificate Let's Encrypt thành công (Ctrl+C thoát log).

**Firewall VPS** phải mở: **22** (SSH), **80**, **443**. Không cần mở port 8080 ra internet.

---

### A10. Kiểm tra API hoạt động

```bash
curl -s https://api-dev.viethung.uk/api/login \
  -X POST -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

curl -s https://api-staging.viethung.uk/api/login \
  -X POST -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

Kỳ vọng: JSON có `code: 200` và `token`.

---

### A11. Vercel (frontend)

Trên [Vercel](https://vercel.com) → Project → **Settings → Environment Variables**:

| Environment Vercel | Biến | Giá trị |
|--------------------|------|---------|
| Development / Preview (dev) | `NEXT_PUBLIC_API_BASE_URL` | `https://api-dev.viethung.uk/api` |
| Preview hoặc Production (staging) | `NEXT_PUBLIC_API_BASE_URL` | `https://api-staging.viethung.uk/api` |

**Không** có khoảng trắng đầu/cuối URL.

Sau khi lưu → **Redeploy** project Vercel.

---

## Phần B — Cấu hình CI/CD GitHub (một lần)

Sau phần B, mỗi khi anh **push code BE** lên nhánh `develop` hoặc `pre`, GitHub tự:

1. Build image Docker API.
2. Đẩy lên GitHub Container Registry (GHCR).
3. SSH vào VPS → `docker compose pull` → restart container API.
4. **Không xóa database** — chỉ cập nhật app; migration chạy lại khi container start.

### B1. Bật quyền workflow

Repo GitHub **CleverSchool** → **Settings → Actions → General**:

- **Workflow permissions** → chọn **Read and write permissions**.
- Save.

### B2. Tạo Personal Access Token (pull image trên VPS)

1. GitHub → **Settings → Developer settings → Personal access tokens → Tokens (classic)**.
2. **Generate new token**:
   - Note: `cleverschool-ghcr-pull`
   - Expiration: tùy anh (90 ngày / no expiration).
   - Scopes: tick **`read:packages`** (và `repo` nếu repo private).
3. Copy token (chỉ hiện một lần).

### B3. Cho phép VPS pull package GHCR (repo private)

**Cách đơn giản:** Repo → **Packages** → package `cleverschool-api` → **Package settings → Change visibility** → Public (nếu được phép).

**Hoặc** giữ private và dùng token ở B2 trên VPS:

```bash
echo "<PAT_read_packages>" | docker login ghcr.io -u <github-username> --password-stdin
```

### B4. Thêm Secrets vào repo

**Settings → Secrets and variables → Actions → New repository secret:**

| Secret | Giá trị |
|--------|---------|
| `VPS_HOST` | `160.250.4.181` |
| `VPS_USER` | User SSH (vd. `root` hoặc `ubuntu`) |
| `VPS_SSH_KEY` | Toàn bộ nội dung file **private key** (bắt đầu `-----BEGIN ... KEY-----`) |
| `GHCR_PULL_TOKEN` | PAT ở bước B2 |
| `VPS_PORT` | (tùy chọn) `22` |
| `VPS_DEPLOY_PATH_DEV` | (tùy chọn) `/opt/cleverschool-dev` |
| `VPS_DEPLOY_PATH_STAGING` | (tùy chọn) `/opt/cleverschool-staging` |

### B5. Kiểm tra workflow có trong repo

File phải nằm tại:

```text
CleverSchool/.github/workflows/deploy-api.yml
```

**Không** dùng `be/.github/workflows/` (Harbor/K8s cũ — GitHub không chạy workflow trong subfolder đó).

### B6. Chạy thử CI/CD lần đầu

1. Commit & push lên nhánh **`develop`** (có thay đổi trong `be/` hoặc `deploy/`).
2. Vào tab **Actions** trên GitHub → workflow **Deploy API (dev / staging)**.
3. Chờ job **xanh** (build-and-push → deploy).

Hoặc chạy tay: **Actions → Deploy API → Run workflow** → chọn `dev` hoặc `staging`.

**Lưu ý lần đầu:** File `.env` trên VPS **phải tồn tại** trước khi CI deploy (phần A6). CI **không** tạo `.env` và **không** ghi đè mật khẩu DB.

---

## Phần C — Hàng ngày (khi có code mới)

### Dev

```bash
git checkout develop
# ... sửa code trong be/ ...
git add .
git commit -m "fix: ..."
git push origin develop
```

→ GitHub Actions deploy **dev** → `https://api-dev.viethung.uk`

### Staging

```bash
git checkout pre
git merge develop   # hoặc cherry-pick tùy quy trình team
git push origin pre
```

→ GitHub Actions deploy **staging** → `https://api-staging.viethung.uk`

### Anh không cần SSH thủ công mỗi lần

Trừ khi:

- Đổi biến `.env` trên VPS → `docker compose restart api` trong thư mục deploy tương ứng.
- Xem log: `docker compose logs api -f`
- Seed lại admin: `./scripts/seed-admin.sh`

---

## Phần D — Sơ đồ tổng thể

```text
┌─────────────────┐     push develop/pre      ┌──────────────────┐
│  GitHub Code    │ ────────────────────────► │ GitHub Actions   │
└─────────────────┘                           │ build + GHCR     │
                                              └────────┬─────────┘
                                                       │ SSH deploy
                                                       ▼
┌─────────────────┐   HTTPS    ┌──────────────────────────────────────┐
│ Vercel (FE)     │ ─────────► │ VPS 160.250.4.181                    │
└─────────────────┘            │  Caddy :443                          │
                               │    ├─ api-dev → csdev-api + DB dev   │
                               │    └─ api-staging → csstaging + DB   │
                               └──────────────────────────────────────┘
```

---

## Phần E — Backup database

**Dev:**

```bash
docker exec csdev-postgres pg_dump -U lms_user lms_db_dev > ~/backup-dev-$(date +%F).sql
```

**Staging:**

```bash
docker exec csstaging-postgres pg_dump -U lms_user lms_db_staging > ~/backup-staging-$(date +%F).sql
```

---

## Phần F — Xử lý lỗi thường gặp

| Triệu chứng | Nguyên nhân / Cách xử lý |
|-------------|---------------------------|
| Cloudflare 502 | API/Caddy chưa chạy: `docker ps`, `docker compose logs api` |
| CORS từ Vercel | Thiếu URL trong `ALLOW_ORIGINS` → sửa `.env` → `docker compose restart api` |
| Actions không chạy | Push nhầm nhánh; hoặc chỉ đổi `fe/` (workflow chỉ theo dõi `be/`, `deploy/`) |
| Deploy fail: thiếu `.env` | Tạo `.env` trên VPS (A6) |
| Pull image 401 | `GHCR_PULL_TOKEN` sai hoặc chưa `docker login ghcr.io` trên VPS |
| Migration lỗi | `docker compose logs api`; kiểm tra `postgres` healthy |
| Nhánh tên `dev` thay vì `develop` | Sửa `branches:` trong `.github/workflows/deploy-api.yml` |

---

## Phần G — Checklist hoàn tất

**Setup VPS (một lần)**

- [ ] DNS `api-dev`, `api-staging` → `160.250.4.181` (Proxied)
- [ ] Docker + network `cleverschool-edge`
- [ ] `.env` dev và staging (password **khác nhau**)
- [ ] `docker compose up -d` cả hai môi trường
- [ ] Seed admin dev + staging
- [ ] Caddy proxy HTTPS
- [ ] `curl` login OK cả hai URL

**Vercel**

- [ ] `NEXT_PUBLIC_API_BASE_URL` đúng từng môi trường
- [ ] Redeploy FE

**GitHub CI/CD**

- [ ] Workflow permissions Read and write
- [ ] Secrets: `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`, `GHCR_PULL_TOKEN`
- [ ] Push `develop` / `pre` → Actions xanh
- [ ] Đổi mật khẩu `admin123`

---

## File liên quan trong repo

| File | Mục đích |
|------|----------|
| `deploy/docker-compose.yml` | Postgres + Redis + API (mỗi env một bản copy) |
| `deploy/env.dev.example` | Mẫu `.env` dev |
| `deploy/env.staging.example` | Mẫu `.env` staging |
| `deploy/proxy/` | Caddy HTTPS |
| `deploy/scripts/seed-admin.sh` | Tạo user admin |
| `.github/workflows/deploy-api.yml` | CI/CD tự động |

---

*Nếu nhánh dev của team là `dev` thay vì `develop`, báo team dev để cập nhật trigger trong `deploy-api.yml`.*
