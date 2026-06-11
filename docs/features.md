# Tài Liệu Chức Năng Hệ Thống — CleverSchool LMS Backend

Tài liệu mô tả **các chức năng đang có trong code backend** (`be-lms`), căn theo route, controller và job thực tế. Cập nhật lần cuối: **2026-06-04**.

**Liên quan:** [requirement.md](./requirement.md) (nhu cầu), [architecture.md](./architecture.md) (kiến trúc), [database/README.md](./database/README.md) (dữ liệu).

---

## 1. Tổng Quan

CleverSchool LMS là hệ thống quản lý học tập cho trung tâm / trường, gồm:

- Quản trị tổ chức (trường, lớp, khóa, bài học).
- Ngân hàng câu hỏi và đánh giá (bài kiểm tra, bài về nhà, bài luyện tập).
- Luồng học sinh làm bài, chấm tự động / chấm tay.
- Dashboard và báo cáo.
- Học liệu số (media, H5P, SCORM, PowerPoint).
- Chat lớp học realtime, flashcard, cuộc thi (contest).
- Job nền và API nội bộ vận hành.

Backend cung cấp **REST API** (Gin), không render UI. Frontend (Next.js) gọi API và hiển thị.

---

## 2. Nhóm Người Dùng & Vai Trò

| Vai trò (DB) | ID mặc định | Mô tả chức năng |
|--------------|-------------|-----------------|
| Admin | 1 | Toàn hệ thống: trường, user, role, permission, cấu hình |
| Teacher | 2 | Khóa/lớp được giao: nội dung, giao bài, chấm, dashboard giáo viên |
| Student | 3 | Học, làm bài, nộp bài, xem điểm; một phiên đăng nhập (token mới đá token cũ) |
| School | 4 | Quản trị trong phạm vi trường / chi nhánh |
| Read Only | 5 | Xem dữ liệu, hạn chế thao tác ghi |

- Một user có **nhiều role** (`user_ref_roles`).
- API quản trị `/api/manage/*` yêu cầu đăng nhập + permission dạng `resource.action` (ví dụ `courses.update`).
- Học sinh có thể liên kết **phụ huynh** (`parent_id`); API `GET /api/student-parent/:parent_id` lấy danh sách học sinh theo phụ huynh.

---

## 3. Bản Đồ API (Theo Nhóm Route)

| Nhóm | Prefix | Xác thực | Mục đích |
|------|--------|----------|----------|
| Auth | `/api` | Một phần public | Đăng ký, đăng nhập, profile, session |
| Quản trị | `/api/manage` | Token + permission | CRUD và nghiệp vụ quản lý |
| Học tập | `/api/study` | Token | Lưu điểm, xem đáp án, comment |
| Dashboard học tập | `/api/dashboard` | Token | Thống kê HS/GV, ranking |
| Dashboard báo cáo | `/api/dashboard`, `/api/dashboard/report` | Token | Báo cáo trường/khóa, export |
| Media | `/api/medias`, `/api/upload-file` | Token + quyền | Thư viện file, upload |
| H5P | `/api/h5p` | Token | Nội dung H5P, điểm, user data |
| SCORM | `/api/scorm`, `/scorm` | Một phần public | Gói SCORM, runtime API |
| Upload | `/upload`, `/api/tus-uploads` | Tus cần quyền | Upload thường / resumable |
| PowerPoint | `/power-point`, `api/power-point` | — | Serve và scan gói PPT |
| Chat SSE | `/api/manage/courses/.../chat/events` | Tự xác thực trong controller | Realtime tin nhắn |
| Nội bộ | `/api/internal`, `/api/internal/command` | `internal.command` | Thống kê, đồng bộ, maintenance |
| Khác | `/swagger`, `/console`, migration | Tùy cấu hình | Docs, CLI dev, upload public |

---

## 4. Chức Năng Chi Tiết Theo Module

### 4.1 Xác Thực & Tài Khoản (`/api`)

| Chức năng | Method & path | Ghi chú |
|-----------|---------------|---------|
| Đăng ký | `POST /register` | Rate limit |
| Đăng nhập | `POST /login` | Trả token; HS: token mới trong Redis |
| Quên mật khẩu | `POST /forgot-password` | Rate limit |
| Đặt lại mật khẩu | `POST /reset-password` | |
| Danh sách session | `GET /sessions` | |
| Đăng xuất / tất cả thiết bị | `POST /logout`, `POST /logout-all` | |
| Profile | `GET /profile` | |
| Đổi mật khẩu | `PUT /change-password` | |
| Chọn role đang dùng | `PUT /accept-roles/:role` | Khi user nhiều role |

**Quản trị user** (`/api/manage/users` — CRUD + import/export/restore):

- Tạo/sửa/xóa/khôi phục tài khoản, gán trường (`school_id`).
- `GET /users/export`, `GET /users/export-pdf` — xuất danh sách.
- `POST /users/import` — import.
- `PUT /users/:id/reset-password` — reset mật khẩu (admin).
- `GET /users/activated` — user đang kích hoạt.
- `GET /users/username-exists/:username` — kiểm tra trùng username.
- `GET /users/permissions` — quyền của user hiện tại.
- `PUT /profile`, `GET /profile` (manage) — hồ sơ trong portal quản trị.

---

### 4.2 Phân Quyền RBAC (`/api/manage`)

| Chức năng | API |
|-----------|-----|
| CRUD vai trò | `/roles` (index, show, store, update, destroy) |
| Xem/gán permission cho role | `GET /roles/:id/permissions`, `GET /permissions`, `POST /permissions` |
| Permission hiển thị trên UI | `GET/POST /permissions/display` |

Quy ước: mỗi module CRUD đăng ký qua `RegisterModuleRoute` → permission `users.index`, `courses.store`, …

---

### 4.3 Tổ Chức & Multi-Tenant

#### Trường / trung tâm — `schools`

- CRUD, soft delete, restore, **export/import**.
- Thông tin liên hệ, địa chỉ (`ward_code`), logo JSONB, thống kê lớp/HS.
- `GET /school-dashboard-starts` — số liệu tóm tắt dashboard trường.

#### Lớp hành chính — `classes`

- CRUD, restore, export/import.
- Gán `school_id`, `grade_id`, giáo viên chủ nhiệm (`teacher_info` JSONB).
- `GET/POST/PUT /classes/:id/users` — học sinh / giáo viên trong lớp.
- `PUT /class/user-relation` — cập nhật quan hệ user–class.

#### Khối — `grades`

- CRUD, restore.

#### Môn học — `subjects`

- CRUD, restore.

#### Chương trình — `programs`

- CRUD, restore — khung chương trình chuẩn.

#### Khóa học — `courses`

- CRUD, restore; liên kết `subject`, `program`, trường (`course_schools`).
- Trạng thái `state`: coming / active / finished.
- Clone thông tin (`clone_info`).
- `GET/POST/PUT /courses/:id/users` — GV/HS trong khóa.
- `GET /courses/:id/score` — điểm tổng hợp khóa.
- `POST /courses/resync-schedules` — đồng bộ lại lịch học.

#### Chương & bài học — `chapters`, `lessons`

- CRUD chapter/lesson, restore (lesson).
- `PUT /chapters/:id/sort-lessons` — sắp xếp thứ tự bài.
- `PUT /lessons/:id/completion` — đánh dấu hoàn thành bài (học sinh).
- Gắn tag, topic, skill; phụ thuộc bài (`lesson_dependencies`).

#### Địa giới — `provinces`

- `GET /provinces`, `GET /provinces/:code`, `GET /provinces/:code/wards`.

---

### 4.4 Lịch Học & Học Kỳ

| Thực thể | Chức năng |
|----------|-----------|
| Học kỳ `semesters` | CRUD đầy đủ (custom routes) |
| Ngày nghỉ `holidays` | CRUD |
| Ca học `study-shifts` | CRUD module |
| Tuần `weeks` | `GET /weeks`, `GET /weeks/by-date` |
| Lịch bài `lesson-schedules` | `GET /lesson-schedules`, gắn course/lesson/shift/week |
| Lịch theo lesson | `GET/POST /lessons/schedules` |
| Sao chép lịch | `POST /lesson-schedules/copy` |

---

### 4.5 Nội Dung Học — Giáo Án

| Module | CRUD | Chức năng đặc biệt |
|--------|------|---------------------|
| `lesson-plans` | Có | `PUT /lesson-plans/:id/complete` |
| `lesson-plan-parts` | Có | Phần trong giáo án |

**Gắn nội dung vào bài học theo khóa** (`/lessons/.../:lesson_id/:course_id`):

- Xem/ghi: giáo án, exam, homework, exercise cho cặp lesson + course.

---

### 4.6 Ngân Hàng Câu Hỏi

#### Câu hỏi — `questions`

- CRUD, restore, **export/import**.
- `PUT /questions/sync-keywords` — đồng bộ từ khóa (job liên quan).
- Loại câu (`question_type`): multiple_choice, fill_in_blanks, ordering, matching, drag_drop, labeling, category, speaking, writing.
- Dạng hiển thị (`display`): horizontal, vertical, word_count.
- File đính kèm: `file_info`, `file_infos` (JSONB).

#### Cấu trúc đáp án (theo loại câu)

- Trắc nghiệm: `answers`.
- Sắp xếp / kéo thả: `answer_positions`.
- Nối cặp: `answer_matchings`.
- Labeling tọa độ: `answer_coordinates`.
- Phân loại nhóm: `answer_groups`, `group_answers`.

#### Phân loại & nguồn

- `source-questions` — CRUD, restore.
- `question-attributes` — CRUD; `GET /question-attributes/parent`.
- `POST /question-relation` — quan hệ giữa câu hỏi.
- `tags`, `topics`, `skills` — CRUD, restore; gắn lesson qua bảng ref.

#### HR / metadata nhân sự (tùy trung tâm)

- `departments`, `employee-positions`, `degrees`, `certificates` — CRUD.

---

### 4.7 Đánh Giá Học Tập — Exam, Homework, Exercise

Ba loại bài dùng chung pattern: **định nghĩa đề → gắn lesson/khóa → giao cho HS → làm bài → chấm**.

| Loại | CRUD manage | Đặc điểm |
|------|-------------|----------|
| **Exam** | `/exams` | Có `time_limit`, `deadline`, `max_score` |
| **Homework** | `/homeworks` | Tiến độ nộp, `status_scoring`, skip câu |
| **Exercise** | `/exercises` | Luyện tập, tương tự exam |

**Chức năng chung (mỗi loại):**

| Chức năng | API pattern |
|-----------|-------------|
| Clone đề | `POST /{type}/:id/cloned` |
| Giao bài | `POST /{type}/:id/assigned` |
| Xem bài đã gắn lesson | `GET /{type}/:id/assigned-lessons` |

Câu hỏi trong đề lưu snapshot tại **`cloned_questions`** (JSON), không junction `*_questions`.

**API bảo trì:** `POST /api/homework/update-all-total-questions` — cập nhật `total_questions` hàng loạt.

---

### 4.8 Luồng Học Sinh & Chấm Điểm (`/api/study`)

#### Lưu điểm tự động (theo loại câu)

| Loại câu | API |
|----------|-----|
| Trắc nghiệm | `POST /save-score/multiple-choice` |
| Điền khuyết | `POST /save-score/fill-in-blank` |
| Sắp xếp / drag-drop | `POST /save-score/ordering-and-dragdrop` |
| Nối cặp | `POST /save-score/matching` |
| Labeling | `POST /save-score/labeling` |
| Category | `POST /save-score/category` |
| Gộp nhiều câu | `POST /save-score/bulk` |

#### Chấm tay (writing / speaking)

- `POST /save-answer/manual-scoring` — GV lưu nhận xét/điểm.
- `POST /save-score/manual-scoring` — chốt điểm chấm tay.

#### Bài về nhà

- `POST /save-score/submit-homework` — nộp bài.
- `POST /save-score/skip-question` — bỏ qua câu.
- `GET /save-score/check-submit-homework` — kiểm tra đủ điều kiện nộp.

#### Xem kết quả & danh sách

| Đối tượng | API |
|-----------|-----|
| GV xem bài thi HS | `GET /teacher/exam-answers` |
| HS xem bài thi của mình | `GET /student/exam-answers` |
| Tương tự exercise | `/teacher/exercise-answers`, `/student/exercise-answers` |
| HS bài thi theo tuần | `GET /student/exams-by-week` |
| Danh sách HS làm exam/exercise/homework | `/exam-students`, `/exercise-students`, `/homework-students` |
| Chi tiết exam theo course | `GET /exam-courses` |
| GV/HS homework answers | `/teacher/homework-answers`, `/student/homework-answers` |

#### Comment / nhận xét bài làm

- `POST /exam-comment`, `/homework-comment`, `/exercise-comment`.

---

### 4.9 Dashboard

#### Dashboard tổng quan & export (`/api/dashboard` — không prefix `/manage`)

| Chức năng | API |
|-----------|-----|
| Dashboard chính | `GET /dashboard` |
| Export | `GET /dashboard/export` |
| Báo cáo tổng hợp | `GET /dashboard/report` |
| Báo cáo trường (proto) | `GET /dashboard/report/schools` |
| Export trường | `GET /dashboard/report/schools/export` |
| Báo cáo khóa (proto) | `GET /dashboard/report/courses` |
| Export khóa | `GET /dashboard/report/courses/export` |

#### Dashboard học tập (`/api/dashboard` — nhóm trong `routes.go`)

**Học sinh:**

- `GET /student/exam`, `/student/homework` — thống kê.
- `GET /student/exam-list`, `/student/homework-list` — danh sách.

**Giáo viên — bài thi:**

- `/teacher/exam-scored`, `/teacher/exam-unscored`, `/teacher/exam-overview`.

**Giáo viên — bài về nhà:**

- `/teacher/homework`, `/teacher/homework/:homework_id`.
- `/teacher/homework-overview`, `/teacher/homework-overview-grade`.
- `/teacher/homework-unscored`, `/teacher/homework-scored`.

**Chung:**

- Filter list: `/exam-list`, `/homework-list`, `/school-list`, `/course-list`, `/lesson-list`, `/teacher-list`, `/subject-list`.
- `GET /exam/ranking` — xếp hạng bài thi.

**Chưa implement (comment trong code):** dashboard contest ranking/stats.

---

### 4.10 Thư Viện Media & Upload

| Chức năng | API / path |
|-----------|------------|
| Duyệt thư mục | `GET /medias/folders` |
| Duyệt file trong folder | `GET /medias/files/:folder_id` |
| Upload file | `POST /upload-file` (cần `medias.update`) |
| Tạo/xóa folder | `POST /medias/folders`, `DELETE ...` |
| Xóa file / cả cây | `DELETE /medias/files/:id`, `DELETE .../folder-and-files/:id` |
| Trang upload (HTML) | `GET /upload`, `GET /upload/s3` |
| Thống kê upload | `GET /api/upload/stats` |
| Upload resumable (tus) | `/api/tus-uploads/*` |

Lưu trữ: local public và/hoặc S3 (cấu hình `config/storage`).

---

### 4.11 H5P (Nội Dung Tương Tác)

| Chức năng | API |
|-----------|-----|
| Danh sách / chi tiết content | `GET /h5p/content`, `GET /h5p/content/:id` |
| Sửa / xóa | `PUT`, `DELETE /h5p/content/:id` |
| Lưu trạng thái học | `POST /h5p/content-user-data` |
| Lưu điểm | `POST /h5p/content-score` |
| Đọc user data | `GET /h5p/content-data/...`, `GET /h5p/content-user-data/...` |

---

### 4.12 SCORM

| Chức năng | API / path |
|-----------|------------|
| Launch bài SCORM | `GET /api/scorm/launch/:id` |
| Attempt | `GET /api/scorm/attempt/:id` |
| Danh sách activity | `GET /api/scorm/activities` |
| Upload gói | `POST /api/scorm/upload`, form `/scorm-upload` |
| Quản lý gói | `GET/DELETE /api/scorm/packages` |
| Runtime SCORM 1.2 / 2004 | `/api/scorm/v12/*`, `/api/scorm/v2004/*` (get/set/commit/terminate) |
| Serve file gói | `/scorm/*`, `/static/scorm/*`, `/scorm-viewer` |

---

### 4.13 PowerPoint

- Serve gói PowerPoint đã convert: `GET /power-point/*`.
- Quét/sync theo path: `PUT api/power-point/scan-by-path/:path`.

---

### 4.14 Chat Khóa Học

| Chức năng | API |
|-----------|-----|
| Gửi tin (kèm media) | `POST /courses/:id/chat/messages` |
| Upload file → media | `POST .../upload-files-to-medias` |
| Danh sách tin | `GET .../messages` |
| Xóa tin | `DELETE .../messages/:messageId` |
| Ghim | `POST .../pin`, `GET .../pinned` |
| Đếm / tin gần đây | `GET .../count`, `GET .../recent` |
| Reaction | `POST/DELETE/GET .../reactions` |
| Reply | `POST .../replies`, `GET .../with-replies` |
| Realtime SSE | `GET /api/manage/courses/:id/chat/events` (và multi-course, user events) |
| Kênh active | `GET /chat/channels`, `/chat/channels/stats` |

Hỗ trợ: text, file, image, system; soft delete; đọc tin (`chat_message_reads`).

---

### 4.15 Flashcard / Từ Vựng

| Chức năng | API |
|-----------|-----|
| CRUD từ | `POST/GET /flashcard/vocabularies` |
| Gắn từ vào lesson | `GET/POST /flashcard/lessons/:lessonId/vocabularies` |
| Phiên học | `POST /flashcard/sessions`, activities, complete, resume |
| Tiến độ | `GET .../progress`, `PUT /flashcard/vocabulary-progress/:id` |
| Session đang học | `GET .../active-session` |

---

### 4.16 Cuộc Thi (Contest)

| Chức năng | API |
|-----------|-----|
| CRUD contest | `/contests` + restore |
| CRUD vòng | `/contest-rounds` + restore |
| Contest + rounds | `GET /contests/:id/with-rounds`, `GET /contests/:id/rounds` |
| Vòng theo contest | `GET /contests/:id/contest-rounds` |
| Thí sinh vòng | `GET /contest-rounds/:id/users` |
| Joiner (school/province/class/person) | `GET/POST/DELETE .../joiners` |

Chấm điểm contest: bảng `contest_round_question_user_*` (tương tự exam). Dashboard contest: **chưa có API** (TODO trong routes).

---

### 4.17 AI Chấm Bài (Thử Nghiệm)

- `POST /api/manage/ai/grading/writing`
- `POST /api/manage/ai/grading/speaking`

---

### 4.18 Phản Hồi Hệ Thống — Feedback

- CRUD `feedbacks` + `PATCH .../status` (pending, approved, rejected, resolved).

---

### 4.19 Cảnh Báo Vận Hành — Warning System

- `GET /warnings/active-users-count`, `/warnings/active-users`
- `GET /warnings/failed-logins-count`, `/warnings/failed-logins`

Dùng cho admin theo dõi đăng nhập / user đang hoạt động.

---

### 4.20 API Nội Bộ & Lệnh Bảo Trì

**Yêu cầu permission:** `internal.command`

| Nhóm | Chức năng |
|------|-----------|
| Thống kê trường | Generate / generate-last-week / GET / clear |
| Thống kê khóa | Tương tự |
| Daily jobs | `POST /daily/school-statistics`, `course-statistics`, `all-statistics` |
| Liệt kê job | `GET /internal/jobs` |
| Command | Copy homework → exercise, clear exercise, sync homework lesson ids, sync homework status scoring |

---

### 4.21 Khác

| Chức năng | Mô tả |
|-----------|--------|
| Activity log | Middleware ghi log request vào DB |
| i18n | Thông báo lỗi đa ngôn ngữ (mặc định `vi`) |
| Swagger | `ENABLE_SWAGGER=true` → `/swagger/*` |
| Rate limit | Redis — login, forgot-password, global |
| Migration upload public | `POST /api/migration/upload-public` |
| Console dev | `POST /console/run-cli` — lệnh whitelist |
| CLI `main.go` | `migrate`, `migrate:version`, `migrate:force` |

---

## 5. Job Nền (Cron — Tự Chạy Khi Server Start)

| Job | Mục đích |
|-----|----------|
| `StartCleanupCronJob` | Dọn dữ liệu tạm / cũ |
| `StartSyncMediaCronJob` | Đồng bộ media |
| `StartDashboardCacheCronJob` | Cache dashboard (chỉ khi `ENABLE_DASHBOARD_JOBS=true`; không bắt buộc MVP) |
| `StartHistoryUseCronJob` | Lịch sử sử dụng / hoạt động |
| `StartSyncKeywordCronJob` | Đồng bộ keywords câu hỏi |
| `StartClearExportFilesCronJob` | Xóa file export cũ |
| `StartDatabaseBackupCronJob` | Backup database |
| `StartS3CleanupCronJob` | Dọn file S3 |
| `StartHomeworkStatusScoringCronJob` | Đồng bộ trạng thái chấm homework |
| `StartDailySchoolStatisticsCronJob` | Thống kê trường theo ngày |
| `StartDailyCourseStatisticsCronJob` | Thống kê khóa theo ngày |

---

## 6. Phân Loại Theo Mức Ưu Tiên (Tham Chiếu requirement.md)

### MVP — Lõi nghiệp vụ đang có API đầy đủ

- Auth, user, role, permission, school, class, course, lesson, chapter.
- Question bank, exam, homework, exercise, save-score, manual scoring.
- Clone / assigned, lesson schedule, semester/holiday.
- Dashboard HS/GV và report schools/courses (API theo request; job nền tắt mặc định trên VPS).
- Media upload cơ bản.

### Mở rộng — Đã có code, có thể tắt khi triển khai gọn

- Contest (chưa dashboard).
- Chat + SSE realtime.
- Flashcard.
- H5P, SCORM, PowerPoint.
- AI grading.
- Warning system, history use chi tiết.
- HR: department, degree, certificate, province/ward.
- Level test (model + DB, ít route public).

---

## 7. Luồng Nghiệp Vụ Tiêu Biểu (End-to-End)

### 7.1 Giáo viên giao bài kiểm tra

1. Tạo câu hỏi trong ngân hàng (`POST /questions`).
2. Tạo exam (`POST /exams`), gắn câu hỏi (cloned_questions).
3. Gắn exam vào lesson + course (`POST /lessons/exams/:lesson_id/:course_id` hoặc assigned).
4. `POST /exams/:id/assigned` — giao cho nhóm HS.

### 7.2 Học sinh làm bài

1. Đăng nhập → token.
2. Lấy đề (manage/show hoặc study APIs).
3. Làm từng câu → `POST /api/study/save-score/...`.
4. Nộp (homework: `submit-homework`).

### 7.3 Giáo viên chấm & theo dõi

1. `GET /api/study/teacher/*-answers` — xem bài làm.
2. Chấm tự luận: `manual-scoring`.
3. Dashboard: `/api/dashboard/teacher/...`.

---

## 8. Ràng Buộc Kỹ Thuật Quan Trọng

- **Token header:** `Token` (không phải Bearer mặc định).
- **Permission:** Backend quyết định quyền; frontend chỉ ẩn/hiện UI.
- **Multi-tenant:** Lọc theo `school_id` / course_schools tùy repository.
- **Replica DB:** Đọc có thể qua replica nếu cấu hình.
- **Redis:** Cache permission, rate limit, token HS; app vẫn chạy nếu Redis lỗi lúc startup (một số tính năng suy giảm).

---

## 9. Cập Nhật Tài Liệu Này

Khi thêm/sửa/xóa route hoặc chức năng nghiệp vụ:

1. Sửa file **`be/docs/features.md`** (mục tương ứng).
2. Nếu đổi API contract → cập nhật Swagger / proto nếu có.
3. Ghi ngày cập nhật ở đầu file.

Checklist đầy đủ cho schema: [database/maintenance.md](./database/maintenance.md).

---

## 10. Bảng Tra Cứu Module CRUD (`/api/manage`)

| Resource path | Export | Import | Restore |
|---------------|--------|--------|---------|
| questions | ✓ | ✓ | ✓ |
| classes | ✓ | ✓ | ✓ |
| lessons | — | — | ✓ |
| lesson-plans, lesson-plan-parts | — | — | — |
| exams, exercises, homeworks | — | — | — |
| contests, contest-rounds | — | — | ✓ |
| programs, courses, schools, subjects, chapters | schools/courses: ✓ | schools/classes: ✓ | ✓ |
| roles, users | users: ✓ | users: ✓ | users: ✓ |
| source-questions, tags, topics, skills, grades, question-attributes, study-shifts | — | — | ✓ |
| departments, employee-positions, degrees, certificates | — | — | — |

*Cột “—” = module chỉ có CRUD cơ bản qua `RegisterModuleRoute`.*
