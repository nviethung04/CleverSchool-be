-- Persist metadata fields for writing/speaking question types (proto QuestionMetaData).

ALTER TABLE questions ADD COLUMN IF NOT EXISTS max_characters INT DEFAULT 0;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS allow_image_upload BOOLEAN DEFAULT FALSE;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS max_recording_time INT DEFAULT 0;

