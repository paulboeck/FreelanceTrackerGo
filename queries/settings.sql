-- name: GetSetting :one
SELECT key, value, data_type, description, created_at, updated_at 
FROM setting 
WHERE key = ?;

-- name: GetAllSettings :many
SELECT key, value, data_type, description, created_at, updated_at 
FROM setting 
ORDER BY key;

-- name: UpdateSetting :exec
UPDATE setting 
SET value = ?, updated_at = CURRENT_TIMESTAMP 
WHERE key = ?;