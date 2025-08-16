-- +goose Up
-- Create timesheet table with SQLite syntax
CREATE TABLE timesheet (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    work_date DATE NOT NULL,
    hours_worked DECIMAL(5,2) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    FOREIGN KEY (project_id) REFERENCES project(id)
);

-- Create indexes for performance
CREATE INDEX idx_timesheet_project_id ON timesheet(project_id);
CREATE INDEX idx_timesheet_work_date ON timesheet(work_date);
CREATE INDEX idx_timesheet_deleted_at ON timesheet(deleted_at);

-- +goose Down
-- Drop the timesheet table and indexes
DROP INDEX IF EXISTS idx_timesheet_deleted_at;
DROP INDEX IF EXISTS idx_timesheet_work_date;
DROP INDEX IF EXISTS idx_timesheet_project_id;
DROP TABLE IF EXISTS timesheet;