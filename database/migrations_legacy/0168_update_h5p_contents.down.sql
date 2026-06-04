ALTER TABLE h5p_contents
    DROP CONSTRAINT IF EXISTS h5p_contents_content_id_unique;

ALTER TABLE h5p_contents
    DROP COLUMN IF EXISTS content_id;

ALTER TABLE h5p_contents
    DROP COLUMN IF EXISTS library;
