DROP TABLE IF EXISTS deliveries;
ALTER TABLE driver_profiles
  DROP CONSTRAINT IF EXISTS driver_profiles_company_id_id_unique;
