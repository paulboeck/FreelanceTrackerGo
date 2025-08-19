-- +goose Up
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    data_type TEXT NOT NULL CHECK (data_type IN ('string', 'int', 'float', 'decimal', 'bool')),
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert default settings
INSERT INTO settings (key, value, data_type, description) VALUES 
    ('default_hourly_rate', '85.00', 'decimal', 'Default hourly rate for new projects'),
    ('invoice_title', 'Invoice for Academic Editing', 'string', 'Title displayed on generated invoices'),
    ('freelancer_name', 'Your Name Here', 'string', 'Freelancer name for invoices'),
    ('freelancer_address', 'Your Address', 'string', 'Freelancer address for invoices'),
    ('freelancer_phone', 'Your Phone', 'string', 'Freelancer phone for invoices'),
    ('freelancer_email', 'your.email@example.com', 'string', 'Freelancer email for invoices');

-- +goose Down
DROP TABLE settings;