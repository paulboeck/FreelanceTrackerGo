-- +goose Up
INSERT INTO settings (key, value, data_type, description) VALUES 
    ('list_page_size', '10', 'string', 'Number of items to display per page on list pages');

-- +goose Down
DELETE FROM settings WHERE key = 'list_page_size';