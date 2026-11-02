ALTER TABLE "manifest_items"
  DROP COLUMN IF EXISTS "delivered_at",
  DROP COLUMN IF EXISTS "delivered_by";

-- Postgres cannot remove a single value from an enum type, so 'delivered'
-- stays on manifest_item_status; no row can reference it once the columns and
-- the status update are gone.
