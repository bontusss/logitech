CREATE TABLE driver_locations (
  driver_id       uuid PRIMARY KEY,
  company_id      uuid NOT NULL,
  latitude        double precision NOT NULL CHECK (latitude BETWEEN -90 AND 90),
  longitude       double precision NOT NULL CHECK (longitude BETWEEN -180 AND 180),
  accuracy_m      double precision CHECK (accuracy_m IS NULL OR accuracy_m >= 0),
  heading         double precision CHECK (heading IS NULL OR (heading >= 0 AND heading < 360)),
  speed_kph       double precision CHECK (speed_kph IS NULL OR speed_kph >= 0),
  is_broadcasting boolean NOT NULL DEFAULT true,
  updated_at      timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (company_id, driver_id)
    REFERENCES driver_profiles (company_id, id) ON DELETE CASCADE
);

CREATE INDEX driver_locations_company_updated_idx
  ON driver_locations (company_id, updated_at DESC);
