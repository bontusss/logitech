-- Granular functional permissions for company management. Existing coarse
-- permissions are retained so earlier endpoints and tokens keep working.

INSERT INTO permissions (key, description) VALUES
  ('company.view',            'View the company profile'),
  ('company.update',          'Update the company profile'),
  ('company.settings.manage', 'Manage company settings'),
  ('branches.view',           'View branches'),
  ('branches.create',         'Create branches'),
  ('branches.update',         'Update branches and assign managers'),
  ('branches.deactivate',     'Activate or deactivate branches'),
  ('staff.view',              'View the staff directory'),
  ('staff.invite',            'Invite staff'),
  ('staff.update',            'Update staff details and branch assignments'),
  ('staff.deactivate',        'Suspend or remove staff'),
  ('staff.roles.manage',      'Assign roles and manage custom roles'),
  ('shipments.view',          'View shipments'),
  ('shipments.create',        'Create shipments'),
  ('shipments.update',        'Update shipments'),
  ('rates.view',              'View shipping rates'),
  ('rates.create',            'Create shipping rates'),
  ('rates.update',            'Update shipping rates'),
  ('rates.publish',           'Publish shipping rates'),
  ('reports.view',            'View reports'),
  ('reports.export',          'Export reports'),
  ('branding.manage',         'Manage company branding'),
  ('subdomain.manage',        'Manage the company subdomain')
ON CONFLICT (key) DO NOTHING;

INSERT INTO roles (company_id, key, name, description, is_system)
SELECT NULL, 'customer_service', 'Customer service', 'Handles customer enquiries and shipment visibility', true
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE company_id IS NULL AND key = 'customer_service');

-- Owner: every permission, including future additions.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.company_id IS NULL AND r.key = 'owner'
ON CONFLICT DO NOTHING;

-- Administrator: every operational permission except ownership itself.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN (
  'company.view', 'company.update', 'company.settings.manage',
  'branches.view', 'branches.create', 'branches.update', 'branches.deactivate',
  'staff.view', 'staff.invite', 'staff.update', 'staff.deactivate', 'staff.roles.manage',
  'shipments.view', 'shipments.create', 'shipments.update',
  'rates.view', 'rates.create', 'rates.update', 'rates.publish',
  'reports.view', 'reports.export', 'branding.manage', 'subdomain.manage'
)
WHERE r.company_id IS NULL AND r.key = 'company_admin'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN (
  'company.view', 'branches.view', 'branches.update', 'staff.view',
  'shipments.view', 'shipments.create', 'shipments.update',
  'rates.view', 'reports.view'
)
WHERE r.company_id IS NULL AND r.key = 'branch_manager'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN (
  'company.view', 'branches.view', 'shipments.view', 'shipments.create',
  'shipments.update', 'rates.view'
)
WHERE r.company_id IS NULL AND r.key = 'operations_staff'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN (
  'company.view', 'branches.view', 'staff.view', 'shipments.view',
  'rates.view', 'reports.view'
)
WHERE r.company_id IS NULL AND r.key = 'customer_service'
ON CONFLICT DO NOTHING;
