-- The last stage of a record's journey.
--
--   received   the destination branch confirmed the goods reached them
--   delivered  the receiving staff handed the goods over to the customer
--
-- "delivered" is added to the existing enum so the stage is a real status
-- rather than a flag on "received".
ALTER TYPE "manifest_item_status" ADD VALUE IF NOT EXISTS 'delivered';

ALTER TABLE "manifest_items"
  -- Staff member at the destination branch who handed the goods over.
  ADD COLUMN "delivered_by" uuid REFERENCES "staff" ("id") ON DELETE SET NULL,
  ADD COLUMN "delivered_at" timestamptz;
