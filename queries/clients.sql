-- name: InsertClient :execlastid
INSERT INTO client (name) 
VALUES (?);

-- name: GetClient :one
SELECT id, name, updated_at, created_at, deleted_at 
FROM client 
WHERE id = ? AND deleted_at IS NULL;

-- name: GetAllClients :many
SELECT id, name, updated_at, created_at, deleted_at 
FROM client 
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateClient :exec
UPDATE client 
SET name = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: DeleteClient :exec
UPDATE client 
SET deleted_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;