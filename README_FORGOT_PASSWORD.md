# Chức năng Quên mật khẩu - LMS System

## Tổng quan

Đã triển khai thành công chức năng quên mật khẩu với gửi email thông báo cho hệ thống LMS. Chức năng này cho phép người dùng đặt lại mật khẩu khi quên mật khẩu hiện tại thông qua email.

## Các tính năng đã triển khai

### ✅ Core Features

- [x] Yêu cầu đặt lại mật khẩu qua email
- [x] Gửi email với template HTML đẹp
- [x] Token bảo mật với thời hạn 2 giờ
- [x] Đặt lại mật khẩu với validation
- [x] Rate limiting để tránh spam
- [x] Xóa tự động các token hết hạn

### ✅ Security Features

- [x] Token một lần sử dụng
- [x] Validation email tồn tại
- [x] Giới hạn tần suất yêu cầu (10 phút)
- [x] Mật khẩu tối thiểu 6 ký tự
- [x] Rate limiting API (3 requests/minute)

### ✅ Database & Models

- [x] Model PasswordReset
- [x] Migration cho bảng password_resets
- [x] Repository pattern
- [x] Cron job dọn dẹp dữ liệu

## Cấu trúc Files

```
be-lms/
├── models/
│   └── password_reset.go              # Model PasswordReset
├── database/migrations/
│   ├── 0108_create_password_resets.up.sql
│   └── 0108_create_password_resets.down.sql
├── services/
│   ├── auth_service.go                # Updated với forgot password
│   └── email_service.go               # Service gửi email
├── repositories/
│   ├── auth_repository.go             # Updated với GetByEmail
│   └── password_reset_repository.go   # Repository cho password reset
├── controllers/
│   ├── auth_controller.go             # Updated với forgot password
│   └── auth_controller_test.go        # Unit tests
├── routes/
│   └── auth.go                        # Updated routes
├── prot/
│   └── user.proto                     # Updated protobuf messages
├── i18n/languages/vi/
│   └── translations.vi.json           # Updated messages
├── jobs/
│   └── clear_cron_job.go              # Updated với cleanup
└── docs/
    └── FORGOT_PASSWORD.md             # Documentation
```

## API Endpoints

### 1. Yêu cầu đặt lại mật khẩu

```http
POST /api/forgot-password
Content-Type: application/json

{
  "username": "user@example.com"
}
```

### 2. Đặt lại mật khẩu

```http
POST /api/reset-password
Content-Type: application/json

{
  "token": "abc123...",
  "new_password": "newpassword123",
  "confirm_password": "newpassword123"
}
```

## Cấu hình

### 1. Environment Variables

Thêm vào file `.env`:

```env
# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=your-email@gmail.com
FROM_NAME=LMS System
FRONTEND_URL=http://localhost:3000
```

### 2. Database Migration

Chạy migration để tạo bảng:

```bash
# Từ thư mục database
migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" up
```

### 3. Protobuf Generation

Generate lại protobuf files:

```bash
cd prot
protoc --go_out=. --go_opt=paths=source_relative user.proto
```

## Sử dụng

### 1. Yêu cầu đặt lại mật khẩu

```bash
curl -X POST http://localhost:8080/api/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com"}'
```

### 2. Đặt lại mật khẩu

```bash
curl -X POST http://localhost:8080/api/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token": "abc123...",
    "new_password": "newpassword123",
    "confirm_password": "newpassword123"
  }'
```

## Bảo mật

### Rate Limiting

- **Forgot Password:** 3 requests per minute
- **Login:** 2 requests per second

### Token Security

- **Expiration:** 2 giờ
- **One-time use:** Token chỉ sử dụng được 1 lần
- **Auto cleanup:** Xóa token hết hạn mỗi giờ

### Email Validation

- Kiểm tra email tồn tại trong hệ thống
- Giới hạn tần suất yêu cầu (10 phút)
- Template email bảo mật

## Testing

Chạy unit tests:

```bash
go test ./controllers -v
```

## Monitoring

### Logs

- Email gửi thành công/thất bại
- Token tạo/xóa
- Rate limit violations

### Metrics

- Số lượng yêu cầu forgot password
- Số lượng email gửi thành công
- Số lượng password reset thành công

## Troubleshooting

### Email không gửi được

1. Kiểm tra cấu hình SMTP
2. Đảm bảo Gmail App Password đúng
3. Kiểm tra firewall/network

### Token không hợp lệ

1. Kiểm tra thời gian hết hạn
2. Đảm bảo token chưa được sử dụng
3. Kiểm tra database connection

### Rate limit errors

1. Đợi thời gian cooldown
2. Kiểm tra Redis connection
3. Kiểm tra rate limit configuration

## Future Enhancements

- [ ] SMS notification option
- [ ] Multiple email templates
- [ ] Admin dashboard for password reset requests
- [ ] Audit logging
- [ ] Email queue system
- [ ] Custom token expiration per user role

## Support

Nếu gặp vấn đề, vui lòng:

1. Kiểm tra logs
2. Chạy unit tests
3. Kiểm tra cấu hình environment variables
