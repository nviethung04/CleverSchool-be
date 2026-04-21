CREATE TABLE IF NOT EXISTS meeting_delete_logs (
    id SERIAL PRIMARY KEY,
    meeting_id BIGINT NOT NULL,
    deleted_by BIGINT NOT NULL,
    deleted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_meeting_delete_logs_meeting_id ON meeting_delete_logs(meeting_id);
