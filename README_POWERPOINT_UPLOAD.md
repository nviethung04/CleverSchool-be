# PowerPoint Upload System

Hệ thống upload và quản lý file PowerPoint HTML được nén trong file ZIP hoặc RAR.

## Tính năng

- ✅ Upload file ZIP hoặc RAR chứa PowerPoint HTML
- ✅ Tự động giải nén và tổ chức file
- ✅ Quản lý danh sách presentations
- ✅ Xem thông tin chi tiết presentation
- ✅ Xóa presentation
- ✅ Giao diện web thân thiện với drag & drop

## Yêu cầu hệ thống

### Để hỗ trợ file RAR:

- **macOS**: `brew install unar`
- **Ubuntu/Debian**: `sudo apt-get install unrar`
- **CentOS/RHEL**: `sudo yum install unrar` hoặc `sudo dnf install unrar`
- **Windows**: Cài đặt WinRAR hoặc 7-Zip

## Cấu trúc file nén

File ZIP hoặc RAR phải có cấu trúc như sau:

```
presentation.zip (hoặc .rar)
├── index.html          # File HTML chính
└── data/               # Thư mục chứa assets
    ├── thmb1.png
    ├── thmb2.png
    ├── slide1.css
    ├── slide1.js
    ├── video1.mp4
    └── ...
```

## API Endpoints

### 1. Upload PowerPoint

```
POST /api/upload-file
Content-Type: multipart/form-data

Parameters:
- file: File ZIP hoặc RAR (required)
- path: power_point
```

**Response:**

### 1. Xem Presentation

```
GET /power-point/{folder_name}/index.html
```

### 2. Truy cập Assets

```
GET /power-point/{folder_name}/data/{filepath}
```

## Cách sử dụng

### 1. Chuẩn bị file nén

- Tạo file ZIP hoặc RAR chứa PowerPoint HTML và thư mục `data`
- Đảm bảo có file `index.html` ở root của file nén

### 2. Upload qua API

```bash
# Upload file ZIP
curl -X POST http://localhost:8080/power-point/upload \
  -F "file=@presentation.zip" \
  -F "folder_name=my-presentation"

# Upload file RAR
curl -X POST http://localhost:8080/power-point/upload \
  -F "file=@presentation.rar" \
  -F "folder_name=my-presentation"
```

### 3. Xem presentation

```bash
curl http://localhost:8080/power-point/{folder_name}/index.html
```

## Validation

### File validation:

- Chỉ chấp nhận file `.zip` và `.rar`
- Kiểm tra zip slip vulnerability
- Tự động tạo tên thư mục nếu không được cung cấp

### Folder name validation:

- Chỉ cho phép chữ cái, số, dấu gạch ngang và gạch dưới
- Độ dài tối đa 50 ký tự
- Tự động thêm timestamp để đảm bảo unique

## Bảo mật

- Kiểm tra zip slip vulnerability
- Validate file type
- Sanitize folder names
- Kiểm tra quyền truy cập file

## Lưu ý

1. File sẽ được lưu trong thư mục `power_point/` tại root của project
2. Mỗi presentation sẽ có thư mục riêng với tên unique
3. Assets sẽ được serve qua API endpoint `/power-point/{folder}/data/`
4. File HTML sẽ được modify để thay thế đường dẫn assets thành API paths
5. Để hỗ trợ RAR, cần cài đặt unrar command line tool

## Troubleshooting

### Lỗi "File not found"

- Kiểm tra file nén có đúng cấu trúc không
- Đảm bảo có file `index.html` ở root

### Lỗi "Invalid file type"

- Chỉ chấp nhận file `.zip` và `.rar`
- Kiểm tra extension file

### Lỗi "Failed to extract RAR file"

- Cài đặt unrar command line tool
- Kiểm tra file RAR có bị corrupt không

### Lỗi "Folder already exists"

- Chọn tên thư mục khác
- Hoặc để trống để tự động tạo tên unique

### Lỗi "Failed to extract"

- Kiểm tra file nén có bị corrupt không
- Đảm bảo có đủ quyền ghi vào thư mục
- Đối với RAR: kiểm tra unrar đã được cài đặt

## Cài đặt unrar

### macOS:

```bash
brew install unar
```

### Ubuntu/Debian:

```bash
sudo apt-get update
sudo apt-get install unrar
```

### CentOS/RHEL:

```bash
sudo yum install unrar
# hoặc
sudo dnf install unrar
```

### Windows:

- Tải và cài đặt WinRAR từ https://www.win-rar.com/
- Hoặc cài đặt 7-Zip từ https://7-zip.org/
- Đảm bảo thêm vào PATH system
