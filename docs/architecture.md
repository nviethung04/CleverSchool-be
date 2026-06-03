# Architecture

File này mô tả kiến trúc backend ở mức hệ thống. Chi tiết endpoint xem Swagger và `routes/`.

## Bức tranh tổng quan

Backend dùng Go, Gin, GORM, Postgres, Redis, Swagger, Protobuf và các tích hợp ngoài như S3, Firebase, Google, Microsoft, Zoom, H5P, SCORM.

Luồng chính:

```text
Client -> Gin route -> middleware -> controller -> service -> repository -> database
                                      -> resource/DTO/protobuf -> response
```

Luồng nền:

```text
Cron job / command / internal API -> service/repository -> database/cache/provider ngoài
```

## Điểm vào

`main.go` là điểm vào chính.

- Không có argument: chạy app server.
- `migrate`: chạy migration.
- `migrate:version`: xem version migration.
- `migrate:force <version>`: ép version migration.
- Các command thống kê, đồng bộ và tính lại dữ liệu: chạy tác vụ vận hành.

## Khởi động app

`app/app_service.go` chịu trách nhiệm:

- load `.env`
- init JWT, disk, logger
- kết nối Postgres
- kết nối Redis nếu bật
- init Firebase nếu có credentials
- tạo Gin server
- gắn session, recovery, i18n, rate limit, CORS, activity log
- đăng ký routes
- bật Swagger nếu cấu hình
- đăng ký GORM observer
- chạy cron job
- start HTTP server

## Layer chính

### Routes

`routes/routes.go` gom API thành các nhóm lớn:

- public endpoint
- `/api/manage`: API quản trị, có auth và role
- `/api/study`: API học tập, có auth
- `/api/dashboard`: API dashboard, có auth
- `/api/internal` và `/api/internal/command`: API vận hành nội bộ
- WebSocket, media, H5P, SCORM, migration, static

`routes/module_routes.go` tạo CRUD route chung cho các module.

### Controller

Controller nhận request, đọc params/body, gọi service hoặc repository, rồi trả response. Controller không nên ôm quá nhiều luật nghiệp vụ.

### Service

Service chứa logic nghiệp vụ chính khi module có nhiều bước xử lý. Ví dụ auth, homework, contest, upload, meeting hoặc scoring.

### Repository

Repository gom query database, transaction và thao tác GORM. Đây là nơi cần cẩn thận với filter, preload, soft delete, index và phân trang.

### Model

Model mô tả dữ liệu lưu trong database. Model có thể được GORM dùng để map bảng, field, relation và soft delete.

### Resource, DTO, Protobuf

Resource và DTO định dạng dữ liệu trả ra hoặc chuyển giữa các lớp. Protobuf định nghĩa kiểu dữ liệu và sinh file `.pb.go`.

## Middleware

Các middleware quan trọng:

- Auth middleware: xác thực người dùng.
- Role middleware: kiểm tra permission cho route quản trị.
- Timeout middleware: giới hạn thời gian xử lý một số API.
- I18n middleware: xử lý ngôn ngữ.
- Activity log middleware: ghi hoạt động.
- Rate limit middleware: giới hạn tần suất request bằng Redis.
- Custom recovery: bắt panic và trả response ổn định hơn.

## Tác vụ nền

`app.CronJob()` khởi động nhiều job:

- cleanup
- sync media
- dashboard cache
- history use
- sync keyword
- clear export files
- database backup
- S3 cleanup
- homework status scoring
- daily school/course statistics

Command trong `command/` dùng cho tác vụ chạy thủ công hoặc internal API.

## Tích hợp ngoài

- Postgres: lưu dữ liệu chính.
- Redis: cache, rate limit, session hoặc dữ liệu tạm.
- Firebase: push notification.
- S3: lưu file nếu cấu hình upload S3.
- Google/Microsoft/Zoom: meeting, attendance, recording.
- H5P/SCORM: nội dung học tương tác.
- Swagger: tài liệu API.
- Buf/Protobuf: sinh Go code từ `.proto`.

## Điểm cần chú ý khi sửa

- Route quản trị cần đúng permission.
- Response nên đi qua helper thống nhất.
- Dashboard cần chú ý query nặng và index.
- Scoring có nhiều loại câu hỏi, không sửa một loại mà làm lệch loại khác.
- Homework, exam, exercise có nhiều logic giống nhau nhưng không hoàn toàn giống.
- Migration cũ có thể đã chạy ở môi trường thật.
- Observer chạy ngầm, dễ tạo tác dụng phụ nếu không hiểu rõ.
- Job nền có thể sửa dữ liệu ngoài luồng request.
- WebSocket cần cấu hình timeout và CORS riêng hơn request thường.

## Ghi chú thuật ngữ

- Architecture: cách các phần của hệ thống nối với nhau.
- Layer: lớp trách nhiệm trong code, ví dụ route, controller, service.
- Middleware: đoạn xử lý chen giữa request và controller.
- Repository: lớp thao tác database.
- Observer: hook tự động khi GORM create, update, delete.
- Cron job: tác vụ chạy theo lịch.
- Provider ngoài: dịch vụ bên ngoài như Firebase, S3, Google, Microsoft, Zoom.
- Cache: dữ liệu tạm để đọc nhanh hơn hoặc giảm query.
- Contract: thỏa thuận input/output giữa backend và frontend.
