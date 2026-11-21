-- name: ListPublicCompanyBranches :many
SELECT
  id,
  name,
  address,
  city,
  state,
  phone,
  COALESCE(email::text, '')::text AS email,
  operating_hours
FROM branches
WHERE company_id = $1
  AND status = 'active'
  AND deleted_at IS NULL
ORDER BY name
LIMIT 20;

-- name: ListCompanyReviews :many
SELECT id, reviewer_name, rating, comment, created_at
FROM company_reviews
WHERE company_id = $1
ORDER BY created_at DESC
LIMIT 8;

-- name: GetCompanyReviewSummary :one
SELECT
  COUNT(*)::bigint AS review_count,
  COALESCE(ROUND(AVG(rating)::numeric, 1), 0)::float8 AS average_rating
FROM company_reviews
WHERE company_id = $1;

-- name: CreateCompanyReview :one
INSERT INTO company_reviews (company_id, customer_id, reviewer_name, rating, comment)
VALUES ($1, sqlc.narg(customer_id), $2, $3, $4)
RETURNING id, reviewer_name, rating, comment, created_at;
