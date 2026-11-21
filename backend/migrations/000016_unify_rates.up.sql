-- Collapse the two parallel pricing stacks onto the operational one.
--
-- Providers configure intercity rates on company_booking_routes (per-km prices
-- per delivery speed, minimum charge, transit/estimate/pickup metadata) and
-- last-mile rates on last_mile_service_areas/service_zones/pricing_zone_rates.
-- The legacy Task 3 scaffolding (shipping_services, company_routes, rate_cards,
-- pricing_rules, shipping_quotes) was never wired to any UI, public flow or
-- test, so it is removed rather than kept as a second source of truth.

ALTER TABLE deliveries
  DROP COLUMN IF EXISTS selected_rate_card_id;

ALTER TABLE company_booking_routes
  DROP COLUMN IF EXISTS transportation_method,
  DROP COLUMN IF EXISTS pricing_model,
  DROP COLUMN IF EXISTS origin_branch_id,
  DROP COLUMN IF EXISTS destination_branch_id;

DROP TABLE IF EXISTS pricing_weight_brackets;
DROP TABLE IF EXISTS pricing_size_categories;
DROP TABLE IF EXISTS pricing_quote_versions;

DROP TABLE IF EXISTS quote_charges;
DROP TABLE IF EXISTS shipping_quotes;
DROP TABLE IF EXISTS rate_card_history;
DROP TABLE IF EXISTS additional_charges;
DROP TABLE IF EXISTS pricing_rules;
DROP TABLE IF EXISTS pricing_tiers;
DROP TABLE IF EXISTS rate_card_versions;
DROP TABLE IF EXISTS rate_cards;
DROP TABLE IF EXISTS package_categories;
DROP TABLE IF EXISTS company_routes;
DROP TABLE IF EXISTS shipping_services;
DROP TABLE IF EXISTS company_service_areas;
DROP FUNCTION IF EXISTS handle_rate_card_version_updated_at();
