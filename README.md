# 📦 Hướng dẫn Cài đặt Hệ thống LMS
cd đến docker-project
Chạy lần lượt:

cp .env-example .env (có thể sửa đổi các thông số db trong .env)
docker compose up -d
docker compose exec postgres-master bash

cp /etc/postgresql/pg_hba.conf /var/lib/postgresql/data/pg_hba.conf

Ctrl D
docker compose restart
docker compose up -d

file .env trong be-lms cấu hình các thông số giống với dile .env trong docker

## 📥 Vào thư mục chứa migrates
    cd database/migrations

## 📥 Chạy Migration
    migrate -path . -database "postgres://admin:admin123@localhost:5433/lms?sslmode=disable" up
---
Vào data base, chạy lệnh
CREATE EXTENSION IF NOT EXISTS unaccent; (cho phép tìm kiếm bỏ qua dấu)

# Hướng dẫn deploy lên sever
cd vào be-lms
Với hdh Window: chạy deploy.bat
Với hdh ios/liux: chạy ./deploy-mac.sh
Lưu ý phải đặt Key là địa chỉ đến file ssh-key trên máy cá nhân

Trên sever cần có 1 file deploy.sh nằm cùng cấp với file thực thi, có nội dung:

###################################################################################
#!/bin/bash

# --- Đặt biến cấu hình ---
APP_NAME="myapp"
APP_DIR="/var/www/source-code-lms/be-lms"
TEMP_DIR="/home/enspire/backend-lms"
LOG_FILE="app.log"

echo "🚀 Bắt đầu deploy $APP_NAME..."

# Bước 0: Xóa file cũ và thư mục cũ
echo "🧹 Xóa file và thư mục cũ nếu có..."
sudo -n rm -f "$APP_DIR/$APP_NAME" && echo "✔️ Đã xóa $APP_NAME cũ" || echo "ℹ️ Không có file $APP_NAME để xóa"
sudo -n rm -rf "$APP_DIR/templates" && echo "✔️ Đã xóa thư mục templates cũ" || echo "ℹ️ Không có thư mục templates để xóa"
sudo -n rm -f "$APP_DIR/package.json" && echo "✔️ Đã xóa package.json cũ" || echo "ℹ️ Không có package.json để xóa"
sudo -n rm -f "$APP_DIR/package-lock.json" && echo "✔️ Đã xóa package-lock.json cũ" || echo "ℹ️ Không có package-lock.json để xóa"

# Bước 1: Copy file và thư mục mới từ TEMP_DIR vào APP_DIR
echo "📂 Di chuyển file mới từ $TEMP_DIR sang $APP_DIR..."
sudo -n mv "$TEMP_DIR/$APP_NAME" "$APP_DIR/" && echo "✔️ Đã di chuyển $APP_NAME mới" || echo "❌ Lỗi di chuyển $APP_NAME"
sudo -n mv "$TEMP_DIR/templates" "$APP_DIR/" && echo "✔️ Đã di chuyển templates mới" || echo "❌ Lỗi di chuyển templates"
sudo -n mv "$TEMP_DIR/package.json" "$APP_DIR/" && echo "✔️ Đã di chuyển package.json mới" || echo "ℹ️ Không có package.json mới"
sudo -n mv "$TEMP_DIR/package-lock.json" "$APP_DIR/" && echo "✔️ Đã di chuyển package-lock.json mới" || echo "ℹ️ Không có package-lock.json mới"

# Bước 2: Vào thư mục deploy
cd "$APP_DIR" || { echo "❌ Không tìm thấy thư mục $APP_DIR"; exit 1; }

# Bước 3: Stop app cũ trước
echo "🛑 Dừng app cũ nếu đang chạy..."
pid=$(ps aux | grep "./$APP_NAME" | grep -v grep | awk '{print $2}')
if [ -n "$pid" ]; then
kill "$pid"
echo "✅ Đã kill tiến trình cũ: $pid"
sleep 2
else
echo "ℹ️  Không có tiến trình cũ đang chạy."
fi


# Bước 4: Cấp quyền thực thi cho file mới
echo "🔒 Cấp quyền thực thi cho $APP_NAME..."
chmod +x "./$APP_NAME"

# Bước 5: Cài đặt node modules nếu có package.json
if [ -f package.json ]; then
echo "📦 Cài đặt Node.js dependencies (production only)..."
npm install --omit=dev || echo "❌ npm install thất bại"
else
echo "⚠️ package.json không tồn tại, bỏ qua bước npm install."
fi
# Bước 6: Chạy migration
echo "🚀 Chạy database migration..."
./$APP_NAME migrate
MIGRATE_STATUS=$?

if [ $MIGRATE_STATUS -ne 0 ]; then
echo "⚠️ Migration thất bại. Kiểm tra có phải database đang ở trạng thái 'dirty' không..."

    # Force version nếu dirty
    ./$APP_NAME migrate version | grep "dirty=true" > /dev/null
    if [ $? -eq 0 ]; then
        echo "🧹 Database đang ở trạng thái dirty. Thực hiện force fix..."

        CURRENT_VERSION=$(./$APP_NAME migrate:version | grep -oE '[0-9]+' | head -1)
        ./$APP_NAME migrate:force "$CURRENT_VERSION"
        
        echo "🔁 Thử chạy lại migration..."
        ./$APP_NAME migrate

        if [ $? -ne 0 ]; then
            echo "❌ Migration vẫn thất bại sau khi force fix. Dừng deploy."
            exit 1
        else
            echo "✅ Migration thành công sau khi force."
        fi
    else
        echo "❌ Migration thất bại vì lý do khác. Dừng deploy."
        exit 1
    fi
else
echo "✅ Migration thành công."
fi

# Bước 7: Khởi động lại app
echo "▶️ Khởi động app mới..."
nohup "./$APP_NAME" > "$LOG_FILE" 2>&1 &

echo "✅ Deploy hoàn tất!"

###################################################################################