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
