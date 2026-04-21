CREATE TABLE IF NOT EXISTS meeting_attendances (
    id SERIAL PRIMARY KEY,
    meeting_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP,
    duration_minutes INTEGER DEFAULT 0,
    join_method VARCHAR(50) DEFAULT 'link',
    device_info TEXT,
    ip_address VARCHAR(45),
    is_present BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_meeting_attendances_meeting FOREIGN KEY (meeting_id) REFERENCES microsoft_meetings(id) ON DELETE CASCADE,
    CONSTRAINT fk_meeting_attendances_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_meeting_attendances_meeting_id ON meeting_attendances(meeting_id);
CREATE INDEX idx_meeting_attendances_user_id ON meeting_attendances(user_id);
CREATE INDEX idx_meeting_attendances_joined_at ON meeting_attendances(joined_at);
CREATE INDEX idx_meeting_attendances_deleted_at ON meeting_attendances(deleted_at);
