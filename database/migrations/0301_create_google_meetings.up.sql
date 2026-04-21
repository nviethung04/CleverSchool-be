CREATE TABLE IF NOT EXISTS google_meetings (
    id SERIAL PRIMARY KEY,
    calendar_event_id VARCHAR(255) NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    course_id INTEGER NULL,
    lesson_id INTEGER NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    meet_url VARCHAR(500) NOT NULL,
    calendar_url VARCHAR(500),
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
CREATE INDEX IF NOT EXISTS idx_google_meetings_user_id ON google_meetings(user_id);
CREATE INDEX IF NOT EXISTS idx_google_meetings_course_id ON google_meetings(course_id);
CREATE INDEX IF NOT EXISTS idx_google_meetings_lesson_id ON google_meetings(lesson_id);
CREATE INDEX IF NOT EXISTS idx_google_meetings_start_time ON google_meetings(start_time);
CREATE INDEX IF NOT EXISTS idx_google_meetings_status ON google_meetings(status);
CREATE INDEX IF NOT EXISTS idx_google_meetings_deleted_at ON google_meetings(deleted_at);