-- Rollback: restore the original placeholder values used in migration 000003.
-- Note: pin_hash is intentionally NOT restored because it was never set before.
UPDATE staff
SET
    otp_hash   = 'DEMO_HASH_NOT_FOR_LOGIN',
    pin_hash   = NULL,
    updated_at = now()
WHERE staff_id IN ('GAL-001', 'GAL-002', 'GAL-003', 'GAL-004')
  AND first_login_done = true;
