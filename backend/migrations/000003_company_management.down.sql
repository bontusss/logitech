DROP TABLE IF EXISTS company_subdomains;
DROP TABLE IF EXISTS invitation_branches;
DROP TABLE IF EXISTS branch_members;
DROP TABLE IF EXISTS branches;

ALTER TABLE company_invitations DROP COLUMN IF EXISTS accepted_by_user_id;

ALTER TABLE company_branding
  DROP COLUMN IF EXISTS cover_image_url,
  DROP COLUMN IF EXISTS public_name,
  DROP COLUMN IF EXISTS description,
  DROP COLUMN IF EXISTS pwa_short_name;

ALTER TABLE companies
  DROP COLUMN IF EXISTS description,
  DROP COLUMN IF EXISTS cover_image_url,
  DROP COLUMN IF EXISTS operating_hours;
