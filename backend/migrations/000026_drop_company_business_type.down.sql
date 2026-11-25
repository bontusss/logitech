ALTER TABLE companies
  ADD COLUMN IF NOT EXISTS business_type text NOT NULL DEFAULT 'intercity';
