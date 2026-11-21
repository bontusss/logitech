ALTER TABLE driver_profiles
  DROP CONSTRAINT IF EXISTS driver_profiles_status_check;

ALTER TABLE driver_profiles
  ADD CONSTRAINT driver_profiles_status_check
  CHECK (status IN ('active', 'suspended', 'inactive'));

ALTER TABLE companies
  ADD COLUMN accepted_item_categories text[] NOT NULL
    DEFAULT ARRAY['cloth', 'electronics', 'document', 'food', 'jewelry', 'other']::text[],
  ADD COLUMN prohibited_item_categories text[] NOT NULL DEFAULT '{}'::text[];
