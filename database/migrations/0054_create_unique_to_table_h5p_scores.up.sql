CREATE UNIQUE INDEX idx_h5p_scores_conflict 
ON h5p_scores (user_id, lesson_id, sub_id, parent_sub_id);
