# Danh mục tất cả bảng PostgreSQL — CleverSchool BE

Tra cứu **danh sách bảng** (không mô tả cột chi tiết). Chi tiết cột: [tables-reference.md](./tables-reference.md). Quan hệ ER: [design.md](./design.md).

**Nguồn:** `be/database/migrations/` (baseline `0001`–`0051`, migration mới nhất **0051**).  
**Cập nhật:** 2026-06-23

**Tổng:** **150 bảng** (không tính `schema_migrations` của golang-migrate).

---

## Bảng tra cứu nhanh (A–Z)

| # | Bảng | Nhóm | Migration gốc |
|---|------|------|----------------|
| 1 | `activity_logs` | Hệ thống / audit | 0019 |
| 2 | `administrative_regions` | Địa giới hành chính | 0018 |
| 3 | `administrative_units` | Địa giới hành chính | 0018 |
| 4 | `answer_coordinates` | Ngân hàng câu hỏi | 0006 |
| 5 | `answer_groups` | Ngân hàng câu hỏi | 0006 |
| 6 | `answer_matchings` | Ngân hàng câu hỏi | 0006 |
| 7 | `answer_positions` | Ngân hàng câu hỏi | 0006 |
| 8 | `answers` | Ngân hàng câu hỏi | 0006 |
| 9 | `assessment_criteria` | Đánh giá / báo cáo | 0048 |
| 10 | `assessment_criteria_groups` | Đánh giá / báo cáo | 0048 |
| 11 | `assessment_publishes` | Đánh giá / báo cáo | 0048 |
| 12 | `assessment_ref_lessons` | Bài giao ↔ bài học | 0045, 0050 |
| 13 | `assessment_score_details` | Đánh giá / báo cáo | 0048 |
| 14 | `assessment_scores` | Đánh giá / báo cáo | 0048 |
| 15 | `assessment_subcriteria` | Đánh giá / báo cáo | 0048 |
| 16 | `assessments` | Bài kiểm tra đánh giá | 0040 |
| 17 | `certificates` | HR | 0004 |
| 18 | `chapters` | Tổ chức / CT | 0002 |
| 19 | `chat_message_reactions` | Chat khóa học | 0015 |
| 20 | `chat_message_reads` | Chat khóa học | 0015 |
| 21 | `chat_messages` | Chat khóa học | 0015 |
| 22 | `classes` | Tổ chức / CT | 0002 |
| 23 | `cloned_questions` | Bài tập / đề | 0008 |
| 24 | `contest_round_joiner_classes` | Contest | 0017 |
| 25 | `contest_round_joiner_persons` | Contest | 0017 |
| 26 | `contest_round_joiner_provinces` | Contest | 0017 |
| 27 | `contest_round_joiner_schools` | Contest | 0017 |
| 28 | `contest_round_question_user_fill_in_blanks` | Contest | 0017 |
| 29 | `contest_round_question_user_groups` | Contest | 0017 |
| 30 | `contest_round_question_user_labelings` | Contest | 0017 |
| 31 | `contest_round_question_user_manual_scoring` | Contest | 0017 |
| 32 | `contest_round_question_user_matchings` | Contest | 0017 |
| 33 | `contest_round_question_user_multiple_choices` | Contest | 0017 |
| 34 | `contest_round_question_user_positions` | Contest | 0017 |
| 35 | `contest_round_users` | Contest | 0017 |
| 36 | `contest_rounds` | Contest | 0017 |
| 37 | `contests` | Contest | 0017 |
| 38 | `course_ref_semesters` | Lịch học | 0011 |
| 39 | `course_ref_study_shifts` | Lịch học | 0011 |
| 40 | `course_schools` | Tổ chức / CT | 0002 |
| 41 | `courses` | Tổ chức / CT | 0002 |
| 42 | `dashboard_report_courses` | Dashboard | 0012 |
| 43 | `dashboard_report_schools` | Dashboard | 0012 |
| 44 | `degrees` | HR | 0004 |
| 45 | `departments` | HR | 0004 |
| 46 | `employee_positions` | HR | 0004 |
| 47 | `exam_comments` | Chấm điểm / nhận xét | 0010 |
| 48 | `exam_question_user_fill_in_blanks` | Chấm điểm | 0009 |
| 49 | `exam_question_user_groups` | Chấm điểm | 0009 |
| 50 | `exam_question_user_labelings` | Chấm điểm | 0009 |
| 51 | `exam_question_user_manual_scoring` | Chấm điểm | 0009 |
| 52 | `exam_question_user_matchings` | Chấm điểm | 0009 |
| 53 | `exam_question_user_positions` | Chấm điểm | 0009 |
| 54 | `exam_question_users` | Chấm điểm | 0009 |
| 55 | `exam_ref_lessons` | Bài giao ↔ bài học | 0008, 0046 |
| 56 | `exam_users` | Chấm điểm | 0009 |
| 57 | `exams` | Bài kiểm tra | 0008 |
| 58 | `exercise_comments` | Chấm điểm / nhận xét | 0010 |
| 59 | `exercise_question_user_fill_in_blanks` | Chấm điểm | 0009 |
| 60 | `exercise_question_user_groups` | Chấm điểm | 0009 |
| 61 | `exercise_question_user_labelings` | Chấm điểm | 0009 |
| 62 | `exercise_question_user_manual_scoring` | Chấm điểm | 0009 |
| 63 | `exercise_question_user_matchings` | Chấm điểm | 0009 |
| 64 | `exercise_question_user_positions` | Chấm điểm | 0009 |
| 65 | `exercise_question_users` | Chấm điểm | 0009 |
| 66 | `exercise_ref_lessons` | Bài giao ↔ bài học | 0008, 0046 |
| 67 | `exercise_users` | Chấm điểm | 0009 |
| 68 | `exercises` | Bài luyện tập | 0008 |
| 69 | `feedbacks` | Phản hồi | 0013 |
| 70 | `flashcard_activities` | Flashcard | 0016 |
| 71 | `flashcard_sessions` | Flashcard | 0016 |
| 72 | `grades` | Tổ chức / CT | 0002 |
| 73 | `group_answers` | Ngân hàng câu hỏi | 0006 |
| 74 | `h5p_content_scores` | H5P | 0014 |
| 75 | `h5p_content_user_data` | H5P | 0014 |
| 76 | `h5p_contents` | H5P | 0014 |
| 77 | `history_uses` | Hệ thống | 0013 |
| 78 | `holidays` | Lịch học | 0011 |
| 79 | `homework_comments` | Chấm điểm / nhận xét | 0010 |
| 80 | `homework_question_user_fill_in_blanks` | Chấm điểm | 0009 |
| 81 | `homework_question_user_groups` | Chấm điểm | 0009 |
| 82 | `homework_question_user_labelings` | Chấm điểm | 0009 |
| 83 | `homework_question_user_manual_scoring` | Chấm điểm | 0009 |
| 84 | `homework_question_user_matchings` | Chấm điểm | 0009 |
| 85 | `homework_question_user_positions` | Chấm điểm | 0009 |
| 86 | `homework_question_users` | Chấm điểm | 0009 |
| 87 | `homework_ref_lessons` | Bài giao ↔ bài học | 0008, 0046 |
| 88 | `homework_user_skip_questions` | Chấm điểm | 0009 |
| 89 | `homework_users` | Chấm điểm | 0009 |
| 90 | `homeworks` | Bài tập về nhà | 0008 |
| 91 | `lesson_completions` | Tiến độ bài học | 0013 |
| 92 | `lesson_dependencies` | Metadata bài học | 0005 |
| 93 | `lesson_plan_completes` | Giáo án | 0007 |
| 94 | `lesson_plan_parts` | Giáo án | 0007 |
| 95 | `lesson_plan_ref_lessons` | Giáo án | 0007 |
| 96 | `lesson_plans` | Giáo án | 0007 |
| 97 | `lesson_plans_lessons` | Giáo án | 0007 |
| 98 | `lesson_ref_skills` | Metadata bài học | 0005 |
| 99 | `lesson_ref_tags` | Metadata bài học | 0005 |
| 100 | `lesson_ref_topics` | Metadata bài học | 0005 |
| 101 | `lesson_schedules` | Lịch học | 0011 |
| 102 | `lesson_vocabularies` | Flashcard | 0016 |
| 103 | `lessons` | Tổ chức / CT | 0002 |
| 104 | `level_standards` | Level test | 0009 |
| 105 | `level_test_question_users` | Level test | 0009 |
| 106 | `level_test_questions` | Level test | 0009 |
| 107 | `level_tests` | Level test | 0009 |
| 108 | `level_users` | Level test | 0009 |
| 109 | `levels` | Level test | 0009 |
| 110 | `medias` | Media / file | 0013, 0022 |
| 111 | `message_medias` | Chat khóa học | 0015 |
| 112 | `password_resets` | Auth | 0013 |
| 113 | `permissions` | Auth / phân quyền | 0003 |
| 114 | `programs` | Tổ chức / CT | 0002 |
| 115 | `provinces` | Địa giới hành chính | 0018 |
| 116 | `question_attributes` | Ngân hàng câu hỏi | 0006 |
| 117 | `question_ref_attributes` | Ngân hàng câu hỏi | 0006 |
| 118 | `questions` | Ngân hàng câu hỏi | 0006 |
| 119 | `role_permissions` | Auth / phân quyền | 0003 |
| 120 | `roles` | Auth / phân quyền | 0003 |
| 121 | `schools` | Tổ chức / CT | 0002 |
| 122 | `scorm_activities` | SCORM | 0014 |
| 123 | `scorm_attempts` | SCORM | 0014 |
| 124 | `scorm_cmi` | SCORM | 0014 |
| 125 | `scorm_sessions` | SCORM | 0014 |
| 126 | `semester_ref_holidays` | Lịch học | 0011 |
| 127 | `semesters` | Lịch học | 0011 |
| 128 | `settings` | Cài đặt hệ thống | 0043 |
| 129 | `skills` | Metadata bài học | 0005 |
| 130 | `source_questions` | Ngân hàng câu hỏi | 0006 |
| 131 | `student_vocabulary_progress` | Flashcard | 0016 |
| 132 | `study_report_criterias` | Đánh giá / báo cáo | 0048 |
| 133 | `study_report_publishes` | Đánh giá / báo cáo | 0048 |
| 134 | `study_reports` | Đánh giá / báo cáo | 0048 |
| 135 | `study_shifts` | Lịch học | 0011 |
| 136 | `subjects` | Tổ chức / CT | 0002 |
| 137 | `tags` | Metadata bài học | 0005 |
| 138 | `topics` | Metadata bài học | 0005 |
| 139 | `user_address` | User / HR | 0004 |
| 140 | `user_classes` | User | 0004 |
| 141 | `user_courses` | User | 0004, 0042 |
| 142 | `user_departments` | HR | 0004 |
| 143 | `user_positions` | HR | 0004 |
| 144 | `user_ref_roles` | Auth | 0003 |
| 145 | `user_ref_subjects` | User | 0004 |
| 146 | `user_sessions` | Auth / session | 0023 |
| 147 | `users` | Auth | 0003 |
| 148 | `vocabularies` | Flashcard | 0016 |
| 149 | `wards` | Địa giới hành chính | 0018 |
| 150 | `weeks` | Lịch học | 0011 |

---

## Theo nhóm nghiệp vụ

### 1. Tổ chức & chương trình học (10)
`schools`, `grades`, `classes`, `subjects`, `programs`, `courses`, `course_schools`, `chapters`, `lessons`

### 2. Auth & phân quyền (7)
`roles`, `permissions`, `role_permissions`, `users`, `user_ref_roles`, `password_resets`, `user_sessions`

### 3. Liên kết user & HR (11)
`user_classes`, `user_courses`, `user_ref_subjects`, `user_address`, `departments`, `employee_positions`, `user_departments`, `user_positions`, `degrees`, `certificates`

### 4. Metadata bài học (7)
`tags`, `topics`, `skills`, `lesson_ref_tags`, `lesson_ref_topics`, `lesson_ref_skills`, `lesson_dependencies`

### 5. Ngân hàng câu hỏi (11)
`source_questions`, `questions`, `answers`, `answer_positions`, `answer_matchings`, `answer_coordinates`, `answer_groups`, `group_answers`, `question_attributes`, `question_ref_attributes`, `cloned_questions`

### 6. Giáo án (5)
`lesson_plans`, `lesson_plan_parts`, `lesson_plan_ref_lessons`, `lesson_plans_lessons`, `lesson_plan_completes`

### 7. Bài giao (exam / homework / exercise / assessment) (8)
`exams`, `homeworks`, `exercises`, `assessments`, `exam_ref_lessons`, `homework_ref_lessons`, `exercise_ref_lessons`, `assessment_ref_lessons`

### 8. Chấm điểm — exam (9)
`exam_users`, `exam_question_users`, `exam_question_user_fill_in_blanks`, `exam_question_user_positions`, `exam_question_user_matchings`, `exam_question_user_labelings`, `exam_question_user_groups`, `exam_question_user_manual_scoring`, `exam_comments`

### 9. Chấm điểm — homework (10)
`homework_users`, `homework_question_users`, `homework_question_user_fill_in_blanks`, `homework_question_user_positions`, `homework_question_user_matchings`, `homework_question_user_labelings`, `homework_question_user_groups`, `homework_question_user_manual_scoring`, `homework_user_skip_questions`, `homework_comments`

### 10. Chấm điểm — exercise (9)
`exercise_users`, `exercise_question_users`, `exercise_question_user_fill_in_blanks`, `exercise_question_user_positions`, `exercise_question_user_matchings`, `exercise_question_user_labelings`, `exercise_question_user_groups`, `exercise_question_user_manual_scoring`, `exercise_comments`

### 11. Đánh giá & báo cáo học tập (9)
`assessment_criteria_groups`, `assessment_criteria`, `assessment_subcriteria`, `study_report_criterias`, `assessment_scores`, `assessment_score_details`, `assessment_publishes`, `study_reports`, `study_report_publishes`

### 12. Level test (6)
`level_standards`, `levels`, `level_tests`, `level_test_questions`, `level_test_question_users`, `level_users`

### 13. Lịch học (8)
`semesters`, `holidays`, `semester_ref_holidays`, `course_ref_semesters`, `study_shifts`, `course_ref_study_shifts`, `weeks`, `lesson_schedules`

### 14. Dashboard (2)
`dashboard_report_schools`, `dashboard_report_courses`

### 15. Media & tương tác (11)
`medias`, `feedbacks`, `history_uses`, `lesson_completions`, `h5p_contents`, `h5p_content_scores`, `h5p_content_user_data`, `scorm_activities`, `scorm_attempts`, `scorm_cmi`, `scorm_sessions`

### 16. Chat (4)
`chat_messages`, `message_medias`, `chat_message_reactions`, `chat_message_reads`

### 17. Flashcard (5)
`vocabularies`, `lesson_vocabularies`, `student_vocabulary_progress`, `flashcard_sessions`, `flashcard_activities`

### 18. Contest (15)
`contests`, `contest_rounds`, `contest_round_users`, `contest_round_joiner_schools`, `contest_round_joiner_provinces`, `contest_round_joiner_persons`, `contest_round_joiner_classes`, `contest_round_question_user_multiple_choices`, `contest_round_question_user_fill_in_blanks`, `contest_round_question_user_positions`, `contest_round_question_user_matchings`, `contest_round_question_user_labelings`, `contest_round_question_user_groups`, `contest_round_question_user_manual_scoring`

### 19. Địa giới hành chính (4)
`administrative_regions`, `administrative_units`, `provinces`, `wards`

### 20. Hệ thống (3)
`settings`, `activity_logs`, `user_sessions`

---

## Pattern bảng chấm điểm theo loại câu

Với mỗi prefix `exam_`, `homework_`, `exercise_` (và tương tự `contest_round_question_user_*`):

| Hậu tố bảng | Dạng câu |
|-------------|----------|
| `*_users` | Tổng điểm / trạng thái nộp bài |
| `*_question_users` | Trắc nghiệm |
| `*_question_user_fill_in_blanks` | Điền khuyết |
| `*_question_user_positions` | Sắp xếp / kéo thả vị trí |
| `*_question_user_matchings` | Ghép cặp |
| `*_question_user_labelings` | Gắn nhãn |
| `*_question_user_groups` | Phân loại |
| `*_question_user_manual_scoring` | Viết / nói (chấm tay) |

---

## Ghi chú

- **`schema_migrations`**: bảng nội bộ của golang-migrate, không liệt kê ở trên.
- **`dashboard_views`**: được nhắc trong tài liệu cũ; **không** có trong migration greenfield hiện tại — dùng bảng `dashboard_report_*` và query trực tiếp.
- **Partition `activity_logs`**: migration `0019` tạo bảng gốc; partition theo tháng có thể bổ sung sau (xem `be/docs/database/maintenance.md`).
- Khi thêm migration mới: cập nhật file này **và** [tables-reference.md](./tables-reference.md) **và** [CHANGELOG.md](./CHANGELOG.md).

---

## Lệnh kiểm tra trên DB (tùy chọn)

```sql
SELECT tablename
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY tablename;
```

So sánh kết quả với bảng trên để xác nhận môi trường đã migrate đủ (version = `EXPECTED_MIGRATION_VERSION` trong `be/deploy/`).
