CREATE TABLE schools (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255),
    short_name VARCHAR(255),
    type TEXT NOT NULL DEFAULT 'school',
    ward_code VARCHAR(20),
    address_vn TEXT,
    address_en TEXT,
    parent_school_id BIGINT REFERENCES schools(id) ON DELETE SET NULL,
    contact_name VARCHAR(255),
    contact_phone VARCHAR(50),
    contact_mail VARCHAR(255),
    status BOOLEAN NOT NULL DEFAULT TRUE,
    logo_info JSONB,
    student_count INT DEFAULT 0,
    class_count INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);
CREATE INDEX idx_schools_name ON schools(name);
CREATE INDEX idx_schools_deleted_at ON schools(deleted_at);

CREATE TABLE grades (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE classes (
    id BIGSERIAL PRIMARY KEY,
    school_id BIGINT REFERENCES schools(id) ON DELETE CASCADE,
    grade_id BIGINT REFERENCES grades(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    status BOOLEAN DEFAULT TRUE,
    sort_position INT DEFAULT 0,
    current_students INT DEFAULT 0,
    max_students INT DEFAULT 0,
    teacher_info JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);
CREATE INDEX idx_classes_school_id ON classes(school_id);

CREATE TABLE subjects (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE programs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    duration INT DEFAULT 0,
    status BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE courses (
    id BIGSERIAL PRIMARY KEY,
    subject_id BIGINT REFERENCES subjects(id) ON DELETE SET NULL,
    program_id BIGINT REFERENCES programs(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    object_title VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT,
    duration INT DEFAULT 0,
    type VARCHAR(255) NOT NULL DEFAULT '',
    image_info JSONB,
    level VARCHAR(255) NOT NULL DEFAULT '',
    status BOOLEAN,
    current_students INT DEFAULT 0,
    target TEXT,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    time VARCHAR(255),
    state course_state NOT NULL DEFAULT 'coming',
    clone_info JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);
CREATE INDEX idx_courses_subject_id ON courses(subject_id);
CREATE INDEX idx_courses_state ON courses(state);

CREATE TABLE course_schools (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    school_id BIGINT NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    UNIQUE (course_id, school_id)
);

CREATE TABLE chapters (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title VARCHAR(255),
    object_title VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT,
    sort_position INT DEFAULT 0,
    clone_info JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);
CREATE INDEX idx_chapters_course_id ON chapters(course_id);

CREATE TABLE lessons (
    id BIGSERIAL PRIMARY KEY,
    chapter_id BIGINT NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    title VARCHAR(255),
    object_title VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT,
    status BOOLEAN DEFAULT TRUE,
    sort_position INT DEFAULT 0,
    views INT DEFAULT 0,
    author_id BIGINT,
    clone_info JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);
CREATE INDEX idx_lessons_chapter_id ON lessons(chapter_id);
