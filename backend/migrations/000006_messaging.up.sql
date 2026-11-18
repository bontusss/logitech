CREATE TABLE email_jobs (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  company_id    uuid REFERENCES companies (id) ON DELETE CASCADE,
  kind          text NOT NULL CHECK (kind IN ('welcome_customer', 'welcome_provider', 'provider_onboarding')),
  sequence      smallint NOT NULL DEFAULT 0,
  status        text NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'sending', 'sent', 'cancelled', 'failed')),
  scheduled_at  timestamptz NOT NULL,
  claimed_at    timestamptz,
  sent_at       timestamptz,
  attempts      smallint NOT NULL DEFAULT 0,
  last_error    text,
  created_at    timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, kind, sequence)
);

CREATE INDEX email_jobs_ready_idx ON email_jobs (scheduled_at)
  WHERE status IN ('pending', 'sending');

CREATE TABLE user_phone_verifications (
  user_id          uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
  phone            text NOT NULL,
  token_hash       text NOT NULL,
  expires_at       timestamptz NOT NULL,
  verified_at      timestamptz,
  last_sent_at     timestamptz NOT NULL,
  window_started_at timestamptz NOT NULL,
  sends_in_window  smallint NOT NULL DEFAULT 1,
  attempts         smallint NOT NULL DEFAULT 0,
  created_at       timestamptz NOT NULL DEFAULT now()
);
