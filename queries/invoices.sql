-- name: InsertInvoice :execlastid
INSERT INTO invoice (project_id, invoice_date, date_paid, payment_terms, amount_due, display_details) 
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetInvoice :one
SELECT id, project_id, invoice_date, date_paid, payment_terms, amount_due, display_details, updated_at, created_at, deleted_at 
FROM invoice 
WHERE id = ? AND deleted_at IS NULL;

-- name: GetInvoicesByProject :many
SELECT id, project_id, invoice_date, date_paid, payment_terms, amount_due, display_details, updated_at, created_at, deleted_at 
FROM invoice 
WHERE project_id = ? AND deleted_at IS NULL
ORDER BY invoice_date DESC, created_at DESC;

-- name: UpdateInvoice :exec
UPDATE invoice 
SET invoice_date = ?, date_paid = ?, payment_terms = ?, amount_due = ?, display_details = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: DeleteInvoice :exec
UPDATE invoice 
SET deleted_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: GetInvoiceForPDF :one
SELECT 
    i.id, i.project_id, i.invoice_date, i.date_paid, i.payment_terms, i.amount_due, i.display_details,
    i.updated_at, i.created_at, i.deleted_at,
    p.name as project_name,
    c.name as client_name
FROM invoice i
JOIN project p ON i.project_id = p.id
JOIN client c ON p.client_id = c.id
WHERE i.id = ? AND i.deleted_at IS NULL;

-- name: GetInvoiceComprehensiveForPDF :one
SELECT 
    i.id, i.project_id, i.invoice_date, i.date_paid, i.payment_terms, i.amount_due, i.display_details,
    i.updated_at, i.created_at, i.deleted_at,
    p.name as project_name, p.status as project_status, p.hourly_rate as project_hourly_rate,
    p.discount_percent, p.discount_reason, p.adjustment_amount, p.adjustment_reason,
    p.currency_display, p.currency_conversion_rate, p.flat_fee_invoice,
    p.additional_info as project_additional_info, p.additional_info2 as project_additional_info2,
    c.id as client_id, c.name as client_name, c.email as client_email,
    c.phone as client_phone, c.address1 as client_address1, c.address2 as client_address2, 
    c.address3 as client_address3, c.city as client_city, c.state as client_state, 
    c.zip_code as client_zip_code, c.bill_to as client_bill_to,
    c.include_address_on_invoice, c.university_affiliation,
    c.additional_info as client_additional_info, c.additional_info2 as client_additional_info2
FROM invoice i
JOIN project p ON i.project_id = p.id
JOIN client c ON p.client_id = c.id
WHERE i.id = ? AND i.deleted_at IS NULL;