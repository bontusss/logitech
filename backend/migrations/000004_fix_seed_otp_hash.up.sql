-- Fix seeded staff accounts that were inserted with a non-NULL otp_hash placeholder
-- ('DEMO_HASH_NOT_FOR_LOGIN') and first_login_done = true.
--
-- Root cause: the backend evaluates
--   otpPending := !staff.FirstLoginDone || staff.OtpHash.Valid
-- so a non-NULL otp_hash forces every login through the OTP path even when the
-- user has already completed first sign-in, producing the infinite "ask for a
-- new pin" / "That one-time password didn't work" loop.
--
-- Fix: clear otp_hash and set a real bcrypt pin_hash (hash of "demo1234") for
-- the four seed accounts.  Any staff added through the normal flow already has
-- the correct state and is unaffected.

UPDATE staff
SET
    pin_hash  = '$2a$10$MsSIuOMf2c0nTzNGo3ZABOc7bGfU2JsI45.B.sXRdSH7NKyWAgB6W',
    otp_hash  = NULL,
    updated_at = now()
WHERE staff_id IN ('GAL-001', 'GAL-002', 'GAL-003', 'GAL-004')
  AND first_login_done = true
  AND otp_hash = 'DEMO_HASH_NOT_FOR_LOGIN';
