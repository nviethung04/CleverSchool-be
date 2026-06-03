# Course Statistics System

## Mô tả
Hệ thống thống kê theo khóa học (courses) tương tự như school statistics nhưng tính toán dữ liệu theo từng khóa học thay vì theo trường.

## Cấu trúc Database

### Bảng `dashboard_report_courses`
```sql
CREATE TABLE dashboard_report_courses (
    id SERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL,
    total_students BIGINT DEFAULT 0,
    total_teachers BIGINT DEFAULT 0,
    active_students BIGINT DEFAULT 0,
    active_teachers BIGINT DEFAULT 0,
    students_completed_homework BIGINT DEFAULT 0,
    teacher_ids TEXT DEFAULT '', -- Danh sách ID các teacher thuộc khóa, phân cách bằng dấu phẩy
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## Logic Tính Toán

### 1. **Total Students**
- Đếm học sinh thuộc khóa học qua bảng `user_courses`
- Chỉ tính user có `role_id = 3` (học sinh)
- Chỉ tính user được tạo trước cuối ngày `end_date`
- Filter `deleted_at IS NULL`

### 2. **Total Teachers**
- Đếm giáo viên thuộc khóa học qua bảng `user_courses`
- Chỉ tính user có `role_id = 2` (giáo viên)
- Chỉ tính user được tạo trước cuối ngày `end_date`
- Filter `deleted_at IS NULL`
- Lưu danh sách `teacher_ids` (phân cách bằng dấu phẩy)

### 3. **Active Students**
- Đếm học sinh có hoạt động trong khoảng thời gian
- Sử dụng UNION ALL các bảng `activity_logs` theo tháng
- Filter theo `role_id = 3` và `created_at` trong khoảng thời gian
- Join với `user_courses` để lấy học sinh thuộc khóa

### 4. **Active Teachers**
- Đếm giáo viên có hoạt động trong khoảng thời gian
- Sử dụng UNION ALL các bảng `activity_logs` theo tháng
- Filter theo `role_id = 2` và `created_at` trong khoảng thời gian
- Join với `user_courses` để lấy giáo viên thuộc khóa

### 5. **Students Completed Homework**
- Đếm học sinh hoàn thành ít nhất 1 bài tập trong khoảng thời gian
- Join `homework_users` với `homeworks` để kiểm tra `questions_completed >= total_questions`
- Join với `user_courses` để lấy học sinh thuộc khóa
- Filter theo `created_at` trong khoảng thời gian

## Cách Sử Dụng

### 1. Chạy Migration
```bash
go run main.go migrate
```

### 2. Command Line
```bash
# Tạo thống kê cho khoảng thời gian cụ thể
go run main.go generate-course-statistics -start=2025-01-01 -end=2025-01-07

# Xem hướng dẫn
go run main.go generate-course-statistics -help
```

### 3. API Endpoints

#### Tạo thống kê cho khoảng thời gian cụ thể
**POST** `/api/internal/course-statistics/generate`
```json
{
  "start_date": "2025-01-01",
  "end_date": "2025-01-07"
}
```

#### Tạo thống kê cho tuần trước
**POST** `/api/internal/course-statistics/generate-last-week`

#### Lấy thống kê đã tạo
**GET** `/api/internal/course-statistics?start_date=2025-01-01&end_date=2025-01-07`
**GET** `/api/internal/course-statistics?start_date=2025-01-01&end_date=2025-01-07&course_id=1`

### 4. Response Format
```json
{
  "id": 1,
  "course_id": 1,
  "total_students": 100,
  "total_teachers": 5,
  "active_students": 80,
  "active_teachers": 4,
  "students_completed_homework": 75,
  "teacher_ids": "1,2,3,4,5",
  "start_date": "2025-01-01",
  "end_date": "2025-01-07",
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

## Khác Biệt Với School Statistics

| Aspect | School Statistics | Course Statistics |
|--------|------------------|-------------------|
| **Đơn vị tính** | Theo trường (schools) | Theo khóa học (courses) |
| **Học sinh** | `users.school_id` | `user_courses.course_id` |
| **Giáo viên** | Qua `user_courses → courses → course_schools` | Trực tiếp qua `user_courses` |
| **Teacher IDs** | Không lưu | Lưu danh sách teacher_ids |
| **Bảng lưu trữ** | `dashboard_report_schools` | `dashboard_report_courses` |

## Files Đã Tạo

1. **Migration**: `database/migrations/0505_create_dashboard_report_courses.up.sql`
2. **Model**: `models/dashboard_report_courses.go`
3. **Repository**: `repositories/dashboard_report_courses_repository.go`
4. **Service**: `services/dashboard_report_courses_service.go`
5. **Job**: `jobs/course_statistics_cron_job.go`
6. **Command**: `command/generate_course_statistics.go`
7. **Controller**: `controllers/internal_command_controller.go` (thêm methods)
8. **Routes**: `routes/routes.go` (thêm routes)

## Lưu Ý

- Cần có quyền `internal.command` để sử dụng API
- Sử dụng JWT authentication
- Dữ liệu được lưu vào bảng `dashboard_report_courses`
- Teacher IDs được lưu dưới dạng string phân cách bằng dấu phẩy
- Tất cả queries đều filter `deleted_at IS NULL`
