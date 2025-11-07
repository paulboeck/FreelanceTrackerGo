-- +goose Up
-- Add invoice_num column to invoice table
ALTER TABLE invoice ADD COLUMN invoice_num INTEGER;

-- Set invoice_num to id for existing records
UPDATE invoice SET invoice_num = id WHERE invoice_num IS NULL;

-- +goose Down
-- Remove invoice_num column
ALTER TABLE invoice DROP COLUMN invoice_num;
