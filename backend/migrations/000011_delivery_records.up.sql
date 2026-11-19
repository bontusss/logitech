ALTER TABLE driver_profiles
  ADD CONSTRAINT driver_profiles_company_id_id_unique UNIQUE (company_id, id);

CREATE TABLE deliveries (
  id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id            uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  sender_name           text NOT NULL,
  sender_phone          text NOT NULL,
  receiver_name         text NOT NULL,
  receiver_phone        text NOT NULL,
  pickup_address        text NOT NULL,
  dropoff_address       text NOT NULL,
  item_description      text NOT NULL,
  declared_value_minor  bigint NOT NULL CHECK (declared_value_minor >= 0),
  delivery_price_minor  bigint NOT NULL CHECK (delivery_price_minor >= 0),
  payer                 text NOT NULL CHECK (payer IN ('sender', 'receiver')),
  is_paid               boolean NOT NULL DEFAULT false,
  payment_method        text CHECK (payment_method IN ('cash', 'pos', 'bank_transfer')),
  assigned_driver_id    uuid,
  status                text NOT NULL DEFAULT 'created'
                        CHECK (status IN ('created', 'assigned', 'picked_up', 'in_transit', 'delivered', 'cancelled')),
  created_by            uuid NOT NULL REFERENCES users (id),
  created_at            timestamptz NOT NULL DEFAULT now(),
  updated_at            timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (company_id, assigned_driver_id)
    REFERENCES driver_profiles (company_id, id),
  CHECK ((is_paid AND payment_method IS NOT NULL) OR (NOT is_paid AND payment_method IS NULL))
);

CREATE INDEX deliveries_company_created_idx ON deliveries (company_id, created_at DESC);
CREATE INDEX deliveries_company_status_idx ON deliveries (company_id, status);
CREATE INDEX deliveries_assigned_driver_idx ON deliveries (assigned_driver_id) WHERE assigned_driver_id IS NOT NULL;

CREATE TRIGGER deliveries_set_updated_at
BEFORE UPDATE ON deliveries
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
