-- +goose Up
-- Add additional project attributes
ALTER TABLE project ADD COLUMN status TEXT NOT NULL DEFAULT 'Estimating';
ALTER TABLE project ADD COLUMN hourly_rate REAL NOT NULL DEFAULT 0.00;
ALTER TABLE project ADD COLUMN deadline TEXT;
ALTER TABLE project ADD COLUMN scheduled_start TEXT;
ALTER TABLE project ADD COLUMN invoice_cc_email TEXT;
ALTER TABLE project ADD COLUMN invoice_cc_description TEXT;
ALTER TABLE project ADD COLUMN schedule_comments TEXT;
ALTER TABLE project ADD COLUMN additional_info TEXT;
ALTER TABLE project ADD COLUMN additional_info2 TEXT;
ALTER TABLE project ADD COLUMN discount_percent REAL;
ALTER TABLE project ADD COLUMN discount_reason TEXT;
ALTER TABLE project ADD COLUMN adjustment_amount REAL;
ALTER TABLE project ADD COLUMN adjustment_reason TEXT;
ALTER TABLE project ADD COLUMN currency_display TEXT NOT NULL DEFAULT 'USD';
ALTER TABLE project ADD COLUMN currency_conversion_rate REAL NOT NULL DEFAULT 1.00000;
ALTER TABLE project ADD COLUMN flat_fee_invoice INTEGER NOT NULL DEFAULT 0;
ALTER TABLE project ADD COLUMN notes TEXT;

-- Create indexes for performance on commonly queried fields
CREATE INDEX idx_project_status ON project(status);
CREATE INDEX idx_project_deadline ON project(deadline);
CREATE INDEX idx_project_scheduled_start ON project(scheduled_start);

-- +goose Down
-- Drop the added columns and indexes
DROP INDEX IF EXISTS idx_project_scheduled_start;
DROP INDEX IF EXISTS idx_project_deadline;
DROP INDEX IF EXISTS idx_project_status;

ALTER TABLE project DROP COLUMN notes;
ALTER TABLE project DROP COLUMN flat_fee_invoice;
ALTER TABLE project DROP COLUMN currency_conversion_rate;
ALTER TABLE project DROP COLUMN currency_display;
ALTER TABLE project DROP COLUMN adjustment_reason;
ALTER TABLE project DROP COLUMN adjustment_amount;
ALTER TABLE project DROP COLUMN discount_reason;
ALTER TABLE project DROP COLUMN discount_percent;
ALTER TABLE project DROP COLUMN additional_info2;
ALTER TABLE project DROP COLUMN additional_info;
ALTER TABLE project DROP COLUMN schedule_comments;
ALTER TABLE project DROP COLUMN invoice_cc_description;
ALTER TABLE project DROP COLUMN invoice_cc_email;
ALTER TABLE project DROP COLUMN scheduled_start;
ALTER TABLE project DROP COLUMN deadline;
ALTER TABLE project DROP COLUMN hourly_rate;
ALTER TABLE project DROP COLUMN status;