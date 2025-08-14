-- +goose Up
-- Create client table with SQLite syntax
CREATE TABLE client (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index for performance
CREATE INDEX idx_client_name ON client(name);

-- +goose Down
-- Drop the client table
DROP INDEX IF EXISTS idx_client_name;
DROP TABLE IF EXISTS client;