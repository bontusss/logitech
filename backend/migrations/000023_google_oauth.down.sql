DROP INDEX IF EXISTS users_google_subject_unique;

ALTER TABLE users
  DROP COLUMN IF EXISTS google_subject;
