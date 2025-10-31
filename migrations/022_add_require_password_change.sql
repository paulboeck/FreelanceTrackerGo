-- +goose Up
-- Add require_password_change flag to user table
ALTER TABLE user ADD COLUMN require_password_change INTEGER NOT NULL DEFAULT 0;

-- +goose Down
-- Remove require_password_change column
ALTER TABLE user DROP COLUMN require_password_change;
