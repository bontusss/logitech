UPDATE driver_profiles SET status = 'suspended' WHERE status = 'inactive';

ALTER TABLE driver_profiles
  DROP CONSTRAINT IF EXISTS driver_profiles_status_check;

ALTER TABLE driver_profiles
  ADD CONSTRAINT driver_profiles_status_check
  CHECK (status IN ('active', 'suspended'));

ALTER TABLE companies
  DROP COLUMN IF EXISTS prohibited_item_categories,
  DROP COLUMN IF EXISTS accepted_item_categories;
