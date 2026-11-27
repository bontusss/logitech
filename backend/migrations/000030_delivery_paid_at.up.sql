-- Record the moment a delivery payment is captured so dashboards can report
-- collections by day instead of inferring them from the last update time.
ALTER TABLE deliveries
  ADD COLUMN paid_at timestamptz;

-- Existing paid rows have no payment timestamp. Their last update is the
-- closest available approximation; new rows are stamped by the trigger below.
UPDATE deliveries
SET paid_at = updated_at
WHERE is_paid AND paid_at IS NULL;

CREATE OR REPLACE FUNCTION set_delivery_paid_at()
RETURNS trigger AS $$
BEGIN
  IF NEW.is_paid THEN
    IF TG_OP = 'INSERT' OR OLD.is_paid IS DISTINCT FROM true THEN
      NEW.paid_at := COALESCE(NEW.paid_at, now());
    END IF;
  ELSE
    NEW.paid_at := NULL;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER deliveries_set_paid_at
BEFORE INSERT OR UPDATE ON deliveries
FOR EACH ROW EXECUTE FUNCTION set_delivery_paid_at();
