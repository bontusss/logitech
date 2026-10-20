CREATE TABLE IF NOT EXISTS health_checks (
    id         bigserial PRIMARY KEY,
    checked_at timestamptz NOT NULL DEFAULT now(),
    note       text
);
