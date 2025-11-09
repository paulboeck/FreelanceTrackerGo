-- name: InsertAPIKey :execlastid
INSERT INTO api_keys (user_id, name, key_hash, key_prefix, scopes)
VALUES (?, ?, ?, ?, ?);

-- name: GetAPIKeyByHash :one
SELECT id, user_id, name, key_hash, key_prefix, scopes, last_used_at, created, updated, deleted_at
FROM api_keys
WHERE key_hash = ? AND deleted_at IS NULL;

-- name: GetAPIKeyByID :one
SELECT id, user_id, name, key_hash, key_prefix, scopes, last_used_at, created, updated, deleted_at
FROM api_keys
WHERE id = ? AND deleted_at IS NULL;

-- name: GetAPIKeysByUserID :many
SELECT id, user_id, name, key_hash, key_prefix, scopes, last_used_at, created, updated, deleted_at
FROM api_keys
WHERE user_id = ? AND deleted_at IS NULL
ORDER BY created DESC;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys
SET last_used_at = CURRENT_TIMESTAMP, updated = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateAPIKey :exec
UPDATE api_keys
SET name = ?, scopes = ?, updated = CURRENT_TIMESTAMP
WHERE id = ? AND deleted_at IS NULL;

-- name: SoftDeleteAPIKey :exec
UPDATE api_keys
SET deleted_at = CURRENT_TIMESTAMP, updated = CURRENT_TIMESTAMP
WHERE id = ? AND deleted_at IS NULL;

-- name: CountAPIKeysByUserID :one
SELECT COUNT(*)
FROM api_keys
WHERE user_id = ? AND deleted_at IS NULL;

-- name: GetAPIKeysByPrefix :many
SELECT id, user_id, name, key_hash, key_prefix, scopes, last_used_at, created, updated, deleted_at
FROM api_keys
WHERE key_prefix = ? AND deleted_at IS NULL;
