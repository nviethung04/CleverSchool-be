ALTER TABLE h5p_content_scores
ADD CONSTRAINT unique_content_user
UNIQUE (content_id, user_id);
