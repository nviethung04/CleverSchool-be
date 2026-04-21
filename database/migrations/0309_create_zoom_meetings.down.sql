-- Drop zoom_meetings table
DROP INDEX IF EXISTS idx_zoom_meetings_zoom_id;
DROP INDEX IF EXISTS idx_zoom_meetings_deleted_at;
DROP INDEX IF EXISTS idx_zoom_meetings_start_time;
DROP INDEX IF EXISTS idx_zoom_meetings_status;
DROP INDEX IF EXISTS idx_zoom_meetings_lesson_id;
DROP INDEX IF EXISTS idx_zoom_meetings_course_id;
DROP INDEX IF EXISTS idx_zoom_meetings_user_id;

DROP TABLE IF EXISTS zoom_meetings;