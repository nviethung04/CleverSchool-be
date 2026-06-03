# Requirement

File này ghi yêu cầu hệ thống ở mức tổng quan. Không dùng để thay thế ticket chi tiết hoặc Swagger.

## Mục tiêu sản phẩm

CleverSchool Backend phục vụ hệ thống học tập và quản trị đào tạo. Hệ thống cần hỗ trợ quản trị dữ liệu trường học, tổ chức lớp và khóa học, xây nội dung học, giao bài, chấm điểm, theo dõi tiến độ, báo cáo và tích hợp nội dung học tương tác.

## Nhóm người dùng

- Quản trị hệ thống: quản lý user, role, permission, cấu hình và dữ liệu nền.
- Quản trị trường hoặc đơn vị: quản lý school, faculty, class, course, subject, program.
- Giáo viên: quản lý bài học, bài tập, bài thi, đánh giá, điểm và báo cáo lớp.
- Học viên: học bài, làm bài, xem kết quả, tham gia contest hoặc meeting.
- Hệ thống nội bộ: chạy job, command, đồng bộ dữ liệu, tạo báo cáo và dọn dữ liệu.

## Yêu cầu chức năng chính

### Xác thực và phân quyền

Hệ thống cần đăng nhập, đăng xuất, đổi mật khẩu, reset mật khẩu, quản lý session và lấy profile. API quản trị phải kiểm tra token và permission theo từng hành động.

### Quản trị danh mục học tập

Hệ thống cần quản lý school, faculty, department, class, course, program, subject, chapter, lesson, heading, tag, topic, skill, grade, semester, holiday và training level.

### Nội dung học và câu hỏi

Hệ thống cần quản lý lesson plan, lesson plan part, question, source question, question attribute, answer và các kiểu câu hỏi liên quan. Nội dung có thể gắn vào homework, exam, exercise hoặc bài học.

### Bài tập, bài thi và luyện tập

Hệ thống cần tạo, cập nhật, xóa, khôi phục, import/export và lấy danh sách homework, exam, exercise. Học viên cần lưu câu trả lời, nộp bài, bỏ qua câu hỏi, xem đáp án theo quyền.

### Chấm điểm và đánh giá

Hệ thống cần hỗ trợ chấm tự động, chấm thủ công, chấm nhiều loại câu hỏi, đánh giá của giáo viên, lưu điểm theo học viên và tính lại dữ liệu khi cần.

### Dashboard và báo cáo

Hệ thống cần cung cấp dashboard cho học viên, giáo viên, trường và khóa học. Dữ liệu dashboard có thể cần job tính toán, cache, index và API export.

### Media và upload

Hệ thống cần upload file, upload S3, tus resumable upload, quản lý media, dọn file và hỗ trợ tài nguyên tĩnh cho H5P, SCORM, PowerPoint nếu có.

### Meeting và thông báo

Hệ thống cần hỗ trợ Google, Microsoft, Zoom meeting, attendance, recording, notification log, notice và push notification qua Firebase khi được cấu hình.

### Contest và flashcard

Hệ thống cần hỗ trợ contest, contest round, joiner, question, result, ranking. Flashcard cần hỗ trợ vocabulary, lesson vocabulary, study session, progress và activity.

### Tác vụ nội bộ

Hệ thống cần command hoặc internal API để migrate, sync class main, sync homework status scoring, recalculate total questions, recalculate homework users, generate statistics và clear dữ liệu báo cáo khi cần.

## Yêu cầu phi chức năng

- API cần trả response thống nhất qua helper.
- API quản trị cần có auth và role middleware.
- Hệ thống cần chịu được Redis tắt nếu chức năng đó không bắt buộc Redis.
- App cần không chết khi Firebase credentials thiếu; chỉ tắt push notification.
- Migration cần chạy theo version và có cách xem version.
- Job nền không được làm nghẽn request chính.
- Dashboard và report cần chú ý hiệu năng query.
- Upload lớn cần tránh timeout và giới hạn bộ nhớ bất hợp lý.
- WebSocket cần giữ kết nối ổn định, không bị timeout HTTP thông thường cắt ngang.
- Log cần đủ để debug nhưng không lộ secret.

## Biên hệ thống

Backend này chịu trách nhiệm API, dữ liệu, auth, permission, scoring, job và tích hợp ngoài. Frontend chịu trách nhiệm giao diện. Database chịu trách nhiệm lưu trữ. Redis chịu trách nhiệm cache/session/rate limit khi bật. Provider ngoài chịu trách nhiệm meeting, storage hoặc push notification theo cấu hình.

## Ghi chú thuật ngữ

- Requirement: yêu cầu hệ thống cần đáp ứng.
- Chức năng: việc người dùng hoặc hệ thống cần làm được.
- Phi chức năng: yêu cầu về hiệu năng, bảo mật, ổn định, vận hành.
- Auth: xác thực người dùng bằng token/session.
- Permission: quyền cụ thể như `courses.index` hoặc `users.update`.
- Scoring: quá trình lưu và tính điểm.
- Dashboard: màn hình tổng hợp số liệu.
- Export/import: xuất hoặc nhập dữ liệu qua file.
- Internal API: API chỉ dành cho vận hành hoặc hệ thống nội bộ.
