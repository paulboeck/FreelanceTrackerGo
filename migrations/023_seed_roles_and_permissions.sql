-- +goose Up
-- Seed default roles
INSERT INTO role (name, description) VALUES
    ('Administrator', 'Full system access with all permissions'),
    ('Freelancer', 'Standard user with access to clients, projects, timesheets, invoices, and reports');

-- Seed all application permissions
INSERT INTO permission (name, description) VALUES
    -- Client Management
    ('clients.list', 'View client list'),
    ('clients.view', 'View individual client details'),
    ('clients.create', 'Create new clients'),
    ('clients.edit', 'Edit existing clients'),
    ('clients.delete', 'Delete clients'),

    -- Project Management
    ('projects.list', 'View project list'),
    ('projects.view', 'View individual project details'),
    ('projects.create', 'Create new projects'),
    ('projects.edit', 'Edit existing projects'),
    ('projects.delete', 'Delete projects'),

    -- Timesheet Management
    ('timesheets.create', 'Create new timesheets'),
    ('timesheets.edit', 'Edit existing timesheets'),
    ('timesheets.delete', 'Delete timesheets'),

    -- Invoice Management
    ('invoices.create', 'Create new invoices'),
    ('invoices.edit', 'Edit existing invoices'),
    ('invoices.delete', 'Delete invoices'),
    ('invoices.print', 'Print/view invoice PDFs'),
    ('invoices.email', 'Email invoices to clients'),

    -- Reports
    ('reports.view', 'View income and other reports'),

    -- Settings
    ('settings.view', 'View application settings'),
    ('settings.edit', 'Edit application settings'),

    -- User Management
    ('users.create', 'Create new users'),
    ('users.list', 'View user list'),
    ('users.edit', 'Edit existing users'),
    ('users.delete', 'Delete users');

-- Assign ALL permissions to Administrator role
INSERT INTO role_permission (role_id, permission_id)
SELECT
    (SELECT id FROM role WHERE name = 'Administrator'),
    id
FROM permission;

-- Assign non-admin permissions to Freelancer role (all except settings.edit and users.*)
INSERT INTO role_permission (role_id, permission_id)
SELECT
    (SELECT id FROM role WHERE name = 'Freelancer'),
    id
FROM permission
WHERE name IN (
    'clients.list', 'clients.view', 'clients.create', 'clients.edit', 'clients.delete',
    'projects.list', 'projects.view', 'projects.create', 'projects.edit', 'projects.delete',
    'timesheets.create', 'timesheets.edit', 'timesheets.delete',
    'invoices.create', 'invoices.edit', 'invoices.delete', 'invoices.print', 'invoices.email',
    'reports.view',
    'settings.view'
);

-- +goose Down
-- Remove all role_permission assignments
DELETE FROM role_permission;

-- Remove all permissions
DELETE FROM permission;

-- Remove all roles
DELETE FROM role;
