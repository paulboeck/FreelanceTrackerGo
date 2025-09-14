-- +goose Up
-- Create user table with SQLite syntax
CREATE TABLE user (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    hashed_password TEXT(60) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

-- Create indexes for performance
CREATE INDEX idx_user_email ON user(email);
CREATE INDEX idx_user_deleted_at ON user(deleted_at);

-- +goose Down
-- Drop the user table
DROP INDEX IF EXISTS idx_user_email;
DROP INDEX IF EXISTS idx_user_deleted_at;
DROP TABLE IF EXISTS user;