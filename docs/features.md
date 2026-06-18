# Features — CleverSchool LMS

Danh mục chức năng backend và điểm nối frontend. **Mỗi thay đổi API, schema hoặc hành vi nghiệp vụ phải cập nhật file này** (xem `be/AGENTS.md`).

Cập nhật: **2026-06-16**

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

### 3.2 Settings — đọc công khai theo key

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/api/manage/settings/by-key/:key` | Không | Dùng cho maintenance mode, popup lễ (`fe/proxy.ts`, `fe/components/holiday-popup.tsx`) |

### 3.3 Khóa học — người dùng trong khóa

| Method | Path | Permission | Ghi chú |
|--------|------|------------|---------|
| GET | `/api/manage/courses/:id/users` | `courses.show` | Cần cột `user_courses.main_teacher` (migration 0042) |

---

## 4. Sửa schema / hành vi lưu (0039–0044)

| Migration | Nội dung | Ảnh hưởng FE/BE |
|-----------|----------|-----------------|
| **0039** | `grades`: thêm `number`, `name_vn`, `name_en`; seed Khối 1–12 | Tạo lớp: không gửi `grade_id=0` (FK). BE: `class_repository` Omit `GradeId` khi 0. FE: `normalizeGradeIdValue` |
| **0040–0041** | Bảng `assessments` + permissions | POST `/api/manage/assessments` |
| **0042** | `user_courses`: `main_teacher`, `is_current`, `start_time`, `end_time` | GET course users, sắp xếp GV chính |
| **0043–0044** | Bảng `settings` + permissions | GET/POST settings, đọc by-key |
| *(course)* | `course_ref_semesters` composite PK; model GORM | POST tạo khóa không lỗi `RETURNING id` |

Chi tiết cột: [database/tables-reference.md](./database/tables-reference.md).  
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

| Slug ẩn | Ghi chú |
|---------|---------|
| `admin/settings` | Settings API đã có; bỏ slug khỏi mảng để hiện menu |
| `criteria-assessment`, `criteria-report` | Báo cáo tiêu chí — ngoài MVP |
| H5P, SCORM, contest, … | Mở rộng sau MVP |

Core hiển thị: Tổng quan, Người dùng, Trường, Môn học, Chương trình, Khóa học, Học liệu, Học kỳ.

---

## 7. Checklist khi thêm chức năng

1. Route + middleware + permission (`config/permission.go` + migration seed).
2. Model, repository, service, controller, proto (nếu đổi contract).
3. Migration `.up` / `.down` nếu đổi schema.
4. Cập nhật: **file này**, `tables-reference.md`, `CHANGELOG.md`, `AGENTS.md` (nếu quy ước mới).
5. FE: page/API client + `fe/docs/ui-design.md` hoặc `fe/AGENTS.md` nếu đổi UX.
6. **VPS:** tăng `be/deploy/EXPECTED_MIGRATION_VERSION`, cập nhật `docs/vps-release-checklist.md` §3 nếu cần.
7. `go run . refresh-permissions` trên dev; trên VPS: `post-deploy-check.sh` sau CI.
8. Restart backend sau deploy migration.

**Release lên VPS / Vercel:** [vps-release-checklist.md](./vps-release-checklist.md)

---

## 8. Lỗi đã xử lý (2026-06-16)

| Triệu chứng | Nguyên nhân | Cách xử lý |
|-------------|-------------|------------|
| POST class FK `classes_grade_id_fkey` | `grade_id=0` | Migration 0039 + Omit 0 ở BE + normalize FE |
| POST assessments 404 | Chưa có route | Module assessments 0040–0041 |
| GET course users 500 | Thiếu `main_teacher` | Migration 0042 |
| GET settings 404 | Chưa có bảng/route | Module settings 0043–0044 |
| GET course-schedule/family 404 | Chưa có route | `course_family_service` |
| React controlled→uncontrolled (program edit) | `form.cover` undefined + input Link | Bỏ Link; `cover ?? ""` khi load |
| POST course `RETURNING id` trên `course_ref_semesters` | Model PK sai | Composite PK trên model |
