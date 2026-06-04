ALTER TABLE questions ADD COLUMN is_display_vertical BOOLEAN DEFAULT FALSE;

ALTER TABLE questions DROP COLUMN display;

DROP TYPE display_enum;
