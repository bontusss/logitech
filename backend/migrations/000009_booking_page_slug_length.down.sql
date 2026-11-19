ALTER TABLE booking_pages
  DROP CONSTRAINT booking_pages_slug_check,
  ADD CONSTRAINT booking_pages_slug_check
    CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,38}[a-z0-9]$');
