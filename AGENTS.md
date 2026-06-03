# AGENTS.md

## Vai trò của file này
File này là chỉ dẫn chung cho người hoặc AI agent khi làm việc trong backend CleverSchool.
Khi trả lời anh Hùng, hãy mở đầu tự nhiên bằng "Chào anh Hùng" nếu phù hợp ngữ cảnh.
Mục tiêu là giúp đọc project theo hệ thống, không sửa vội theo từng file rời rạc.
Không dùng file này để ghi chi tiết API, schema, hoặc ticket cụ thể.
Chi tiết nghiệp vụ nên nằm trong `docs/requirement.md`.
Chi tiết kiến trúc nên nằm trong `docs/architecture.md`.
Chi tiết dữ liệu nên nằm trong `docs/database.md`.
Chi tiết endpoint đã có thể tra trong `docs/swagger.yaml` hoặc các file Swagger con.

## Nguyên tắc làm việc
Luôn hiểu luồng chính trước khi sửa code.
Luôn kiểm tra thay đổi sẵn có trong worktree trước khi chỉnh file.
Không revert thay đổi của người khác nếu không được yêu cầu rõ.
Ưu tiên sửa nhỏ, đúng vùng, dễ kiểm chứng.
Không tạo abstraction mới nếu code hiện tại đã có pattern đủ dùng.
Không đổi tên folder, package, hoặc route nếu chưa thấy lý do thật sự.
Không chỉnh migration cũ tùy tiện vì có thể ảnh hưởng database đã chạy.
Không thêm dependency mới nếu có thể dùng thư viện hiện có.
Không hard-code secret, token, password, webhook, hoặc domain môi trường.
Không log dữ liệu nhạy cảm như mật khẩu, token, thông tin định danh quan trọng.
Khi lỗi liên quan quyền truy cập, kiểm tra middleware trước khi kiểm tra service.
Khi lỗi liên quan dữ liệu, kiểm tra request, model, repository, migration theo thứ tự.
Khi lỗi liên quan response, kiểm tra resource, DTO, protobuf, và helper response.
Khi sửa logic học tập, kiểm tra cả homework, exam, exercise nếu chúng dùng chung luồng.
Khi sửa route quản trị, kiểm tra role permission tương ứng.
Khi sửa upload/media, kiểm tra local file, S3, tusd, và cleanup job.
Khi sửa meeting, kiểm tra provider cụ thể: Google, Microsoft, hoặc Zoom.
Khi sửa dashboard, kiểm tra cache, job tính toán, index database, và phân quyền.
Khi sửa migration, luôn xem cặp `.up.sql` và `.down.sql`.
Khi tạo migration mới, đặt số tiếp theo và tên mô tả rõ việc thay đổi.
Khi generate protobuf, dùng `make gen-proto` hoặc `buf generate` theo môi trường.
Khi build production, nhớ Dockerfile có bước generate protobuf.
Khi chạy app local, cần `.env`, Postgres, Redis nếu bật, và port hợp lệ.
Khi Redis tắt, code nên vẫn chạy được với các phần không bắt buộc Redis.
Khi Firebase thiếu credentials, push notification có thể tắt nhưng app không nên chết.
Khi DB replica không có, hệ thống có thể fallback về master.
Khi Swagger bật, endpoint tài liệu nằm dưới `/swagger/*any`.
Khi WebSocket có lỗi, kiểm tra timeout, CORS header, hub, client, và route ws.
Khi H5P hoặc SCORM lỗi, kiểm tra static assets, template, route, và config URL.
Khi có encoding tiếng Việt lạ trong file cũ, không mở rộng lỗi bằng chỉnh sửa lan man.
Khi viết tài liệu mới, ưu tiên câu ngắn, từ dễ hiểu, ghi rõ thuật ngữ.
Khi review code, nêu rủi ro trước, sau đó mới nêu phần đã ổn.
Khi chưa chắc, đọc thêm file thay vì đoán.

## Trình tự đọc project
Bước 1: đọc `README.md` để nắm mục tiêu và cách chạy tổng quan.
Bước 2: đọc `go.mod` để biết framework và thư viện chính.
Bước 3: đọc `main.go` để biết app chạy server hay command.
Bước 4: đọc `app/app_service.go` để hiểu khởi động hệ thống.
Bước 5: đọc `config/config.go` để biết biến môi trường và config runtime.
Bước 6: đọc `database/db/db.go` để hiểu kết nối Postgres, Redis, callback.
Bước 7: đọc `routes/routes.go` để hiểu các nhóm API lớn.
Bước 8: đọc `routes/module_routes.go` để hiểu pattern CRUD chung.
Bước 9: đọc controller của module cần sửa.
Bước 10: đọc service của module đó nếu có xử lý nghiệp vụ riêng.
Bước 11: đọc repository để hiểu query và transaction.
Bước 12: đọc model để hiểu bảng, quan hệ, field, soft delete.
Bước 13: đọc request/protobuf để hiểu input.
Bước 14: đọc resource/DTO để hiểu output.
Bước 15: đọc migration liên quan trước khi thay database.
Bước 16: đọc jobs nếu chức năng có chạy nền hoặc dữ liệu cache.
Bước 17: đọc middleware nếu chức năng liên quan auth, role, timeout, i18n, log.
Bước 18: đọc utils nếu chức năng liên quan JWT, response, upload, scoring adapter.
Bước 19: đọc docs Swagger nếu cần đối chiếu endpoint.
Bước 20: chạy test hoặc build phù hợp sau khi sửa.

## Cách nhìn hệ thống
Hãy xem project như một backend LMS, không chỉ là tập hợp controller.
Người dùng đi qua auth, role, route, controller, service, repository, database.
Dữ liệu học tập đi qua lesson, question, homework, exam, exercise, score, report.
Dữ liệu quản trị đi qua school, class, course, role, permission, user.
Dữ liệu hiển thị đi qua resource, DTO, protobuf, và response helper.
Dữ liệu nền đi qua cron job, command nội bộ, cache, và migration.
Tích hợp ngoài gồm S3, Firebase, Google, Microsoft, Zoom, H5P, SCORM.
Một thay đổi nhỏ ở model có thể ảnh hưởng route, dashboard, import/export, và job.
Một thay đổi nhỏ ở permission có thể làm API đúng logic nhưng người dùng không gọi được.
Một thay đổi nhỏ ở migration có thể làm dữ liệu production khó rollback.
Vì vậy luôn hỏi: thay đổi này đi vào hệ thống từ đâu và đi ra ở đâu?

## Quy tắc code
Giữ style Go hiện có của repo.
Controller nên mỏng, nhận request và trả response.
Service nên chứa nghiệp vụ chính khi module có service.
Repository nên chứa query và thao tác database.
Model nên mô tả dữ liệu, không nhồi quá nhiều logic nghiệp vụ.
Resource và DTO nên định dạng dữ liệu trả ra.
Middleware nên xử lý việc cắt ngang request như auth, role, timeout, i18n.
Command nên dùng cho tác vụ chạy một lần hoặc nội bộ.
Job nên dùng cho tác vụ lặp lại theo lịch.
Observer nên dùng cẩn thận vì nó chạy ngầm theo GORM callback.
Không copy-paste route khi có thể dùng `RegisterModuleRoute`.
Không bỏ qua `RoleMiddleware` ở route quản trị.
Không trả raw error nhạy cảm cho client.
Không để transaction nửa vời khi thao tác nhiều bảng phụ thuộc nhau.
Không dùng query quá rộng cho dashboard nếu có thể lọc và index.
Không thay đổi response contract nếu frontend đang phụ thuộc.
Không đổi protobuf mà quên generate file `.pb.go`.
Không đổi Swagger mà quên kiểm tra endpoint thực tế.

## Quy tắc database
Postgres là database chính.
GORM là ORM chính.
Migration là nguồn lịch sử thay đổi schema.
Master DB dùng cho ghi và đọc chính.
Replica DB có thể dùng cho đọc nếu cấu hình.
Redis dùng cho cache, session, rate limit, hoặc các tác vụ phụ trợ.
Không sửa migration đã chạy ở môi trường thật nếu không có kế hoạch rõ.
Migration `.up.sql` phải đi cùng `.down.sql` khi có thể rollback hợp lý.
Index cần được cân nhắc cho dashboard, ranking, report, và query theo user/course.
Soft delete cần được hiểu trước khi viết query đếm hoặc lọc.
Foreign key có thể không đầy đủ, nên kiểm tra logic liên kết trong code.
Tên bảng thường đi theo model và migration, không đoán từ tên route.

## Quy tắc tài liệu
Tài liệu phải giúp người mới đọc nhanh hơn.
Tài liệu không nên biến thành nơi chép lại toàn bộ code.
Mỗi file tài liệu nên có phần ghi chú thuật ngữ nếu có thuật ngữ riêng.
Khi thêm module lớn, cập nhật requirement, architecture, database nếu liên quan.
Khi thêm command hoặc job quan trọng, ghi vào architecture.
Khi thêm bảng hoặc quan hệ chính, ghi vào database.
Khi thêm nghiệp vụ mới, ghi vào requirement.
Ngôn ngữ nên rõ, gần với cách team nói chuyện.
Tránh từ quá học thuật nếu có cách nói đơn giản hơn.
Ưu tiên ví dụ đường dẫn file thật hơn mô tả chung chung.

## Ghi chú thuật ngữ
Agent: người hoặc công cụ AI đang đọc và sửa project.
PRJ: viết tắt của project.
Module: một nhóm chức năng như users, courses, homeworks.
Route: đường dẫn API được Gin đăng ký.
Controller: nơi nhận request và gọi lớp xử lý tiếp theo.
Service: nơi chứa luật nghiệp vụ chính.
Repository: nơi gom truy vấn database.
Model: cấu trúc Go đại diện cho bảng hoặc dữ liệu lưu trữ.
Resource: lớp định dạng dữ liệu trả về client.
DTO: cấu trúc dữ liệu dùng để chuyển dữ liệu giữa các lớp.
Migration: file SQL thay đổi cấu trúc hoặc dữ liệu database theo version.
Master DB: database chính, thường dùng để ghi.
Replica DB: database phụ, thường dùng để đọc nếu có.
Middleware: lớp xử lý request trước hoặc sau controller.
Observer: hook chạy tự động quanh thao tác GORM.
Job: tác vụ chạy nền theo lịch.
Command: tác vụ chạy bằng CLI hoặc API nội bộ.
Swagger: tài liệu API dạng máy đọc và người đọc được.
Protobuf: định nghĩa kiểu dữ liệu để sinh code Go cho request/response.
H5P: nội dung học tập tương tác.
SCORM: chuẩn gói bài học e-learning.
