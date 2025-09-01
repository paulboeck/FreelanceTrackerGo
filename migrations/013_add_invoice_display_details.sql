-- +goose Up
-- Add display_details boolean field to invoice table
ALTER TABLE invoice ADD COLUMN display_details BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
-- Remove display_details field from invoice table
ALTER TABLE invoice DROP COLUMN display_details;