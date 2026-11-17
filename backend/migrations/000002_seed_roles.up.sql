INSERT INTO permissions (key, description) VALUES
  ('company.manage',    'Update company profile and verification'),
  ('company.settings',  'Update company settings'),
  ('company.branding',  'Update company branding'),
  ('members.manage',    'Invite, update and remove company members'),
  ('shipments.manage',  'Create and manage shipments'),
  ('rates.publish',     'Publish and update shipping rates'),
  ('bookings.manage',   'Manage marketplace bookings'),
  ('reports.view',      'View company reports and analytics'),
  ('audit.view',        'View the company audit trail')
ON CONFLICT (key) DO NOTHING;

INSERT INTO roles (company_id, key, name, description, is_system)
SELECT NULL, 'owner', 'Company owner', 'Full control of the company account', true
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE company_id IS NULL AND key = 'owner');

INSERT INTO roles (company_id, key, name, description, is_system)
SELECT NULL, 'company_admin', 'Company administrator', 'Runs day-to-day company administration', true
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE company_id IS NULL AND key = 'company_admin');

INSERT INTO roles (company_id, key, name, description, is_system)
SELECT NULL, 'branch_manager', 'Branch manager', 'Manages a branch and its shipments', true
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE company_id IS NULL AND key = 'branch_manager');

INSERT INTO roles (company_id, key, name, description, is_system)
SELECT NULL, 'operations_staff', 'Operations staff', 'Handles shipments and deliveries', true
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE company_id IS NULL AND key = 'operations_staff');

-- Owner: every company permission.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.company_id IS NULL AND r.key = 'owner'
ON CONFLICT DO NOTHING;

-- Company administrator: everything except ownership-level profile control.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN (
  'company.settings', 'company.branding', 'members.manage', 'shipments.manage',
  'rates.publish', 'bookings.manage', 'reports.view', 'audit.view'
)
WHERE r.company_id IS NULL AND r.key = 'company_admin'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN ('shipments.manage', 'reports.view')
WHERE r.company_id IS NULL AND r.key = 'branch_manager'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN ('shipments.manage')
WHERE r.company_id IS NULL AND r.key = 'operations_staff'
ON CONFLICT DO NOTHING;
