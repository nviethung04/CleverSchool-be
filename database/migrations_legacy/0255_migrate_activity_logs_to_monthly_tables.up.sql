-- Migration: Migrate activity_logs to monthly tables (V3 - Simple and Safe)
-- This migration creates monthly tables and migrates existing data with simple JSONB conversion

-- 1. Tạo bảng cho tháng hiện tại (nếu chưa có)
DO $$
DECLARE
    current_table_name TEXT;
    current_month INTEGER;
    current_year INTEGER;
BEGIN
    current_month := EXTRACT(MONTH FROM NOW());
    current_year := EXTRACT(YEAR FROM NOW());
    current_table_name := 'activity_logs_' || LPAD(current_month::TEXT, 2, '0') || '_' || current_year;
    
    -- Kiểm tra xem bảng đã tồn tại chưa
    IF NOT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = current_table_name
    ) THEN
        -- Tạo bảng mới
        EXECUTE format('
            CREATE TABLE %I (
                id SERIAL PRIMARY KEY,
                user_id BIGINT,
                role_id INTEGER,
                method VARCHAR(10) NOT NULL,
                path TEXT NOT NULL,
                status_code INTEGER NOT NULL,
                client_ip VARCHAR(45),
                device VARCHAR(20),
                latency_ms INTEGER,
                query_params JSONB,
                request_body JSONB,
                response_body JSONB,
                attribute JSONB,
                session_id TEXT,
                agent TEXT,
                created_at TIMESTAMP DEFAULT NOW()
            )
        ', current_table_name);
        
        -- Tạo indexes
        EXECUTE format('
            CREATE INDEX IF NOT EXISTS idx_%I_created_at ON %I(created_at DESC)
        ', current_table_name, current_table_name);
        
        EXECUTE format('
            CREATE INDEX IF NOT EXISTS idx_%I_user_id ON %I(user_id)
        ', current_table_name, current_table_name);
        
        EXECUTE format('
            CREATE INDEX IF NOT EXISTS idx_%I_user_created_at ON %I(user_id, created_at DESC)
        ', current_table_name, current_table_name);
        
        EXECUTE format('
            CREATE INDEX IF NOT EXISTS idx_%I_device ON %I(device)
        ', current_table_name, current_table_name);
        
        EXECUTE format('
            CREATE INDEX IF NOT EXISTS idx_%I_device_created_at ON %I(device, created_at DESC)
        ', current_table_name, current_table_name);
        
        RAISE NOTICE 'Created table: %', current_table_name;
    ELSE
        RAISE NOTICE 'Table already exists: %', current_table_name;
    END IF;
END $$;

-- 2. Chuyển dữ liệu từ bảng cũ sang bảng theo tháng (chỉ nếu bảng cũ tồn tại)
DO $$
DECLARE
    rec RECORD;
    target_table_name TEXT;
    month_val INTEGER;
    year_val INTEGER;
    insert_sql TEXT;
    total_moved INTEGER := 0;
    old_table_exists BOOLEAN;
BEGIN
    -- Kiểm tra xem bảng cũ có tồn tại không
    SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = 'activity_logs'
    ) INTO old_table_exists;
    
    IF NOT old_table_exists THEN
        RAISE NOTICE 'Old activity_logs table does not exist, skipping data migration';
        RETURN;
    END IF;
    
    -- Lấy tất cả dữ liệu từ bảng cũ và nhóm theo tháng
    FOR rec IN 
        SELECT 
            EXTRACT(MONTH FROM created_at) as month,
            EXTRACT(YEAR FROM created_at) as year,
            COUNT(*) as count
        FROM activity_logs 
        GROUP BY EXTRACT(MONTH FROM created_at), EXTRACT(YEAR FROM created_at)
        ORDER BY year, month
    LOOP
        month_val := rec.month;
        year_val := rec.year;
        target_table_name := 'activity_logs_' || LPAD(month_val::TEXT, 2, '0') || '_' || year_val;
        
        -- Tạo bảng cho tháng này nếu chưa có
        IF NOT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = target_table_name
        ) THEN
            EXECUTE format('
                CREATE TABLE %I (
                    id SERIAL PRIMARY KEY,
                    user_id BIGINT,
                    role_id INTEGER,
                    method VARCHAR(10) NOT NULL,
                    path TEXT NOT NULL,
                    status_code INTEGER NOT NULL,
                    client_ip VARCHAR(45),
                    device VARCHAR(20),
                    latency_ms INTEGER,
                    query_params JSONB,
                    request_body JSONB,
                    response_body JSONB,
                    attribute JSONB,
                    session_id TEXT,
                    agent TEXT,
                    created_at TIMESTAMP DEFAULT NOW()
                )
            ', target_table_name);
            
            -- Tạo indexes
            EXECUTE format('
                CREATE INDEX IF NOT EXISTS idx_%I_created_at ON %I(created_at DESC)
            ', target_table_name, target_table_name);
            
            EXECUTE format('
                CREATE INDEX IF NOT EXISTS idx_%I_user_id ON %I(user_id)
            ', target_table_name, target_table_name);
            
            EXECUTE format('
                CREATE INDEX IF NOT EXISTS idx_%I_user_created_at ON %I(user_id, created_at DESC)
            ', target_table_name, target_table_name);
            
            EXECUTE format('
                CREATE INDEX IF NOT EXISTS idx_%I_device ON %I(device)
            ', target_table_name, target_table_name);
            
            EXECUTE format('
                CREATE INDEX IF NOT EXISTS idx_%I_device_created_at ON %I(device, created_at DESC)
            ', target_table_name, target_table_name);
            
            RAISE NOTICE 'Created table: %', target_table_name;
        END IF;
        
        -- Chuyển dữ liệu với simple JSONB conversion (chỉ convert nếu là JSON hợp lệ)
        insert_sql := format('
            INSERT INTO %I (
                user_id, role_id, method, path, status_code, client_ip, device, 
                latency_ms, query_params, request_body, response_body, attribute, 
                session_id, agent, created_at
            )
            SELECT 
                user_id, role_id, method, path, status_code, client_ip, device,
                latency_ms, 
                -- Simple JSONB conversion - only convert if valid JSON
                CASE 
                    WHEN query_params IS NULL OR query_params = '''' THEN NULL
                    WHEN query_params ~ ''^[\s]*[\[\{]'' THEN 
                        CASE 
                            WHEN query_params::jsonb IS NOT NULL THEN query_params::jsonb
                            ELSE NULL
                        END
                    ELSE NULL
                END as query_params,
                CASE 
                    WHEN request_body IS NULL OR request_body = '''' THEN NULL
                    WHEN request_body ~ ''^[\s]*[\[\{]'' THEN 
                        CASE 
                            WHEN request_body::jsonb IS NOT NULL THEN request_body::jsonb
                            ELSE NULL
                        END
                    ELSE NULL
                END as request_body,
                CASE 
                    WHEN response_body IS NULL OR response_body = '''' THEN NULL
                    WHEN response_body ~ ''^[\s]*[\[\{]'' THEN 
                        CASE 
                            WHEN response_body::jsonb IS NOT NULL THEN response_body::jsonb
                            ELSE NULL
                        END
                    ELSE NULL
                END as response_body,
                CASE 
                    WHEN attribute IS NULL OR attribute = '''' THEN NULL
                    WHEN attribute ~ ''^[\s]*[\[\{]'' THEN 
                        CASE 
                            WHEN attribute::jsonb IS NOT NULL THEN attribute::jsonb
                            ELSE NULL
                        END
                    ELSE NULL
                END as attribute,
                session_id, agent, created_at
            FROM activity_logs 
            WHERE EXTRACT(MONTH FROM created_at) = %s 
            AND EXTRACT(YEAR FROM created_at) = %s
        ', target_table_name, month_val, year_val);
        
        -- Sử dụng exception handling để bỏ qua lỗi JSON
        BEGIN
            EXECUTE insert_sql;
            GET DIAGNOSTICS total_moved = ROW_COUNT;
            RAISE NOTICE 'Moved % records to table: %', total_moved, target_table_name;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Error migrating data to table %: %', target_table_name, SQLERRM;
            -- Continue with next table
        END;
    END LOOP;
    
    RAISE NOTICE 'Data migration completed!';
END $$;

-- 3. Tạo function để tự động tạo bảng cho tháng mới
CREATE OR REPLACE FUNCTION create_monthly_activity_logs_table(target_date DATE DEFAULT NULL)
RETURNS TEXT AS $$
DECLARE
    table_name TEXT;
    month_val INTEGER;
    year_val INTEGER;
    target_date_val DATE;
BEGIN
    -- Sử dụng ngày hiện tại nếu không có tham số
    IF target_date IS NULL THEN
        target_date_val := CURRENT_DATE;
    ELSE
        target_date_val := target_date;
    END IF;
    
    month_val := EXTRACT(MONTH FROM target_date_val);
    year_val := EXTRACT(YEAR FROM target_date_val);
    table_name := 'activity_logs_' || LPAD(month_val::TEXT, 2, '0') || '_' || year_val;
    
    -- Kiểm tra xem bảng đã tồn tại chưa
    IF EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = table_name
    ) THEN
        RETURN 'Table ' || table_name || ' already exists';
    END IF;
    
    -- Tạo bảng
    EXECUTE format('
        CREATE TABLE %I (
            id SERIAL PRIMARY KEY,
            user_id BIGINT,
            role_id INTEGER,
            method VARCHAR(10) NOT NULL,
            path TEXT NOT NULL,
            status_code INTEGER NOT NULL,
            client_ip VARCHAR(45),
            device VARCHAR(20),
            latency_ms INTEGER,
            query_params JSONB,
            request_body JSONB,
            response_body JSONB,
            attribute JSONB,
            session_id TEXT,
            agent TEXT,
            created_at TIMESTAMP DEFAULT NOW()
        )
    ', table_name);
    
    -- Tạo indexes
    EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_%I_created_at ON %I(created_at DESC)
    ', table_name, table_name);
    
    EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_%I_user_id ON %I(user_id)
    ', table_name, table_name);
    
    EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_%I_user_created_at ON %I(user_id, created_at DESC)
    ', table_name, table_name);
    
    EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_%I_device ON %I(device)
    ', table_name, table_name);
    
    EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_%I_device_created_at ON %I(device, created_at DESC)
    ', table_name, table_name);
    
    RETURN 'Created table: ' || table_name;
END;
$$ LANGUAGE plpgsql;

-- 4. Tạo function để xóa bảng cũ (chạy sau khi đã xác nhận migration thành công)
CREATE OR REPLACE FUNCTION cleanup_old_activity_logs_table()
RETURNS TEXT AS $$
DECLARE
    table_exists BOOLEAN;
BEGIN
    -- Kiểm tra xem bảng cũ có tồn tại không
    SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = 'activity_logs'
    ) INTO table_exists;
    
    IF NOT table_exists THEN
        RETURN 'Table activity_logs does not exist';
    END IF;
    
    -- Xóa bảng cũ
    DROP TABLE activity_logs;
    
    RETURN 'Successfully dropped old activity_logs table';
END;
$$ LANGUAGE plpgsql;
