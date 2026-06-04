-- +migrate Up
DROP SEQUENCE IF EXISTS h5p_content_user_data_id_seq;
DROP SEQUENCE IF EXISTS h5p_content_scores_id_seq;

DROP TABLE IF EXISTS public.h5p_content_user_data;
DROP TABLE IF EXISTS public.h5p_content_scores;
DROP TABLE IF EXISTS public.h5p_contents;

CREATE UNIQUE INDEX IF NOT EXISTS idx_h5p_scores_conflict 
ON h5p_scores (user_id, lesson_id, sub_id, parent_sub_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_h5p_scores_conflict;

DROP TABLE IF EXISTS public.h5p_content_user_data;
DROP TABLE IF EXISTS public.h5p_content_scores;
DROP TABLE IF EXISTS public.h5p_contents;

DROP SEQUENCE IF EXISTS h5p_content_user_data_id_seq;
DROP SEQUENCE IF EXISTS h5p_content_scores_id_seq;
