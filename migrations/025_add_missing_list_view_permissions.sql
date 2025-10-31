-- +goose Up
-- Add missing list and view permissions for timesheets and invoices
INSERT INTO permission (name, description) VALUES
    ('timesheets.list', 'View timesheet list'),
    ('timesheets.view', 'View individual timesheet details'),
    ('invoices.list', 'View invoice list'),
    ('invoices.view', 'View individual invoice details');

-- Assign these permissions to Administrator role
INSERT INTO role_permission (role_id, permission_id)
SELECT
    (SELECT id FROM role WHERE name = 'Administrator'),
    id
FROM permission
WHERE name IN ('timesheets.list', 'timesheets.view', 'invoices.list', 'invoices.view');

-- Assign these permissions to Freelancer role
INSERT INTO role_permission (role_id, permission_id)
SELECT
    (SELECT id FROM role WHERE name = 'Freelancer'),
    id
FROM permission
WHERE name IN ('timesheets.list', 'timesheets.view', 'invoices.list', 'invoices.view');

-- +goose Down
-- Remove the role_permission assignments
DELETE FROM role_permission
WHERE permission_id IN (
    SELECT id FROM permission
    WHERE name IN ('timesheets.list', 'timesheets.view', 'invoices.list', 'invoices.view')
);

-- Remove the permissions
DELETE FROM permission
WHERE name IN ('timesheets.list', 'timesheets.view', 'invoices.list', 'invoices.view');
