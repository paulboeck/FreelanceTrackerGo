-- name: InsertClient :one
INSERT INTO client (name) 
VALUES (?)
RETURNING id;

-- name: GetClient :one
SELECT id, name, updated_at, created_at 
FROM client 
WHERE id = ?;

-- name: GetAllClients :many
SELECT id, name, updated_at, created_at 
FROM client 
ORDER BY created_at DESC;