-- name: InsertProject :execlastid
INSERT INTO project (
    name, client_id, status, hourly_rate, deadline, scheduled_start,
    invoice_cc_email, invoice_cc_description, schedule_comments,
    additional_info, additional_info2, discount_percent, discount_reason,
    adjustment_amount, adjustment_reason, currency_display, 
    currency_conversion_rate, flat_fee_invoice, notes
) 
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetProject :one
SELECT id, name, client_id, status, hourly_rate, deadline, scheduled_start,
       invoice_cc_email, invoice_cc_description, schedule_comments,
       additional_info, additional_info2, discount_percent, discount_reason,
       adjustment_amount, adjustment_reason, currency_display, 
       currency_conversion_rate, flat_fee_invoice, notes,
       updated_at, created_at, deleted_at 
FROM project 
WHERE id = ? AND deleted_at IS NULL;

-- name: GetProjectsByClient :many
SELECT id, name, client_id, status, hourly_rate, deadline, scheduled_start,
       invoice_cc_email, invoice_cc_description, schedule_comments,
       additional_info, additional_info2, discount_percent, discount_reason,
       adjustment_amount, adjustment_reason, currency_display, 
       currency_conversion_rate, flat_fee_invoice, notes,
       updated_at, created_at, deleted_at 
FROM project 
WHERE client_id = ? AND deleted_at IS NULL
ORDER BY updated_at DESC;

-- name: UpdateProject :exec
UPDATE project 
SET name = ?, status = ?, hourly_rate = ?, deadline = ?, scheduled_start = ?,
    invoice_cc_email = ?, invoice_cc_description = ?, schedule_comments = ?,
    additional_info = ?, additional_info2 = ?, discount_percent = ?, discount_reason = ?,
    adjustment_amount = ?, adjustment_reason = ?, currency_display = ?, 
    currency_conversion_rate = ?, flat_fee_invoice = ?, notes = ?,
    updated_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: DeleteProject :exec
UPDATE project 
SET deleted_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: GetAllProjectsWithClient :many
SELECT p.id, p.name, p.client_id, p.status, p.hourly_rate, p.deadline, p.scheduled_start,
       p.invoice_cc_email, p.invoice_cc_description, p.schedule_comments,
       p.additional_info, p.additional_info2, p.discount_percent, p.discount_reason,
       p.adjustment_amount, p.adjustment_reason, p.currency_display, 
       p.currency_conversion_rate, p.flat_fee_invoice, p.notes,
       p.updated_at, p.created_at, p.deleted_at,
       c.name as client_name
FROM project p
JOIN client c ON p.client_id = c.id
WHERE p.deleted_at IS NULL AND c.deleted_at IS NULL
ORDER BY p.updated_at DESC;

-- name: GetProjectsWithClientPagination :many
SELECT p.id, p.name, p.client_id, p.status, p.hourly_rate, p.deadline, p.scheduled_start,
       p.invoice_cc_email, p.invoice_cc_description, p.schedule_comments,
       p.additional_info, p.additional_info2, p.discount_percent, p.discount_reason,
       p.adjustment_amount, p.adjustment_reason, p.currency_display, 
       p.currency_conversion_rate, p.flat_fee_invoice, p.notes,
       p.updated_at, p.created_at, p.deleted_at,
       c.name as client_name
FROM project p
JOIN client c ON p.client_id = c.id
WHERE p.deleted_at IS NULL AND c.deleted_at IS NULL
ORDER BY p.updated_at DESC
LIMIT ? OFFSET ?;

-- name: GetProjectsCount :one
SELECT COUNT(*) 
FROM project p
JOIN client c ON p.client_id = c.id
WHERE p.deleted_at IS NULL AND c.deleted_at IS NULL;