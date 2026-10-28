-- =============================================================================
-- Logitech — complete schema + demo data (design reference).
--
-- There are no manifests any more: a "manifest item" is one shipment.
-- 1. A customer brings items in; staff (the sender) record the item, the
--    amount charged and the payment taken. Whatever is unpaid is the balance.
-- 2. The item reaches the destination branch; staff there collect any balance
--    and mark it received, or raise an exception.
-- 3. The owner sees shipments, revenue, payments, daily collections, balances
--    and exceptions across every branch.
--
-- This file mirrors backend/migrations/000001_logitech_schema.up.sql and
-- backend/migrations/000002_demo_data.up.sql; keep the three in sync.
-- =============================================================================

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


-- Demo data.
--
-- Every staff account has PIN "demo1234" (already first-login complete), so the
-- app can be used immediately:
--
--   GAL-001  owner        Chinedu Okafor   Lagos
--   GAL-002  manager      Ada Nwosu        Lagos
--   GAL-003  staff        Tunde Bello      Lagos   (operations)
--   GAL-004  staff        Mary James       Abuja   (receiving)
--   GAL-005  manager      Halima Bello     Abuja
--   GAL-006  staff        Ifeoma Nnaji     Ibadan  (receiving)
--   GAL-007  staff        Suleiman Audu    Abuja   (operations)
--
-- The shipments cover every state and every payment state: paid in full, partly
-- paid with a balance, unpaid, cancelled, in transit, received and flagged with
-- an exception. Dates are relative to now() so the daily reports always show
-- recent activity.
--
-- amount_paid and payment_status are NOT inserted: the payments trigger derives
-- them (and the generated balance column) from the payments below.

BEGIN;

INSERT INTO branches (id, name, code, address, state, phone)
VALUES
('11111111-1111-4111-8111-111111111111', 'Lagos Main Branch', 'LAG-01', '12 Apapa Road', 'Lagos', '08030000001'),
('22222222-2222-4222-8222-222222222222', 'Abuja Branch', 'ABJ-01', '15 Garki Road', 'FCT', '08030000002'),
('33333333-3333-4333-8333-333333333333', 'Ibadan Branch', 'IBA-01', '8 Ring Road', 'Oyo', '08030000003');

-- pin_hash is bcrypt("demo1234", cost=10). otp_hash is NULL and
-- first_login_done is true: these accounts are ready to log in.
INSERT INTO staff (
    id, branch_id, full_name, phone, email, staff_id,
    pin_hash, otp_hash, role, job_role, status, first_login_done, first_login_at
)
VALUES
('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1',
 '11111111-1111-4111-8111-111111111111',
 'Chinedu Okafor', '08031110001', 'owner@galopin.example',
 'GAL-001', '$2a$10$MsSIuOMf2c0nTzNGo3ZABOc7bGfU2JsI45.B.sXRdSH7NKyWAgB6W', NULL,
 'owner', NULL, 'active', true, now() - interval '30 days'),

('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2',
 '11111111-1111-4111-8111-111111111111',
 'Ada Nwosu', '08031110002', 'manager@galopin.example',
 'GAL-002', '$2a$10$MsSIuOMf2c0nTzNGo3ZABOc7bGfU2JsI45.B.sXRdSH7NKyWAgB6W', NULL,
 'manager', NULL, 'active', true, now() - interval '30 days'),

('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 '11111111-1111-4111-8111-111111111111',
 'Tunde Bello', '08031110003', 'staff1@galopin.example',
 'GAL-003', '$2a$10$MsSIuOMf2c0nTzNGo3ZABOc7bGfU2JsI45.B.sXRdSH7NKyWAgB6W', NULL,
 'staff', 'operations', 'active', true, now() - interval '30 days'),

('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4',
 '22222222-2222-4222-8222-222222222222',
 'Mary James', '08031110004', 'staff2@galopin.example',
 'GAL-004', '$2a$10$MsSIuOMf2c0nTzNGo3ZABOc7bGfU2JsI45.B.sXRdSH7NKyWAgB6W', NULL,
 'staff', 'receiving', 'active', true, now() - interval '30 days'),

('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa5',
 '22222222-2222-4222-8222-222222222222',
 'Halima Bello', '08031110005', 'manager2@galopin.example',
 'GAL-005', '$2a$10$MsSIuOMf2c0nTzNGo3ZABOc7bGfU2JsI45.B.sXRdSH7NKyWAgB6W', NULL,
 'manager', NULL, 'active', true, now() - interval '30 days'),

('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa6',
 '33333333-3333-4333-8333-333333333333',
 'Ifeoma Nnaji', '08031110006', 'staff3@galopin.example',
 'GAL-006', '$2a$10$MsSIuOMf2c0nTzNGo3ZABOc7bGfU2JsI45.B.sXRdSH7NKyWAgB6W', NULL,
 'staff', 'receiving', 'active', true, now() - interval '30 days'),

('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa7',
 '22222222-2222-4222-8222-222222222222',
 'Suleiman Audu', '08031110007', 'staff4@galopin.example',
 'GAL-007', '$2a$10$MsSIuOMf2c0nTzNGo3ZABOc7bGfU2JsI45.B.sXRdSH7NKyWAgB6W', NULL,
 'staff', 'operations', 'active', true, now() - interval '30 days');

INSERT INTO customers (id, full_name, phone, email, customer_type)
VALUES
('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', 'Emeka Supplies Ltd', '08032220001', 'emeka@example.com', 'business'),
('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2', 'Amina Yusuf', '08032220002', 'amina@example.com', 'individual'),
('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3', 'Olu Stores', '08032220003', 'olu@example.com', 'business'),
('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb4', 'Ngozi Eze', '08032220004', 'ngozi@example.com', 'individual'),
('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb5', 'Kano Textiles Ltd', '08032220005', 'sales@kanotextiles.example', 'business'),
('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb6', 'Bala Autos', '08032220006', 'bala@example.com', 'business');

INSERT INTO drivers (
    id, branch_id, full_name, phone, license_number, license_expiry, status
)
VALUES
('cccccccc-cccc-4ccc-8ccc-ccccccccccc1', '11111111-1111-4111-8111-111111111111',
 'Musa Ibrahim', '08033330001', 'DEMO-LIC-001', '2027-12-31', 'on_trip'),
('cccccccc-cccc-4ccc-8ccc-ccccccccccc2', '11111111-1111-4111-8111-111111111111',
 'Peter Adewale', '08033330002', 'DEMO-LIC-002', '2028-06-30', 'available'),
('cccccccc-cccc-4ccc-8ccc-ccccccccccc3', '22222222-2222-4222-8222-222222222222',
 'Ibrahim Sani', '08033330003', 'DEMO-LIC-003', '2027-09-30', 'available');

INSERT INTO fleet (
    id, branch_id, registration_number, vehicle_type, make_model, capacity_kg, status
)
VALUES
('dddddddd-dddd-4ddd-8ddd-ddddddddddd1', '11111111-1111-4111-8111-111111111111',
 'LAG-123-XY', 'truck', 'Mitsubishi Canter', 3500, 'on_trip'),
('dddddddd-dddd-4ddd-8ddd-ddddddddddd2', '11111111-1111-4111-8111-111111111111',
 'LAG-456-XY', 'van', 'Toyota Hiace', 1500, 'available'),
('dddddddd-dddd-4ddd-8ddd-ddddddddddd3', '22222222-2222-4222-8222-222222222222',
 'ABJ-789-XY', 'truck', 'Isuzu NPR', 4000, 'available');

INSERT INTO manifest_items (
    id, tracking_number, serial_number, description, quantity,
    sender_customer_id, receiver_customer_id,
    origin_branch_id, destination_branch_id,
    amount, currency, status, recorded_by, notes,
    dispatched_at, received_by, received_at, received_quantity, condition,
    created_at, updated_at
)
VALUES
-- 1. Lagos -> Abuja, received, paid in full over two payments.
('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1', 'LAG-01-10421', '001',
 'Cartons of engine parts', 10,
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2',
 '11111111-1111-4111-8111-111111111111', '22222222-2222-4222-8222-222222222222',
 45000, 'NGN', 'received', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 'Balance collected on delivery',
 now() - interval '5 days', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4', now() - interval '4 days', 10, 'good',
 now() - interval '5 days', now() - interval '4 days'),

-- 2. Lagos -> Abuja, in transit, partly paid, balance outstanding.
('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee2', 'LAG-01-10422', '002',
 'Electrical equipment', 5,
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2',
 '11111111-1111-4111-8111-111111111111', '22222222-2222-4222-8222-222222222222',
 18000, 'NGN', 'in_transit', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 NULL, now() - interval '3 days', NULL, NULL, NULL, NULL,
 now() - interval '3 days', now() - interval '3 days'),

-- 3. Lagos -> Ibadan, received, paid in full (part at the counter, rest on arrival).
('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee3', 'LAG-01-10423', '003',
 'Packaged household goods', 8,
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3',
 '11111111-1111-4111-8111-111111111111', '33333333-3333-4333-8333-333333333333',
 24000, 'NGN', 'received', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 NULL, now() - interval '2 days', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa6', now() - interval '1 day', 8, 'good',
 now() - interval '2 days', now() - interval '1 day'),

-- 4. Abuja -> Lagos, pending, one payment today, balance outstanding.
('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee4', 'ABJ-01-20811', '001',
 'Computer accessories', 4,
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3',
 '22222222-2222-4222-8222-222222222222', '11111111-1111-4111-8111-111111111111',
 32000, 'NGN', 'pending', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa7',
 NULL, NULL, NULL, NULL, NULL, NULL,
 now() - interval '6 hours', now() - interval '6 hours'),

-- 5. Abuja -> Lagos, received short and damaged: exception investigating.
('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee5', 'ABJ-01-20812', '002',
 'Spare tyres', 6,
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb6', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb4',
 '22222222-2222-4222-8222-222222222222', '11111111-1111-4111-8111-111111111111',
 27500, 'NGN', 'exception', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa7',
 'One tyre damaged, one missing', now() - interval '6 days',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3', now() - interval '5 days', 5, 'damaged',
 now() - interval '6 days', now() - interval '5 days'),

-- 6. Ibadan -> Abuja, received, paid in full up front.
('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee6', 'IBA-01-31201', '001',
 'Bags of cocoa', 20,
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb5',
 '33333333-3333-4333-8333-333333333333', '22222222-2222-4222-8222-222222222222',
 52000, 'NGN', 'received', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa6',
 NULL, now() - interval '7 days', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4', now() - interval '6 days', 20, 'good',
 now() - interval '7 days', now() - interval '6 days'),

-- 7. Lagos -> Ibadan, cancelled before dispatch, nothing owed.
('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee7', 'LAG-01-10424', '004',
 'Furniture set', 3,
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb5', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb6',
 '11111111-1111-4111-8111-111111111111', '33333333-3333-4333-8333-333333333333',
 60000, 'NGN', 'cancelled', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2',
 'Customer cancelled before dispatch', NULL, NULL, NULL, NULL, NULL,
 now() - interval '8 days', now() - interval '8 days'),

-- 8. Lagos -> Abuja, pending, paid in full at the counter today.
('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee8', 'LAG-01-10425', '005',
 'Laptop bags', 12,
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb4', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb5',
 '11111111-1111-4111-8111-111111111111', '22222222-2222-4222-8222-222222222222',
 21000, 'NGN', 'pending', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 NULL, NULL, NULL, NULL, NULL, NULL,
 now() - interval '2 hours', now() - interval '2 hours'),

-- 9. Abuja -> Lagos, in transit, partly paid.
('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee9', 'ABJ-01-20813', '003',
 'Boxes of textiles', 15,
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb5', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1',
 '22222222-2222-4222-8222-222222222222', '11111111-1111-4111-8111-111111111111',
 38000, 'NGN', 'in_transit', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa7',
 NULL, now() - interval '2 days', NULL, NULL, NULL, NULL,
 now() - interval '2 days', now() - interval '2 days'),

-- 10. Lagos -> Abuja, received, paid in full before departure.
('eeeeeeee-eeee-4eee-8eee-eeeeeeeeee10', 'LAG-01-10426', '006',
 'Machine spare parts', 18,
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb6', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3',
 '11111111-1111-4111-8111-111111111111', '22222222-2222-4222-8222-222222222222',
 47000, 'NGN', 'received', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 NULL, now() - interval '9 days', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4', now() - interval '8 days', 18, 'good',
 now() - interval '9 days', now() - interval '8 days');

-- Payments. The trigger fills amount_paid/payment_status and the balance.
INSERT INTO payments (
    id, manifest_item_id, amount, currency, method, reference, recorded_by, paid_at, notes
)
VALUES
('17171717-1717-4171-8171-171717171701', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1',
 20000, 'NGN', 'bank_transfer', 'TRF-9001', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 now() - interval '5 days', 'Part payment at the counter'),
('17171717-1717-4171-8171-171717171702', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1',
 25000, 'NGN', 'cash', 'CASH-9002', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4',
 now() - interval '4 days', 'Balance collected on delivery'),

('17171717-1717-4171-8171-171717171703', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee2',
 8000, 'NGN', 'cash', 'CASH-9003', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 now() - interval '3 days', 'Part payment at the counter'),

('17171717-1717-4171-8171-171717171704', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee3',
 10000, 'NGN', 'pos', 'POS-9004', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 now() - interval '2 days', 'Part payment at the counter'),
('17171717-1717-4171-8171-171717171705', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee3',
 14000, 'NGN', 'cash', 'CASH-9005', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa6',
 now() - interval '1 day', 'Balance collected on delivery'),

('17171717-1717-4171-8171-171717171706', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee4',
 10000, 'NGN', 'cash', 'CASH-9006', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa7',
 now() - interval '2 hours', 'Part payment at the counter'),

('17171717-1717-4171-8171-171717171707', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee5',
 15000, 'NGN', 'cash', 'CASH-9007', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa7',
 now() - interval '6 days', 'Part payment at the counter'),

('17171717-1717-4171-8171-171717171708', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee6',
 52000, 'NGN', 'bank_transfer', 'TRF-9008', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa6',
 now() - interval '7 days', 'Paid in full at the counter'),

('17171717-1717-4171-8171-171717171709', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee8',
 21000, 'NGN', 'cash', 'CASH-9009', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 now() - interval '2 hours', 'Paid in full at the counter'),

('17171717-1717-4171-8171-171717171710', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee9',
 12000, 'NGN', 'pos', 'POS-9010', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa7',
 now() - interval '2 days', 'Part payment at the counter'),

('17171717-1717-4171-8171-171717171711', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeee10',
 47000, 'NGN', 'cash', 'CASH-9011', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 now() - interval '9 days', 'Paid in full at the counter');

-- One exception is still being investigated, one was resolved.
INSERT INTO exceptions (
    id, exception_number, type, severity, title, description,
    manifest_item_id, reported_by, assigned_to, branch_id, status,
    resolution_notes, resolved_by, resolved_at, created_at, updated_at
)
VALUES
('14141414-1414-4141-8141-141414141411', 'EXC-0001', 'damaged_goods', 'high',
 'One tyre damaged, one missing',
 'Six tyres were dispatched; five arrived, one of them damaged.',
 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee5', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '11111111-1111-4111-8111-111111111111',
 'investigating', NULL, NULL, NULL,
 now() - interval '5 days', now() - interval '4 days'),

('14141414-1414-4141-8141-141414141412', 'EXC-0002', 'delivery_delay', 'low',
 'Shipment arrived a day late',
 'The vehicle was held up by road works between Lagos and Abuja.',
 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeee10', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa5', '22222222-2222-4222-8222-222222222222',
 'resolved', 'Customer informed; no compensation required.',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa5', now() - interval '7 days',
 now() - interval '8 days', now() - interval '7 days');

INSERT INTO exception_updates (
    id, exception_id, updated_by, comment, previous_status, new_status, created_at
)
VALUES
('15151515-1515-4151-8151-151515151511', '14141414-1414-4141-8141-141414141411',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 'Photos attached; dispatch team asked to trace the missing tyre.',
 'open', 'investigating', now() - interval '4 days'),
('15151515-1515-4151-8151-151515151512', '14141414-1414-4141-8141-141414141412',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa5',
 'Road works cleared; shipment delivered the next morning.',
 'open', 'resolved', now() - interval '7 days');

INSERT INTO activity_logs (id, user_id, entity_type, entity_id, action, details, created_at)
VALUES
('18181818-1818-4181-8181-181818181811', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 'manifest_item', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1', 'record_created',
 '{"tracking_number":"LAG-01-10421","amount":"45000.00","amount_paid":20000}',
 now() - interval '5 days'),
('18181818-1818-4181-8181-181818181812', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4',
 'manifest_item', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1', 'record_received',
 '{"tracking_number":"LAG-01-10421","received_quantity":10,"condition":"good"}',
 now() - interval '4 days'),
('18181818-1818-4181-8181-181818181813', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 'exception', '14141414-1414-4141-8141-141414141411', 'exception_raised',
 '{"exception_number":"EXC-0001","tracking_number":"ABJ-01-20812","type":"damaged_goods","severity":"high"}',
 now() - interval '5 days');

INSERT INTO sync_operations (
    id, user_id, operation_type, entity_type, entity_id,
    idempotency_key, status, synced_at
)
VALUES
('20202020-2020-4020-8020-202020202011', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa7',
 'create', 'manifest_item', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee4',
 'demo-sync-abj-20811', 'completed', now() - interval '6 hours');

COMMIT;

-- Continue numbering after the seeded rows.
SELECT setval('staff_id_seq', 7, true);
SELECT setval('exception_number_seq', 2, true);
