package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/internal/auth"
	"backend/internal/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ExceptionHandler owns the exception (incident) endpoints.
type ExceptionHandler struct {
	db      *sql.DB
	queries *sqlc.Queries
}

// NewExceptionHandler constructs an ExceptionHandler.
func NewExceptionHandler(queries *sqlc.Queries, pool *sql.DB) *ExceptionHandler {
	return &ExceptionHandler{db: pool, queries: queries}
}

// ExceptionUpdateResponse is one comment/status change on an exception.
type ExceptionUpdateResponse struct {
	ID             string    `json:"id"`
	Comment        string    `json:"comment"`
	PreviousStatus string    `json:"previous_status,omitempty"`
	NewStatus      string    `json:"new_status,omitempty"`
	UpdatedBy      string    `json:"updated_by"`
	UpdatedByName  string    `json:"updated_by_name"`
	CreatedAt      time.Time `json:"created_at"`
}

// ExceptionResponse is a raised exception, with its follow-up comments.
type ExceptionResponse struct {
	ID              string                    `json:"id"`
	ExceptionNumber string                    `json:"exception_number"`
	Type            string                    `json:"type"`
	Severity        string                    `json:"severity"`
	Title           string                    `json:"title"`
	Description     string                    `json:"description,omitempty"`
	ManifestItemID  string                    `json:"manifest_item_id"`
	TrackingNumber  string                    `json:"tracking_number"`
	BranchID        string                    `json:"branch_id"`
	BranchName      string                    `json:"branch_name"`
	Status          string                    `json:"status"`
	ReportedBy      string                    `json:"reported_by"`
	ReportedByName  string                    `json:"reported_by_name"`
	AssignedTo      string                    `json:"assigned_to,omitempty"`
	AssignedToName  string                    `json:"assigned_to_name,omitempty"`
	ResolutionNotes string                    `json:"resolution_notes,omitempty"`
	ResolvedBy      string                    `json:"resolved_by,omitempty"`
	ResolvedByName  string                    `json:"resolved_by_name,omitempty"`
	ResolvedAt      *time.Time                `json:"resolved_at,omitempty"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
	Updates         []ExceptionUpdateResponse `json:"updates,omitempty"`
}

// validExceptionTypes bounds the exception type filter.
var validExceptionTypes = map[string]bool{
	"missing_goods":       true,
	"damaged_goods":       true,
	"quantity_mismatch":   true,
	"wrong_goods":         true,
	"wrong_destination":   true,
	"delivery_delay":      true,
	"document_mismatch":   true,
	"payment_discrepancy": true,
	"other":               true,
}

// validExceptionSeverities bounds the severity filter.
var validExceptionSeverities = map[string]bool{
	"low":      true,
	"medium":   true,
	"high":     true,
	"critical": true,
}

// validExceptionStatuses bounds the status filter.
var validExceptionStatuses = map[string]bool{
	"open":          true,
	"investigating": true,
	"resolved":      true,
	"closed":        true,
}

// CreateExceptionRequest is the body of POST /api/records/:id/exceptions.
type CreateExceptionRequest struct {
	Type        string `json:"type" example:"quantity_mismatch"`
	Severity    string `json:"severity,omitempty" example:"high"`
	Title       string `json:"title" example:"Two packages missing"`
	Description string `json:"description,omitempty"`
	AssignedTo  string `json:"assigned_to,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

// CreateException godoc
// @Summary Raise an exception on a record
// @Description The receiver-specific alternative to receiving cleanly: records what went wrong (missing, damaged, quantity mismatch, wrong goods ...) and flags the shipment as "exception". Defaults to the destination branch, or the origin branch when the sender raises it before dispatch.
// @Tags exceptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Record UUID"
// @Param request body CreateExceptionRequest true "Exception details"
// @Success 201 {object} ExceptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/records/{id}/exceptions [post]
func (h *ExceptionHandler) CreateException(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	staffID, claims, ok := claimsStaffID(c)
	if !ok {
		return
	}
	recordID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "record id must be a valid UUID"})
		return
	}
	var req CreateExceptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	exceptionType := strings.ToLower(strings.TrimSpace(req.Type))
	if !validExceptionTypes[exceptionType] {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "type must be one of missing_goods, damaged_goods, quantity_mismatch, wrong_goods, wrong_destination, delivery_delay, document_mismatch, payment_discrepancy, other"})
		return
	}
	severity := strings.ToLower(strings.TrimSpace(req.Severity))
	if severity == "" {
		severity = "medium"
	}
	if !validExceptionSeverities[severity] {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "severity must be low, medium, high or critical"})
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "title is required"})
		return
	}

	ctx := c.Request.Context()
	record, err := h.queries.GetManifestItemByID(ctx, recordID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "record not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the record: " + err.Error()})
		return
	}
	if !canViewRecord(claims, record) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "record not found"})
		return
	}
	if record.Status == sqlc.ManifestItemStatusCancelled {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "a cancelled record cannot raise an exception"})
		return
	}

	assignedTo := uuid.NullUUID{}
	if raw := strings.TrimSpace(req.AssignedTo); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "assigned_to must be a valid staff UUID"})
			return
		}
		if _, err := h.queries.GetStaffByID(ctx, id); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "assigned_to staff not found"})
			return
		}
		assignedTo = uuid.NullUUID{UUID: id, Valid: true}
	}

	branchID := record.DestinationBranchID
	if claims.BranchID == record.OriginBranchID.String() && record.Status == sqlc.ManifestItemStatusPending {
		branchID = record.OriginBranchID
	}

	sequence, err := h.queries.NextExceptionNumber(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to number the exception: " + err.Error()})
		return
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to start transaction: " + err.Error()})
		return
	}
	defer func() { _ = tx.Rollback() }()
	txq := h.queries.WithTx(tx)

	created, err := txq.CreateException(ctx, sqlc.CreateExceptionParams{
		ID:              uuid.New(),
		ExceptionNumber: fmt.Sprintf("EXC-%04d", sequence),
		Type:            sqlc.ExceptionType(exceptionType),
		Severity:        sqlc.ExceptionSeverity(severity),
		Title:           strings.TrimSpace(req.Title),
		Description:     nullIfEmpty(req.Description),
		ManifestItemID:  record.ID,
		ReportedBy:      staffID,
		AssignedTo:      assignedTo,
		BranchID:        branchID,
		Status:          sqlc.ExceptionStatusOpen,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to raise the exception: " + err.Error()})
		return
	}

	if comment := strings.TrimSpace(req.Comment); comment != "" {
		if _, err := txq.CreateExceptionUpdate(ctx, sqlc.CreateExceptionUpdateParams{
			ID:             uuid.New(),
			ExceptionID:    created.ID,
			UpdatedBy:      staffID,
			Comment:        comment,
			PreviousStatus: sqlc.NullExceptionStatus{},
			NewStatus:      sqlc.NullExceptionStatus{ExceptionStatus: sqlc.ExceptionStatusOpen, Valid: true},
		}); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save the exception comment: " + err.Error()})
			return
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE manifest_items
		SET status = 'exception', updated_at = now()
		WHERE id = $1 AND status <> 'exception'`, record.ID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to flag the record: " + err.Error()})
		return
	}

	if err := logActivity(ctx, tx, staffID, "exception", created.ID, "exception_raised", map[string]any{
		"exception_number": created.ExceptionNumber,
		"tracking_number":  record.TrackingNumber,
		"type":             exceptionType,
		"severity":         severity,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to log the exception: " + err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save the exception: " + err.Error()})
		return
	}

	detail, err := h.loadException(ctx, created.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the exception: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, detail)
}

// ListExceptions godoc
// @Summary List exceptions
// @Description Lists exceptions newest first. Owners and the system admin see all branches; managers and staff see exceptions touching their own branch. Filters: status, severity, type, branch_id, from/to (created_at), limit.
// @Tags exceptions
// @Produce json
// @Security BearerAuth
// @Param status query string false "open, investigating, resolved or closed"
// @Param severity query string false "low, medium, high or critical"
// @Param type query string false "Exception type"
// @Param branch_id query string false "Branch UUID"
// @Param from query string false "Created on or after (YYYY-MM-DD)"
// @Param to query string false "Created on or before (YYYY-MM-DD)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Success 200 {array} ExceptionResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/exceptions [get]
func (h *ExceptionHandler) ListExceptions(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	args := make([]any, 0, 8)
	where := " WHERE 1 = 1"

	add := func(condition string, value any) {
		args = append(args, value)
		where += fmt.Sprintf(" "+condition, len(args))
	}

	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		if !validExceptionStatuses[raw] {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "status must be open, investigating, resolved or closed"})
			return
		}
		add("AND e.status::text = $%d", raw)
	}
	if raw := strings.TrimSpace(c.Query("severity")); raw != "" {
		if !validExceptionSeverities[raw] {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "severity must be low, medium, high or critical"})
			return
		}
		add("AND e.severity::text = $%d", raw)
	}
	if raw := strings.TrimSpace(c.Query("type")); raw != "" {
		if !validExceptionTypes[raw] {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "unknown exception type"})
			return
		}
		add("AND e.type::text = $%d", raw)
	}
	if raw := strings.TrimSpace(c.Query("branch_id")); raw != "" {
		if _, err := uuid.Parse(raw); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "branch_id must be a valid UUID"})
			return
		}
		add("AND e.branch_id = $%d::uuid", raw)
	}
	if from, err := parseOptionalDate(c.Query("from")); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "from must be YYYY-MM-DD"})
		return
	} else if from != nil {
		add("AND e.created_at >= $%d", *from)
	}
	if to, err := parseOptionalDate(c.Query("to")); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "to must be YYYY-MM-DD"})
		return
	} else if to != nil {
		add("AND e.created_at < ($%d::date + 1)", *to)
	}

	if !canSeeAllBranches(claims.Role) {
		if claims.BranchID == "" {
			where += " AND FALSE"
		} else {
			args = append(args, claims.BranchID)
			n := len(args)
			where += fmt.Sprintf(" AND (mi.origin_branch_id = $%d::uuid OR mi.destination_branch_id = $%d::uuid OR e.branch_id = $%d::uuid)", n, n, n)
		}
	}

	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "limit must be a positive integer"})
			return
		}
		if parsed > 200 {
			parsed = 200
		}
		limit = parsed
	}
	args = append(args, limit)
	limitPlaceholder := len(args)

	query := `
		SELECT e.id::text, e.exception_number, e.type::text, e.severity::text, e.title,
		       COALESCE(e.description, ''), e.manifest_item_id::text, mi.tracking_number,
		       e.branch_id::text, COALESCE(b.name, ''), e.status::text,
		       e.reported_by::text, COALESCE(reporter.full_name, ''),
		       COALESCE(e.assigned_to::text, ''), COALESCE(assignee.full_name, ''),
		       COALESCE(e.resolution_notes, ''),
		       COALESCE(e.resolved_by::text, ''), COALESCE(resolver.full_name, ''),
		       e.resolved_at, e.created_at, e.updated_at
		FROM exceptions e
		JOIN manifest_items mi ON mi.id = e.manifest_item_id
		LEFT JOIN branches b ON b.id = e.branch_id
		LEFT JOIN staff reporter ON reporter.id = e.reported_by
		LEFT JOIN staff assignee ON assignee.id = e.assigned_to
		LEFT JOIN staff resolver ON resolver.id = e.resolved_by` + where + fmt.Sprintf(`
		ORDER BY e.created_at DESC, e.exception_number DESC
		LIMIT $%d`, limitPlaceholder)

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list exceptions: " + err.Error()})
		return
	}
	defer rows.Close()

	results := make([]ExceptionResponse, 0)
	for rows.Next() {
		var item ExceptionResponse
		var resolvedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.ExceptionNumber, &item.Type, &item.Severity, &item.Title,
			&item.Description, &item.ManifestItemID, &item.TrackingNumber,
			&item.BranchID, &item.BranchName, &item.Status,
			&item.ReportedBy, &item.ReportedByName,
			&item.AssignedTo, &item.AssignedToName,
			&item.ResolutionNotes,
			&item.ResolvedBy, &item.ResolvedByName,
			&resolvedAt, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read exceptions: " + err.Error()})
			return
		}
		if resolvedAt.Valid {
			item.ResolvedAt = &resolvedAt.Time
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read exceptions: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// GetException godoc
// @Summary Get an exception with its history
// @Tags exceptions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Exception UUID"
// @Success 200 {object} ExceptionResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/exceptions/{id} [get]
func (h *ExceptionHandler) GetException(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "exception id must be a valid UUID"})
		return
	}

	item, err := h.loadException(c.Request.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "exception not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the exception: " + err.Error()})
		return
	}
	record, err := h.queries.GetManifestItemByID(c.Request.Context(), uuid.MustParse(item.ManifestItemID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the linked record: " + err.Error()})
		return
	}
	if !canViewRecord(claims, record) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "exception not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// UpdateExceptionRequest is the body of POST /api/exceptions/:id/updates.
type UpdateExceptionRequest struct {
	Status          string `json:"status,omitempty" example:"investigating"`
	Comment         string `json:"comment,omitempty"`
	ResolutionNotes string `json:"resolution_notes,omitempty"`
	AssignedTo      string `json:"assigned_to,omitempty"`
}

// UpdateException godoc
// @Summary Comment on, reassign or resolve an exception
// @Description Adds a comment to the exception history and optionally moves it to another status or reassigns it. Resolving or closing the last open exception on a record returns the record to its delivery status (received, or in transit when it never arrived).
// @Tags exceptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Exception UUID"
// @Param request body UpdateExceptionRequest true "Update"
// @Success 200 {object} ExceptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/exceptions/{id}/updates [post]
func (h *ExceptionHandler) UpdateException(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	staffID, claims, ok := claimsStaffID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "exception id must be a valid UUID"})
		return
	}
	var req UpdateExceptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status != "" && !validExceptionStatuses[status] {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "status must be open, investigating, resolved or closed"})
		return
	}
	comment := strings.TrimSpace(req.Comment)
	if status == "" && comment == "" && strings.TrimSpace(req.ResolutionNotes) == "" && strings.TrimSpace(req.AssignedTo) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "provide a status, comment, assigned_to or resolution_notes"})
		return
	}
	if status == "resolved" && strings.TrimSpace(req.ResolutionNotes) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "resolution_notes are required to resolve an exception"})
		return
	}

	ctx := c.Request.Context()
	exception, err := h.queries.GetExceptionByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "exception not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the exception: " + err.Error()})
		return
	}
	record, err := h.queries.GetManifestItemByID(ctx, exception.ManifestItemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the linked record: " + err.Error()})
		return
	}
	if !canViewRecord(claims, record) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "exception not found"})
		return
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to start transaction: " + err.Error()})
		return
	}
	defer func() { _ = tx.Rollback() }()
	txq := h.queries.WithTx(tx)

	previousStatus := exception.Status
	newStatus := previousStatus
	if status != "" {
		newStatus = sqlc.ExceptionStatus(status)
	}

	if status != "" || strings.TrimSpace(req.ResolutionNotes) != "" {
		params := sqlc.UpdateExceptionStatusParams{
			ID:              exception.ID,
			Status:          newStatus,
			ResolutionNotes: nullIfEmpty(req.ResolutionNotes),
		}
		if newStatus == sqlc.ExceptionStatusResolved || newStatus == sqlc.ExceptionStatusClosed {
			params.ResolvedBy = uuid.NullUUID{UUID: staffID, Valid: true}
		}
		if _, err := txq.UpdateExceptionStatus(ctx, params); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update the exception: " + err.Error()})
			return
		}
	}

	if raw := strings.TrimSpace(req.AssignedTo); raw != "" {
		assignee, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "assigned_to must be a valid staff UUID"})
			return
		}
		if _, err := h.queries.GetStaffByID(ctx, assignee); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "assigned_to staff not found"})
			return
		}
		if _, err := txq.AssignException(ctx, sqlc.AssignExceptionParams{
			ID:         exception.ID,
			AssignedTo: uuid.NullUUID{UUID: assignee, Valid: true},
		}); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to reassign the exception: " + err.Error()})
			return
		}
	}

	if comment != "" {
		if _, err := txq.CreateExceptionUpdate(ctx, sqlc.CreateExceptionUpdateParams{
			ID:             uuid.New(),
			ExceptionID:    exception.ID,
			UpdatedBy:      staffID,
			Comment:        comment,
			PreviousStatus: sqlc.NullExceptionStatus{ExceptionStatus: previousStatus, Valid: true},
			NewStatus:      sqlc.NullExceptionStatus{ExceptionStatus: newStatus, Valid: true},
		}); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save the comment: " + err.Error()})
			return
		}
	}

	// Once nothing is open any more the record can leave the "exception" state.
	if newStatus == sqlc.ExceptionStatusResolved || newStatus == sqlc.ExceptionStatusClosed {
		if _, err := tx.ExecContext(ctx, `
			UPDATE manifest_items mi
			SET status = CASE
			        WHEN mi.received_at IS NOT NULL THEN 'received'::manifest_item_status
			        WHEN mi.dispatched_at IS NOT NULL THEN 'in_transit'::manifest_item_status
			        ELSE 'pending'::manifest_item_status
			    END,
			    updated_at = now()
			WHERE mi.id = $1
			  AND mi.status = 'exception'
			  AND NOT EXISTS (
			        SELECT 1 FROM exceptions e
			        WHERE e.manifest_item_id = mi.id
			          AND e.status IN ('open', 'investigating')
			  )`, record.ID); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update the record: " + err.Error()})
			return
		}
	}

	if err := logActivity(ctx, tx, staffID, "exception", exception.ID, "exception_updated", map[string]any{
		"exception_number": exception.ExceptionNumber,
		"previous_status":  string(previousStatus),
		"new_status":       string(newStatus),
	}); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to log the update: " + err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save the update: " + err.Error()})
		return
	}

	detail, err := h.loadException(ctx, exception.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the exception: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}

// ListRecordExceptions godoc
// @Summary List the exceptions of a record
// @Tags exceptions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Record UUID"
// @Success 200 {array} ExceptionResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/records/{id}/exceptions [get]
func (h *ExceptionHandler) ListRecordExceptions(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}
	recordID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "record id must be a valid UUID"})
		return
	}
	record, err := h.queries.GetManifestItemByID(c.Request.Context(), recordID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "record not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the record: " + err.Error()})
		return
	}
	if !canViewRecord(claims, record) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "record not found"})
		return
	}

	rows, err := h.queries.ListExceptionsByItem(c.Request.Context(), recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list exceptions: " + err.Error()})
		return
	}
	results := make([]ExceptionResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, exceptionResponseFromListRow(row))
	}
	c.JSON(http.StatusOK, results)
}

// exceptionResponseFromListRow maps a joined exception row onto the response.
func exceptionResponseFromListRow(row sqlc.ListExceptionsByItemRow) ExceptionResponse {
	return ExceptionResponse{
		ID:              row.ID.String(),
		ExceptionNumber: row.ExceptionNumber,
		Type:            string(row.Type),
		Severity:        string(row.Severity),
		Title:           row.Title,
		Description:     row.Description.String,
		ManifestItemID:  row.ManifestItemID.String(),
		BranchID:        row.BranchID.String(),
		BranchName:      row.BranchName,
		Status:          string(row.Status),
		ReportedBy:      row.ReportedBy.String(),
		ReportedByName:  row.ReportedByName,
		AssignedTo:      nullUUIDString(row.AssignedTo),
		AssignedToName:  row.AssignedToName,
		ResolutionNotes: row.ResolutionNotes.String,
		ResolvedBy:      nullUUIDString(row.ResolvedBy),
		ResolvedByName:  row.ResolvedByName,
		ResolvedAt:      nullTimePtr(row.ResolvedAt),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

// nullTimePtr converts a nullable timestamp into a pointer.
func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

// loadException returns an exception with its tracking number, names and history.
func (h *ExceptionHandler) loadException(ctx context.Context, id uuid.UUID) (ExceptionResponse, error) {
	exception, err := h.queries.GetExceptionByID(ctx, id)
	if err != nil {
		return ExceptionResponse{}, err
	}

	rows, err := h.queries.ListExceptionsByItem(ctx, exception.ManifestItemID)
	if err != nil {
		return ExceptionResponse{}, err
	}
	var result ExceptionResponse
	found := false
	for _, row := range rows {
		if row.ID == id {
			result = exceptionResponseFromListRow(row)
			found = true
			break
		}
	}
	if !found {
		return ExceptionResponse{}, sql.ErrNoRows
	}

	updates, err := h.queries.ListExceptionUpdates(ctx, id)
	if err != nil {
		return ExceptionResponse{}, err
	}
	result.Updates = make([]ExceptionUpdateResponse, 0, len(updates))
	for _, u := range updates {
		result.Updates = append(result.Updates, ExceptionUpdateResponse{
			ID:             u.ID.String(),
			Comment:        u.Comment,
			PreviousStatus: nullExceptionStatus(u.PreviousStatus),
			NewStatus:      nullExceptionStatus(u.NewStatus),
			UpdatedBy:      u.UpdatedBy.String(),
			UpdatedByName:  u.UpdatedByName,
			CreatedAt:      u.CreatedAt,
		})
	}

	// The tracking number is not part of the exception row returned above.
	var trackingNumber string
	if err := h.db.QueryRowContext(ctx, `SELECT tracking_number FROM manifest_items WHERE id = $1`, exception.ManifestItemID).Scan(&trackingNumber); err != nil {
		return ExceptionResponse{}, err
	}
	result.TrackingNumber = trackingNumber

	return result, nil
}

// nullExceptionStatus renders a nullable enum status as a string.
func nullExceptionStatus(value sqlc.NullExceptionStatus) string {
	if !value.Valid {
		return ""
	}
	return string(value.ExceptionStatus)
}
