ALTER TABLE h5p_content_scores
ADD CONSTRAINT h5p_content_scores_content_user_unique
UNIQUE (content_id, user_id);
