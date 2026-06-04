-- +migrate Up

CREATE TABLE role_permissions
(
    permission_id INT NOT NULL,
    role_id       INT NOT NULL,
    PRIMARY KEY (permission_id, role_id),
    FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE
);

CREATE INDEX idx_role_permissions_permission ON role_permissions (permission_id);
CREATE INDEX idx_role_permissions_role ON role_permissions (role_id);