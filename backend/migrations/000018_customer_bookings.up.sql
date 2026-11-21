-- Customer self-service bookings: a booking made on a provider's public page
-- can optionally be linked to a customer account. Guests keep their tracking
-- link; registered customers see the booking in their dashboard.

ALTER TABLE deliveries
  ADD COLUMN customer_id uuid REFERENCES users (id) ON DELETE SET NULL;

CREATE INDEX deliveries_customer_created_idx
  ON deliveries (customer_id, created_at DESC)
  WHERE customer_id IS NOT NULL;

-- Every booking-page request gets a tracking code up front so the guest can
-- copy a working tracking URL as soon as the request is submitted.
UPDATE deliveries
SET tracking_code = 'TRK-' || upper(substr(replace(id::text, '-', ''), 1, 12))
WHERE tracking_code IS NULL
  AND booking_source = 'booking_page';
