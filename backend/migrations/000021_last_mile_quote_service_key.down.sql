DELETE FROM marketplace_quotes WHERE service_key = 'last_mile';

ALTER TABLE marketplace_quotes
  DROP CONSTRAINT IF EXISTS marketplace_quotes_service_key_check;

ALTER TABLE marketplace_quotes
  ADD CONSTRAINT marketplace_quotes_service_key_check
  CHECK (service_key IN ('standard', 'express', 'same_day'));
