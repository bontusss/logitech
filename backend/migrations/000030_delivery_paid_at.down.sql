DROP TRIGGER IF EXISTS deliveries_set_paid_at ON deliveries;
DROP FUNCTION IF EXISTS set_delivery_paid_at();
ALTER TABLE deliveries DROP COLUMN IF EXISTS paid_at;
