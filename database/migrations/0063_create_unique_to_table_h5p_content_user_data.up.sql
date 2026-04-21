ALTER TABLE h5p_content_user_data
ADD CONSTRAINT unique_user_data_per_content
UNIQUE (content_id, context_id, sub_content_id, user_id);
