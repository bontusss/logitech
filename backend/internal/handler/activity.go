package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ActivityHandler serves the audit trail every record, payment and exception
// change writes: who acted, on which entity, and when.
type ActivityHandler struct {
	db *sql.DB
}

// NewActivityHandler constructs an ActivityHandler.
func NewActivityHandler(pool *sql.DB) *ActivityHandler {
	return &ActivityHandler{db: pool}
}

// ActivityLog is one audit entry.
type ActivityLog struct {
	ID         string          `json:"id"`
	UserID     string          `json:"user_id"`
	UserName   string          `json:"user_name"`
	BranchID   string          `json:"branch_id,omitempty"`
	BranchName string          `json:"branch_name,omitempty"`
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Action     string          `json:"action"`
	Details    json.RawMessage `json:"details,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// List godoc
// @Summary Activity log
// @Description The audit trail, newest first: who recorded, edited, dispatched, received, delivered or cancelled a shipment, took a payment, or raised and updated an exception. The owner and the system admin see every branch; a branch manager sees only the actions of their own branch's staff.
// @Tags activity
// @Produce json
// @Security BearerAuth
// @Param branch_id query string false "Branch UUID (owner / system admin only)"
// @Param entity_type query string false "manifest_item or exception"
// @Param action query string false "e.g. record_created, payment_recorded, exception_raised"
// @Param limit query int false "Page size (default 100, max 200)"
// @Param offset query int false "Offset"
// @Success 200 {array} ActivityLog
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/activity [get]
func (h *ActivityHandler) List(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}
	if !canSeeAllBranches(claims.Role) && claims.Role != auth.RoleManager {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "activity is available to managers, the owner and the system administrator"})
		return
	}

	// The owner and the system admin may narrow to one branch; a manager is
	// always pinned to their own branch, whatever the query says.
	var branchFilter any
	if canSeeAllBranches(claims.Role) {
		if raw := strings.TrimSpace(c.Query("branch_id")); raw != "" {
			if _, err := uuid.Parse(raw); err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "branch_id must be a valid UUID"})
				return
			}
			branchFilter = raw
		}
	} else {
		if claims.BranchID == "" {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "your account is not attached to a branch"})
			return
		}
		branchFilter = claims.BranchID
	}

	entityType := strings.TrimSpace(c.Query("entity_type"))
	action := strings.TrimSpace(c.Query("action"))

	limit := 100
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
	offset := 0
	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "offset must be zero or positive"})
			return
		}
		offset = parsed
	}

	rows, err := h.db.QueryContext(c.Request.Context(), `
		SELECT al.id::text, al.user_id::text, COALESCE(s.full_name, ''),
		       COALESCE(s.branch_id::text, ''), COALESCE(b.name, ''),
		       al.entity_type, al.entity_id::text, al.action, al.details, al.created_at
		FROM activity_logs al
		LEFT JOIN staff s ON s.id = al.user_id
		LEFT JOIN branches b ON b.id = s.branch_id
		WHERE ($1::uuid IS NULL OR s.branch_id = $1::uuid)
		  AND ($2::text = '' OR al.entity_type = $2::text)
		  AND ($3::text = '' OR al.action = $3::text)
		ORDER BY al.created_at DESC
		LIMIT $4 OFFSET $5`, branchFilter, entityType, action, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the activity log: " + err.Error()})
		return
	}
	defer rows.Close()

	entries := make([]ActivityLog, 0)
	for rows.Next() {
		var entry ActivityLog
		var details []byte
		if err := rows.Scan(
			&entry.ID, &entry.UserID, &entry.UserName,
			&entry.BranchID, &entry.BranchName,
			&entry.EntityType, &entry.EntityID, &entry.Action,
			&details, &entry.CreatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read the activity log: " + err.Error()})
			return
		}
		if len(details) > 0 {
			entry.Details = json.RawMessage(details)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read the activity log: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}
