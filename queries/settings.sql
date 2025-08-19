-- name: GetSetting :one
SELECT key, value, data_type, description, created_at, updated_at 
FROM settings 
WHERE key = ?;

-- name: GetAllSettings :many
SELECT key, value, data_type, description, created_at, updated_at 
FROM settings 
ORDER BY key;

-- name: UpdateSetting :exec
UPDATE settings 
SET value = ?, updated_at = CURRENT_TIMESTAMP 
WHERE key = ?;