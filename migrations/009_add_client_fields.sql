-- +goose Up
-- Add additional fields to client table
ALTER TABLE client ADD COLUMN email TEXT NOT NULL DEFAULT '';
ALTER TABLE client ADD COLUMN phone TEXT;
ALTER TABLE client ADD COLUMN address TEXT;
ALTER TABLE client ADD COLUMN hourly_rate DECIMAL(10,2) NOT NULL DEFAULT 0.00;
ALTER TABLE client ADD COLUMN notes TEXT;
ALTER TABLE client ADD COLUMN additional_info TEXT;
ALTER TABLE client ADD COLUMN additional_info2 TEXT;
ALTER TABLE client ADD COLUMN bill_to TEXT;
ALTER TABLE client ADD COLUMN include_address_on_invoice BOOLEAN NOT NULL DEFAULT 1;
ALTER TABLE client ADD COLUMN invoice_cc_email TEXT;
ALTER TABLE client ADD COLUMN invoice_cc_description TEXT;
ALTER TABLE client ADD COLUMN university_affiliation TEXT;

-- Create index for email lookups
CREATE INDEX idx_client_email ON client(email);

-- +goose Down
-- Remove the added columns
DROP INDEX IF EXISTS idx_client_email;
ALTER TABLE client DROP COLUMN university_affiliation;
ALTER TABLE client DROP COLUMN invoice_cc_description;
ALTER TABLE client DROP COLUMN invoice_cc_email;
ALTER TABLE client DROP COLUMN include_address_on_invoice;
ALTER TABLE client DROP COLUMN bill_to;
ALTER TABLE client DROP COLUMN additional_info2;
ALTER TABLE client DROP COLUMN additional_info;
ALTER TABLE client DROP COLUMN notes;
ALTER TABLE client DROP COLUMN hourly_rate;
ALTER TABLE client DROP COLUMN address;
ALTER TABLE client DROP COLUMN phone;
ALTER TABLE client DROP COLUMN email;