-- +goose Up
-- Create invoice table with SQLite syntax
CREATE TABLE invoice (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    invoice_date DATE NOT NULL,
    date_paid DATE NULL,
    payment_terms TEXT NOT NULL,
    amount_due DECIMAL(10,2) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    FOREIGN KEY (project_id) REFERENCES project(id)
);

-- Create indexes for performance
CREATE INDEX idx_invoice_project_id ON invoice(project_id);
CREATE INDEX idx_invoice_invoice_date ON invoice(invoice_date);
CREATE INDEX idx_invoice_date_paid ON invoice(date_paid);
CREATE INDEX idx_invoice_deleted_at ON invoice(deleted_at);

-- +goose Down
-- Drop the invoice table and indexes
DROP INDEX IF EXISTS idx_invoice_deleted_at;
DROP INDEX IF EXISTS idx_invoice_date_paid;
DROP INDEX IF EXISTS idx_invoice_invoice_date;
DROP INDEX IF EXISTS idx_invoice_project_id;
DROP TABLE IF EXISTS invoice;