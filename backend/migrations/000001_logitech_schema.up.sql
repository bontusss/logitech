-- Logitech logistics schema.
--
-- There are no manifests any more: a "manifest item" IS a shipment record.
-- 1. A customer brings item(s) to a branch. A staff member (the sender) records
--    the item: who sent it, who receives it, the destination branch, the amount
--    charged and the payment taken at the counter. Whatever is left unpaid is
--    the item's balance.
-- 2. The item travels to the destination branch. A staff member there (the
--    receiver) collects any remaining balance and either marks the item
--    received (with the quantity and condition actually received) or raises an
--    exception.
-- 3. The owner sees total shipments, revenue, payments, daily collections,
--    outstanding balances and exceptions across every branch.

CREATE TYPE "staff_role" AS ENUM (
  'owner',
  'manager',
  'staff'
);

-- Job sub-roles for staff: operations, dispatch, receiving.
-- Managers can perform all of them.
CREATE TYPE "staff_job_role" AS ENUM (
  'operations',
  'dispatch',
  'receiving'
);

CREATE TYPE "record_status" AS ENUM (
  'active',
  'inactive'
);

CREATE TYPE "driver_status" AS ENUM (
  'available',
  'on_trip',
  'inactive'
);

CREATE TYPE "vehicle_status" AS ENUM (
  'available',
  'on_trip',
  'maintenance',
  'inactive'
);

-- Lifecycle of a single shipment (manifest item).
--   pending     recorded at the origin branch, not yet on the road
--   in_transit  handed to transport / left the origin branch
--   received    confirmed at the destination branch
--   exception   received but flagged (missing / damaged / wrong goods ...)
--   cancelled   voided before delivery
CREATE TYPE "manifest_item_status" AS ENUM (
  'pending',
  'in_transit',
  'received',
  'exception',
  'cancelled'
);

CREATE TYPE "exception_type" AS ENUM (
  'missing_goods',
  'damaged_goods',
  'quantity_mismatch',
  'wrong_goods',
  'wrong_destination',
  'delivery_delay',
  'document_mismatch',
  'payment_discrepancy',
  'other'
);

CREATE TYPE "exception_severity" AS ENUM (
  'low',
  'medium',
  'high',
  'critical'
);

CREATE TYPE "exception_status" AS ENUM (
  'open',
  'investigating',
  'resolved',
  'closed'
);

CREATE TYPE "payment_method" AS ENUM (
  'cash',
  'bank_transfer',
  'pos',
  'other'
);

-- Derived from the money paid against an item:
--   pending  nothing paid yet (or the item is free)
--   partial  something paid, a balance remains
--   paid     the amount is settled in full
CREATE TYPE "payment_status" AS ENUM (
  'pending',
  'partial',
  'paid'
);

CREATE TYPE "customer_type" AS ENUM (
  'individual',
  'business'
);

CREATE TYPE "item_condition" AS ENUM (
  'good',
  'damaged',
  'missing'
);

CREATE TYPE "sync_status" AS ENUM (
  'pending',
  'processing',
  'completed',
  'failed'
);

CREATE TABLE "branches" (
  "id" uuid PRIMARY KEY,
  "name" varchar(150) NOT NULL,
  "code" varchar(30) UNIQUE NOT NULL,
  "address" text,
  "state" varchar(100),
  "phone" varchar(30),
  "status" record_status NOT NULL DEFAULT 'active',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "staff" (
  "id" uuid PRIMARY KEY,
  "branch_id" uuid NOT NULL,
  "full_name" varchar(150) NOT NULL,
  "phone" varchar(30),
  "email" varchar(255) UNIQUE,
  -- Auto-generated on creation, format: <STAFF_ID_PREFIX>-NNN (e.g. GAL-001);
  -- prefix is configurable via the STAFF_ID_PREFIX env var.
  "staff_id" varchar(10) UNIQUE NOT NULL,
  -- bcrypt hash of the pending one-time password (issued at creation and on PIN
  -- reset). NULL once consumed, i.e. after the user sets their PIN.
  "otp_hash" varchar(255),
  -- bcrypt hash of the personal PIN chosen by the user on first login (max 10 chars).
  "pin_hash" varchar(255),
  "role" staff_role NOT NULL,
  -- Only meaningful when role = 'staff'.
  "job_role" staff_job_role,
  "status" record_status NOT NULL DEFAULT 'active',
  "first_login_done" boolean NOT NULL DEFAULT false,
  "first_login_at" timestamptz,
  "failed_login_attempts" int NOT NULL DEFAULT 0,
  -- Set after 3 failed login attempts; cleared by the branch manager or the owner.
  "locked_at" timestamptz,
  -- Staff are created by their branch manager; managers by the owner.
  "created_by" uuid,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "staff_job_role_required" CHECK ("role" <> 'staff' OR "job_role" IS NOT NULL)
);

CREATE TABLE "customers" (
  "id" uuid PRIMARY KEY,
  "full_name" varchar(150) NOT NULL,
  "phone" varchar(30),
  "email" varchar(255),
  "address" text,
  "customer_type" customer_type NOT NULL DEFAULT 'individual',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "drivers" (
  "id" uuid PRIMARY KEY,
  "branch_id" uuid NOT NULL,
  "full_name" varchar(150) NOT NULL,
  "phone" varchar(30) NOT NULL,
  "license_number" varchar(100),
  "license_expiry" date,
  "status" driver_status NOT NULL DEFAULT 'available',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "fleet" (
  "id" uuid PRIMARY KEY,
  "branch_id" uuid NOT NULL,
  "registration_number" varchar(50) UNIQUE NOT NULL,
  "vehicle_type" varchar(50) NOT NULL,
  "make_model" varchar(100),
  "capacity_kg" decimal(10,2),
  "status" vehicle_status NOT NULL DEFAULT 'available',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

-- One row per shipment. "balance" is always amount - amount_paid and is kept
-- in sync by the payments trigger (see refresh_manifest_item_balance below).
CREATE TABLE "manifest_items" (
  "id" uuid PRIMARY KEY,
  -- Public identifier printed on the waybill, unique per branch and worldwide
  -- (e.g. LAG-01-48213). Generated by the API when not supplied.
  "tracking_number" varchar(30) UNIQUE NOT NULL,
  "serial_number" varchar(50),
  "description" text NOT NULL,
  "quantity" int NOT NULL DEFAULT 1,
  "sender_customer_id" uuid NOT NULL,
  "receiver_customer_id" uuid NOT NULL,
  "origin_branch_id" uuid NOT NULL,
  "destination_branch_id" uuid NOT NULL,
  "amount" decimal(12,2) NOT NULL DEFAULT 0,
  -- Sum of the item's payments; maintained by the payments trigger.
  "amount_paid" decimal(12,2) NOT NULL DEFAULT 0,
  -- Amount still owed on this shipment.
  "balance" decimal(12,2) GENERATED ALWAYS AS ("amount" - "amount_paid") STORED,
  "currency" char(3) NOT NULL DEFAULT 'NGN',
  -- Derived from the money columns so it can never drift from the balance.
  "payment_status" payment_status GENERATED ALWAYS AS (
    CASE
      WHEN "amount" <= 0 OR "amount_paid" >= "amount" THEN 'paid'::payment_status
      WHEN "amount_paid" > 0 THEN 'partial'::payment_status
      ELSE 'pending'::payment_status
    END
  ) STORED,
  "status" manifest_item_status NOT NULL DEFAULT 'pending',
  -- Staff member (sender) who recorded the shipment.
  "recorded_by" uuid NOT NULL,
  "photo_url" text,
  "notes" text,
  "dispatched_at" timestamptz,
  -- Staff member (receiver) at the destination branch.
  "received_by" uuid,
  "received_at" timestamptz,
  "received_quantity" int,
  "condition" item_condition,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "manifest_items_quantity_positive" CHECK ("quantity" > 0),
  CONSTRAINT "manifest_items_amount_positive" CHECK ("amount" >= 0),
  CONSTRAINT "manifest_items_paid_range" CHECK ("amount_paid" >= 0 AND "amount_paid" <= "amount"),
  CONSTRAINT "manifest_items_received_quantity_positive" CHECK ("received_quantity" IS NULL OR "received_quantity" >= 0)
);

-- Every payment made against one shipment: the counter payment taken when the
-- item is recorded and the balance collected at the destination branch.
CREATE TABLE "payments" (
  "id" uuid PRIMARY KEY,
  "manifest_item_id" uuid NOT NULL,
  "amount" decimal(12,2) NOT NULL,
  "currency" char(3) NOT NULL DEFAULT 'NGN',
  "method" payment_method NOT NULL,
  "reference" varchar(150),
  "recorded_by" uuid NOT NULL,
  "paid_at" timestamptz NOT NULL DEFAULT (now()),
  "notes" text,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "payments_amount_positive" CHECK ("amount" > 0)
);

CREATE TABLE "exceptions" (
  "id" uuid PRIMARY KEY,
  "exception_number" varchar(30) UNIQUE NOT NULL,
  "type" exception_type NOT NULL,
  "severity" exception_severity NOT NULL DEFAULT 'medium',
  "title" varchar(200) NOT NULL,
  "description" text,
  "manifest_item_id" uuid NOT NULL,
  "reported_by" uuid NOT NULL,
  "assigned_to" uuid,
  "branch_id" uuid NOT NULL,
  "status" exception_status NOT NULL DEFAULT 'open',
  "resolution_notes" text,
  "resolved_by" uuid,
  "resolved_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "exception_updates" (
  "id" uuid PRIMARY KEY,
  "exception_id" uuid NOT NULL,
  "updated_by" uuid NOT NULL,
  "comment" text NOT NULL,
  "previous_status" exception_status,
  "new_status" exception_status,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "activity_logs" (
  "id" uuid PRIMARY KEY,
  "user_id" uuid NOT NULL,
  "entity_type" varchar(50) NOT NULL,
  "entity_id" uuid NOT NULL,
  "action" varchar(100) NOT NULL,
  "details" jsonb,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sync_operations" (
  "id" uuid PRIMARY KEY,
  "user_id" uuid NOT NULL,
  "operation_type" varchar(50) NOT NULL,
  "entity_type" varchar(50) NOT NULL,
  "entity_id" uuid,
  "idempotency_key" varchar(150) UNIQUE NOT NULL,
  "status" sync_status NOT NULL DEFAULT 'pending',
  "error_message" text,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "synced_at" timestamptz
);

CREATE UNIQUE INDEX ON "drivers" ("license_number");

-- Customers are matched by phone when a shipment is recorded, so one row per
-- phone number is enforced (rows without a phone are unaffected).
CREATE UNIQUE INDEX "customers_phone_unique" ON "customers" ("phone") WHERE "phone" IS NOT NULL;

CREATE INDEX ON "manifest_items" ("origin_branch_id");

CREATE INDEX ON "manifest_items" ("destination_branch_id");

CREATE INDEX ON "manifest_items" ("status");

CREATE INDEX ON "manifest_items" ("payment_status");

CREATE INDEX ON "manifest_items" ("created_at");

CREATE INDEX ON "manifest_items" ("received_at");

CREATE INDEX ON "manifest_items" ("sender_customer_id");

CREATE INDEX ON "manifest_items" ("receiver_customer_id");

CREATE INDEX ON "payments" ("manifest_item_id");

CREATE INDEX ON "payments" ("paid_at");

CREATE INDEX ON "exceptions" ("manifest_item_id");

CREATE INDEX ON "exceptions" ("status");

CREATE INDEX ON "exceptions" ("branch_id");

CREATE INDEX ON "exceptions" ("type");

CREATE INDEX ON "exceptions" ("severity");

CREATE INDEX ON "exceptions" ("created_at");

CREATE INDEX ON "activity_logs" ("entity_type", "entity_id");

CREATE INDEX ON "activity_logs" ("user_id");

CREATE INDEX ON "activity_logs" ("created_at");

COMMENT ON COLUMN "staff"."otp_hash" IS 'One-time password issued at creation / on PIN reset; NULL once consumed';
COMMENT ON COLUMN "staff"."pin_hash" IS 'Personal PIN hash, chosen by the user on first login (max 10 chars)';
COMMENT ON COLUMN "staff"."job_role" IS 'operations | dispatch | receiving; required when role = staff';
COMMENT ON COLUMN "staff"."locked_at" IS 'Set after 3 failed login attempts';
COMMENT ON COLUMN "manifest_items"."tracking_number" IS 'Waybill number printed for the customer, unique across branches (e.g. LAG-01-48213)';
COMMENT ON COLUMN "manifest_items"."balance" IS 'Generated: amount - amount_paid';
COMMENT ON COLUMN "manifest_items"."recorded_by" IS 'Staff (sender) who recorded the shipment at the origin branch';
COMMENT ON COLUMN "manifest_items"."received_by" IS 'Staff (receiver) who confirmed the shipment at the destination branch';

ALTER TABLE "staff" ADD FOREIGN KEY ("branch_id") REFERENCES "branches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "staff" ADD FOREIGN KEY ("created_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "drivers" ADD FOREIGN KEY ("branch_id") REFERENCES "branches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "fleet" ADD FOREIGN KEY ("branch_id") REFERENCES "branches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "manifest_items" ADD FOREIGN KEY ("sender_customer_id") REFERENCES "customers" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "manifest_items" ADD FOREIGN KEY ("receiver_customer_id") REFERENCES "customers" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "manifest_items" ADD FOREIGN KEY ("origin_branch_id") REFERENCES "branches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "manifest_items" ADD FOREIGN KEY ("destination_branch_id") REFERENCES "branches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "manifest_items" ADD FOREIGN KEY ("recorded_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "manifest_items" ADD FOREIGN KEY ("received_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "payments" ADD FOREIGN KEY ("manifest_item_id") REFERENCES "manifest_items" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "payments" ADD FOREIGN KEY ("recorded_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exceptions" ADD FOREIGN KEY ("manifest_item_id") REFERENCES "manifest_items" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exceptions" ADD FOREIGN KEY ("reported_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exceptions" ADD FOREIGN KEY ("assigned_to") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exceptions" ADD FOREIGN KEY ("branch_id") REFERENCES "branches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exceptions" ADD FOREIGN KEY ("resolved_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exception_updates" ADD FOREIGN KEY ("exception_id") REFERENCES "exceptions" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exception_updates" ADD FOREIGN KEY ("updated_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "activity_logs" ADD FOREIGN KEY ("user_id") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "sync_operations" ADD FOREIGN KEY ("user_id") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

-- Money bookkeeping ---------------------------------------------------------
-- The amount paid, the payment status and the balance are all derived: the
-- balance and status are generated columns, and this trigger keeps amount_paid
-- in step with the payments taken against the shipment.

CREATE OR REPLACE FUNCTION refresh_manifest_item_balance() RETURNS trigger AS $$
DECLARE
  item_ids uuid[];
  item_id  uuid;
  paid     numeric(12,2);
BEGIN
  IF TG_OP = 'DELETE' THEN
    item_ids := ARRAY[OLD.manifest_item_id];
  ELSIF TG_OP = 'UPDATE' AND OLD.manifest_item_id <> NEW.manifest_item_id THEN
    item_ids := ARRAY[OLD.manifest_item_id, NEW.manifest_item_id];
  ELSE
    item_ids := ARRAY[NEW.manifest_item_id];
  END IF;

  FOREACH item_id IN ARRAY item_ids LOOP
    SELECT COALESCE(SUM(p.amount), 0) INTO paid
    FROM payments p
    WHERE p.manifest_item_id = item_id;

    UPDATE manifest_items mi
    SET amount_paid = paid,
        updated_at = now()
    WHERE mi.id = item_id;
  END LOOP;

  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER payments_refresh_balance
AFTER INSERT OR UPDATE OR DELETE ON payments
FOR EACH ROW EXECUTE FUNCTION refresh_manifest_item_balance();

-- Numbering sequences -------------------------------------------------------

CREATE SEQUENCE IF NOT EXISTS staff_id_seq START WITH 1;

CREATE SEQUENCE IF NOT EXISTS exception_number_seq START WITH 1;
