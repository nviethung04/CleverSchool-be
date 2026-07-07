# Features — CleverSchool LMS

Danh mục chức năng backend và điểm nối frontend. **Mỗi thay đổi API, schema hoặc hành vi nghiệp vụ phải cập nhật file này** (xem `be/AGENTS.md`).

Cập nhật: **2026-06-21**

---

## 1. Nhóm route

| Prefix | Mục đích |
|--------|----------|
| `/api/manage` | Quản trị: CRUD lõi, nội dung, người dùng, quyền |
| `/api/study` | Học sinh: nộp bài, lưu điểm, comment |
| `/api/dashboard` | Thống kê, báo cáo |
| `/api/internal` | Lệnh nội bộ (cron, rebuild cache) |
| Upload | tus, S3, media, SCORM, H5P, PowerPoint |

Chi tiết kiến trúc: [architecture.md](./architecture.md).

---

## 2. Module CRUD chuẩn (`RegisterModuleRoute`)

Pattern: `GET/POST /api/manage/{resource}`, `GET/PUT/DELETE /api/manage/{resource}/:id`.  
Permission: `{resource}.index|show|store|update|destroy`.

### Đã có từ baseline (0001–0021)

users, roles, permissions, schools, grades, classes, subjects, programs, courses, chapters, lessons, semesters, holidays, exams, homeworks, exercises, lesson-plans, tags, topics, skills, medias, flashcards, contests, feedbacks, …

### Bổ sung sau baseline

| Resource | Migration | Permission | Ghi chú |
|----------|-----------|------------|---------|
| `assessments` | 0040, 0041 | `assessments.*` | Bài kiểm tra đánh giá (mini_test, tiêu chí). FE: `/manage/assessments` |
| `settings` | 0043, 0044 | `settings.*` | Cài đặt hệ thống key/value JSONB. FE: `/admin/settings` (menu có thể ẩn — xem §6) |

Sau khi thêm permission mới, chạy:

```bash
cd be && go run . refresh-permissions
```

---

## 3. Route tùy biến (không CRUD generic)

### 3.1 Lịch học — nhóm khóa clone (`course-schedule`)

| Method | Path | Permission | Mô tả |
|--------|------|------------|-------|
| GET | `/api/manage/course-schedule/family` | `courses.show` | Danh sách khóa “anh em” cùng `clone_info` (parent + children) |
| POST | `/api/manage/course-schedule/sync-all-family` | `courses.update` | Đồng bộ **lịch học** (lesson schedules) từ khóa hiện tại sang các khóa clone cùng nhóm. SSE progress. **Không** đồng bộ tên, học sinh, khối |

Code: `be/services/course_family_service.go`, `be/routes/routes.go`.

Query: `course_id` (bắt buộc).

### 3.2 Dashboard học sinh — bài tập về nhà (homework / BTVN)

Tab **Bài tập về nhà** trên FE `/progress-report` (học sinh) gọi API **homework** (BTVN), không dùng exercise (luyện tập trong bài học).

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/api/dashboard/student/homework` | Token | Tổng quan + biểu đồ theo tuần: BTVN giao / đã làm, tỉ lệ đúng, câu hoàn thành. Query: `course_id`, `start_date`, `end_date` (unix) |
| GET | `/api/dashboard/student/homework-list` | Token | Bảng chi tiết từng BTVN. Query: `course_id`, `start_date`, `end_date`, `page`, `limit` |

**Nguồn dữ liệu:** `homeworks`, `homework_ref_lessons` (đã giao: `assigned_by > 0`), `homework_users`, `user_courses`. **Không bắt buộc** `lesson_schedules` / bảng `weeks` — nếu chưa xếp lịch tuần vẫn đếm BTVN đã giao theo `homework_ref_lessons` + khóa học.

Code: `be/repositories/dashboard_student_homework*.go`, FE: `fe/app/[locale]/progress-report/components/homework-tab.tsx`.

**Trang Bài tập HS** (`/assignments`): `GET /api/study/student/exams-by-week` — query `week_id` (có thể `0` khi chưa xếp lịch tuần), `course_id?`, `page`, `limit`. Trả khóa học + danh sách bài học kèm exam/homework/exercise/**assessment** đã giao (`assigned_by > 0` trên `*_ref_lessons`). **Danh sách bài học:** ưu tiên `lesson_schedules` theo tuần; nếu trống → mọi bài thuộc chương trình khóa (`chapters.program_id` hoặc `chapters.course_id` legacy); fallback cuối: bài có bài tập đã giao qua `*_ref_lessons`. Code: `be/repositories/exam_student_repository.go` (`getAssignedLessonsForCourse`, `GetAssessmentsByLesson`). FE: `fe/app/[locale]/assignments/page.tsx`, `assignments-page.tsx` (assessment → `/progress-report?tab=assessment`).

### 3.2.1 Dashboard học sinh — bài luyện tập (exercise)

API exercise (`/dashboard/student/exercise`, `exercise-list`) dùng cho **luyện tập trong bài học**, không thay tab BTVN trên báo cáo HS.

### 3.2.2 Dashboard học sinh — bài đánh giá (assessment)

Tab **Đánh giá** trên FE `/progress-report` (học sinh). **Không** gọi `/api/manage/assessments` (403 — chỉ role admin/GV).

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/api/dashboard/student/assessment-list` | Token | Dropdown chọn bài KT đánh giá. Query: `course_id` (bắt buộc), `subject_id?`, `type?` (`final_exam` / `mini_test`), `page`, `limit` |
| GET | `/api/dashboard/student/assessments` | Token | Bảng điểm / tiêu chí theo bài đã chọn. Query: `course_id` (bắt buộc), `assessment_id?`, `type?`, `page`, `limit` |
| POST | `/api/study/assessment-submit` | Token (HS) | Nộp file bài làm (`assessment_id`, `course_id`, `file_infos`). **Chỉ** khi bài đã giao (`assessment_ref_lessons.assigned_by > 0`) |
| POST | `/api/manage/assessments/:id/assigned` | Token (GV/Admin) | Giao / thu hồi assessment theo khóa. Body: `{ lesson_id, is_assigned, course_id }` |
| GET | `/api/manage/assessments/:id/assigned-lessons` | Token | Danh sách bài học đã giao của assessment |

**Nguồn dữ liệu:** `assessments`, `assessment_ref_lessons`, `assessment_scores`, `assessment_score_details`, `assessment_publishes`, `user_courses`.

**Điểm HS chỉ hiện sau khi GV publish** (`PUT /api/manage/publish-assessments`). Nhận xét: `study_reports` + `PUT /api/manage/publish-study-report`.

Code: `be/services/assessment_scoring_service.go`, `be/repositories/assessment_scoring_repository.go`, FE: `assessment-tab.tsx`, `fe/config/featureFlags.ts` (`ASSESSMENTS_ENABLED`).

### 3.2.3 Module assessment — quản trị, chấm điểm, nhận xét (2026-06-22)

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| POST | `/api/assessments/save-score/bulk` | Token (GV/Admin) | Lưu điểm hàng loạt theo tiêu chí |
| GET | `/api/dashboard/teacher/assessment/students` | Token | Danh sách HS + điểm. Query: `course_id`, `assessment_id`, `get_score` |
| GET/PUT | `/api/manage/publish-assessments` | Token | Trạng thái / xuất bản bảng điểm |
| GET/POST/PUT | `/api/manage/study-reports` | Token | CRUD báo cáo học tập (nhận xét sao/checklist) |
| GET | `/api/manage/study-reports/evaluates/:course_id` | Token | Danh sách HS đã/chưa đánh giá |
| PUT | `/api/manage/publish-study-report` | Token | Xuất bản nhận xét cho HS |
| POST/GET/PUT/DELETE | `/api/manage/assessment-criteria-group(s)/*` | Token (Admin) | Nhóm tiêu chí chấm điểm |
| POST/GET/PUT/DELETE | `/api/manage/study-report-criterias/*` | Token (Admin) | Mẫu tiêu chí nhận xét |

Migration **0048–0049**. Tab admin: `ASSESSMENTS_ENABLED = true`. Tab GV báo cáo: `TEACHER_REPORTS_ASSESSMENT_TAB_ENABLED = true`.

### 3.2.1 Dashboard GV — báo cáo BTVN (`/teacher/reports`)

Tab **Bài tập về nhà** (homework/BTVN, không phải exercise/luyện tập). Tab *Kiểm tra và đánh giá* / *Xuất mẫu* ẩn — `fe/config/hiddenModules.ts`.

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/api/dashboard/teacher/homework` | Token | Thống kê HS theo khóa (tổng BTVN, hoàn thành, …). Query: `course_id`, `start_date`, `end_date`, `chapter_id?`, `lesson_ids?`, `homework_ids?`, `page`, `limit`, `order_by` |
| GET | `/api/dashboard/teacher/homework/list-homeworks` | Token | Danh sách BTVN đã giao của **một HS** (modal khi bấm tên HS). Query: `course_id`, `student_id`, `start_date`, `end_date`, `chapter_id?`, `lesson_ids?`, `homework_ids?`, `page`, `limit`, `order_by`. Response JSON `{ homeworks[], total }` (không proto) |

Code: `be/repositories/dashboard_teacher_homework_list_repository.go`, FE: `fe/app/[locale]/teacher/reports/components/homework-tab.tsx`.

**Sau deploy:** restart backend (route `list-homeworks` mới).

### 3.2.2 Dashboard GV — báo cáo luyện tập & kiểm tra (`/teacher/reports`)

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/api/dashboard/exercise-list` | Token | Danh sách bài luyện tập theo khóa (filter tab báo cáo). Query: `course_id`, `lesson_ids?`, `start_date`, `end_date`, `is_assigned?` |
| GET | `/api/dashboard/teacher/exercise` | Token | Thống kê HS theo khóa (luyện tập). Query: `course_id`, `page`, `limit`, `lesson_id?` |
| GET | `/api/dashboard/teacher/exercise/list-exercises` | Token | Danh sách luyện tập đã giao của **một HS** (modal). Query: `course_id`, `student_id`, `start_date`, `end_date`, `chapter_id?`, `lesson_ids?`, `exercise_ids?`, `page`, `limit` |
| GET | `/api/dashboard/exam/ranking` | Token | Bảng xếp hạng / thống kê bài kiểm tra theo khóa. Query: `course_id`, `start_date`, `end_date`, `exam_id?`, `page`, `limit`, `order_by` |
| GET | `/api/dashboard/exam-list` | Token | Danh sách bài kiểm tra theo khóa (filter tab kiểm tra) |

FE: tab **Bài luyện tập** → `exercise-tab.tsx`; tab **Bài kiểm tra** → `exam-tab.tsx` (`/teacher/reports?tab=exercise|exam`).

**Quy ước thống kê (2026-06-22):**

- **Luyện tập:** *Hoàn thành* = có bản ghi `exercise_users` (đã nộp); *Đang làm* = có câu trả lời nhưng chưa nộp. Không lọc theo `lesson_id` trên `exercise_users` / `exercise_question_users`.
- **Kiểm tra:** Xếp hạng lấy `AVG(exam_users.ratio)` qua subquery — chỉ bài đã giao (`exam_ref_lessons.assigned_by > 0`, `course_id` khớp hoặc NULL/0). Biểu đồ phân bố đếm **lượt nộp bài** theo khoảng điểm %. FE hiển thị điểm dạng `%` (làm tròn). Bộ lọc ngày: tùy chọn; cuối ngày tính đến 23:59:59.

Code: `be/repositories/dashboard_teacher_exercise_student_repository.go`, `dashboard_exam_ranking_repository.go`, FE `exam-tab.tsx`.

### 3.3 Bài học — cập nhật nội dung & từ vựng

| Method | Path | Permission | Mô tả |
|--------|------|------------|-------|
| PUT | `/api/manage/lessons/:id` | `lessons.update` | Cập nhật metadata + gắn exam / homework / exercise / **assessment** / lesson_plan ở **mức chương trình** (`course_id IS NULL` trong bảng `*_ref_lessons`). Query `course_id` trên PUT **không** đổi chỗ lưu nội dung mẫu (2026-06-18) |
| PUT | `/api/manage/flashcard/lessons/:lessonId/vocabularies` | (flashcard) | **Đồng bộ** toàn bộ từ vựng của bài học — body `{ vocabulary_ids: number[] }`. FE gọi sau khi lưu form bài học |

**Mô hình dữ liệu**

- Chương + bài học thuộc **chương trình** (`chapters.program_id`). Khóa học chỉ tham chiếu `program_id` — không clone riêng từng bài khi sửa CT.
- `PUT /api/manage/chapters/:id`: FE thường chỉ gửi `title`, `description`, `lessons[]`. BE merge field còn lại từ bản ghi cũ. **`course_id` legacy** có thể `NULL` (migration `0030`); khi cập nhật/xóa, repository **không ghi** `course_id=0` (tránh vi phạm FK `chapters_course_id_fkey`) — giống `Create`.
- Liên kết bài tập/KT/LT/đánh giá: `exam_ref_lessons`, `homework_ref_lessons`, `exercise_ref_lessons`, `assessment_ref_lessons`. **Mức chương trình:** `course_id IS NULL` (không ghi `0` — FK `courses`). **Mức khóa:** `course_id` = id khóa hợp lệ.
- Từ vựng: bảng `lesson_vocabularies` — **không** qua `PUT /lessons`; dùng API flashcard ở trên. Body field `assessments` (số nhiều) trên `PUT /lessons`.

Code: `be/services/lesson_service.go` (`Update`), `be/repositories/lesson_repository.go`, `be/services/flashcard_service.go` (`SyncLessonVocabularies`).

FE: `fe/components/lesson-form.tsx`, `fe/lib/api/lessons.ts` (`syncLessonVocabularies`).

### 3.5 Giáo viên — danh sách HS & chấm bài (`/api/study`)

| Method | Path | Query | Mô tả |
|--------|------|-------|-------|
| GET | `/api/study/exam-students` | `exam_id`, `course_id?`, `page`, `limit` | Danh sách HS + trạng thái nộp/chấm bài kiểm tra. **Khi có `course_id`:** chỉ HS thuộc khóa đó và bài đã giao (`assigned_by > 0`) |
| GET | `/api/study/exercise-students` | `exercise_id`, `course_id?`, … | Tương tự cho luyện tập |
| POST | `/api/study/exercise-reset` | body: `exercise_id`, `user_id`, `lesson_id?` (optional, không lọc xóa) | GV reset lượt làm: xóa **mọi** `exercise_users` + đáp án + nhận xét của HS cho bài đó |
| GET | `/api/study/homework-students` | `homework_id`, `course_id?` | Danh sách HS + tiến độ bài luyện tập |
| GET | `/api/study/teacher/homework-answers` | `homework_id`, `user_id` | Chi tiết đáp án + điểm 1 HS (luyện tập) |
| GET | `/api/study/teacher/exam-answers` | `exam_id`, `user_id` | Chi tiết bài làm 1 HS (GV chấm Writing/Speaking) |
| POST | `/api/study/save-score/manual-scoring` | body: `exam_id`, `user_id`, `score_list` | Lưu điểm chấm tay |
| POST | `/api/study/save-score/bulk` | body: `exercise_id` **+ `lesson_id`**, `time`, `list_answers[]` | HS nộp bài luyện tập (trắc nghiệm). `lesson_id` bắt buộc — ghi `exercise_users` và đáp án theo bài học |

**Lưu đáp án HS (migration 0009):** `homework_users` / `exam_users` / `exercise_users` (tổng điểm, trạng thái); chi tiết từng câu: `*_question_users`, `*_question_user_fill_in_blanks`, …, `*_question_user_manual_scoring` (Writing/Speaking). Bài luyện tập: join báo cáo GV theo `(exercise_id, user_id, lesson_id)`.

FE luồng chấm: `/teacher/lessons/[id]` → tab Luyện tập / Kiểm tra → **Chấm bài** → danh sách HS → bấm HS → trang đáp án (`.../homework|exam|exercise/[id]/student/[studentId]`).

**Chưa có BE (404):** `/api/manage/study-reports/*` (tab **Kiểm tra và đánh giá** trong `/teacher/reports` — báo cáo theo tiêu chí, khác chấm bài luyện tập trong bài học). Module `criteria-report` ngoài MVP.

FE: `/teacher/lessons/[id]/exam/[examId]?course_id=...` → nút **Chấm điểm** / **Xem điểm**. Repository: `be/repositories/exam_student_repository.go`, `exercise_student_repository.go` (join từ `users`, lọc `user_courses.course_id` khi có `course_id`).

### 3.6 Cài đặt hệ thống (public by-key)

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/api/manage/settings/by-key/:key` | Không | Dùng cho maintenance mode, popup lễ (`fe/proxy.ts`, `fe/components/holiday-popup.tsx`) |

### 3.7 Khóa học — người dùng trong khóa

| Method | Path | Permission | Ghi chú |
|--------|------|------------|---------|
| GET | `/api/manage/courses/:id/users` | `courses.show` | Cần cột `user_courses.main_teacher` (migration 0042) |
| GET | `/api/manage/courses/:id/score` | `courses.show` | **HS** (role 3): bảng điểm khóa — `exam_results[]` (điểm thang 10 từ `ratio/10`), chỉ bài kiểm tra **đã giao** (`exam_ref_lessons.assigned_by > 0`). FE `/courses/[id]` tab Điểm đọc field `exam_results` (không phải `exam_ressults`). |

---

## 4. Sửa schema / hành vi lưu (0039–0044)

| Migration | Nội dung | Ảnh hưởng FE/BE |
|-----------|----------|-----------------|
| **0039** | `grades`: thêm `number`, `name_vn`, `name_en`; seed Khối 1–12 | Tạo lớp: không gửi `grade_id=0` (FK). BE: `class_repository` Omit `GradeId` khi 0. FE: `normalizeGradeIdValue` |
| **0040–0041** | Bảng `assessments` + permissions | POST `/api/manage/assessments` |
| **0042** | `user_courses`: `main_teacher`, `is_current`, `start_time`, `end_time` | GET course users, sắp xếp GV chính |
| **0043–0044** | Bảng `settings` + permissions | GET/POST settings, đọc by-key |
| **0045** | `assessment_ref_lessons` — liên kết assessment ↔ lesson | `PUT /lessons` field `assessments`; GET lesson trả `assessments` |
| **0046** | `*_ref_lessons`: partial unique theo `course_id` (CT vs khóa) | `POST /homeworks|exams|exercises/:id/assigned` — giao bài theo khóa |
| **0050** | `assessment_ref_lessons`: `assigned_at`, `assigned_by`; partial unique theo `course_id` | `POST /assessments/:id/assigned`; HS thấy assessment trên `exams-by-week` + `/assignments` |
| **0047** | `exercises.*` permissions cho role học sinh (3) | HS xem/làm exercise; cache Redis — chạy `refresh-permissions` |
| *(course)* | `course_ref_semesters` composite PK; model GORM | POST tạo khóa không lỗi `RETURNING id` |

Chi tiết cột: [database/tables-reference.md](./database/tables-reference.md).  
Danh sách tất cả bảng: [database/all-tables.md](./database/all-tables.md).  
Lịch sử: [database/CHANGELOG.md](./database/CHANGELOG.md).

---

## 5. Upload ảnh bìa (program / course)

- **Chỉ upload file**, không nhập URL trên form admin program.
- Folder upload: `programs/covers` (program), tương tự pattern course khi có ảnh.
- API upload: tus / upload service hiện có (`be/config/storage.go`, `be/README_UPLOAD.md`).
- FE quy ước: [fe/docs/ui-design.md](../../fe/docs/ui-design.md) §11.

Trang đã áp dụng upload-only:

- `fe/app/[locale]/manage/programs/create/page.tsx`
- `fe/app/[locale]/manage/programs/[id]/edit/page.tsx`

Trang **chưa** đồng bộ (vẫn có Upload/Link): `fe/app/[locale]/teacher/courses/[id]/edit/page.tsx`.

---

## 6. MVP — menu frontend

`fe/config/hiddenModules.ts` ẩn module khỏi sidebar; trang vẫn mở được bằng URL.

| Trạng thái | Module |
|------------|--------|
| **Đã bật menu** (2026-06-23) | Xếp hạng HS (`/ranking`), Phản hồi HS/GV (`/feedback`) |
| Tab Quản lý trường (role 4) | **Ẩn** — `HIDDEN_USER_LIST_TABS` gồm `schools` trong `fe/config/hiddenModules.ts` (role School chưa lọc dữ liệu theo trường; BE vẫn seed role 4) |
| **BE tương ứng** (migration `0051`) | `feedbacks.*`, `h5p-contents.*` trong `permission.go` + `role_permissions`; role **School (4)** seed qua `GetSchoolPermissions()`; route `/manage/feedbacks` và `/api/h5p/content*` có `RoleMiddleware` |
| **Redis** | `REDIS_ENABLED=true` trong `.env` / `be/deploy/.env.example` — session + cache permission; sau migrate chạy `go run . refresh-permissions` |
| Slug vẫn ẩn | `admin/subjects` (mặc định chỉ seed 1 môn *Tiếng Anh*), `admin/h5p`, `admin/settings`, `admin/export`, `admin/notices`, `admin/feedback`, `criteria-*`, `lecture-bank`, `contest`, `scorm`, HR extensions |
| Tab GV báo cáo | Assessment: bật (`TEACHER_REPORTS_ASSESSMENT_TAB_ENABLED`); Xuất mẫu: tắt |
| Sửa câu hỏi clone trong bài tập | Tắt (`CLONE_QUESTION_EDIT_ENABLED = false`) — sửa từ form BTVN/kiểm tra/luyện tập luôn cập nhật ngân hàng câu hỏi gốc, không gửi `homework_id`/`exam_id`/`exercise_id` |
| Nút Xuất/Nhập dữ liệu (Excel) | Tắt (`EXPORT_IMPORT_DATA_UI_ENABLED = false`) — admin users/schools, chi tiết trường (lớp/HS), ngân hàng câu hỏi GV; dashboard `DASHBOARD_FILTERS.export = false` |
| Dialog media câu hỏi (URL / Thư viện) | Tắt (`QUESTION_MEDIA_URL_LIBRARY_UI_ENABLED = false`) — tạo/sửa câu hỏi chỉ còn tab **Tải lên file** trong `MediaUploadDialog` |
| Trang Quản lý tập tin (`/admin/images`) | **Ẩn menu** — slug `admin/images` trong `HIDDEN_MODULES`; API `upload-file` và upload trong form câu hỏi/khóa vẫn dùng bình thường |
| Học kỳ & lịch tuần | **Ẩn** — `admin/semesters` trong `HIDDEN_MODULES`; `SEMESTER_SCHEDULE_UI_ENABLED = false` (chọn kỳ trên khóa, Schedule, chọn tuần HS/GV). Giao/làm bài theo khóa vẫn chạy (`week_id=0`) |

Core hiển thị: Tổng quan, Người dùng (tab HS / GV / Admin; tab Quản lý trường ẩn), Trường, Khóa học, Chương trình, Học liệu. (Học kỳ, Quản lý tập tin ẩn menu.) (Môn học ẩn menu vì mặc định chỉ có 1 môn; H5P ẩn menu). Thứ tự menu: khóa học (vận hành) trước chương trình (mẫu nội dung).

---

## 7. Checklist khi thêm chức năng

1. Route + middleware + permission (`config/permission.go` + migration seed).
2. Model, repository, service, controller, proto (nếu đổi contract).
3. Migration `.up` / `.down` nếu đổi schema.
4. Cập nhật: **file này**, `all-tables.md`, `tables-reference.md`, `CHANGELOG.md`, `AGENTS.md` (nếu quy ước mới).
5. FE: page/API client + `fe/docs/ui-design.md` hoặc `fe/AGENTS.md` nếu đổi UX.
6. **VPS:** tăng `be/deploy/EXPECTED_MIGRATION_VERSION`, cập nhật `docs/vps-release-checklist.md` §3 nếu cần.
7. `go run . refresh-permissions` trên dev; trên VPS: `post-deploy-check.sh` sau CI.
8. Restart backend sau deploy migration.

**Release lên VPS / Vercel:** [vps-release-checklist.md](./vps-release-checklist.md)

---

## 8. Lỗi đã xử lý (2026-06-16 – 2026-06-18)

| Triệu chứng | Nguyên nhân | Cách xử lý |
|-------------|-------------|------------|
| POST class FK `classes_grade_id_fkey` | `grade_id=0` | Migration 0039 + Omit 0 ở BE + normalize FE |
| POST assessments 404 | Chưa có route | Module assessments 0040–0041 |
| GET course users 500 | Thiếu `main_teacher` | Migration 0042 |
| GET settings 404 | Chưa có bảng/route | Module settings 0043–0044 |
| GET course-schedule/family 404 | Chưa có route | `course_family_service` |
| React controlled→uncontrolled (program edit) | `form.cover` undefined + input Link | Bỏ Link; `cover ?? ""` khi load |
| POST course `RETURNING id` trên `course_ref_semesters` | Model PK sai | Composite PK trên model |
| Lưu bài học “thành công” nhưng mất exam/homework/exercise | PUT kèm `course_id` → ghi ref theo khóa; xem lại không có `course_id` | BE luôn lưu ref ở `course_id=0`; FE không gửi `course_id` khi `UpdateLesson` |
| Từ vựng không lưu sau sửa bài học | `PUT /lessons` không xử lý vocab | `PUT /flashcard/lessons/:id/vocabularies` + FE `syncLessonVocabularies` (kể cả mảng rỗng để xóa hết) |
| Assessment không lưu trên form bài học | Thiếu bảng ref + FE gửi sai field `assessment` | Migration 0045 `assessment_ref_lessons`; BE/FE dùng `assessments` |
| POST `/*/assigned` 500 `ON CONFLICT` | DB chỉ UQ `(lesson_id, *_id)`, code conflict 3 cột | Migration 0046 partial unique theo `course_id` |
| PUT lesson 500 `column "dependency_id" does not exist` | GORM dùng sai tên cột | DB: `dependency_lesson_id`; model + repository đã map đúng |
| PUT lesson 500 duplicate `lesson_plan_ref_lessons` | Insert lại link đã tồn tại | `UpdateLessonPlan` skip nếu đã có `(lesson_plan_id, lesson_id)` |
| PUT lesson 500 `homework_ref_lessons_course_id_fkey` | Ghi `course_id = 0` vi phạm FK `courses` | Mức CT: `Omit("CourseId")` → NULL trong DB |
| `NOT IN (NULL)` khi xóa ref lesson rỗng | `DeleteOld*` với mảng ID rỗng | Chỉ thêm `NOT IN` khi `len(ids) > 0` |
| Báo cáo HS tab Bài luyện tập trống / 404 `exercise-list` | FE gọi homework API hoặc BE chưa restart | API exercise §3.2; restart BE sau deploy |
| HS 403 khi mở exercise từ bài học | Thiếu `exercises.show` trong Redis | Migration 0047 + `refresh-permissions`; đăng nhập lại |
| Câu hỏi **category** (vd. id 14) không hiện danh mục / chấm sai / đáp án đúng trống | Snapshot `cloned_questions` rỗng (`options`/`correct_answers`) sau khi sửa ngân hàng câu hỏi | `mergeCloneQuestionOptions` + `EnrichClonedQuestionMaps` (homework-answers); tự heal snapshot khi `GetCloned`; chấm qua `categoryQuestionForScoring` — **restart BE** |
