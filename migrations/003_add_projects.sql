-- +goose Up
-- Create project table with SQLite syntax
CREATE TABLE project (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    client_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    FOREIGN KEY (client_id) REFERENCES client(id)
);

-- Create indexes for performance
CREATE INDEX idx_project_client_id ON project(client_id);
CREATE INDEX idx_project_deleted_at ON project(deleted_at);
CREATE INDEX idx_project_name ON project(name);

-- +goose Down
-- Drop the project table and indexes
DROP INDEX IF EXISTS idx_project_name;
DROP INDEX IF EXISTS idx_project_deleted_at;
DROP INDEX IF EXISTS idx_project_client_id;
DROP TABLE IF EXISTS project;