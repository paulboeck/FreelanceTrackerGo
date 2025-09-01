-- +goose Up
-- Add hourly_rate column to timesheet table
ALTER TABLE timesheet ADD COLUMN hourly_rate REAL NOT NULL DEFAULT 0.00;

-- +goose Down
-- Remove the hourly_rate column
ALTER TABLE timesheet DROP COLUMN hourly_rate;