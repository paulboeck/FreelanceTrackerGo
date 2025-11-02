-- +goose Up
-- Add invoice_num column to invoice table
ALTER TABLE invoice ADD COLUMN invoice_num INTEGER;

-- Set invoice_num to id for existing records
UPDATE invoice SET invoice_num = id;

-- Create trigger to automatically set invoice_num to id on insert
CREATE TRIGGER set_invoice_num_after_insert
AFTER INSERT ON invoice
FOR EACH ROW
BEGIN
    UPDATE invoice SET invoice_num = NEW.id WHERE id = NEW.id;
END;

-- +goose Down
-- Drop the trigger
DROP TRIGGER IF EXISTS set_invoice_num_after_insert;

-- Remove invoice_num column
ALTER TABLE invoice DROP COLUMN invoice_num;
