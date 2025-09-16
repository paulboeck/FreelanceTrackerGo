-- +goose Up
-- Add email configuration settings
INSERT INTO setting (key, value, data_type, description) VALUES 
    ('email_enabled', 'false', 'bool', 'Enable email functionality'),
    ('smtp_host', 'smtp.gmail.com', 'string', 'SMTP server hostname'),
    ('smtp_port', '587', 'int', 'SMTP server port'),
    ('smtp_username', '', 'string', 'SMTP username (email address)'),
    ('smtp_password', '', 'string', 'SMTP password (app password for Gmail)'),
    ('smtp_from_name', 'FreelanceTracker', 'string', 'From name for outgoing emails'),
    ('smtp_use_tls', 'true', 'bool', 'Use TLS encryption for SMTP');

-- +goose Down
-- Remove email settings
DELETE FROM setting WHERE key IN (
    'email_enabled',
    'smtp_host', 
    'smtp_port',
    'smtp_username',
    'smtp_password',
    'smtp_from_name',
    'smtp_use_tls'
);