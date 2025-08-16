-- name: InsertProject :execlastid
INSERT INTO project (name, client_id) 
VALUES (?, ?);

-- name: GetProject :one
SELECT id, name, client_id, updated_at, created_at, deleted_at 
FROM project 
WHERE id = ? AND deleted_at IS NULL;

-- name: GetProjectsByClient :many
SELECT id, name, client_id, updated_at, created_at, deleted_at 
FROM project 
WHERE client_id = ? AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateProject :exec
UPDATE project 
SET name = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: DeleteProject :exec
UPDATE project 
SET deleted_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;