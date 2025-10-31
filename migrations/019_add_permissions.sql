-- +goose Up
-- Create permission table
CREATE TABLE permission (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

-- Create indexes for performance
CREATE INDEX idx_permission_name ON permission(name);
CREATE INDEX idx_permission_deleted_at ON permission(deleted_at);

-- +goose Down
-- Drop the permission table
DROP INDEX IF EXISTS idx_permission_name;
DROP INDEX IF EXISTS idx_permission_deleted_at;
DROP TABLE IF EXISTS permission;
