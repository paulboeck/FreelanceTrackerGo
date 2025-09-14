-- +goose Up
INSERT INTO setting (key, value, data_type, description) VALUES 
    ('list_page_size', '10', 'string', 'Number of items to display per page on list pages');

-- +goose Down
DELETE FROM setting WHERE key = 'list_page_size';