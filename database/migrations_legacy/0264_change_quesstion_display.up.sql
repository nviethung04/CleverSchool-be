CREATE TYPE display_enum AS ENUM ('word_count', 'vertical', 'horizontal');

ALTER TABLE questions ADD COLUMN display display_enum DEFAULT 'word_count';

ALTER TABLE questions DROP COLUMN is_display_vertical;
