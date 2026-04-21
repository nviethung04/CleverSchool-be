CREATE TABLE IF NOT EXISTS microsoft_meetings (
    id SERIAL PRIMARY KEY,
    online_meeting_id VARCHAR(255) NOT NULL UNIQUE,
    calendar_event_id VARCHAR(255) UNIQUE,
    user_id INTEGER NOT NULL,
    course_id INTEGER NULL,
    lesson_id INTEGER NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    join_url VARCHAR(500) NOT NULL,
    join_web_url VARCHAR(500),
    conference_id VARCHAR(100),
    toll_number VARCHAR(50),
    toll_free_number VARCHAR(50),
    time_zone VARCHAR(100) DEFAULT 'UTC',
    status VARCHAR(50) DEFAULT 'scheduled',
    is_recurring BOOLEAN DEFAULT FALSE,
    recurrence_rule TEXT,
    attendees TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE SET NULL ON UPDATE CASCADE,
    FOREIGN KEY (lesson_id) REFERENCES lessons(id) ON DELETE SET NULL ON UPDATE CASCADE
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_microsoft_meetings_user_id ON microsoft_meetings(user_id);
CREATE INDEX IF NOT EXISTS idx_microsoft_meetings_course_id ON microsoft_meetings(course_id);
CREATE INDEX IF NOT EXISTS idx_microsoft_meetings_lesson_id ON microsoft_meetings(lesson_id);
CREATE INDEX IF NOT EXISTS idx_microsoft_meetings_start_time ON microsoft_meetings(start_time);
CREATE INDEX IF NOT EXISTS idx_microsoft_meetings_status ON microsoft_meetings(status);
CREATE INDEX IF NOT EXISTS idx_microsoft_meetings_deleted_at ON microsoft_meetings(deleted_at);