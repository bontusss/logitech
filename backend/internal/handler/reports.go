package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"backend/internal/auth"
	"backend/internal/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ReportHandler serves the owner/manager analytics: shipments, revenue,
// collections, outstanding balances and exceptions.
//
// Definitions used by every report:
//   - billed      amount charged on records that originated from the branch
//     (or every branch for the owner) and were created in the period
//   - collected   payments taken in the period by staff of that branch
//   - outstanding unpaid balance on every live record that originated from the
//     branch (all time, not just the period)
type ReportHandler struct {
	db      *sql.DB
	queries *sqlc.Queries
}

// NewReportHandler constructs a ReportHandler.
func NewReportHandler(queries *sqlc.Queries, pool *sql.DB) *ReportHandler {
	return &ReportHandler{db: pool, queries: queries}
}

// ReportPeriod is the inclusive date range of a report.
type ReportPeriod struct {
	From string `json:"from" example:"2026-09-29"`
	To   string `json:"to" example:"2026-10-28"`
}

// ReportShipmentStats is the shipment breakdown of the overview.
type ReportShipmentStats struct {
	Total           int64            `json:"total"`
	Quantity        int64            `json:"quantity"`
	Received        int64            `json:"received"`
	ByStatus        map[string]int64 `json:"by_status"`
	ByPaymentStatus map[string]int64 `json:"by_payment_status"`
}

// ReportMoneyStats is the money breakdown of the overview.
type ReportMoneyStats struct {
	Billed         float64 `json:"billed"`
	Collected      float64 `json:"collected"`
	Outstanding    float64 `json:"outstanding"`
	CollectionRate float64 `json:"collection_rate"`
	PaymentsCount  int64   `json:"payments_count"`
	CollectedToday float64 `json:"collected_today"`
}

// ReportMethodTotal is the amount collected through one payment method.
type ReportMethodTotal struct {
	Method string  `json:"method"`
	Count  int64   `json:"count"`
	Amount float64 `json:"amount"`
}

// ReportExceptionStats counts exceptions by status.
type ReportExceptionStats struct {
	Open          int64 `json:"open"`
	Investigating int64 `json:"investigating"`
	Resolved      int64 `json:"resolved"`
	Closed        int64 `json:"closed"`
	Total         int64 `json:"total"`
}

// ReportOverview is the owner dashboard payload.
type ReportOverview struct {
	Period           ReportPeriod         `json:"period"`
	Waybills         ReportShipmentStats  `json:"waybills"`
	Shipments        ReportShipmentStats  `json:"shipments"`
	Money            ReportMoneyStats     `json:"money"`
	PaymentsByMethod []ReportMethodTotal  `json:"payments_by_method"`
	Exceptions       ReportExceptionStats `json:"exceptions"`
}

// DailyReportRow is one day of the daily collection report.
type DailyReportRow struct {
	Date      string  `json:"date" example:"2026-10-28"`
	Shipments int64   `json:"shipments"`
	Billed    float64 `json:"billed"`
	Payments  int64   `json:"payments"`
	Collected float64 `json:"collected"`
}

// BranchReportRow is one branch of the per-branch report.
type BranchReportRow struct {
	BranchID          string  `json:"branch_id"`
	BranchName        string  `json:"branch_name"`
	BranchCode        string  `json:"branch_code"`
	ShipmentsSent     int64   `json:"shipments_sent"`
	ShipmentsReceived int64   `json:"shipments_received"`
	Billed            float64 `json:"billed"`
	Collected         float64 `json:"collected"`
	Outstanding       float64 `json:"outstanding"`
	OpenExceptions    int64   `json:"open_exceptions"`
}

// reportScope is the branch restriction of a report (empty for the owner).
type reportScope struct {
	all      bool
	branchID *string
}

// reportScopeFor enforces who may read reports and how far they reach.
func reportScopeFor(c *gin.Context) (reportScope, bool) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return reportScope{}, false
	}
	if canSeeAllBranches(claims.Role) {
		if raw := strings.TrimSpace(c.Query("branch_id")); raw != "" {
			if _, err := uuid.Parse(raw); err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "branch_id must be a valid UUID"})
				return reportScope{}, false
			}
			return reportScope{all: false, branchID: &raw}, true
		}
		return reportScope{all: true}, true
	}
	if claims.Role != auth.RoleManager {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "reports are available to managers and the owner"})
		return reportScope{}, false
	}
	if claims.BranchID == "" {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "your account is not attached to a branch"})
		return reportScope{}, false
	}
	branchID := claims.BranchID
	return reportScope{branchID: &branchID}, true
}

// parseReportPeriod reads from/to, defaulting to the last 30 days.
func parseReportPeriod(c *gin.Context) (time.Time, time.Time, error) {
	to := time.Now().UTC().Truncate(24 * time.Hour)
	from := to.AddDate(0, 0, -29)

	parsedFrom, err := parseOptionalDate(c.Query("from"))
	if err != nil {
		return from, to, errors.New("from must be YYYY-MM-DD")
	}
	parsedTo, err := parseOptionalDate(c.Query("to"))
	if err != nil {
		return from, to, errors.New("to must be YYYY-MM-DD")
	}
	if parsedFrom != nil {
		from = *parsedFrom
	}
	if parsedTo != nil {
		to = *parsedTo
	}
	if to.Before(from) {
		return from, to, errors.New("to must be on or after from")
	}
	if to.Sub(from) > 366*24*time.Hour {
		return from, to, errors.New("the period cannot exceed one year")
	}
	return from, to, nil
}

// Overview godoc
// @Summary Owner dashboard: shipments, revenue, collections, balances, exceptions
// @Description Totals for the selected period (default: the last 30 days). The owner and system admin see every branch; a manager sees their own branch. "billed" is what was charged on records created in the period, "collected" is the payments taken in the period, and "outstanding" is the unpaid balance on all live records, plus today's collections and the exception backlog.
// @Tags reports
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date (YYYY-MM-DD), default 29 days ago"
// @Param to query string false "End date (YYYY-MM-DD), default today"
// @Param branch_id query string false "Branch UUID (owner / system admin only)"
// @Success 200 {object} ReportOverview
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/reports/overview [get]
func (h *ReportHandler) Overview(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	scope, ok := reportScopeFor(c)
	if !ok {
		return
	}
	from, to, err := parseReportPeriod(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	branch := scope.branchArg()
	ctx := c.Request.Context()
	result := ReportOverview{
		Period: ReportPeriod{From: from.Format("2006-01-02"), To: to.Format("2006-01-02")},
		Shipments: ReportShipmentStats{
			// Zero-filled so dashboards get every key even on an empty period.
			ByStatus: map[string]int64{
				"pending": 0, "in_transit": 0, "received": 0, "exception": 0, "cancelled": 0,
			},
			ByPaymentStatus: map[string]int64{"pending": 0, "partial": 0, "paid": 0},
		},
		PaymentsByMethod: make([]ReportMethodTotal, 0),
	}

	// Shipments created in the period.
	rows, err := h.db.QueryContext(ctx, `
		SELECT mi.status::text, mi.payment_status::text, COUNT(*), COALESCE(SUM(mi.quantity), 0),
		       COUNT(*) FILTER (WHERE mi.received_at IS NOT NULL)
		FROM manifest_items mi
		WHERE ($1::uuid IS NULL OR mi.origin_branch_id = $1::uuid)
		  AND mi.created_at >= $2 AND mi.created_at < ($3::date + 1)
		GROUP BY mi.status, mi.payment_status`, branch, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load shipment totals: " + err.Error()})
		return
	}
	for rows.Next() {
		var status, paymentStatus string
		var count, quantity, received int64
		if err := rows.Scan(&status, &paymentStatus, &count, &quantity, &received); err != nil {
			rows.Close()
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read shipment totals: " + err.Error()})
			return
		}
		result.Shipments.Total += count
		result.Shipments.Quantity += quantity
		if status == "received" || status == "exception" {
			// received_at is only set once the destination branch confirms.
			result.Shipments.Received += received
		}
		result.Shipments.ByStatus[status] += count
		result.Shipments.ByPaymentStatus[paymentStatus] += count
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read shipment totals: " + err.Error()})
		return
	}
	rows.Close()

	// Money charged in the period and the balance still owed (all time).
	if err := h.db.QueryRowContext(ctx, `
		SELECT
		  COALESCE(SUM(mi.amount) FILTER (WHERE mi.created_at >= $2 AND mi.created_at < ($3::date + 1)), 0)::float8,
		  COALESCE(SUM(mi.balance) FILTER (WHERE mi.status <> 'cancelled'), 0)::float8
		FROM manifest_items mi
		WHERE ($1::uuid IS NULL OR mi.origin_branch_id = $1::uuid)
		  AND mi.status <> 'cancelled'`, branch, from, to).Scan(&result.Money.Billed, &result.Money.Outstanding); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load money totals: " + err.Error()})
		return
	}

	// Payments taken in the period, by the scoped branches' staff.
	if err := h.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(p.amount), 0)::float8, COUNT(*)
		FROM payments p
		JOIN staff s ON s.id = p.recorded_by
		WHERE ($1::uuid IS NULL OR s.branch_id = $1::uuid)
		  AND p.paid_at >= $2 AND p.paid_at < ($3::date + 1)`, branch, from, to).
		Scan(&result.Money.Collected, &result.Money.PaymentsCount); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load collections: " + err.Error()})
		return
	}

	if result.Money.Billed > 0 {
		result.Money.CollectionRate = roundMoney(result.Money.Collected / result.Money.Billed * 100)
	}

	if err := h.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(p.amount), 0)::float8
		FROM payments p
		JOIN staff s ON s.id = p.recorded_by
		WHERE ($1::uuid IS NULL OR s.branch_id = $1::uuid)
		  AND p.paid_at >= current_date AND p.paid_at < current_date + 1`, branch).
		Scan(&result.Money.CollectedToday); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load today's collections: " + err.Error()})
		return
	}

	methodRows, err := h.db.QueryContext(ctx, `
		SELECT p.method::text, COUNT(*), COALESCE(SUM(p.amount), 0)::float8
		FROM payments p
		JOIN staff s ON s.id = p.recorded_by
		WHERE ($1::uuid IS NULL OR s.branch_id = $1::uuid)
		  AND p.paid_at >= $2 AND p.paid_at < ($3::date + 1)
		GROUP BY p.method
		ORDER BY COUNT(*) DESC`, branch, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load payment methods: " + err.Error()})
		return
	}
	defer methodRows.Close()
	for methodRows.Next() {
		var item ReportMethodTotal
		if err := methodRows.Scan(&item.Method, &item.Count, &item.Amount); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read payment methods: " + err.Error()})
			return
		}
		result.PaymentsByMethod = append(result.PaymentsByMethod, item)
	}
	if err := methodRows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read payment methods: " + err.Error()})
		return
	}

	exceptionRows, err := h.db.QueryContext(ctx, `
		SELECT e.status::text, COUNT(*)
		FROM exceptions e
		JOIN manifest_items mi ON mi.id = e.manifest_item_id
		WHERE ($1::uuid IS NULL OR mi.origin_branch_id = $1::uuid OR mi.destination_branch_id = $1::uuid)
		GROUP BY e.status`, branch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load exceptions: " + err.Error()})
		return
	}
	defer exceptionRows.Close()
	for exceptionRows.Next() {
		var status string
		var count int64
		if err := exceptionRows.Scan(&status, &count); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read exceptions: " + err.Error()})
			return
		}
		switch status {
		case "open":
			result.Exceptions.Open += count
		case "investigating":
			result.Exceptions.Investigating += count
		case "resolved":
			result.Exceptions.Resolved += count
		case "closed":
			result.Exceptions.Closed += count
		}
		result.Exceptions.Total += count
	}
	if err := exceptionRows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read exceptions: " + err.Error()})
		return
	}

	result.Waybills = result.Shipments
	c.JSON(http.StatusOK, result)
}

// Daily godoc
// @Summary Daily shipments and collections
// @Description One row per day in the period: how many shipments were recorded, how much was billed, how many payments were taken and how much was collected. The owner sees all branches; a manager sees their own branch.
// @Tags reports
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date (YYYY-MM-DD), default 29 days ago"
// @Param to query string false "End date (YYYY-MM-DD), default today"
// @Param branch_id query string false "Branch UUID (owner / system admin only)"
// @Success 200 {array} DailyReportRow
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/reports/daily [get]
func (h *ReportHandler) Daily(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	scope, ok := reportScopeFor(c)
	if !ok {
		return
	}
	from, to, err := parseReportPeriod(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(), `
		WITH days AS (
		  SELECT generate_series($2::date, $3::date, interval '1 day')::date AS day
		)
		SELECT to_char(d.day, 'YYYY-MM-DD'),
		       (SELECT COUNT(*) FROM manifest_items mi
		        WHERE ($1::uuid IS NULL OR mi.origin_branch_id = $1::uuid)
		          AND mi.created_at >= d.day AND mi.created_at < d.day + 1),
		       (SELECT COALESCE(SUM(mi.amount), 0)::float8 FROM manifest_items mi
		        WHERE ($1::uuid IS NULL OR mi.origin_branch_id = $1::uuid)
		          AND mi.status <> 'cancelled'
		          AND mi.created_at >= d.day AND mi.created_at < d.day + 1),
		       (SELECT COUNT(*) FROM payments p
		        JOIN staff s ON s.id = p.recorded_by
		        WHERE ($1::uuid IS NULL OR s.branch_id = $1::uuid)
		          AND p.paid_at >= d.day AND p.paid_at < d.day + 1),
		       (SELECT COALESCE(SUM(p.amount), 0)::float8 FROM payments p
		        JOIN staff s ON s.id = p.recorded_by
		        WHERE ($1::uuid IS NULL OR s.branch_id = $1::uuid)
		          AND p.paid_at >= d.day AND p.paid_at < d.day + 1)
		FROM days d
		ORDER BY d.day`, scope.branchArg(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the daily report: " + err.Error()})
		return
	}
	defer rows.Close()

	results := make([]DailyReportRow, 0)
	for rows.Next() {
		var item DailyReportRow
		if err := rows.Scan(&item.Date, &item.Shipments, &item.Billed, &item.Payments, &item.Collected); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read the daily report: " + err.Error()})
			return
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read the daily report: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// Branches godoc
// @Summary Per-branch performance
// @Description Shipments sent and received, amounts billed, payments collected, unpaid balances and open exceptions for every branch. The owner sees all branches; a manager sees only their own.
// @Tags reports
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date (YYYY-MM-DD), default 29 days ago"
// @Param to query string false "End date (YYYY-MM-DD), default today"
// @Param branch_id query string false "Branch UUID (owner / system admin only)"
// @Success 200 {array} BranchReportRow
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/reports/branches [get]
func (h *ReportHandler) Branches(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	scope, ok := reportScopeFor(c)
	if !ok {
		return
	}
	from, to, err := parseReportPeriod(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(), `
		SELECT b.id::text, b.name, b.code,
		       (SELECT COUNT(*) FROM manifest_items mi
		        WHERE mi.origin_branch_id = b.id
		          AND mi.created_at >= $2 AND mi.created_at < ($3::date + 1)),
		       (SELECT COUNT(*) FROM manifest_items mi
		        WHERE mi.destination_branch_id = b.id
		          AND mi.received_at >= $2 AND mi.received_at < ($3::date + 1)),
		       (SELECT COALESCE(SUM(mi.amount), 0)::float8 FROM manifest_items mi
		        WHERE mi.origin_branch_id = b.id AND mi.status <> 'cancelled'
		          AND mi.created_at >= $2 AND mi.created_at < ($3::date + 1)),
		       (SELECT COALESCE(SUM(p.amount), 0)::float8 FROM payments p
		        JOIN staff s ON s.id = p.recorded_by
		        WHERE s.branch_id = b.id
		          AND p.paid_at >= $2 AND p.paid_at < ($3::date + 1)),
		       (SELECT COALESCE(SUM(mi.balance), 0)::float8 FROM manifest_items mi
		        WHERE mi.origin_branch_id = b.id AND mi.status <> 'cancelled'),
		       (SELECT COUNT(*) FROM exceptions e
		        WHERE e.branch_id = b.id AND e.status IN ('open', 'investigating'))
		FROM branches b
		WHERE b.status = 'active'
		  AND ($1::uuid IS NULL OR b.id = $1::uuid)
		ORDER BY b.name`, scope.branchArg(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the branch report: " + err.Error()})
		return
	}
	defer rows.Close()

	results := make([]BranchReportRow, 0)
	for rows.Next() {
		var item BranchReportRow
		if err := rows.Scan(
			&item.BranchID, &item.BranchName, &item.BranchCode,
			&item.ShipmentsSent, &item.ShipmentsReceived,
			&item.Billed, &item.Collected, &item.Outstanding,
			&item.OpenExceptions,
		); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read the branch report: " + err.Error()})
			return
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read the branch report: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// branchArg renders the scope as a nullable branch UUID argument for SQL.
func (s reportScope) branchArg() any {
	if s.branchID == nil || strings.TrimSpace(*s.branchID) == "" {
		return nil
	}
	return *s.branchID
}
