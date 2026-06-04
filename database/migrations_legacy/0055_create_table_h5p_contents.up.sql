CREATE TABLE public.h5p_content_user_data (
    id BIGSERIAL PRIMARY KEY,
    content_id BIGINT,
    user_id VARCHAR(255) NOT NULL,
    sub_content_id INT,
    data_type VARCHAR(255) NOT NULL,
    preload BOOLEAN DEFAULT false,
    invalidate BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    context_id VARCHAR(255),
    user_state JSONB
);

CREATE TABLE public.h5p_content_scores (
    id BIGSERIAL PRIMARY KEY,
    content_id BIGINT,
    user_id VARCHAR(255) NOT NULL,
    score INT NOT NULL,
    max_score INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    opened TIMESTAMP,
    finished TIMESTAMP,
    time TIMESTAMP
);

CREATE TABLE public.h5p_contents (
    id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    parameters JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    PRIMARY KEY (id)
);
