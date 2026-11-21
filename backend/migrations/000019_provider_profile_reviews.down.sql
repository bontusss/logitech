DROP TABLE IF EXISTS company_reviews;

ALTER TABLE companies
  DROP COLUMN IF EXISTS photo_urls,
  DROP COLUMN IF EXISTS year_established,
  DROP COLUMN IF EXISTS linkedin_url,
  DROP COLUMN IF EXISTS twitter_url,
  DROP COLUMN IF EXISTS facebook_url,
  DROP COLUMN IF EXISTS instagram_url,
  DROP COLUMN IF EXISTS website_url;
