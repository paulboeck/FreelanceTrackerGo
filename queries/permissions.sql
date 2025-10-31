-- name: GetPermissionByID :one
SELECT id, name, description, created_at, updated_at, deleted_at
FROM permission
WHERE id = ? AND deleted_at IS NULL;

-- name: GetPermissionByName :one
SELECT id, name, description, created_at, updated_at, deleted_at
FROM permission
WHERE name = ? AND deleted_at IS NULL;

-- name: GetAllPermissions :many
SELECT id, name, description, created_at, updated_at, deleted_at
FROM permission
WHERE deleted_at IS NULL
ORDER BY name;

-- name: GetRolePermissions :many
SELECT p.id, p.name, p.description, p.created_at, p.updated_at, p.deleted_at
FROM permission p
INNER JOIN role_permission rp ON p.id = rp.permission_id
WHERE rp.role_id = ? AND p.deleted_at IS NULL
ORDER BY p.name;

-- name: GetUserPermissions :many
SELECT DISTINCT p.id, p.name, p.description, p.created_at, p.updated_at, p.deleted_at
FROM permission p
INNER JOIN role_permission rp ON p.id = rp.permission_id
INNER JOIN user_role ur ON rp.role_id = ur.role_id
WHERE ur.user_id = ? AND p.deleted_at IS NULL
ORDER BY p.name;

-- name: AssignPermissionToRole :exec
INSERT INTO role_permission (role_id, permission_id)
VALUES (?, ?);

-- name: RemovePermissionFromRole :exec
DELETE FROM role_permission
WHERE role_id = ? AND permission_id = ?;

-- name: RemoveAllPermissionsFromRole :exec
DELETE FROM role_permission
WHERE role_id = ?;
