-- Align question_attributes with models + seed default metadata for question bank filters.

ALTER TABLE question_attributes
    ADD COLUMN IF NOT EXISTS subject_id BIGINT REFERENCES subjects(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS level INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS weight NUMERIC(10, 2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS status BOOLEAN DEFAULT TRUE;

-- Seed Kỹ năng / Mức độ / Mức độ nhận thức for the first subject (usually id=1).
DO $$
DECLARE
    sid BIGINT;
    skill_parent_id BIGINT;
    level_parent_id BIGINT;
    cognition_parent_id BIGINT;
BEGIN
    SELECT id INTO sid FROM subjects WHERE deleted_at IS NULL ORDER BY id LIMIT 1;
    IF sid IS NULL THEN
        RETURN;
    END IF;

    -- Parent: Kỹ năng
    INSERT INTO question_attributes (parent_id, subject_id, name, level, weight, status)
    SELECT 0, sid, 'Kỹ năng', 1, 1, TRUE
    WHERE NOT EXISTS (
        SELECT 1 FROM question_attributes
        WHERE name = 'Kỹ năng' AND subject_id = sid AND COALESCE(parent_id, 0) = 0
    );

    SELECT id INTO skill_parent_id FROM question_attributes
    WHERE name = 'Kỹ năng' AND subject_id = sid AND COALESCE(parent_id, 0) = 0
    ORDER BY id LIMIT 1;

    INSERT INTO question_attributes (parent_id, subject_id, name, level, weight, status)
    SELECT skill_parent_id, sid, v.name, 2, v.weight, TRUE
    FROM (VALUES
        ('Vocabulary', 1::NUMERIC),
        ('Structure', 2::NUMERIC),
        ('Phonics', 3::NUMERIC)
    ) AS v(name, weight)
    WHERE skill_parent_id IS NOT NULL
      AND NOT EXISTS (
        SELECT 1 FROM question_attributes c
        WHERE c.parent_id = skill_parent_id AND c.name = v.name
      );

    -- Parent: Mức độ
    INSERT INTO question_attributes (parent_id, subject_id, name, level, weight, status)
    SELECT 0, sid, 'Mức độ', 1, 1, TRUE
    WHERE NOT EXISTS (
        SELECT 1 FROM question_attributes
        WHERE name = 'Mức độ' AND subject_id = sid AND COALESCE(parent_id, 0) = 0
    );

    SELECT id INTO level_parent_id FROM question_attributes
    WHERE name = 'Mức độ' AND subject_id = sid AND COALESCE(parent_id, 0) = 0
    ORDER BY id LIMIT 1;

    INSERT INTO question_attributes (parent_id, subject_id, name, level, weight, status)
    SELECT level_parent_id, sid, v.name, 2, v.weight, TRUE
    FROM (VALUES
        ('Easy', 1::NUMERIC),
        ('Medium', 2::NUMERIC),
        ('Hard', 3::NUMERIC)
    ) AS v(name, weight)
    WHERE level_parent_id IS NOT NULL
      AND NOT EXISTS (
        SELECT 1 FROM question_attributes c
        WHERE c.parent_id = level_parent_id AND c.name = v.name
      );

    -- Parent: Mức độ nhận thức
    INSERT INTO question_attributes (parent_id, subject_id, name, level, weight, status)
    SELECT 0, sid, 'Mức độ nhận thức', 1, 1, TRUE
    WHERE NOT EXISTS (
        SELECT 1 FROM question_attributes
        WHERE name = 'Mức độ nhận thức' AND subject_id = sid AND COALESCE(parent_id, 0) = 0
    );

    SELECT id INTO cognition_parent_id FROM question_attributes
    WHERE name = 'Mức độ nhận thức' AND subject_id = sid AND COALESCE(parent_id, 0) = 0
    ORDER BY id LIMIT 1;

    INSERT INTO question_attributes (parent_id, subject_id, name, level, weight, status)
    SELECT cognition_parent_id, sid, v.name, 2, v.weight, TRUE
    FROM (VALUES
        ('Nhận diện', 1::NUMERIC),
        ('Ghi nhớ', 1::NUMERIC),
        ('Vận dụng', 2::NUMERIC),
        ('Vận dụng nâng cao/sáng tạo', 1::NUMERIC)
    ) AS v(name, weight)
    WHERE cognition_parent_id IS NOT NULL
      AND NOT EXISTS (
        SELECT 1 FROM question_attributes c
        WHERE c.parent_id = cognition_parent_id AND c.name = v.name
      );
END $$;
