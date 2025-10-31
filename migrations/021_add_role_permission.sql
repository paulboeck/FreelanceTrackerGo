-- +goose Up
-- Create role_permission junction table for many-to-many relationship between roles and permissions
CREATE TABLE role_permission (
    role_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permission(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_role_permission_role_id ON role_permission(role_id);
CREATE INDEX idx_role_permission_permission_id ON role_permission(permission_id);

-- +goose Down
-- Drop the role_permission table
DROP INDEX IF EXISTS idx_role_permission_role_id;
DROP INDEX IF EXISTS idx_role_permission_permission_id;
DROP TABLE IF EXISTS role_permission;
