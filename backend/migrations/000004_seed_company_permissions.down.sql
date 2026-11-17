DELETE FROM roles WHERE company_id IS NULL AND key = 'customer_service';

DELETE FROM permissions WHERE key IN (
  'company.view', 'company.update', 'company.settings.manage',
  'branches.view', 'branches.create', 'branches.update', 'branches.deactivate',
  'staff.view', 'staff.invite', 'staff.update', 'staff.deactivate', 'staff.roles.manage',
  'shipments.view', 'shipments.create', 'shipments.update',
  'rates.view', 'rates.create', 'rates.update', 'rates.publish',
  'reports.view', 'reports.export', 'branding.manage', 'subdomain.manage'
);
