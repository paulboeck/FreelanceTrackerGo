-- +goose Up
-- Add description column to timesheet table
ALTER TABLE timesheet ADD COLUMN description VARCHAR(255);

-- +goose Down
-- Remove the description column
ALTER TABLE timesheet DROP COLUMN description;