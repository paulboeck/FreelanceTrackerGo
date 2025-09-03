-- name: InsertClient :execlastid
INSERT INTO client (name, email, phone, address1, address2, address3, city, state, zip_code, hourly_rate, notes, additional_info, additional_info2, bill_to, include_address_on_invoice, invoice_cc_email, invoice_cc_description, university_affiliation) 
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetClient :one
SELECT id, name, email, phone, address1, address2, address3, city, state, zip_code, hourly_rate, notes, additional_info, additional_info2, bill_to, include_address_on_invoice, invoice_cc_email, invoice_cc_description, university_affiliation, updated_at, created_at, deleted_at 
FROM client 
WHERE id = ? AND deleted_at IS NULL;

-- name: GetAllClients :many
SELECT id, name, email, phone, address1, address2, address3, city, state, zip_code, hourly_rate, notes, additional_info, additional_info2, bill_to, include_address_on_invoice, invoice_cc_email, invoice_cc_description, university_affiliation, updated_at, created_at, deleted_at 
FROM client 
WHERE deleted_at IS NULL
ORDER BY updated_at DESC;

-- name: GetClientsWithPagination :many
SELECT id, name, email, phone, address1, address2, address3, city, state, zip_code, hourly_rate, notes, additional_info, additional_info2, bill_to, include_address_on_invoice, invoice_cc_email, invoice_cc_description, university_affiliation, updated_at, created_at, deleted_at 
FROM client 
WHERE deleted_at IS NULL
ORDER BY updated_at DESC
LIMIT ? OFFSET ?;

-- name: GetClientsCount :one
SELECT COUNT(*) 
FROM client 
WHERE deleted_at IS NULL;

-- name: UpdateClient :exec
UPDATE client 
SET name = ?, email = ?, phone = ?, address1 = ?, address2 = ?, address3 = ?, city = ?, state = ?, zip_code = ?, hourly_rate = ?, notes = ?, additional_info = ?, additional_info2 = ?, bill_to = ?, include_address_on_invoice = ?, invoice_cc_email = ?, invoice_cc_description = ?, university_affiliation = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;

-- name: DeleteClient :exec
UPDATE client 
SET deleted_at = CURRENT_TIMESTAMP 
WHERE id = ? AND deleted_at IS NULL;