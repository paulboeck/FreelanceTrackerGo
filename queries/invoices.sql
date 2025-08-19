-- name: InsertInvoice :execlastid
INSERT INTO invoice (project_id, invoice_date, date_paid, payment_terms, amount_due) 
VALUES (?, ?, ?, ?, ?);

-- name: GetInvoice :one
SELECT id, project_id, invoice_date, date_paid, payment_terms, amount_due, updated_at, created_at, deleted_at 
FROM invoice 
WHERE id = ? AND deleted_at IS NULL;

-- name: GetInvoicesByProject :many
SELECT id, project_id, invoice_date, date_paid, payment_terms, amount_due, updated_at, created_at, deleted_at 
FROM invoice 
WHERE project_id = ? AND deleted_at IS NULL
ORDER BY invoice_date DESC, created_at DESC;

-- name: UpdateInvoice :exec
UPDATE invoice 
SET invoice_date = ?, date_paid = ?, payment_terms = ?, amount_due = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: DeleteInvoice :exec
UPDATE invoice 
SET deleted_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;