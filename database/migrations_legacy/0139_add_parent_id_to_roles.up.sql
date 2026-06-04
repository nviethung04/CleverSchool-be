ALTER TABLE roles
ADD COLUMN parent_id BIGINT DEFAULT 0;

INSERT INTO roles (name, status, parent_id)
VALUES ('school', true, 0);
