-- name: GetRoleByID :one
SELECT id, name, description, created_at, updated_at, deleted_at
FROM role
WHERE id = ? AND deleted_at IS NULL;

-- name: GetRoleByName :one
SELECT id, name, description, created_at, updated_at, deleted_at
FROM role
WHERE name = ? AND deleted_at IS NULL;

-- name: GetAllRoles :many
SELECT id, name, description, created_at, updated_at, deleted_at
FROM role
WHERE deleted_at IS NULL
ORDER BY name;

-- name: InsertRole :execlastid
INSERT INTO role (name, description)
VALUES (?, ?);

-- name: UpdateRole :exec
UPDATE role
SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND deleted_at IS NULL;

-- name: DeleteRole :exec
UPDATE role
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ? AND deleted_at IS NULL;

-- name: GetUserRoles :many
SELECT r.id, r.name, r.description, r.created_at, r.updated_at, r.deleted_at
FROM role r
INNER JOIN user_role ur ON r.id = ur.role_id
WHERE ur.user_id = ? AND r.deleted_at IS NULL
ORDER BY r.name;

-- name: AssignRoleToUser :exec
INSERT INTO user_role (user_id, role_id)
VALUES (?, ?);

-- name: RemoveRoleFromUser :exec
DELETE FROM user_role
WHERE user_id = ? AND role_id = ?;

-- name: RemoveAllRolesFromUser :exec
DELETE FROM user_role
WHERE user_id = ?;
