CREATE TABLE IF NOT EXISTS user_star_exp (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    total_star BIGINT,
    change_star BIGINT,
    total_exp NUMERIC(15,2),
    change_exp NUMERIC(10,2),
    is_current BOOLEAN,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    description TEXT
);

CREATE INDEX IF NOT EXISTS idx_user_star_exp_user_id ON user_star_exp(user_id);

