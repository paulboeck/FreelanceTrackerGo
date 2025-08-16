-- name: InsertTimesheet :execlastid
INSERT INTO timesheet (project_id, work_date, hours_worked) 
VALUES (?, ?, ?);

-- name: GetTimesheet :one
SELECT id, project_id, work_date, hours_worked, updated_at, created_at, deleted_at 
FROM timesheet 
WHERE id = ? AND deleted_at IS NULL;

-- name: GetTimesheetsByProject :many
SELECT id, project_id, work_date, hours_worked, updated_at, created_at, deleted_at 
FROM timesheet 
WHERE project_id = ? AND deleted_at IS NULL
ORDER BY work_date DESC, created_at DESC;

-- name: UpdateTimesheet :exec
UPDATE timesheet 
SET work_date = ?, hours_worked = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: DeleteTimesheet :exec
UPDATE timesheet 
SET deleted_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;