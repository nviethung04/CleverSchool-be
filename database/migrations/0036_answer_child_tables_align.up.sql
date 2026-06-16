-- Align answer_positions / answer_matchings / answer_coordinates with Go models (GORM).

-- answer_positions (fill_in_blanks, ordering, drag_drop)
ALTER TABLE answer_positions ADD COLUMN IF NOT EXISTS kind kind_enum;
ALTER TABLE answer_positions ADD COLUMN IF NOT EXISTS point NUMERIC(10,2) DEFAULT 0;
ALTER TABLE answer_positions ADD COLUMN IF NOT EXISTS is_correct BOOLEAN DEFAULT FALSE;
ALTER TABLE answer_positions ADD COLUMN IF NOT EXISTS correct_position INT DEFAULT 0;

UPDATE answer_positions SET kind = 'text' WHERE kind IS NULL;
ALTER TABLE answer_positions ALTER COLUMN kind SET DEFAULT 'text';
ALTER TABLE answer_positions ALTER COLUMN kind SET NOT NULL;

-- answer_matchings (matching)
ALTER TABLE answer_matchings ADD COLUMN IF NOT EXISTS content TEXT;
ALTER TABLE answer_matchings ADD COLUMN IF NOT EXISTS matching_content TEXT;
ALTER TABLE answer_matchings ADD COLUMN IF NOT EXISTS file_info JSONB;
ALTER TABLE answer_matchings ADD COLUMN IF NOT EXISTS matching_file_info JSONB;
ALTER TABLE answer_matchings ADD COLUMN IF NOT EXISTS kind kind_enum;
ALTER TABLE answer_matchings ADD COLUMN IF NOT EXISTS matching_kind kind_enum;
ALTER TABLE answer_matchings ADD COLUMN IF NOT EXISTS point NUMERIC(10,2) DEFAULT 0;
ALTER TABLE answer_matchings ADD COLUMN IF NOT EXISTS is_correct BOOLEAN DEFAULT FALSE;
ALTER TABLE answer_matchings ADD COLUMN IF NOT EXISTS correct_position INT DEFAULT 0;

UPDATE answer_matchings
SET
    content = COALESCE(content, first_item_content),
    matching_content = COALESCE(matching_content, second_item_content),
    file_info = COALESCE(file_info, first_item_file_info),
    matching_file_info = COALESCE(matching_file_info, second_item_file_info)
WHERE first_item_content IS NOT NULL
   OR second_item_content IS NOT NULL
   OR first_item_file_info IS NOT NULL
   OR second_item_file_info IS NOT NULL;

UPDATE answer_matchings SET kind = 'text' WHERE kind IS NULL;
UPDATE answer_matchings SET matching_kind = 'text' WHERE matching_kind IS NULL;

ALTER TABLE answer_matchings ALTER COLUMN kind SET DEFAULT 'text';
ALTER TABLE answer_matchings ALTER COLUMN matching_kind SET DEFAULT 'text';
ALTER TABLE answer_matchings ALTER COLUMN kind SET NOT NULL;
ALTER TABLE answer_matchings ALTER COLUMN matching_kind SET NOT NULL;

ALTER TABLE answer_matchings DROP COLUMN IF EXISTS first_item_content;
ALTER TABLE answer_matchings DROP COLUMN IF EXISTS second_item_content;
ALTER TABLE answer_matchings DROP COLUMN IF EXISTS first_item_file_info;
ALTER TABLE answer_matchings DROP COLUMN IF EXISTS second_item_file_info;

-- answer_coordinates (labeling)
ALTER TABLE answer_coordinates ADD COLUMN IF NOT EXISTS kind kind_enum;
ALTER TABLE answer_coordinates ADD COLUMN IF NOT EXISTS point NUMERIC(10,2) DEFAULT 0;
ALTER TABLE answer_coordinates ADD COLUMN IF NOT EXISTS file_info JSONB;
ALTER TABLE answer_coordinates ADD COLUMN IF NOT EXISTS sort_position INT DEFAULT 0;
ALTER TABLE answer_coordinates ADD COLUMN IF NOT EXISTS position_x INT;
ALTER TABLE answer_coordinates ADD COLUMN IF NOT EXISTS position_y INT;
ALTER TABLE answer_coordinates ADD COLUMN IF NOT EXISTS position_width INT;
ALTER TABLE answer_coordinates ADD COLUMN IF NOT EXISTS position_height INT;

UPDATE answer_coordinates
SET
    position_x = COALESCE(position_x, x),
    position_y = COALESCE(position_y, y),
    position_width = COALESCE(position_width, width),
    position_height = COALESCE(position_height, height),
    kind = COALESCE(kind, 'text'::kind_enum)
WHERE x IS NOT NULL OR y IS NOT NULL OR width IS NOT NULL OR height IS NOT NULL OR kind IS NULL;

UPDATE answer_coordinates SET kind = 'text' WHERE kind IS NULL;
ALTER TABLE answer_coordinates ALTER COLUMN kind SET DEFAULT 'text';
ALTER TABLE answer_coordinates ALTER COLUMN kind SET NOT NULL;
