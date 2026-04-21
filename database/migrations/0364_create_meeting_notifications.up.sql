CREATE TABLE IF NOT EXISTS meeting_notifications (
    id BIGSERIAL PRIMARY KEY,
    meeting_id BIGINT REFERENCES microsoft_meetings(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    course_id BIGINT,
    lesson_id BIGINT,
    title TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    read_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_meeting_notifications_user_read ON meeting_notifications (user_id, read_at);
CREATE INDEX IF NOT EXISTS idx_meeting_notifications_meeting ON meeting_notifications (meeting_id);
CREATE INDEX IF NOT EXISTS idx_meeting_notifications_course ON meeting_notifications (course_id);
