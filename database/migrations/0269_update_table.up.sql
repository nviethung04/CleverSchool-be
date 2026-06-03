CREATE TYPE question_form_enum AS ENUM ('write', 'question');

ALTER TABLE homeworks ADD COLUMN question_form question_form_enum DEFAULT 'question';
ALTER TABLE exams ADD COLUMN question_form question_form_enum DEFAULT 'question';
ALTER TABLE exercises ADD COLUMN question_form question_form_enum DEFAULT 'question';
ALTER TABLE contest_rounds ADD COLUMN question_form question_form_enum DEFAULT 'question';

ALTER TABLE homeworks ADD COLUMN file_infos jsonb;
ALTER TABLE exams ADD COLUMN file_infos jsonb;
ALTER TABLE exercises ADD COLUMN file_infos jsonb;
ALTER TABLE contest_rounds ADD COLUMN file_infos jsonb;

ALTER TABLE homework_users ADD COLUMN file_infos jsonb;
ALTER TABLE exam_users ADD COLUMN file_infos jsonb;
ALTER TABLE exercise_users ADD COLUMN file_infos jsonb;
ALTER TABLE contest_round_users ADD COLUMN file_infos jsonb;
