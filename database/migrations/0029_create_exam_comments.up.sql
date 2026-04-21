-- +migrate Up

CREATE TABLE exam_comments
(
    exams_id   BIGINT NOT NULL,
    student_id BIGINT NOT NULL,
    teacher_id BIGINT,
    content    TEXT,
    type       TEXT   NOT NULL DEFAULT 'exam_comment',
    created_at TIMESTAMP       DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP       DEFAULT CURRENT_TIMESTAMP,
    created_by INT,
    updated_by INT,
    PRIMARY KEY (exams_id, student_id),
    FOREIGN KEY (exams_id) REFERENCES exams (id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (teacher_id) REFERENCES users (id) ON DELETE SET NULL
);

CREATE INDEX idx_exam_comments_exam ON exam_comments (exams_id);
CREATE INDEX idx_exam_comments_student ON exam_comments (student_id);
CREATE INDEX idx_exam_comments_type ON exam_comments (type);