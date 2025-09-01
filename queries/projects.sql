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
ORDER BY created_at DESC;

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