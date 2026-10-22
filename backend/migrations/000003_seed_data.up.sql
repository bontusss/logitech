BEGIN;

INSERT INTO branches (id, name, code, address, state, phone)
VALUES
('11111111-1111-4111-8111-111111111111', 'Lagos Main Branch', 'LAG-01', '12 Apapa Road', 'Lagos', '08030000001'),
('22222222-2222-4222-8222-222222222222', 'Abuja Branch', 'ABJ-01', '15 Garki Road', 'FCT', '08030000002'),
('33333333-3333-4333-8333-333333333333', 'Ibadan Branch', 'IBA-01', '8 Ring Road', 'Oyo', '08030000003');

INSERT INTO staff (
    id, branch_id, full_name, phone, email, staff_id,
    otp_hash, role, job_role, status, first_login_done
)
VALUES
('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1',
 '11111111-1111-4111-8111-111111111111',
 'Chinedu Okafor', '08031110001', 'owner@galopin.example',
 'GAL-001', 'DEMO_HASH_NOT_FOR_LOGIN', 'owner', NULL, 'active', true),

('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2',
 '11111111-1111-4111-8111-111111111111',
 'Ada Nwosu', '08031110002', 'manager@galopin.example',
 'GAL-002', 'DEMO_HASH_NOT_FOR_LOGIN', 'manager', NULL, 'active', true),

('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 '11111111-1111-4111-8111-111111111111',
 'Tunde Bello', '08031110003', 'staff1@galopin.example',
 'GAL-003', 'DEMO_HASH_NOT_FOR_LOGIN', 'staff', 'operations', 'active', true),

('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4',
 '22222222-2222-4222-8222-222222222222',
 'Mary James', '08031110004', 'staff2@galopin.example',
 'GAL-004', 'DEMO_HASH_NOT_FOR_LOGIN', 'staff', 'receiving', 'active', true);

INSERT INTO customers (id, full_name, phone, email, customer_type)
VALUES
('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1',
 'Emeka Supplies Ltd', '08032220001', 'emeka@example.com', 'business'),
('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2',
 'Amina Yusuf', '08032220002', 'amina@example.com', 'individual'),
('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3',
 'Olu Stores', '08032220003', 'olu@example.com', 'business');

INSERT INTO drivers (
    id, branch_id, full_name, phone, license_number, license_expiry, status
)
VALUES
('cccccccc-cccc-4ccc-8ccc-ccccccccccc1',
 '11111111-1111-4111-8111-111111111111',
 'Musa Ibrahim', '08033330001', 'DEMO-LIC-001', '2027-12-31', 'on_trip'),
('cccccccc-cccc-4ccc-8ccc-ccccccccccc2',
 '11111111-1111-4111-8111-111111111111',
 'Peter Adewale', '08033330002', 'DEMO-LIC-002', '2028-06-30', 'available'),
('cccccccc-cccc-4ccc-8ccc-ccccccccccc3',
 '22222222-2222-4222-8222-222222222222',
 'Ibrahim Sani', '08033330003', 'DEMO-LIC-003', '2027-09-30', 'available');

INSERT INTO fleet (
    id, branch_id, registration_number, vehicle_type, make_model,
    capacity_kg, status
)
VALUES
('dddddddd-dddd-4ddd-8ddd-ddddddddddd1',
 '11111111-1111-4111-8111-111111111111',
 'LAG-123-XY', 'truck', 'Mitsubishi Canter', 3500, 'on_trip'),
('dddddddd-dddd-4ddd-8ddd-ddddddddddd2',
 '11111111-1111-4111-8111-111111111111',
 'LAG-456-XY', 'van', 'Toyota Hiace', 1500, 'available'),
('dddddddd-dddd-4ddd-8ddd-ddddddddddd3',
 '22222222-2222-4222-8222-222222222222',
 'ABJ-789-XY', 'truck', 'Isuzu NPR', 4000, 'available');

-- confirmed_by is required, so these demo manifests have confirmation records.
INSERT INTO manifests (
    id, manifest_number, origin_branch_id, created_by,
    manifest_date, status, confirmed_by, confirmed_at
)
VALUES
('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1',
 'WM-2026-001', '11111111-1111-4111-8111-111111111111',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2',
 '2026-10-20', 'dispatched',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '2026-10-20 12:00:00+01'),

('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee2',
 'WM-2026-002', '11111111-1111-4111-8111-111111111111',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 '2026-10-21', 'confirmed',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '2026-10-21 09:00:00+01'),

('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee3',
 'WM-2026-003', '22222222-2222-4222-8222-222222222222',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4',
 '2026-10-21', 'review_required',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4', '2026-10-21 10:00:00+01');

INSERT INTO manifest_items (
    id, manifest_id, serial_number, description, quantity,
    sender_customer_id, receiver_customer_id, destination_branch_id,
    amount, currency, payment_status, ocr_confidence
)
VALUES
('ffffffff-ffff-4fff-8fff-fffffffffff1',
 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1', '001',
 'Cartons of engine parts', 10,
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1',
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2',
 '22222222-2222-4222-8222-222222222222',
 45000, 'NGN', 'paid', 0.9800),

('ffffffff-ffff-4fff-8fff-fffffffffff2',
 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1', '002',
 'Electrical equipment', 5,
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3',
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2',
 '22222222-2222-4222-8222-222222222222',
 18000, 'NGN', 'pending', 0.9100),

('ffffffff-ffff-4fff-8fff-fffffffffff3',
 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee2', '001',
 'Packaged household goods', 8,
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2',
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3',
 '33333333-3333-4333-8333-333333333333',
 24000, 'NGN', 'partial', 0.9500),

('ffffffff-ffff-4fff-8fff-fffffffffff4',
 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee3', '001',
 'Computer accessories', 4,
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1',
 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3',
 '11111111-1111-4111-8111-111111111111',
 32000, 'NGN', 'pending', 0.6200);

INSERT INTO dispatches (
    id, manifest_id, driver_id, vehicle_id,
    origin_branch_id, destination_branch_id, dispatched_by,
    dispatched_at, status
)
VALUES
('12121212-1212-4121-8121-121212121211',
 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1',
 'cccccccc-cccc-4ccc-8ccc-ccccccccccc1',
 'dddddddd-dddd-4ddd-8ddd-ddddddddddd1',
 '11111111-1111-4111-8111-111111111111',
 '22222222-2222-4222-8222-222222222222',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 '2026-10-21 07:30:00+01', 'dispatched');

INSERT INTO receipts (
    id, manifest_item_id, dispatch_id, received_by,
    received_quantity, condition, received_at, notes
)
VALUES
('13131313-1313-4131-8131-131313131311',
 'ffffffff-ffff-4fff-8fff-fffffffffff1',
 '12121212-1212-4121-8121-121212121211',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4',
 8, 'good', '2026-10-21 15:00:00+01',
 'Eight packages verified; two still unaccounted for');

INSERT INTO exceptions (
    id, exception_number, type, severity, title, description,
    manifest_id, manifest_item_id, dispatch_id, receipt_id,
    reported_by, assigned_to, branch_id, status,
    resolved_by
)
VALUES
('14141414-1414-4141-8141-141414141411',
 'EXC-0001', 'quantity_mismatch', 'high',
 'Two packages missing',
 'Manifest lists 10 packages, but only 8 were received.',
 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1',
 'ffffffff-ffff-4fff-8fff-fffffffffff1',
 '12121212-1212-4121-8121-121212121211',
 '13131313-1313-4131-8131-131313131311',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2',
 '22222222-2222-4222-8222-222222222222',
 'investigating',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2');

INSERT INTO exception_updates (
    id, exception_id, updated_by, comment, previous_status, new_status
)
VALUES
('15151515-1515-4151-8151-151515151511',
 '14141414-1414-4141-8141-141414141411',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2',
 'Contacted dispatch team to investigate the missing packages.',
 'open', 'investigating');

INSERT INTO document_extractions (
    id, manifest_id, image_url, status, extracted_data, attempt_number, completed_at
)
VALUES
('16161616-1616-4161-8161-161616161611',
 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee3',
 'https://example.com/demo/manifest-003.jpg',
 'completed',
 '{"manifest_number":"WM-2026-003","items":[{"description":"Computer accessories","quantity":4}]}',
 1, '2026-10-21 10:00:00+01');

INSERT INTO payments (
    id, manifest_item_id, amount, currency, method,
    reference, recorded_by, paid_at
)
VALUES
('17171717-1717-4171-8171-171717171711',
 'ffffffff-ffff-4fff-8fff-fffffffffff1',
 45000, 'NGN', 'bank_transfer', 'DEMO-PAY-001',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3', '2026-10-20 12:00:00+01');

INSERT INTO activity_logs (id, user_id, entity_type, entity_id, action, details)
VALUES
('18181818-1818-4181-8181-181818181811',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2',
 'manifest', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1',
 'manifest_confirmed', '{"manifest_number":"WM-2026-001"}'),

('19191919-1919-4191-8191-191919191911',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4',
 'receipt', '13131313-1313-4131-8131-131313131311',
 'goods_received', '{"received_quantity":8,"expected_quantity":10}');

INSERT INTO sync_operations (
    id, user_id, operation_type, entity_type, entity_id,
    idempotency_key, status, synced_at
)
VALUES
('20202020-2020-4020-8020-202020202011',
 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
 'create', 'manifest', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee2',
 'demo-sync-manifest-002', 'completed', '2026-10-21 09:05:00+01');

COMMIT;

SELECT setval('staff_id_seq', (SELECT COALESCE(MAX(NULLIF(regexp_replace(staff_id, '\D', '', 'g'), '')::bigint), 0) FROM staff) + 1, false);
