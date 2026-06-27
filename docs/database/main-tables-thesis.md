# Các bảng chính — CleverSchool LMS (cho đồ án)

Danh sách **rút gọn** các bảng cốt lõi mô tả trong đồ án / báo cáo. Không gồm bảng phụ (contest, HR, level test, chi tiết từng dạng câu hỏi, …).  
Danh mục đầy đủ: [all-tables.md](./all-tables.md).

**Tổng: 35 bảng chính**

---

## 1. Tổ chức & nội dung học (10 bảng)

| Bảng | Vai trò trong hệ thống |
|------|------------------------|
| `schools` | Trường học / trung tâm |
| `grades` | Khối lớp (1–12) |
| `classes` | Lớp học thuộc trường |
| `subjects` | Môn học |
| `programs` | Chương trình học (mẫu nội dung gốc) |
| `courses` | Khóa học triển khai từ chương trình |
| `course_schools` | Gán khóa học cho trường |
| `chapters` | Chương trong chương trình / khóa |
| `lessons` | Bài học trong chương |
| `lesson_schedules` | Xếp bài học theo tuần / lịch |

---

## 2. Người dùng & phân quyền (6 bảng)

| Bảng | Vai trò trong hệ thống |
|------|------------------------|
| `users` | Tài khoản Admin, Nhà trường, Giáo viên, Học sinh |
| `roles` | Vai trò (Admin, Teacher, Student, School, …) |
| `permissions` | Quyền thao tác (`resource.action`) |
| `role_permissions` | Gán quyền cho vai trò |
| `user_ref_roles` | Gán vai trò cho người dùng |
| `user_classes` | Học sinh / GV thuộc lớp nào |
| `user_courses` | HS / GV thuộc khóa học nào |

*Gộp nhóm 2 thực tế: 7 bảng — `user_courses` là bảng liên kết quan trọng nhất cho luồng dạy–học.*

---

## 3. Ngân hàng câu hỏi (4 bảng)

| Bảng | Vai trò trong hệ thống |
|------|------------------------|
| `source_questions` | Nguồn / bộ đề câu hỏi |
| `questions` | Câu hỏi (9 dạng: TN, điền khuyết, sắp xếp, …) |
| `answers` | Đáp án / lựa chọn của câu hỏi |
| `cloned_questions` | Bản sao câu hỏi gắn vào bài kiểm tra / BTVN / luyện tập |

---

## 4. Bài giao & liên kết bài học (7 bảng)

| Bảng | Vai trò trong hệ thống |
|------|------------------------|
| `exams` | Bài kiểm tra |
| `homeworks` | Bài tập về nhà (BTVN) |
| `exercises` | Bài luyện tập trong bài học |
| `assessments` | Bài kiểm tra đánh giá (mini test, tiêu chí) |
| `exam_ref_lessons` | Gắn / giao bài KT vào bài học (theo CT hoặc khóa) |
| `homework_ref_lessons` | Gắn / giao BTVN vào bài học |
| `exercise_ref_lessons` | Gắn / giao luyện tập vào bài học |
| `assessment_ref_lessons` | Gắn / giao bài đánh giá vào bài học |

*Gộp nhóm 4: 8 bảng — có thể gom 4 bảng `*_ref_lessons` thành một mục “bảng liên kết bài giao” trong đồ án.*

---

## 5. Làm bài & chấm điểm (4 bảng đại diện)

| Bảng | Vai trò trong hệ thống |
|------|------------------------|
| `exam_users` | Kết quả nộp bài kiểm tra của HS |
| `homework_users` | Kết quả nộp BTVN của HS |
| `exercise_users` | Kết quả luyện tập của HS |
| `exam_question_user_manual_scoring` | Chấm tay bài Viết / Nói (đại diện nhóm chấm chi tiết) |

**Ghi chú cho đồ án:** Hệ thống còn các bảng `*_question_users` và `*_question_user_*` theo từng dạng câu (trắc nghiệm, ghép cặp, …) — có thể mô tả chung là *“bảng lưu đáp án chi tiết từng câu”* mà không liệt kê hết.

---

## 6. Đánh giá & báo cáo (5 bảng)

| Bảng | Vai trò trong hệ thống |
|------|------------------------|
| `assessment_scores` | Điểm đánh giá theo tiêu chí |
| `study_reports` | Nhận xét / báo cáo học tập của GV |
| `assessment_publishes` | Xuất bản điểm cho HS xem |
| `dashboard_report_schools` | Thống kê theo trường |
| `dashboard_report_courses` | Thống kê theo khóa học |

---

## 7. Học liệu & phản hồi (3 bảng)

| Bảng | Vai trò trong hệ thống |
|------|------------------------|
| `lesson_plans` | Giáo án / bài giảng |
| `medias` | File đính kèm (ảnh, video, tài liệu) |
| `feedbacks` | Phản hồi của HS / GV |

---

## Sơ đồ quan hệ tóm tắt (mô tả trong đồ án)

```text
schools ── classes ── user_classes ── users ── user_ref_roles ── roles
    │
    └── course_schools ── courses ── user_courses
                              │
programs ── chapters ── lessons ── lesson_schedules
                │           │
                │           ├── exam/homework/exercise/assessment
                │           └── *_ref_lessons (giao bài)
                │
questions ── answers          └── *_users (kết quả HS)
```

---

## Bảng **không** cần đưa vào đồ án (trừ khi chuyên đề riêng)

| Nhóm | Ví dụ | Lý do |
|------|--------|--------|
| Contest | `contests`, `contest_rounds`, … | Module thi đấu, ngoài MVP |
| HR | `departments`, `degrees`, `certificates` | Quản lý nhân sự mở rộng |
| Level test | `level_tests`, `level_users` | Kiểm tra trình độ riêng |
| Chi tiết chấm từng dạng | `homework_question_user_matchings`, … | Quá chi tiết — mô tả pattern |
| SCORM / H5P chi tiết | `scorm_*`, `h5p_content_user_data` | Học liệu tương tác — nhắc tên module là đủ |
| Chat | `chat_messages`, … | Tính năng phụ |
| Flashcard chi tiết | `flashcard_activities`, … | Có thể chỉ nêu `vocabularies` |
| Hệ thống | `activity_logs`, `user_sessions` | Vận hành kỹ thuật |

---

## Đoạn mẫu cho Chương CSDL (copy vào đồ án)

> Cơ sở dữ liệu CleverSchool LMS được thiết kế quanh các nhóm thực thể chính: **(1)** tổ chức đào tạo (`schools`, `classes`, `programs`, `courses`, `lessons`); **(2)** quản lý người dùng và phân quyền (`users`, `roles`, `permissions`, `user_courses`); **(3)** ngân hàng câu hỏi (`questions`, `answers`); **(4)** bài giao học tập (`exams`, `homeworks`, `exercises`, `assessments` và bảng liên kết `*_ref_lessons`); **(5)** lưu kết quả làm bài (`*_users` và bảng đáp án chi tiết); **(6)** báo cáo và đánh giá (`dashboard_report_*`, `study_reports`, `assessment_scores`). Quan hệ giữa chương trình học và khóa học tuân theo mô hình *chương trình làm mẫu, khóa học tham chiếu và giao bài theo từng khóa*.

---

*Cập nhật: 2026-06-23 — đồng bộ migration 0051.*
