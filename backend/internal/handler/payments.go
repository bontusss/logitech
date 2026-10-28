package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"

	"backend/internal/auth"
	"backend/internal/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// addPayment records a payment against a record and returns the record with
// its refreshed amount_paid, payment_status and balance (the payments trigger
// maintains those columns).
//
// The caller must already be inside a transaction; the record row is locked so
// two payments cannot overshoot the amount owed.
func (h *RecordHandler) addPayment(
	ctx context.Context,
	q *sqlc.Queries,
	record sqlc.ManifestItem,
	amount float64,
	method string,
	reference string,
	notes string,
	staffID uuid.UUID,
) (sqlc.ManifestItem, error) {
	if amount <= 0 {
		return record, errors.New("payment amount must be greater than zero")
	}
	method = strings.ToLower(strings.TrimSpace(method))
	if method == "" {
		method = string(sqlc.PaymentMethodCash)
	}
	if !validPaymentMethods[method] {
		return record, errors.New("payment method must be cash, bank_transfer, pos or other")
	}

	current, err := q.GetManifestItemForUpdate(ctx, record.ID)
	if err != nil {
		return record, fmt.Errorf("failed to load the record: %w", err)
	}
	if current.Status == sqlc.ManifestItemStatusCancelled {
		return record, errors.New("a cancelled record cannot take payments")
	}
	balance := moneyFloat(current.Amount) - moneyFloat(current.AmountPaid)
	if amount > balance+0.004 {
		return record, fmt.Errorf("payment of %.2f exceeds the outstanding balance of %.2f %s", amount, balance, current.Currency)
	}

	if _, err := q.CreatePayment(ctx, sqlc.CreatePaymentParams{
		ID:             uuid.New(),
		ManifestItemID: current.ID,
		Amount:         formatMoney(amount),
		Currency:       current.Currency,
		Method:         sqlc.PaymentMethod(method),
		Reference:      nullIfEmpty(reference),
		RecordedBy:     staffID,
		Notes:          nullIfEmpty(notes),
	}); err != nil {
		return record, fmt.Errorf("failed to record the payment: %w", err)
	}

	refreshed, err := q.GetManifestItemByID(ctx, current.ID)
	if err != nil {
		return record, fmt.Errorf("failed to reload the record: %w", err)
	}
	return refreshed, nil
}

// RecordPaymentRequest is the body of POST /api/records/:id/payments.
type RecordPaymentRequest struct {
	Amount    float64 `json:"amount" example:"15000"`
	Method    string  `json:"method" example:"cash"`
	Reference string  `json:"reference,omitempty"`
	Notes     string  `json:"notes,omitempty"`
}

// RecordPayment godoc
// @Summary Record a payment against a record
// @Description Adds a payment for a shipment (at the counter or when the balance is collected at the destination branch) and returns the record with the updated amount paid and remaining balance. A payment can never exceed the outstanding balance.
// @Tags records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Record UUID"
// @Param request body RecordPaymentRequest true "Payment"
// @Success 200 {object} Record
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/records/{id}/payments [post]
func (h *RecordHandler) RecordPayment(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "record id must be a valid UUID"})
		return
	}
	var req RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	ctx := c.Request.Context()
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to start transaction: " + err.Error()})
		return
	}
	defer func() { _ = tx.Rollback() }()
	txq := h.queries.WithTx(tx)

	record, err := txq.GetManifestItemForUpdate(ctx, id)
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

	if _, err := h.addPayment(ctx, txq, record, req.Amount, req.Method, req.Reference, req.Notes, staffID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := logActivity(ctx, tx, staffID, "manifest_item", record.ID, "payment_recorded", map[string]any{
		"tracking_number": record.TrackingNumber,
		"amount":          formatMoney(req.Amount),
		"method":          req.Method,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to log the payment: " + err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save the payment: " + err.Error()})
		return
	}

	records, err := h.loadRecords(ctx, claims, recordFilter{IDs: []uuid.UUID{id}})
	if err != nil || len(records) == 0 {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the record"})
		return
	}
	c.JSON(http.StatusOK, records[0])
}

// ListRecordPayments godoc
// @Summary List the payments of a record
// @Description Returns every payment recorded against a shipment, oldest first.
// @Tags records
// @Produce json
// @Security BearerAuth
// @Param id path string true "Record UUID"
// @Success 200 {array} RecordPaymentResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/records/{id}/payments [get]
func (h *RecordHandler) ListRecordPayments(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "record id must be a valid UUID"})
		return
	}

	record, err := h.queries.GetManifestItemByID(c.Request.Context(), id)
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

	payments, err := h.queries.ListPaymentsByItem(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list payments: " + err.Error()})
		return
	}
	results := make([]RecordPaymentResponse, 0, len(payments))
	for _, p := range payments {
		results = append(results, RecordPaymentResponse{
			ID:             p.ID.String(),
			Amount:         moneyFloat(p.Amount),
			Currency:       p.Currency,
			Method:         string(p.Method),
			Reference:      p.Reference.String,
			RecordedBy:     p.RecordedBy.String(),
			RecordedByName: p.RecordedByName,
			PaidAt:         p.PaidAt,
			Notes:          p.Notes.String,
		})
	}
	c.JSON(http.StatusOK, results)
}

// canViewRecord reports whether claims may see a record: the owner and system
// admin see everything, everyone else only records touching their branch.
func canViewRecord(claims *auth.Claims, record sqlc.ManifestItem) bool {
	if canSeeAllBranches(claims.Role) {
		return true
	}
	if claims.BranchID == "" {
		return false
	}
	return record.OriginBranchID.String() == claims.BranchID || record.DestinationBranchID.String() == claims.BranchID
}

// roundMoney rounds to two decimals, the precision of the money columns.
func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}
