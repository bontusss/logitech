ALTER TABLE deliveries
  ADD COLUMN request_status text NOT NULL DEFAULT 'accepted'
    CHECK (request_status IN ('pending', 'accepted', 'declined')),
  ADD COLUMN tracking_code text,
  ADD COLUMN rider_link_token text,
  ADD COLUMN declined_at timestamptz,
  ADD COLUMN accepted_at timestamptz;

UPDATE deliveries
SET request_status = CASE WHEN status = 'requested' THEN 'pending' ELSE 'accepted' END,
    tracking_code = CASE
      WHEN status <> 'requested' THEN 'TRK-' || upper(substr(replace(id::text, '-', ''), 1, 12))
      ELSE NULL
    END;

CREATE UNIQUE INDEX deliveries_tracking_code_unique
  ON deliveries (tracking_code) WHERE tracking_code IS NOT NULL;
CREATE INDEX deliveries_company_request_status_idx
  ON deliveries (company_id, request_status, created_at DESC);

CREATE TABLE delivery_timeline_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  delivery_id uuid NOT NULL REFERENCES deliveries(id) ON DELETE CASCADE,
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  status text NOT NULL,
  note text,
  created_by uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX delivery_timeline_delivery_created_idx
  ON delivery_timeline_events(delivery_id, created_at DESC);

INSERT INTO delivery_timeline_events (delivery_id, company_id, status, note)
SELECT id, company_id, status, 'Status initialized'
FROM deliveries;
