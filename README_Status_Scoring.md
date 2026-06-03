# Status Scoring System

## Tổng quan
Hệ thống Status Scoring được thiết kế để quản lý trạng thái chấm điểm homework một cách tự động và hiệu quả.

## Cấu trúc Database

### Bảng `homework_users`
- **Cột mới**: `status_scoring` (SMALLINT)
- **Giá trị**:
  - `0`: Không cần chấm thủ công
  - `1`: Chưa chấm xong
  - `2`: Đã chấm xong

## Migration

### File: `0238_add_status_scoring_to_homework_users.up.sql`
- **Công dụng**: Thêm cột `status_scoring` vào bảng `homework_users`
- **Tính năng**: Tạo index để tối ưu performance

### File: `0238_add_status_scoring_to_homework_users.down.sql`
- **Công dụng**: Rollback migration, xóa cột `status_scoring`

## Cron Job

### File: `jobs/homework_status_scoring_cron_job.go`
- **Công dụng**: Tự động sync trạng thái chấm điểm mỗi ngày
- **Chức năng**:
  - `StartHomeworkStatusScoringCronJob()`: Khởi động cron job
  - `SyncHomeworkStatusScoringJob()`: Thực hiện sync
  - `syncHomeworkStatusScoring()`: Logic sync chính

## Command Line Tool

### File: `cmd/sync_homework_status_scoring.go`
- **Công dụng**: Chạy sync thủ công từ command line
- **Sử dụng**: `go run sync_homework_status_scoring.go`

## API Endpoint

### Route: `POST /api/internal/command/sync-homework-status-scoring`
- **Công dụng**: Trigger sync job qua API
- **Authentication**: Yêu cầu `internal.command` role
- **Controller**: `HomeworkStatusScoringController.SyncHomeworkStatusScoring()`

## Tối ưu hóa API

### API được cải thiện:
1. **`/api/dashboard/teacher/homework-overview-grade`**
   - Sử dụng `status_scoring` để đếm
   - Loại bỏ subquery phức tạp
   - Kiểm tra dữ liệu đã xóa

2. **`/api/dashboard/teacher/homework-unscored`**
   - Filter theo `status_scoring = 1`
   - Tối ưu joins
   - Kiểm tra dữ liệu đã xóa

3. **`/api/dashboard/teacher/homework-scored`**
   - Filter theo `status_scoring = 2`
   - Tối ưu joins
   - Kiểm tra dữ liệu đã xóa

## Lợi ích

### Performance
- Giảm số lượng joins không cần thiết
- Sử dụng index trên `status_scoring`
- Loại bỏ subquery phức tạp

### Data Integrity
- Kiểm tra dữ liệu đã xóa (`deleted_at IS NULL`)
- Đảm bảo tính nhất quán giữa các API
- Sync tự động trạng thái chấm điểm

### Maintainability
- Code đơn giản, dễ hiểu
- Logic rõ ràng, dễ maintain
- Tách biệt concerns
