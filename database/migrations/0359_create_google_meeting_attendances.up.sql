CREATE TABLE IF NOT EXISTS google_meeting_attendances (
    id SERIAL PRIMARY KEY,
    meeting_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP NULL,
    duration_minutes INTEGER DEFAULT 0,
    join_method VARCHAR(50) DEFAULT 'link',
    device_info TEXT,
    ip_address VARCHAR(100),
    is_present BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    FOREIGN KEY (meeting_id) REFERENCES google_meetings(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_google_meeting_attendances_meeting_id ON google_meeting_attendances(meeting_id);
CREATE INDEX IF NOT EXISTS idx_google_meeting_attendances_user_id ON google_meeting_attendances(user_id);
CREATE INDEX IF NOT EXISTS idx_google_meeting_attendances_deleted_at ON google_meeting_attendances(deleted_at);
