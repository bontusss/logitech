-- Providers are no longer classified as intercity or last-mile up front.
-- Their configured zones and routes determine what they can serve.
ALTER TABLE companies
  DROP COLUMN IF EXISTS business_type;
