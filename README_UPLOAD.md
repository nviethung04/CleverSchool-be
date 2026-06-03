# Giao Diện Upload File - LMS Enspire

## Tổng quan

Giao diện upload file được thiết kế với giao diện hiện đại, responsive và hỗ trợ upload file với TUS protocol để có thể resume upload khi bị gián đoạn.

## Tính năng

### ✨ Giao diện hiện đại
- Thiết kế responsive, tương thích với mọi thiết bị
- Gradient background đẹp mắt
- Icons từ Font Awesome
- Animation mượt mà

### 🚀 Upload nâng cao
- **Drag & Drop**: Kéo thả file trực tiếp vào vùng upload
- **Multiple files**: Upload nhiều file cùng lúc
- **File preview**: Hiển thị thông tin file (tên, kích thước, loại)
- **Progress tracking**: Theo dõi tiến trình upload real-time
- **Resume upload**: Hỗ trợ tiếp tục upload khi bị gián đoạn (TUS protocol)

### ⚙️ URL Parameters
- **📏 width**: Kích thước chiều rộng giao diện (mặc định: 800)
- **📏 height**: Kích thước chiều cao giao diện (mặc định: 600)
- **📁 path**: Đường dẫn lưu trữ file (mặc định: uploads)
- **🔄 Tự động**: Giao diện tự động điều chỉnh theo các tham số từ URL
- **🎯 Iframe-ready**: Không có background, tối ưu cho việc nhúng iframe

### 📱 Responsive Design
- Tối ưu cho desktop, tablet và mobile
- Layout tự động điều chỉnh theo kích thước màn hình
- Touch-friendly trên thiết bị di động

## Cách sử dụng

### 1. Truy cập giao diện
```
GET /upload?width=800&height=600&path=power_point
```

**Ví dụ URL:**
- `/upload?width=1000&height=700&path=power_point`
- `/upload?width=600&height=400&path=documents` 
- `/upload?width=800&height=600&path=images`



### 2. Chọn file
- **Click vào vùng upload** để mở file picker
- **Kéo thả file** trực tiếp vào vùng upload
- Hỗ trợ chọn nhiều file cùng lúc

### 3. Upload file
- Sau khi chọn file, danh sách file sẽ hiển thị
- **Giao diện tự động điều chỉnh** theo URL parameters (width, height)
- **Không có background** để tối ưu cho việc nhúng iframe
- Click "Bắt đầu Upload" để bắt đầu upload
- Theo dõi tiến trình upload real-time
- **File được lưu vào đường dẫn** được chỉ định trong URL
- Hỗ trợ resume upload nếu bị gián đoạn

### 4. Quản lý file
- Xóa file riêng lẻ bằng nút "X" trên mỗi file
- Xóa tất cả file bằng nút "Xóa tất cả"
- Xem thống kê upload

## Cấu trúc file

```
├── templates/
│   ├── upload.html          # Template HTML chính
│   ├── upload.css           # Styles CSS
│   └── upload.js            # Logic JavaScript
├── controllers/
│   └── upload_controller.go # Controller xử lý upload
└── routes/
    ├── routes.go             # Routes cho upload
    └── static.go             # Static file serving
```

## API Endpoints

### Upload Page
```
GET /upload
```
Hiển thị giao diện upload file

### Upload Stats
```
GET /api/upload/stats
```
Lấy thống kê upload (có thể mở rộng sau)

### TUS Upload
```
POST /api/tus-uploads
```
Endpoint cho TUS protocol upload

### Static Files
```
GET /templates/upload.css
GET /templates/upload.js
```
CSS và JavaScript files từ thư mục templates

## Công nghệ sử dụng

### Frontend
- **HTML5**: Semantic markup
- **CSS3**: Flexbox, Grid, Animations, Gradients
- **JavaScript ES6+**: Classes, Arrow functions, Modern APIs
- **TUS Client**: Resumable uploads
- **Font Awesome**: Icons

### Backend
- **Go/Gin**: Web framework
- **TUS Server**: Resumable upload protocol
- **HTML Templates**: Gin template engine

## Tùy chỉnh

### Thay đổi giao diện
- Chỉnh sửa `templates/upload.html` để thay đổi layout
- Cập nhật `public/static/css/upload.css` để thay đổi styles
- Sửa `public/static/js/upload.js` để thay đổi logic

### Thay đổi endpoint
- Cập nhật `API_DOMAIN` trong file `.env` để thay đổi TUS endpoint
- Sửa routes trong `routes/routes.go` nếu muốn thay đổi URL
- JavaScript tự động sử dụng `window.API_DOMAIN` từ Go template

### Thêm tính năng
- Thêm validation file trong `upload.js`
- Thêm progress callbacks
- Tích hợp với hệ thống authentication nếu cần

## Bảo mật

- Giao diện upload không yêu cầu authentication (có thể thay đổi)
- TUS endpoint được bảo vệ bởi middleware nếu cần
- File được lưu trong thư mục `public/uploads/tus/`

## Troubleshooting

### File không upload được
- Kiểm tra TUS server có chạy không
- Xem console browser để debug
- Kiểm tra network tab để xem request/response

### CSS/JS không load
- Kiểm tra đường dẫn `/templates/` có đúng không
- Xem console browser có lỗi 404 không
- Kiểm tra `StreamTemplateFile` function trong routes/static.go

### Template không render
- Kiểm tra `r.LoadHTMLGlob("templates/*")` trong app_service.go
- Xem logs server có lỗi gì không

## Cài đặt và chạy

### 1. Cấu hình môi trường
Copy file `.env.example` thành `.env` và cập nhật các giá trị:
```bash
cp .env.example .env
```

Cập nhật `API_DOMAIN` trong file `.env`:
```bash
# Development
API_DOMAIN=http://localhost:8080

# Production  
API_DOMAIN=https://lmsx-dev-api.urp.vn
```

### 2. Build ứng dụng
```bash
go build -o be-lms .
```

### 3. Chạy ứng dụng
```bash
./be-lms
```

## Cấu hình Environment Variables

### API_DOMAIN
- **Mô tả**: Domain của API server
- **Mặc định**: `http://localhost:8080`
- **Ví dụ**: 
  - Development: `API_DOMAIN=http://localhost:8080`
  - Production: `API_DOMAIN=https://lmsx-dev-api.urp.vn`

## Phát triển

### Thêm file type mới
1. Cập nhật `getFileIcon()` function trong `upload.js`
2. Thêm CSS class mới trong `upload.css`
3. Test với file type mới

### Thêm validation
1. Sửa `addFiles()` function trong `upload.js`
2. Thêm kiểm tra file size, type, name
3. Hiển thị thông báo lỗi phù hợp

### Tích hợp với backend
1. Thêm API endpoints mới trong `upload_controller.go`
2. Cập nhật frontend để gọi API mới
3. Xử lý response và hiển thị kết quả

## Liên hệ

Nếu có vấn đề hoặc cần hỗ trợ, vui lòng tạo issue hoặc liên hệ team phát triển.
