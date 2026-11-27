-- Saved shipment searches a customer repeats on a schedule. The payload keeps
-- the route, addresses and package details so the customer can run the same
-- search again without re-entering everything.
CREATE TABLE customer_recurring_shipments (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  customer_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  label          text NOT NULL DEFAULT '',
  frequency      text NOT NULL DEFAULT 'monthly'
                 CHECK (frequency IN ('weekly', 'biweekly', 'monthly')),
  search_payload jsonb NOT NULL,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  CHECK (jsonb_typeof(search_payload) = 'object')
);

CREATE INDEX customer_recurring_shipments_customer_idx
  ON customer_recurring_shipments (customer_id, created_at DESC);

CREATE TRIGGER customer_recurring_shipments_set_updated_at
BEFORE UPDATE ON customer_recurring_shipments
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
