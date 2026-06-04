ALTER TABLE
    classes
ADD
    COLUMN grade_id INT;

CREATE INDEX idx_classes_grade_id ON classes (grade_id);
