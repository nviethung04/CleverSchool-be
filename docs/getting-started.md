# Chạy Dự Án CleverSchool Từ Đầu

Hướng dẫn setup **greenfield** (DB mới, chưa production). Cập nhật: 2026-06-04.

**Deploy VPS (Docker + CI/CD GitHub):** xem [deploy-vps.md](./deploy-vps.md).

## Yêu Cầu

| Thành phần | Phiên bản khuyến nghị |
|------------|------------------------|
| Go | 1.24+ |
| PostgreSQL | 14+ (docker-compose dùng 15) |
| Redis | 7+ (bật trong `.env` — cache, permission, rate limit) |
| Node.js + pnpm | Cho frontend (`fe/`) |
| Docker Desktop | Tùy chọn — cách nhanh nhất cho BE + DB + Redis |

---

## Cách 1: Docker (Khuyến nghị — nhanh nhất)

### Bước 1 — Vào thư mục backend

```bash
cd be
```

### Bước 2 — Chạy stack

```bash
docker compose up -d --build
```

Services:

- API: http://localhost:8080  
- PostgreSQL: `localhost:5432` (user `lms_user`, db `lms_db`)  
- Redis: `localhost:6379` (password trong `docker-compose.yml`)

### Bước 3 — Migration

Container `lms-app` tự chạy `./myapp migrate` khi start (`docker-entrypoint.sh`, `RUN_MIGRATIONS=true`).

Kiểm tra log:

```bash
docker compose logs app
```

Kỳ vọng: migration lên version **21**, không `dirty`.

### Bước 4 — Kiểm tra API

```bash
curl http://localhost:8080/api/login
# hoặc bật Swagger: thêm ENABLE_SWAGGER=true vào environment app rồi restart
```

### Bước 5 — Frontend

```bash
cd ../fe
# Sửa fe/.env:
# NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api
pnpm install
pnpm dev
```

Mở http://localhost:3000

---

## Cách 2: Chạy local (Go trực tiếp)

### Bước 1 — PostgreSQL + Redis

Tạo database trống, ví dụ:

- Host: `localhost`, port: `5432`  
- Database: `lms_db`  
- User / password: tùy cấu hình  

Hoặc chỉ chạy DB bằng Docker:

```bash
cd be
docker compose up -d postgres redis
```

### Bước 2 — File `.env` backend

```bash
cd be
copy .env.example .env   # Windows
# cp .env.example .env   # macOS/Linux
```

Chỉnh cho khớp DB/Redis thật, ví dụ khi dùng `docker compose` chỉ postgres/redis:

```env
PORT=8080
DB_MASTER_HOST=localhost
DB_MASTER_PORT=5432
DB_MASTER_USER=lms_user
DB_MASTER_PASSWORD=<password trong docker-compose.yml>
DB_MASTER_NAME=lms_db
DB_REPLICA_HOST=localhost
DB_REPLICA_PORT=5432
DB_REPLICA_USER=lms_user
DB_REPLICA_PASSWORD=<cùng password>
DB_REPLICA_NAME=lms_db
REDIS_ENABLED=true
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=<redis password nếu có>
REDIS_DB=0
JWT_SECRET=<chuỗi bí mật đủ dài>
API_DOMAIN=http://localhost:8080
ALLOW_ORIGINS=http://localhost:3000
ENABLE_SWAGGER=true
FRONTEND_URL=http://localhost:3000
```

### Bước 3 — Migration (baseline v1)

```bash
go run . migrate
go run . migrate:version
```

Nếu DB cũ từ migration legacy: **drop database và tạo lại**, hoặc dùng DB mới.

Lỗi `dirty=true`:

```bash
go run . migrate:version
go run . migrate:force <version_đúng>
go run . migrate
```

### Bước 4 — Chạy server

```bash
go run .
```

Server mặc định port **8080** (biến `PORT`).

### Bước 5 — Frontend

Giống Cách 1, bước 5.

---

## Sau Migration — Seeder & Tài Khoản Admin

### Migration `0020` (tự chạy khi `go run . migrate`)

- Tạo 5 role: Admin (id=1), Teacher, Student, School, Read Only.
- Chỉ seed permission tối thiểu: `internal.command`.

**Chưa đủ** để đăng nhập portal quản trị — cần chạy **seeder Role** bước dưới.

### Seeder Role + Admin (bắt buộc cho dev)

```bash
cd be
go run . migrate          # nếu chưa chạy (hiện tới version 23)
go run ./database/seeder -model Role   # BẮT BUỘC — tạo admin + permissions
```

Nếu bỏ qua bước seeder, login `admin`/`admin123` sẽ báo **Không tìm thấy tài khoản**.

Seeder sẽ:

1. Xóa và tạo lại toàn bộ `permissions` + `role_permissions` (theo `config/permission.go`).
2. Gán full quyền admin từ `GetPermissions()`, quyền GV/HS tương ứng.
3. Tạo user **`admin` / `admin123`** nếu chưa có.
4. Gắn **`user_ref_roles`** (user ↔ role Admin id=1) — **bắt buộc** để `POST /api/login` hoạt động.

Đăng nhập:

```http
POST http://localhost:8080/api/login
Content-Type: application/json

{"username":"admin","password":"admin123"}
```

Response có `token` → gọi API kèm header: `Token: <token>`.

### Seeder khác (tùy chọn)

```bash
go run database/seeder/seeder.go -model Question
go run database/seeder/seeder.go -model QuestionAttribute
go run database/seeder/seeder.go -model Week
go run database/seeder/seeder.go -model Flashcard
```

Chạy không flag → seed câu hỏi mẫu + Role (mặc định cũ).

### Lưu ý bảo mật

- Đổi mật khẩu `admin123` ngay sau khi setup (API đổi mật khẩu hoặc `PUT /api/manage/users/:id/reset-password`).
- Không dùng `admin/admin123` trên staging/production.

---

## Kiểm Tra Nhanh (Smoke Test)

1. `POST /api/login` — nhận token.  
2. Gọi API có auth: header `Token: <jwt>`.  
3. `GET /api/manage/schools` — cần permission `schools.index`.  
4. Frontend login → dashboard.

---

## Lệnh Hữu Ích

| Lệnh | Mô tả |
|------|--------|
| `go run . migrate` | Chạy migration |
| `go run . migrate:version` | Xem version |
| `go run . migrate:force N` | Sửa dirty (cẩn thận) |
| `make gen-proto` | Sinh protobuf (nếu đổi `.proto`) |
| `docker compose down -v` | Xóa container + volume DB (reset hoàn toàn) |

---

## Tài Liệu Liên Quan

- [features.md](./features.md) — chức năng hệ thống  
- [database/migration-guide.md](./database/migration-guide.md) — chi tiết migration  
- [index.md](../index.md) — cửa vào docs backend  

---

## Lỗi Thường Gặp

| Triệu chứng | Hướng xử lý |
|-------------|-------------|
| `No .env file found` | Tạo `be/.env` từ `.env.example` |
| Connection refused Postgres | Kiểm tra host/port; Docker đã `healthy` chưa |
| Redis warning, app vẫn chạy | Bật Redis hoặc `REDIS_ENABLED=false` (một số tính năng giảm) |
| Migration dirty | `migrate:force` + sửa SQL / reset DB |
| CORS frontend | Thêm origin vào `ALLOW_ORIGINS` |
| 401 mọi API manage | Thiếu token hoặc user chưa có permission |
