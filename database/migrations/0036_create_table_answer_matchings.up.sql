CREATE TABLE IF NOT EXISTS answer_matchings (
	id BIGSERIAL PRIMARY KEY,
    question_id INT NOT NULL,

	content TEXT,
    file_url VARCHAR(255),
	is_correct BOOLEAN DEFAULT FALSE,
    kind KIND_ENUM NOT NULL,
    point INTEGER,

    matching_content TEXT,
    matching_file_url VARCHAR(255),
    matching_kind KIND_ENUM NOT NULL,

	sort_position INT,
	correct_position INT,

    CONSTRAINT fk_answer_matchings_question FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE SET NULL
);
