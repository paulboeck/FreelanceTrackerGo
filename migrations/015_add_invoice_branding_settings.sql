-- +goose Up
-- Add additional invoice branding and configuration settings
-- Note: For company_logo_path, PNG format works best. Logo will be scaled to 22.5mm width.
INSERT INTO settings (key, value, data_type, description) VALUES 
    ('freelancer_city_state_zip', 'Your City, State ZIP', 'string', 'Freelancer city, state, and ZIP code for invoices'),
    ('invoice_payment_terms_default', 'Payment is due within 30 days of receipt of this invoice. Thank you for your business!', 'string', 'Default payment terms text for invoices'),
    ('invoice_thank_you_message', 'Thank you for your business!', 'string', 'Thank you message at bottom of invoices'),
    ('invoice_show_individual_timesheets', 'true', 'bool', 'Whether to show individual timesheet line items on invoices'),
    ('invoice_currency_symbol', '$', 'string', 'Currency symbol to display on invoices'),
    ('company_logo_path', './ui/static/img/logo.png', 'string', 'Path to company logo file for invoices (PNG format recommended, displayed at 22.5mm width)');

-- +goose Down
DELETE FROM settings WHERE key IN (
    'freelancer_city_state_zip',
    'invoice_header_decoration', 
    'invoice_payment_terms_default',
    'invoice_thank_you_message',
    'invoice_show_individual_timesheets',
    'invoice_currency_symbol',
    'company_logo_path'
);