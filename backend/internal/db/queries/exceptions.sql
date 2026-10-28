-- name: NextExceptionNumber :one
SELECT nextval('exception_number_seq')::bigint AS next_val;

-- name: CreateException :one
INSERT INTO exceptions (
  id, exception_number, type, severity, title, description,
  manifest_item_id, reported_by, assigned_to, branch_id, status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetExceptionByID :one
SELECT * FROM exceptions
WHERE id = $1 LIMIT 1;

-- name: GetExceptionForUpdate :one
SELECT * FROM exceptions
WHERE id = $1 LIMIT 1
FOR UPDATE;

-- name: ListExceptionsByItem :many
SELECT e.*, COALESCE(b.name, '') AS branch_name,
       COALESCE(reporter.full_name, '') AS reported_by_name,
       COALESCE(assignee.full_name, '') AS assigned_to_name,
       COALESCE(resolver.full_name, '') AS resolved_by_name
FROM exceptions e
LEFT JOIN branches b ON b.id = e.branch_id
LEFT JOIN staff reporter ON reporter.id = e.reported_by
LEFT JOIN staff assignee ON assignee.id = e.assigned_to
LEFT JOIN staff resolver ON resolver.id = e.resolved_by
WHERE e.manifest_item_id = $1
ORDER BY e.created_at DESC;

-- name: UpdateExceptionStatus :one
UPDATE exceptions
SET status = $2,
    resolution_notes = COALESCE(sqlc.narg('resolution_notes'), resolution_notes),
    resolved_by = sqlc.narg('resolved_by'),
    resolved_at = CASE WHEN $2::exception_status IN ('resolved', 'closed') THEN now() ELSE NULL END,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: AssignException :one
UPDATE exceptions
SET assigned_to = sqlc.narg('assigned_to'), updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: CreateExceptionUpdate :one
INSERT INTO exception_updates (
  id, exception_id, updated_by, comment, previous_status, new_status
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: ListExceptionUpdates :many
SELECT eu.*, COALESCE(s.full_name, '') AS updated_by_name
FROM exception_updates eu
LEFT JOIN staff s ON s.id = eu.updated_by
WHERE eu.exception_id = $1
ORDER BY eu.created_at ASC;
