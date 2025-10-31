-- name: InsertUser :execlastid
INSERT INTO user (name, email, hashed_password, require_password_change)
VALUES (?, ?, ?, ?);

-- name: GetUserByEmail :one
SELECT id, name, email, hashed_password, require_password_change, created_at, updated_at, deleted_at
FROM user
WHERE email = ? AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, name, email, hashed_password, require_password_change, created_at, updated_at, deleted_at
FROM user
WHERE id = ? AND deleted_at IS NULL;

-- name: UserExists :one
SELECT EXISTS(SELECT 1 FROM user WHERE email = ? AND deleted_at IS NULL);

-- name: UpdateUser :exec
UPDATE user
SET name = ?, email = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND deleted_at IS NULL;

-- name: UpdateUserPassword :exec
UPDATE user
SET hashed_password = ?, require_password_change = 0, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND deleted_at IS NULL;

-- name: GetAllUsers :many
SELECT id, name, email, hashed_password, require_password_change, created_at, updated_at, deleted_at
FROM user
WHERE deleted_at IS NULL
ORDER BY name;

-- name: DeleteUser :exec
UPDATE user 
SET deleted_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;