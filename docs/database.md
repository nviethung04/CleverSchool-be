# Database

File này mô tả database ở mức hệ thống. Chi tiết cột và bảng cần xem `database/migrations/` và `models/`.

## Công nghệ

Backend dùng Postgres làm database chính và GORM làm ORM. Migration dùng `golang-migrate` với source nhúng từ `database/migrations/*.sql`.

Redis không phải database chính. Redis dùng cho cache, rate limit, session hoặc dữ liệu tạm khi được bật.

## Kết nối

Kết nối database nằm trong `database/db/db.go`.

- `MasterDB`: kết nối Postgres chính.
- `ReplicaDB`: kết nối đọc phụ nếu có cấu hình; nếu lỗi có thể fallback về master.
- `RedisClient`: kết nối Redis nếu `REDIS_ENABLED=true`.

Thông tin kết nối lấy từ `.env` qua `config/config.go`.

## Migration

Migration nằm trong `database/migrations/`.

Quy ước:

- File `.up.sql` dùng để áp thay đổi.
- File `.down.sql` dùng để rollback nếu có thể.
- Số đầu file là version.
- Tên file nên mô tả ngắn việc thay đổi.

Lệnh thường dùng:

```bash
go run . migrate
go run . migrate:version
go run . migrate:force <version>
```

Không nên sửa migration cũ đã chạy ở môi trường thật. Nếu cần đổi schema, tạo migration mới.

## Nhóm dữ liệu chính

### Người dùng và phân quyền

Các model liên quan thường gồm user, role, permission, user role, session, user device, department, position, degree, certificate.

Mục tiêu: quản lý danh tính, phân quyền và hồ sơ người dùng.

### Trường, lớp và khóa học

Các model liên quan thường gồm school, class, class main, course, course family, course school, faculty, program, subject, grade, semester, study shift, holiday.

Mục tiêu: tổ chức dữ liệu đào tạo và lịch học.

### Nội dung học

Các model liên quan thường gồm chapter, lesson, lesson plan, lesson plan part, heading, topic, tag, skill, teaching plan, training level.

Mục tiêu: mô tả nội dung dạy và học.

### Câu hỏi và đáp án

Các model liên quan thường gồm question, source question, question attribute, answer, answer position, answer coordinate, answer group, group answer, matching question, cloned question.

Mục tiêu: lưu ngân hàng câu hỏi, các dạng đáp án và dữ liệu clone.

### Homework, exam, exercise

Các model liên quan thường gồm homework, exam, exercise, các bảng ref lesson, user progress, comments, skip question, scoring fields.

Mục tiêu: giao bài, làm bài, nộp bài, chấm điểm và lưu kết quả.

### Dashboard và báo cáo

Các model liên quan thường gồm dashboard views, dashboard report schools, dashboard report courses, publish report, publish assessment, study report.

Mục tiêu: lưu dữ liệu tổng hợp để frontend đọc nhanh hơn.

### Media, H5P, SCORM

Các model liên quan thường gồm media, h5p content, h5p score, h5p user data, scorm.

Mục tiêu: lưu file, nội dung học tương tác và kết quả tương tác.

### Meeting và thông báo

Các model liên quan thường gồm google account, microsoft account, meeting, attendance, notification log, notice, meeting notification.

Mục tiêu: lưu tích hợp lớp học online, điểm danh, recording và thông báo.

### Contest và flashcard

Các model liên quan thường gồm contest, contest round, contest result, flashcard, vocabulary, session, progress nếu có trong module.

Mục tiêu: hỗ trợ thi đua, vòng thi, luyện từ vựng và tiến độ học.

## Lưu ý query

- Dashboard, ranking và report thường cần index tốt.
- Query theo `user_id`, `course_id`, `class_id`, `school_id`, `lesson_id`, `homework_id`, `exam_id` thường là query nóng.
- Soft delete có thể làm count sai nếu query không rõ điều kiện.
- Import/export có thể tạo tải lớn lên database.
- Job nền có thể ghi dữ liệu cùng lúc với request của người dùng.
- Khi thao tác nhiều bảng liên quan, cân nhắc transaction.
- Khi thêm field mới, kiểm tra resource, DTO, protobuf và frontend contract.

## Ghi chú thuật ngữ

- Database: nơi lưu dữ liệu lâu dài.
- Postgres: hệ quản trị database chính của project.
- GORM: thư viện Go giúp map model với database.
- Migration: lịch sử thay đổi database theo version.
- Schema: cấu trúc bảng, cột, index, constraint.
- Index: cấu trúc giúp query nhanh hơn.
- Transaction: nhóm thao tác database cùng thành công hoặc cùng hủy.
- Soft delete: xóa mềm, dữ liệu còn trong bảng nhưng được đánh dấu đã xóa.
- Master DB: database chính, thường dùng để ghi.
- Replica DB: database phụ, thường dùng để đọc.
- Cache: dữ liệu tạm để đọc nhanh hoặc giảm tải.
- Constraint: ràng buộc dữ liệu như unique hoặc foreign key.
