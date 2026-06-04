# Hướng dẫn deploy BE lên VPS (chi tiết từng bước)

Tài liệu này mô tả **việc anh cần làm** để:

1. Chạy **backend + PostgreSQL + Redis** trên VPS (Docker).
2. **Hai môi trường tách nhau** — dev và staging, **mỗi bên một database riêng**.
3. Domain Cloudflare: `api-dev.viethung.uk`, `api-staging.viethung.uk`.
4. **Tự động CI/CD**: push code lên GitHub → build Docker → deploy VPS (không K8s).

Frontend giữ trên **Vercel**; file này chỉ lo **BE + DB**.

> **Bảo mật:** Không commit mật khẩu thật. File `.env` trên VPS tự tạo từ `env.*.template` (placeholder `change_me_*`). Secret đã lỡ push → **đổi hết** trên VPS và đánh dấu resolved trên GitGuardian.

**VPS IP (của anh):** `160.250.4.181`

| Nhánh Git | Môi trường | API | Thư mục VPS | Database |
|-----------|------------|-----|-------------|----------|
| `develop` | Dev | https://api-dev.viethung.uk | `/opt/cleverschool-dev` | `lms_db_dev` |
| `staging` (hoặc `pre`) | Staging | https://api-staging.viethung.uk | `/opt/cleverschool-staging` | `lms_db_staging` |

Workflow CI/CD: `be/.github/workflows/deploy-api.yml` — push lên repo **`nviethung04/CleverSchool-be`** (nhánh `staging` / `develop`).

> Repo folder `CleverSchool/` (cha) trỏ remote khác — **CI chạy trên repo BE**, không phải folder cha.

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

### A5. Clone repo lần đầu lên VPS (khuyến nghị)

Cấu trúc **đúng** (clone cả repo, không chỉ `deploy/`):

```text
/opt/cleverschool-staging/
├── Dockerfile
├── go.mod, main.go, database/, ...
└── deploy/
    └── .env
```

GitHub **không** nhận mật khẩu tài khoản khi `git clone` — dùng **Personal Access Token** (PAT) hoặc repo **public**.

#### Cách 1 — Clone một lệnh (repo private, có PAT)

Trên VPS:

```bash
apt install -y git rsync
export GITHUB_TOKEN=ghp_xxxxxxxx   # PAT, scope repo (private) hoặc public_repo

git clone -b staging \
  "https://${GITHUB_TOKEN}@github.com/nviethung04/CleverSchool-be.git" \
  /opt/cleverschool-staging

cd /opt/cleverschool-staging/deploy
cp env.staging.viethung.template .env
nano .env    # sửa mật khẩu nếu cần
```

#### Cách 2 — Repo public (không cần token)

```bash
git clone -b staging https://github.com/nviethung04/CleverSchool-be.git /opt/cleverschool-staging
cd /opt/cleverschool-staging/deploy
cp env.staging.viethung.template .env && nano .env
```

#### Cách 3 — Đã clone vào `/tmp` rồi (trường hợp của anh)

Giữ file `.env` cũ, copy **cả repo** lên `/opt/cleverschool-staging`:

```bash
cp /opt/cleverschool-staging/deploy/.env /tmp/.env.staging.bak
rsync -a /tmp/CleverSchool-be/ /opt/cleverschool-staging/
cp /tmp/.env.staging.bak /opt/cleverschool-staging/deploy/.env
ls /opt/cleverschool-staging/Dockerfile   # phải có file này
```

#### Sau khi clone xong

```bash
docker network create cleverschool-edge 2>/dev/null || true
cd /opt/cleverschool-staging/deploy
docker compose up -d --build
docker compose logs api --tail 50
chmod +x scripts/seed-admin.sh && ./scripts/seed-admin.sh
```

**Caddy** (đã chạy rồi thì bỏ qua):

```bash
cp -r /opt/cleverschool-staging/deploy/proxy/* /opt/cleverschool-proxy/
cd /opt/cleverschool-proxy && docker compose up -d
```

#### Lần sau — không cần clone lại

Push nhánh `staging` → GitHub Actions tự deploy (cần Secrets). Hoặc `git pull` trong `/opt/cleverschool-staging` rồi `docker compose up -d --build`.

**SCP từ Windows** — chỉ khi không clone được; phải copy **cả source**, không chỉ `deploy/`.

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
POSTGRES_PASSWORD=change_me_dev_db
REDIS_PASSWORD=change_me_dev_redis
JWT_SECRET=change_me_dev_jwt_min_32_chars
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
POSTGRES_PASSWORD=change_me_staging_db
REDIS_PASSWORD=change_me_staging_redis
JWT_SECRET=change_me_staging_jwt_min_32_chars
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

Sau seed: user `admin` — **đổi mật khẩu ngay** (mật khẩu seed chỉ dùng lần đầu, không ghi trong Git).

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
  -d '{"username":"admin","password":"<your-admin-password>"}'

curl -s https://api-staging.viethung.uk/api/login \
  -X POST -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<your-admin-password>"}'
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

Sau phần B, mỗi khi anh **push code BE** lên nhánh `develop` hoặc `staging`, GitHub tự:

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

## Phần Z — Reset VPS sạch rồi deploy lại

Dùng khi muốn **xóa hết** container, volume DB, thư mục `/opt/cleverschool-*` và làm lại từ đầu.

**Không xóa:** Docker engine, SSH, DNS Cloudflare, firewall.

**Có xóa:** Toàn bộ dữ liệu Postgres staging/dev trên VPS (volume).

### Z1. Trên VPS — dọn sạch

SSH vào VPS, chạy **một trong hai cách**:

**Cách A — script trong repo** (sau khi đã clone hoặc copy script):

```bash
bash /opt/cleverschool-staging/deploy/scripts/vps-reset.sh
# hoặc nếu chưa có repo: dán nội dung từ deploy/scripts/vps-reset.sh
```

**Cách B — lệnh tay (không cần file):**

```bash
cd /opt/cleverschool-staging/deploy 2>/dev/null && docker compose down -v --remove-orphans || true
cd /opt/cleverschool-dev/deploy 2>/dev/null && docker compose down -v --remove-orphans || true
cd /opt/cleverschool-proxy 2>/dev/null && docker compose down -v --remove-orphans || true

docker rm -f cs-caddy csstaging-api csstaging-postgres csstaging-redis \
  csdev-api csdev-postgres csdev-redis 2>/dev/null || true

docker volume rm cleverschool_postgres_data cleverschool_redis_data \
  cleverschool-proxy_caddy_data cleverschool-proxy_caddy_config 2>/dev/null || true

docker network rm cleverschool-edge 2>/dev/null || true

rm -rf /opt/cleverschool-dev /opt/cleverschool-staging /opt/cleverschool-proxy
rm -rf /tmp/CleverSchool-be
```

Gõ `yes` nếu script hỏi xác nhận.

### Z2. Đảm bảo code mới nhất trên GitHub

Trên máy Windows (repo `CleverSchool/be`), push lên **`CleverSchool-be`** nhánh `staging` (có migration `0025`, tắt dashboard/S3 mặc định).

### Z3. Deploy lại staging (khuyến nghị làm trước)

```bash
docker network create cleverschool-edge

# Repo public:
git clone -b staging https://github.com/nviethung04/CleverSchool-be.git /opt/cleverschool-staging

# Repo private:
# export GITHUB_TOKEN=ghp_xxxx
# git clone -b staging "https://${GITHUB_TOKEN}@github.com/nviethung04/CleverSchool-be.git" /opt/cleverschool-staging

cd /opt/cleverschool-staging/deploy
cp env.staging.viethung.template .env
nano .env
# Sửa FRONTEND_URL, ALLOW_ORIGINS = URL Vercel thật
# ENABLE_DASHBOARD_JOBS=false (mặc định trong template)
# Không cần AWS_* cho MVP

docker compose up -d --build
docker compose logs api --tail 50
chmod +x scripts/seed-admin.sh && ./scripts/seed-admin.sh

mkdir -p /opt/cleverschool-proxy
cp -r /opt/cleverschool-staging/deploy/proxy/* /opt/cleverschool-proxy/
cd /opt/cleverschool-proxy && docker compose up -d
```

Hoặc một lệnh (sau reset, cần `GITHUB_TOKEN` nếu private):

```bash
export GITHUB_TOKEN=ghp_xxxx   # bỏ qua nếu public
bash /opt/cleverschool-staging/deploy/scripts/vps-setup-staging.sh
```

### Z4. Kiểm tra

```bash
docker ps
# csstaging-api, csstaging-postgres, csstaging-redis, cs-caddy

curl -s https://api-staging.viethung.uk/api/login \
  -X POST -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<your-admin-password>"}'
```

### Z5. Dev (tùy chọn, sau staging ổn)

Lặp lại A5/A6/A7 với `/opt/cleverschool-dev`, nhánh `develop`, `env.dev.viethung.template`.

| Bước docs | Nội dung |
|-----------|----------|
| A1 | DNS (giữ nguyên nếu đã có) |
| A2 | SSH key (giữ nguyên) |
| A3 | Docker (đã cài thì bỏ qua) |
| Z1 | Reset sạch |
| Z3 | Clone + `.env` + compose + seed + Caddy |
| A10–A11 | curl + Vercel env |

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
| Log `ch.program_id does not exist` | Pull code mới (migration `0025`); hoặc `ENABLE_DASHBOARD_JOBS=false` (mặc định) |
| Log S3 / EC2 IMDS | Không cấu hình AWS trên VPS là bình thường — log chỉ stdout; không cần `AWS_*` cho MVP |
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
- [ ] Push `develop` / `staging` → Actions xanh
- [ ] Đổi mật khẩu admin sau seed

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
