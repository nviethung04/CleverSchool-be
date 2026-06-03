# CleverSchool Backend

Backend CleverSchool là dịch vụ Go cho hệ thống LMS: quản lý trường, lớp, khóa học, bài học, câu hỏi, bài tập, bài thi, đánh giá, báo cáo, media, thông báo, meeting và các tác vụ nền.

## Thành phần chính

- `main.go`: điểm vào của app; chạy server hoặc command như migrate, thống kê, recalculation.
- `app/`: khởi động config, DB, Redis, Firebase, Gin middleware, route, observer và cron job.
- `routes/`: khai báo API public, quản trị, học tập, dashboard, websocket và internal command.
- `controllers/`, `services/`, `repositories/`: luồng xử lý request theo từng module.
- `models/`: model GORM cho dữ liệu chính.
- `requests/`, `dto/`, `resources/`, `prot/`: input, output, format response và protobuf.
- `database/migrations/`: lịch sử thay đổi database.
- `docs/`: Swagger và tài liệu hệ thống.
- `jobs/`, `command/`: tác vụ nền và tác vụ chạy theo lệnh.

## Cách chạy nhanh

Yêu cầu cơ bản:

- Go 1.24
- Postgres
- Redis nếu bật `REDIS_ENABLED=true`
- File `.env` có thông tin DB, JWT, CORS, port và tích hợp ngoài nếu cần

Chạy local:

```bash
go run .
```

Chạy migration:

```bash
go run . migrate
```

Xem version migration:

```bash
go run . migrate:version
```

Generate protobuf:

```bash
make gen-proto
```

Build:

```bash
go build ./...
```

Chạy bằng Docker Compose:

```bash
docker compose up --build
```

## Tài liệu nên đọc

- `AGENTS.md`: nguyên tắc làm việc, trình tự đọc project, cách suy nghĩ theo hệ thống.
- `docs/requirement.md`: yêu cầu và nhóm nghiệp vụ chính.
- `docs/architecture.md`: kiến trúc, luồng request, tác vụ nền, tích hợp ngoài.
- `docs/database.md`: database, migration, nhóm bảng và lưu ý dữ liệu.
- `docs/swagger.yaml`: tài liệu API hiện có.

## Ghi chú thuật ngữ

- LMS: hệ thống quản lý học tập.
- API: cổng backend cho frontend hoặc hệ thống khác gọi.
- CRUD: tạo, đọc, cập nhật, xóa.
- Migration: file SQL thay đổi database theo thứ tự version.
- Protobuf: định nghĩa kiểu dữ liệu để sinh code Go.
- Swagger: tài liệu API.
- Job: tác vụ chạy nền theo lịch.
- Command: tác vụ chạy bằng lệnh hoặc API nội bộ.
- H5P: nội dung học tập tương tác.
- SCORM: chuẩn đóng gói bài học e-learning.
