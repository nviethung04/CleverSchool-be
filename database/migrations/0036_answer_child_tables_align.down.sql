-- Best-effort rollback (data in new columns may be lost).

ALTER TABLE answer_coordinates DROP COLUMN IF EXISTS position_height;
ALTER TABLE answer_coordinates DROP COLUMN IF EXISTS position_width;
ALTER TABLE answer_coordinates DROP COLUMN IF EXISTS position_y;
ALTER TABLE answer_coordinates DROP COLUMN IF EXISTS position_x;
ALTER TABLE answer_coordinates DROP COLUMN IF EXISTS sort_position;
ALTER TABLE answer_coordinates DROP COLUMN IF EXISTS file_info;
ALTER TABLE answer_coordinates DROP COLUMN IF EXISTS point;
ALTER TABLE answer_coordinates DROP COLUMN IF EXISTS kind;

ALTER TABLE answer_matchings ADD COLUMN IF NOT EXISTS first_item_content TEXT;
ALTER TABLE answer_matchings ADD COLUMN IF NOT EXISTS second_item_content TEXT;
ALTER TABLE answer_matchings ADD COLUMN IF NOT EXISTS first_item_file_info JSONB;
ALTER TABLE answer_matchings ADD COLUMN IF NOT EXISTS second_item_file_info JSONB;

UPDATE answer_matchings
SET
    first_item_content = content,
    second_item_content = matching_content,
    first_item_file_info = file_info,
    second_item_file_info = matching_file_info
WHERE content IS NOT NULL OR matching_content IS NOT NULL;

ALTER TABLE answer_matchings DROP COLUMN IF EXISTS correct_position;
ALTER TABLE answer_matchings DROP COLUMN IF EXISTS is_correct;
ALTER TABLE answer_matchings DROP COLUMN IF EXISTS point;
ALTER TABLE answer_matchings DROP COLUMN IF EXISTS matching_kind;
ALTER TABLE answer_matchings DROP COLUMN IF EXISTS kind;
ALTER TABLE answer_matchings DROP COLUMN IF EXISTS matching_file_info;
ALTER TABLE answer_matchings DROP COLUMN IF EXISTS file_info;
ALTER TABLE answer_matchings DROP COLUMN IF EXISTS matching_content;
ALTER TABLE answer_matchings DROP COLUMN IF EXISTS content;

ALTER TABLE answer_positions DROP COLUMN IF EXISTS correct_position;
ALTER TABLE answer_positions DROP COLUMN IF EXISTS is_correct;
ALTER TABLE answer_positions DROP COLUMN IF EXISTS point;
ALTER TABLE answer_positions DROP COLUMN IF EXISTS kind;
