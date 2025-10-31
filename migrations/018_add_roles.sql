-- +goose Up
-- Create role table
CREATE TABLE role (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

-- Create indexes for performance
CREATE INDEX idx_role_name ON role(name);
CREATE INDEX idx_role_deleted_at ON role(deleted_at);

-- +goose Down
-- Drop the role table
DROP INDEX IF EXISTS idx_role_name;
DROP INDEX IF EXISTS idx_role_deleted_at;
DROP TABLE IF EXISTS role;
