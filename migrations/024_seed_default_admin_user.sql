-- +goose Up
-- Insert default admin user with hashed password for 'changeme'
-- Email: admin@example.com
-- Password: changeme (bcrypt hash with cost 12)
INSERT INTO user (name, email, hashed_password, require_password_change)
VALUES ('Administrator', 'admin@example.com', '$2a$12$9IRqjdNfj4rCHBrsSOh2fOsV8Fv3mg9aRr2qhAuP/4GeCWCcMiIcO', 1);

-- Assign Administrator role to the default admin user
INSERT INTO user_role (user_id, role_id)
VALUES (
    (SELECT id FROM user WHERE email = 'admin@example.com'),
    (SELECT id FROM role WHERE name = 'Administrator')
);

-- +goose Down
-- Remove the default admin user's role assignment
DELETE FROM user_role WHERE user_id = (SELECT id FROM user WHERE email = 'admin@example.com');

-- Remove the default admin user
DELETE FROM user WHERE email = 'admin@example.com';
