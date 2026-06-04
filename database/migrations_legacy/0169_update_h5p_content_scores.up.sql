ALTER TABLE h5p_content_scores
ALTER COLUMN time DROP NOT NULL;

ALTER TABLE h5p_content_scores
    ALTER COLUMN user_id TYPE bigint USING user_id::bigint;

ALTER TABLE h5p_content_scores
    ALTER COLUMN content_id TYPE varchar(255);
