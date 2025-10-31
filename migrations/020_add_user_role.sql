-- +goose Up
-- Create user_role junction table for many-to-many relationship between users and roles
CREATE TABLE user_role (
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_user_role_user_id ON user_role(user_id);
CREATE INDEX idx_user_role_role_id ON user_role(role_id);

-- +goose Down
-- Drop the user_role table
DROP INDEX IF EXISTS idx_user_role_user_id;
DROP INDEX IF EXISTS idx_user_role_role_id;
DROP TABLE IF EXISTS user_role;
