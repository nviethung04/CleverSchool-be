-- Create zoom_meetings table
CREATE TABLE IF NOT EXISTS zoom_meetings (
    id BIGSERIAL PRIMARY KEY,
    zoom_meeting_id VARCHAR(255) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    course_id BIGINT NULL,
    lesson_id BIGINT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    start_time TIMESTAMP NOT NULL,
    duration INTEGER NOT NULL DEFAULT 60, -- minutes
    join_url TEXT,
    host_url TEXT,
    password VARCHAR(50),
    meeting_type INTEGER DEFAULT 2, -- 1=instant, 2=scheduled, 3=recurring
    status VARCHAR(50) DEFAULT 'scheduled', -- scheduled, started, ended, cancelled
    is_recording BOOLEAN DEFAULT false,
    waiting_room BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    
    CONSTRAINT fk_zoom_meetings_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_zoom_meetings_course_id FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE SET NULL,
    CONSTRAINT fk_zoom_meetings_lesson_id FOREIGN KEY (lesson_id) REFERENCES lessons(id) ON DELETE SET NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_zoom_meetings_user_id ON zoom_meetings(user_id);
CREATE INDEX IF NOT EXISTS idx_zoom_meetings_course_id ON zoom_meetings(course_id);
CREATE INDEX IF NOT EXISTS idx_zoom_meetings_lesson_id ON zoom_meetings(lesson_id);
CREATE INDEX IF NOT EXISTS idx_zoom_meetings_status ON zoom_meetings(status);
CREATE INDEX IF NOT EXISTS idx_zoom_meetings_start_time ON zoom_meetings(start_time);
