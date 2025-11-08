-- +goose Up
-- Allow NULL values for payment_terms in invoice table
-- SQLite doesn't support ALTER COLUMN, so we need to recreate the table

-- Create new table with payment_terms allowing NULL
CREATE TABLE invoice_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    invoice_date DATE NOT NULL,
    date_paid DATE NULL,
    payment_terms TEXT NULL,
    amount_due DECIMAL(10,2) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    display_details INTEGER NOT NULL DEFAULT 0,
    invoice_num INTEGER,
    FOREIGN KEY (project_id) REFERENCES project(id)
);

-- Copy data from old table
INSERT INTO invoice_new (id, project_id, invoice_date, date_paid, payment_terms, amount_due, created_at, updated_at, deleted_at, display_details, invoice_num)
SELECT id, project_id, invoice_date, date_paid, payment_terms, amount_due, created_at, updated_at, deleted_at, display_details, invoice_num
FROM invoice;

-- Drop old table
DROP TABLE invoice;

-- Rename new table
ALTER TABLE invoice_new RENAME TO invoice;

-- Recreate indexes
CREATE INDEX idx_invoice_project_id ON invoice(project_id);
CREATE INDEX idx_invoice_invoice_date ON invoice(invoice_date);
CREATE INDEX idx_invoice_date_paid ON invoice(date_paid);
CREATE INDEX idx_invoice_deleted_at ON invoice(deleted_at);

-- +goose Down
-- Revert to NOT NULL constraint for payment_terms

-- Create table with NOT NULL constraint
CREATE TABLE invoice_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    invoice_date DATE NOT NULL,
    date_paid DATE NULL,
    payment_terms TEXT NOT NULL,
    amount_due DECIMAL(10,2) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    display_details INTEGER NOT NULL DEFAULT 0,
    invoice_num INTEGER,
    FOREIGN KEY (project_id) REFERENCES project(id)
);

-- Copy data from old table (rows with NULL payment_terms will fail)
INSERT INTO invoice_new (id, project_id, invoice_date, date_paid, payment_terms, amount_due, created_at, updated_at, deleted_at, display_details, invoice_num)
SELECT id, project_id, invoice_date, date_paid, payment_terms, amount_due, created_at, updated_at, deleted_at, display_details, invoice_num
FROM invoice;

-- Drop old table
DROP TABLE invoice;

-- Rename new table
ALTER TABLE invoice_new RENAME TO invoice;

-- Recreate indexes
CREATE INDEX idx_invoice_project_id ON invoice(project_id);
CREATE INDEX idx_invoice_invoice_date ON invoice(invoice_date);
CREATE INDEX idx_invoice_date_paid ON invoice(date_paid);
CREATE INDEX idx_invoice_deleted_at ON invoice(deleted_at);
