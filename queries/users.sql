-- name: InsertUser :execlastid
INSERT INTO user (name, email, hashed_password) 
VALUES (?, ?, ?);

-- name: GetUserByEmail :one
SELECT id, name, email, hashed_password, created_at, updated_at, deleted_at 
FROM user 
WHERE email = ? AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, name, email, hashed_password, created_at, updated_at, deleted_at 
FROM user 
WHERE id = ? AND deleted_at IS NULL;

-- name: UserExists :one
SELECT EXISTS(SELECT 1 FROM user WHERE email = ? AND deleted_at IS NULL);

-- name: UpdateUser :exec
UPDATE user 
SET name = ?, email = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: DeleteUser :exec
UPDATE user 
SET deleted_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;