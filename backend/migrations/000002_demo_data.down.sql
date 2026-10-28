DELETE FROM sync_operations;
DELETE FROM activity_logs;
DELETE FROM exception_updates;
DELETE FROM exceptions;
DELETE FROM payments;
DELETE FROM manifest_items;
DELETE FROM fleet;
DELETE FROM drivers;
DELETE FROM customers;
DELETE FROM staff;
DELETE FROM branches;

SELECT setval('staff_id_seq', 1, false);
SELECT setval('exception_number_seq', 1, false);
