# PowerPoint Presentation Feature

Tính năng hiển thị PowerPoint presentations được lưu dưới dạng HTML files trong hệ thống LMS.

## Cấu trúc thư mục

```
templates/power_point/
├── lesson-sample/
│   ├── index.html
│   └── data/
└── test/
    ├── index.html
    └── data/
```

## API Endpoints

### 1. Hiển thị PowerPoint presentation

```
GET /power-point/:folder/index.html
```

**Parameters:**

- `folder`: Tên thư mục chứa presentation (ví dụ: `test`, `lesson-sample`)

**Response:**

- HTML content của presentation với các đường dẫn assets đã được thay thế

**Ví dụ:**

```
GET /power-point/test/index.html
```

### 2. Serve PowerPoint data

```
GET /power-point/:folder/data/*filepath
```

**Parameters:**

- `folder`: Tên thư mục chứa presentation
- `filepath`: Đường dẫn tới file asset trong thư mục `data/`

**Response:**

- File content với content-type phù hợp

**Ví dụ:**

```
GET /power-point/test/data/slide1.css
GET /power-point/test/data/thmb1.png
GET /power-point/test/data/video1.mp4
```

## Cách tạo PowerPoint folder mới

1. Tạo thư mục mới trong `power_point/power_point/`:

   ```bash
   mkdir power_point/power_point/my-presentation
   mkdir power_point/power_point/my-presentation/data
   ```

2. Copy file `index.html` vào thư mục chính:

   ```bash
   cp your-presentation.html power_point/power_point/my-presentation/index.html
   ```

3. Copy tất cả assets (CSS, JS, images, videos) vào thư mục `data/`:

   ```bash
   cp -r your-assets/* power_point/power_point/my-presentation/data/
   ```

4. Truy cập presentation qua API:
   ```
   GET /power-point/my-presentation/index.html
   ```

## Cách hoạt động

1. **Controller** (`controllers/power_point_controller.go`):

   - Đọc file `index.html` từ thư mục `power_point/power_point/{folder}/`
   - Thay thế các đường dẫn `data/` thành `/power-point/{folder}/data/`
   - Trả về HTML content đã được xử lý

2. **Asset Serving**:

   - Khi browser request assets, controller sẽ đọc file từ `power_point/power_point/{folder}/data/{filepath}`
   - Tự động detect content-type dựa trên file extension
   - Trả về file content với header phù hợp

3. **Service** (`services/power_point_service.go`):
   - Kiểm tra file tồn tại
   - Detect content-type cho các loại file khác nhau

## Content Types được hỗ trợ

- **CSS**: `text/css`
- **JavaScript**: `application/javascript`
- **Images**: `image/png`, `image/jpg`, `image/jpeg`, `image/gif`, `image/webp`
- **Videos**: `video/mp4`, `video/webm`, `video/ogg`
- **Fonts**: `font/woff`, `font/woff2`, `font/ttf`
- **Others**: `application/octet-stream` (default)

## Lưu ý

- Tất cả data phải được đặt trong thư mục `data/` của mỗi presentation
- File `index.html` phải sử dụng đường dẫn tương đối `data/` để reference data
- Hệ thống sẽ tự động thay thế đường dẫn khi serve HTML
- Không cần khai báo static routes - tất cả đều được serve qua API

## Testing

1. **Test presentation chính**:

   ```bash
   curl http://localhost:8080/power-point/test/index.html
   ```

2. **Test data**:

   ```bash
   curl http://localhost:8080/power-point/test/data/slide1.css
   curl http://localhost:8080/power-point/test/data/thmb1.png
   ```

3. **Test trong browser**:
   - Mở `http://localhost:8080/power-point/test/index`
   - Kiểm tra xem tất cả data có load được không
   - Kiểm tra console để đảm bảo không có lỗi 404

## Cấu hình Nginx reverse proxy

1. **Cấu hình Nginx reverse proxy**:

   ```
   sudo nano /etc/nginx/sites-available/lmsx.urp.vn
   ```

2. **Thêm block sau**:

   ```
   location /power-point/ {
      proxy_pass http://127.0.0.1:8080/power-point/;
      proxy_http_version 1.1;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
   }
   ```

3. **Bật HTTPS (Let's Encrypt - certbot)**:

   ```
   sudo apt install certbot python3-certbot-nginx -y
   sudo certbot --nginx -d lmsx.urp.vn
   ```

4. **Kiểm tra và restart Nginx**:
   ```
   sudo nginx -t
   sudo systemctl reload nginx
   ```
