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

CREATE TYPE "manifest_status" AS ENUM (
  'draft',
  'extraction_pending',
  'review_required',
  'confirmed',
  'awaiting_dispatch',
  'dispatched',
  'partially_received',
  'received',
  'cancelled'
);

CREATE TYPE "dispatch_status" AS ENUM (
  'pending',
  'dispatched',
  'arrived',
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

CREATE TYPE "extraction_status" AS ENUM (
  'pending',
  'processing',
  'completed',
  'failed'
);

CREATE TYPE "payment_method" AS ENUM (
  'cash',
  'bank_transfer',
  'pos',
  'other'
);

CREATE TYPE "payment_status" AS ENUM (
  'pending',
  'partial',
  'paid',
  'failed',
  'refunded'
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

CREATE TABLE "manifests" (
  "id" uuid PRIMARY KEY,
  "manifest_number" varchar(100) NOT NULL,
  "origin_branch_id" uuid NOT NULL,
  "created_by" uuid NOT NULL,
  "manifest_date" date NOT NULL,
  "paper_image_url" text,
  "status" manifest_status NOT NULL DEFAULT 'draft',
  "version" int NOT NULL DEFAULT 1,
  "confirmed_by" uuid NOT NULL,
  "confirmed_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "manifest_items" (
  "id" uuid PRIMARY KEY,
  "manifest_id" uuid NOT NULL,
  "serial_number" varchar(50),
  "description" text NOT NULL,
  "quantity" int NOT NULL DEFAULT 1,
  "sender_customer_id" uuid NOT NULL,
  "receiver_customer_id" uuid NOT NULL,
  "destination_branch_id" uuid NOT NULL,
  "amount" decimal(12,2),
  "currency" char(3) NOT NULL DEFAULT 'NGN',
  "payment_status" payment_status NOT NULL DEFAULT 'pending',
  "ocr_confidence" decimal(5,4),
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "dispatches" (
  "id" uuid PRIMARY KEY,
  "manifest_id" uuid NOT NULL,
  "driver_id" uuid NOT NULL,
  "vehicle_id" uuid NOT NULL,
  "origin_branch_id" uuid NOT NULL,
  "destination_branch_id" uuid NOT NULL,
  "dispatched_by" uuid NOT NULL,
  "dispatched_at" timestamptz,
  "arrived_at" timestamptz,
  "status" dispatch_status NOT NULL DEFAULT 'pending',
  "notes" text,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "receipts" (
  "id" uuid PRIMARY KEY,
  "manifest_item_id" uuid NOT NULL,
  "dispatch_id" uuid,
  "received_by" uuid NOT NULL,
  "received_quantity" int NOT NULL DEFAULT 0,
  "condition" item_condition NOT NULL,
  "received_at" timestamptz NOT NULL DEFAULT (now()),
  "notes" text,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "exceptions" (
  "id" uuid PRIMARY KEY,
  "exception_number" varchar(30) UNIQUE NOT NULL,
  "type" exception_type NOT NULL,
  "severity" exception_severity NOT NULL DEFAULT 'medium',
  "title" varchar(200) NOT NULL,
  "description" text,
  "manifest_id" uuid NOT NULL,
  "manifest_item_id" uuid NOT NULL,
  "dispatch_id" uuid NOT NULL,
  "receipt_id" uuid,
  "reported_by" uuid NOT NULL,
  "assigned_to" uuid,
  "branch_id" uuid NOT NULL,
  "status" exception_status NOT NULL DEFAULT 'open',
  "resolution_notes" text,
  "resolved_by" uuid NOT NULL,
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

CREATE TABLE "document_extractions" (
  "id" uuid PRIMARY KEY,
  "manifest_id" uuid NOT NULL,
  "image_url" text NOT NULL,
  "status" extraction_status NOT NULL DEFAULT 'pending',
  "extracted_data" jsonb,
  "error_message" text,
  "attempt_number" int NOT NULL DEFAULT 1,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "completed_at" timestamptz
);

CREATE TABLE "payments" (
  "id" uuid PRIMARY KEY,
  "manifest_item_id" uuid NOT NULL,
  "amount" decimal(12,2) NOT NULL,
  "currency" char(3) NOT NULL DEFAULT 'NGN',
  "method" payment_method NOT NULL,
  "reference" varchar(150),
  "recorded_by" uuid NOT NULL,
  "paid_at" timestamptz NOT NULL DEFAULT (now()),
  "notes" text
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

CREATE UNIQUE INDEX ON "manifests" ("origin_branch_id", "manifest_number");

CREATE INDEX ON "manifest_items" ("manifest_id");

CREATE INDEX ON "manifest_items" ("destination_branch_id");

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

ALTER TABLE "staff" ADD FOREIGN KEY ("branch_id") REFERENCES "branches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "staff" ADD FOREIGN KEY ("created_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "drivers" ADD FOREIGN KEY ("branch_id") REFERENCES "branches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "fleet" ADD FOREIGN KEY ("branch_id") REFERENCES "branches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "manifests" ADD FOREIGN KEY ("origin_branch_id") REFERENCES "branches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "manifests" ADD FOREIGN KEY ("created_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "manifests" ADD FOREIGN KEY ("confirmed_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "manifest_items" ADD FOREIGN KEY ("manifest_id") REFERENCES "manifests" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "manifest_items" ADD FOREIGN KEY ("sender_customer_id") REFERENCES "customers" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "manifest_items" ADD FOREIGN KEY ("receiver_customer_id") REFERENCES "customers" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "manifest_items" ADD FOREIGN KEY ("destination_branch_id") REFERENCES "branches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "dispatches" ADD FOREIGN KEY ("manifest_id") REFERENCES "manifests" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "dispatches" ADD FOREIGN KEY ("driver_id") REFERENCES "drivers" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "dispatches" ADD FOREIGN KEY ("vehicle_id") REFERENCES "fleet" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "dispatches" ADD FOREIGN KEY ("origin_branch_id") REFERENCES "branches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "dispatches" ADD FOREIGN KEY ("destination_branch_id") REFERENCES "branches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "dispatches" ADD FOREIGN KEY ("dispatched_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "receipts" ADD FOREIGN KEY ("manifest_item_id") REFERENCES "manifest_items" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "receipts" ADD FOREIGN KEY ("dispatch_id") REFERENCES "dispatches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "receipts" ADD FOREIGN KEY ("received_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exceptions" ADD FOREIGN KEY ("manifest_id") REFERENCES "manifests" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exceptions" ADD FOREIGN KEY ("manifest_item_id") REFERENCES "manifest_items" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exceptions" ADD FOREIGN KEY ("dispatch_id") REFERENCES "dispatches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exceptions" ADD FOREIGN KEY ("receipt_id") REFERENCES "receipts" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exceptions" ADD FOREIGN KEY ("reported_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exceptions" ADD FOREIGN KEY ("assigned_to") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exceptions" ADD FOREIGN KEY ("branch_id") REFERENCES "branches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exceptions" ADD FOREIGN KEY ("resolved_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exception_updates" ADD FOREIGN KEY ("exception_id") REFERENCES "exceptions" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "exception_updates" ADD FOREIGN KEY ("updated_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "document_extractions" ADD FOREIGN KEY ("manifest_id") REFERENCES "manifests" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "payments" ADD FOREIGN KEY ("manifest_item_id") REFERENCES "manifest_items" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "payments" ADD FOREIGN KEY ("recorded_by") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "activity_logs" ADD FOREIGN KEY ("user_id") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "sync_operations" ADD FOREIGN KEY ("user_id") REFERENCES "staff" ("id") DEFERRABLE INITIALLY IMMEDIATE;


CREATE SEQUENCE IF NOT EXISTS staff_id_seq START WITH 1;
