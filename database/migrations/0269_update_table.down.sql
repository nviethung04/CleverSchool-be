ALTER TABLE IF EXISTS homework_users DROP COLUMN file_infos;
ALTER TABLE IF EXISTS exam_users DROP COLUMN file_infos;
ALTER TABLE IF EXISTS exercise_users DROP COLUMN file_infos;
ALTER TABLE IF EXISTS contest_round_users DROP COLUMN file_infos;

ALTER TABLE homeworks DROP COLUMN file_infos;
ALTER TABLE exams DROP COLUMN file_infos;
ALTER TABLE exercises DROP COLUMN file_infos;
ALTER TABLE contest_rounds DROP COLUMN file_infos;

ALTER TABLE homeworks DROP COLUMN question_form;
ALTER TABLE exams DROP COLUMN question_form;
ALTER TABLE exercises DROP COLUMN question_form;
ALTER TABLE contest_rounds DROP COLUMN question_form;

DROP TYPE question_form_enum;

