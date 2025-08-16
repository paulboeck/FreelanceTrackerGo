-- +goose Up
-- Add deleted_at column for soft delete functionality
ALTER TABLE client ADD COLUMN deleted_at DATETIME NULL;

-- Create index for performance on non-deleted clients
CREATE INDEX idx_client_deleted_at ON client(deleted_at);

-- +goose Down
-- Remove the deleted_at column and index
DROP INDEX IF EXISTS idx_client_deleted_at;
ALTER TABLE client DROP COLUMN deleted_at;