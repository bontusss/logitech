-- The sender and the receiver of a shipment are optional: when a parcel is
-- dropped off nobody may know yet who is collecting it. The party is then left
-- empty instead of being filled with a placeholder customer.
ALTER TABLE "manifest_items" ALTER COLUMN "sender_customer_id" DROP NOT NULL;

ALTER TABLE "manifest_items" ALTER COLUMN "receiver_customer_id" DROP NOT NULL;
