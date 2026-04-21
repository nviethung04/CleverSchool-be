CREATE TABLE user_ref_roles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    CONSTRAINT uq_user_role UNIQUE (user_id, role_id)
);

CREATE INDEX idx_user_ref_roles_user_id ON user_ref_roles(user_id);
CREATE INDEX idx_user_ref_roles_role_id ON user_ref_roles(role_id);

INSERT INTO user_ref_roles (user_id, role_id)
SELECT id AS user_id, role_id
FROM users
WHERE role_id IS NOT NULL
ON CONFLICT (user_id, role_id) DO NOTHING;
