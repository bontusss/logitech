package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// RecordHandler owns shipment records. A record is one manifest item: the
// parcel a customer (sender) hands in at a branch, who it is for (receiver),
// the destination branch, the amount charged, the payments taken against it
// and the remaining balance. The destination branch then receives the record.
type RecordHandler struct {
	cfg     *config.Config
	db      *sql.DB
	queries *sqlc.Queries
}

// NewRecordHandler constructs a RecordHandler. The pool is required for the
// multi-step writes (record + payments) and for the filtered list queries.
func NewRecordHandler(cfg *config.Config, queries *sqlc.Queries, pool *sql.DB) *RecordHandler {
	return &RecordHandler{cfg: cfg, queries: queries, db: pool}
}

// RecordParty is the sender or receiver of a record.
type RecordParty struct {
	FullName string `json:"full_name" example:"Amina Yusuf"`
	Phone    string `json:"phone,omitempty" example:"08032220002"`
}

// PaymentPayload is a payment taken against a record.
type PaymentPayload struct {
	Amount    float64 `json:"amount" example:"15000"`
	Method    string  `json:"method" example:"cash"`
	Reference string  `json:"reference,omitempty" example:"POS-8891"`
	Notes     string  `json:"notes,omitempty"`
}

// RecordPaymentResponse is one payment line of a record.
type RecordPaymentResponse struct {
	ID             string    `json:"id"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
	Method         string    `json:"method"`
	Reference      string    `json:"reference,omitempty"`
	RecordedBy     string    `json:"recorded_by"`
	RecordedByName string    `json:"recorded_by_name"`
	PaidAt         time.Time `json:"paid_at"`
	Notes          string    `json:"notes,omitempty"`
}

// Record is a shipment as returned by the list endpoints.
type Record struct {
	ID                  string                  `json:"id"`
	TrackingNumber      string                  `json:"tracking_number"`
	SerialNumber        string                  `json:"serial_number,omitempty"`
	Description         string                  `json:"description"`
	Quantity            int32                   `json:"quantity"`
	Sender              RecordParty             `json:"sender"`
	Receiver            RecordParty             `json:"receiver"`
	OriginBranchID      string                  `json:"origin_branch_id"`
	OriginBranch        string                  `json:"origin_branch"`
	DestinationBranchID string                  `json:"destination_branch_id"`
	DestinationBranch   string                  `json:"destination_branch"`
	Amount              float64                 `json:"amount"`
	AmountPaid          float64                 `json:"amount_paid"`
	Balance             float64                 `json:"balance"`
	Currency            string                  `json:"currency"`
	PaymentStatus       string                  `json:"payment_status"`
	Status              string                  `json:"status"`
	RecordedBy          string                  `json:"recorded_by"`
	RecordedByName      string                  `json:"recorded_by_name"`
	PhotoURL            string                  `json:"photo_url,omitempty"`
	Notes               string                  `json:"notes,omitempty"`
	DispatchedAt        *time.Time              `json:"dispatched_at,omitempty"`
	ReceivedBy          string                  `json:"received_by,omitempty"`
	ReceivedByName      string                  `json:"received_by_name,omitempty"`
	ReceivedAt          *time.Time              `json:"received_at,omitempty"`
	ReceivedQuantity    *int32                  `json:"received_quantity,omitempty"`
	Condition           string                  `json:"condition,omitempty"`
	CreatedAt           time.Time               `json:"created_at"`
	UpdatedAt           time.Time               `json:"updated_at"`
	Payments            []RecordPaymentResponse `json:"payments,omitempty"`
	Exceptions          []ExceptionResponse     `json:"exceptions,omitempty"`
}

// CreateRecordItem is one item in the POST /api/records body.
type CreateRecordItem struct {
	TrackingNumber      string      `json:"tracking_number,omitempty" example:"LAG-01-48213"`
	SerialNumber        string      `json:"serial_number,omitempty" example:"001"`
	Description         string      `json:"description" example:"Cartons of engine parts"`
	Quantity            int         `json:"quantity" example:"10"`
	Sender              RecordParty `json:"sender"`
	Receiver            RecordParty `json:"receiver"`
	DestinationBranchID string      `json:"destination_branch_id,omitempty"`
	Amount              float64     `json:"amount" example:"45000"`
	Currency            string      `json:"currency,omitempty" example:"NGN"`
	// AmountPaid is the payment taken at the counter when the item is recorded.
	AmountPaid       float64 `json:"amount_paid,omitempty" example:"20000"`
	PaymentMethod    string  `json:"payment_method,omitempty" example:"cash"`
	PaymentReference string  `json:"payment_reference,omitempty"`
	PhotoURL         string  `json:"photo_url,omitempty"`
	Notes            string  `json:"notes,omitempty"`
}

// CreateRecordRequest is the body of POST /api/records. A customer can hand in
// several items at once; they all share the origin branch and, when
// DestinationBranchID is set, the same destination branch.
type CreateRecordRequest struct {
	OriginBranchID      string             `json:"origin_branch_id,omitempty"`
	DestinationBranchID string             `json:"destination_branch_id,omitempty"`
	Items               []CreateRecordItem `json:"items"`
}

// UpdateRecordRequest is the body of PUT /api/records/:id. Only a record that
// has not left the origin branch can be edited.
type UpdateRecordRequest struct {
	SerialNumber        string      `json:"serial_number,omitempty"`
	Description         string      `json:"description,omitempty"`
	Quantity            int         `json:"quantity,omitempty"`
	Receiver            RecordParty `json:"receiver,omitempty"`
	DestinationBranchID string      `json:"destination_branch_id,omitempty"`
	Amount              *float64    `json:"amount,omitempty"`
	Notes               string      `json:"notes,omitempty"`
}

// DispatchRecordRequest is the body of POST /api/records/:id/dispatch.
type DispatchRecordRequest struct {
	Notes string `json:"notes,omitempty"`
}

// ReceiveRecordRequest is the body of POST /api/records/:id/receive. Payment is
// optional: the receiver can collect the remaining balance at the destination
// branch in the same call.
type ReceiveRecordRequest struct {
	ReceivedQuantity *int            `json:"received_quantity,omitempty" example:"10"`
	Condition        string          `json:"condition,omitempty" example:"good"`
	Notes            string          `json:"notes,omitempty"`
	Payment          *PaymentPayload `json:"payment,omitempty"`
}

// trackingNumberPattern is the shape of a tracking number accepted from the client.
var trackingNumberPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{2,29}$`)

// validPaymentMethods is every value of the payment_method enum.
var validPaymentMethods = map[string]bool{
	"cash":          true,
	"bank_transfer": true,
	"pos":           true,
	"other":         true,
}

// validConditions is every value of the item_condition enum.
var validConditions = map[string]bool{
	"good":    true,
	"damaged": true,
	"missing": true,
}

// canSeeAllBranches reports whether a role is scoped to the whole company.
func canSeeAllBranches(role string) bool {
	return role == auth.RoleOwner || role == auth.RoleSystemAdmin
}

// claimsStaffID returns the authenticated staff UUID. The system administrator
// is not a staff record and cannot own shipments.
func claimsStaffID(c *gin.Context) (uuid.UUID, *auth.Claims, bool) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return uuid.Nil, nil, false
	}
	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "only staff accounts can record shipments"})
		return uuid.Nil, nil, false
	}
	return id, claims, true
}

// CreateRecord godoc
// @Summary Record one or more shipments
// @Description The sender-side flow: a staff member records the item(s) a customer brings in, the destination branch, the amount charged and the payment taken at the counter. Any unpaid remainder becomes the record's balance. Senders and receivers are matched on phone number, or created when they are new. When origin_branch_id is omitted the staff member's own branch is used.
// @Tags records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRecordRequest true "Shipment items"
// @Success 201 {array} Record
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/records [post]
func (h *RecordHandler) CreateRecord(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	staffID, claims, ok := claimsStaffID(c)
	if !ok {
		return
	}
	if claims.Role == auth.RoleOwner {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "owners cannot send parcels"})
		return
	}

	var req CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "a record needs at least one item"})
		return
	}

	ctx := c.Request.Context()

	originBranchID, err := h.resolveOriginBranch(ctx, claims, req.OriginBranchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	originBranch, err := h.queries.GetBranchByID(ctx, originBranchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "origin branch not found"})
		return
	}
	if originBranch.Status != sqlc.RecordStatusActive {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "origin branch is inactive"})
		return
	}

	defaultDestination, err := h.optionalBranch(ctx, req.DestinationBranchID, "destination_branch_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Validate every item before writing anything.
	type preparedItem struct {
		payload     CreateRecordItem
		destination uuid.UUID
	}
	prepared := make([]preparedItem, 0, len(req.Items))
	for index, item := range req.Items {
		if msg := validateRecordItem(item, defaultDestination != uuid.Nil); msg != "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("item %d: %s", index+1, msg)})
			return
		}
		destination := defaultDestination
		if destination == uuid.Nil {
			parsed, err := uuid.Parse(strings.TrimSpace(item.DestinationBranchID))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("item %d: destination_branch_id must be a valid UUID", index+1)})
				return
			}
			branch, err := h.queries.GetBranchByID(ctx, parsed)
			if err != nil || branch.Status != sqlc.RecordStatusActive {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("item %d: destination branch not found or inactive", index+1)})
				return
			}
			destination = parsed
		}
		prepared = append(prepared, preparedItem{payload: item, destination: destination})
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to start transaction: " + err.Error()})
		return
	}
	defer func() { _ = tx.Rollback() }()
	txq := h.queries.WithTx(tx)

	createdIDs := make([]uuid.UUID, 0, len(prepared))
	for index, item := range prepared {
		p := item.payload
		senderID, err := resolveCustomer(ctx, txq, p.Sender)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("item %d: sender: %s", index+1, err.Error())})
			return
		}
		receiverID, err := resolveCustomer(ctx, txq, p.Receiver)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("item %d: receiver: %s", index+1, err.Error())})
			return
		}

		trackingNumber, err := h.trackingNumber(ctx, txq, originBranch.Code, p.TrackingNumber)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("item %d: %s", index+1, err.Error())})
			return
		}

		quantity := p.Quantity
		if quantity < 1 {
			quantity = 1
		}
		currency := "NGN"
		if strings.TrimSpace(p.Currency) != "" {
			currency = strings.ToUpper(strings.TrimSpace(p.Currency))
		}

		record, err := txq.CreateManifestItem(ctx, sqlc.CreateManifestItemParams{
			ID:                  uuid.New(),
			TrackingNumber:      trackingNumber,
			SerialNumber:        nullIfEmpty(p.SerialNumber),
			Description:         strings.TrimSpace(p.Description),
			Quantity:            int32(quantity),
			SenderCustomerID:    senderID,
			ReceiverCustomerID:  receiverID,
			OriginBranchID:      originBranchID,
			DestinationBranchID: item.destination,
			Amount:              formatMoney(p.Amount),
			Currency:            currency,
			Status:              sqlc.ManifestItemStatusPending,
			RecordedBy:          staffID,
			PhotoUrl:            nullIfEmpty(p.PhotoURL),
			Notes:               nullIfEmpty(p.Notes),
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("item %d: tracking_number %s is already in use", index+1, trackingNumber)})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save the record: " + err.Error()})
			return
		}

		if p.AmountPaid > 0 {
			if _, err := h.addPayment(ctx, txq, record, p.AmountPaid, p.PaymentMethod, p.PaymentReference, "payment at drop-off", staffID); err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("item %d: %s", index+1, err.Error())})
				return
			}
		}

		if err := logActivity(ctx, tx, staffID, "manifest_item", record.ID, "record_created", map[string]any{
			"tracking_number":       record.TrackingNumber,
			"origin_branch_id":      record.OriginBranchID.String(),
			"destination_branch_id": record.DestinationBranchID.String(),
			"amount":                record.Amount,
			"amount_paid":           p.AmountPaid,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to log the record: " + err.Error()})
			return
		}

		createdIDs = append(createdIDs, record.ID)
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save records: " + err.Error()})
		return
	}

	records, err := h.loadRecords(ctx, claims, recordFilter{IDs: createdIDs})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the saved records: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, records)
}

// RecordSummary is the signed-in account's own waybill counts (dashboard).
type RecordSummary struct {
	Sent    int64 `json:"sent" example:"12"`
	Pending int64 `json:"pending"`
}

// Summary godoc
// @Summary My waybill counts
// @Description Counts the waybills the signed-in account recorded: those that have already been dispatched from the origin branch ("sent") and those still waiting to leave it ("pending"). Cancelled records count as neither.
// @Tags records
// @Produce json
// @Security BearerAuth
// @Success 200 {object} RecordSummary
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/records/summary [get]
func (h *RecordHandler) Summary(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	staffID, _, ok := claimsStaffID(c)
	if !ok {
		return
	}

	var summary RecordSummary
	if err := h.db.QueryRowContext(c.Request.Context(), `
		SELECT COUNT(*) FILTER (WHERE status NOT IN ('pending', 'cancelled')),
		       COUNT(*) FILTER (WHERE status = 'pending')
		FROM manifest_items
		WHERE recorded_by = $1`, staffID).Scan(&summary.Sent, &summary.Pending); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load your waybill counts: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// ListRecords godoc
// @Summary List shipment records
// @Description Lists shipment records newest first with the amount billed, the amount paid and the remaining balance. Owners and the system admin see every branch; managers and staff see records that originate from or are destined to their own branch. Optional filters: free text (tracking number, serial number, description, sender/receiver name or phone), status, payment_status, branch_id (origin or destination), origin_branch_id, destination_branch_id, from/to dates (created_at), and outstanding=true for records with a balance.
// @Tags records
// @Produce json
// @Security BearerAuth
// @Param q query string false "Tracking number, serial number, description, party name or phone"
// @Param status query string false "comma-separated: pending,in_transit,received,exception,cancelled"
// @Param payment_status query string false "pending, partial or paid"
// @Param branch_id query string false "Branch UUID (origin or destination)"
// @Param origin_branch_id query string false "Origin branch UUID"
// @Param destination_branch_id query string false "Destination branch UUID"
// @Param from query string false "Created on or after (YYYY-MM-DD)"
// @Param to query string false "Created on or before (YYYY-MM-DD)"	// @Param outstanding query bool false "Only records with an unpaid balance"
// @Param mine query bool false "Only records the signed-in account recorded"
// @Param limit query int false "Page size (default 50, max 200)"
// @Param offset query int false "Offset"
// @Success 200 {array} Record
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/records [get]
func (h *RecordHandler) ListRecords(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	filter, err := parseRecordFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	records, err := h.loadRecords(c.Request.Context(), claims, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list records: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}

// GetRecord godoc
// @Summary Get a shipment record
// @Description Returns one record with its payment history and any exceptions raised against it.
// @Tags records
// @Produce json
// @Security BearerAuth
// @Param id path string true "Record UUID"
// @Success 200 {object} Record
// @Failure 404 {object} ErrorResponse
// @Router /api/records/{id} [get]
func (h *RecordHandler) GetRecord(c *gin.Context) {
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

	records, err := h.loadRecords(c.Request.Context(), claims, recordFilter{IDs: []uuid.UUID{id}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the record: " + err.Error()})
		return
	}
	if len(records) == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "record not found"})
		return
	}
	c.JSON(http.StatusOK, records[0])
}

// UpdateRecord godoc
// @Summary Edit a record that has not left the origin branch
// @Description Updates the receiver, destination, description, quantity or amount of a pending record. The amount cannot be lowered below what has already been paid. The sender and the origin branch are immutable.
// @Tags records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Record UUID"
// @Param request body UpdateRecordRequest true "Fields to update"
// @Success 200 {object} Record
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/records/{id} [put]
func (h *RecordHandler) UpdateRecord(c *gin.Context) {
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
	var req UpdateRecordRequest
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
	if !h.canEditRecord(claims, record) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "record not found"})
		return
	}
	if record.Status != sqlc.ManifestItemStatusPending {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "only a pending record can be edited"})
		return
	}

	params := sqlc.UpdateManifestItemParams{ID: record.ID}
	if strings.TrimSpace(req.Description) != "" {
		params.Description = nullIfEmpty(req.Description)
	}
	if strings.TrimSpace(req.SerialNumber) != "" {
		params.SerialNumber = nullIfEmpty(req.SerialNumber)
	}
	if req.Quantity > 0 {
		params.Quantity = sql.NullInt32{Int32: int32(req.Quantity), Valid: true}
	}
	if req.Amount != nil {
		if *req.Amount < 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "amount cannot be negative"})
			return
		}
		if *req.Amount+0.004 < moneyFloat(record.AmountPaid) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "amount cannot be lower than the amount already paid"})
			return
		}
		params.Amount = nullIfEmpty(formatMoney(*req.Amount))
	}
	if strings.TrimSpace(req.Notes) != "" {
		params.Notes = nullIfEmpty(req.Notes)
	}
	if strings.TrimSpace(req.DestinationBranchID) != "" {
		destination, err := h.optionalBranch(ctx, req.DestinationBranchID, "destination_branch_id")
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		params.DestinationBranchID = uuid.NullUUID{UUID: destination, Valid: true}
	}
	if strings.TrimSpace(req.Receiver.FullName) != "" {
		receiverID, err := resolveCustomer(ctx, txq, req.Receiver)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "receiver: " + err.Error()})
			return
		}
		params.ReceiverCustomerID = uuid.NullUUID{UUID: receiverID, Valid: true}
	}

	if _, err := txq.UpdateManifestItem(ctx, params); err != nil {
		if strings.Contains(err.Error(), "manifest_items_paid_range") {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "amount cannot be lower than the amount already paid"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update the record: " + err.Error()})
		return
	}

	if err := logActivity(ctx, tx, staffID, "manifest_item", record.ID, "record_updated", map[string]any{
		"tracking_number": record.TrackingNumber,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to log the update: " + err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save the record: " + err.Error()})
		return
	}

	records, err := h.loadRecords(ctx, claims, recordFilter{IDs: []uuid.UUID{id}})
	if err != nil || len(records) == 0 {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the updated record"})
		return
	}
	c.JSON(http.StatusOK, records[0])
}

// DispatchRecord godoc
// @Summary Mark a record as in transit
// @Description Marks a pending record as in_transit when it leaves the origin branch. Origin-branch staff, their manager and the owner can call this.
// @Tags records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Record UUID"
// @Param request body DispatchRecordRequest false "Optional notes"
// @Success 200 {object} Record
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/records/{id}/dispatch [post]
func (h *RecordHandler) DispatchRecord(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	staffID, claims, ok := claimsStaffID(c)
	if !ok {
		return
	}
	if claims.Role == auth.RoleOwner {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "owners cannot dispatch parcels"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "record id must be a valid UUID"})
		return
	}
	var req DispatchRecordRequest
	_ = c.ShouldBindJSON(&req)

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
	if !h.canDispatch(claims, record) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "record not found"})
		return
	}
	if record.Status == sqlc.ManifestItemStatusCancelled {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "a cancelled record cannot be dispatched"})
		return
	}
	if record.Status != sqlc.ManifestItemStatusPending {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "only a pending record can be dispatched"})
		return
	}

	if _, err := txq.MarkManifestItemDispatched(ctx, record.ID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to dispatch the record: " + err.Error()})
		return
	}
	if strings.TrimSpace(req.Notes) != "" {
		if _, err := txq.UpdateManifestItem(ctx, sqlc.UpdateManifestItemParams{ID: record.ID, Notes: nullIfEmpty(req.Notes)}); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save the dispatch notes: " + err.Error()})
			return
		}
	}
	if err := logActivity(ctx, tx, staffID, "manifest_item", record.ID, "record_dispatched", map[string]any{
		"tracking_number": record.TrackingNumber,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to log the dispatch: " + err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save the dispatch: " + err.Error()})
		return
	}

	records, err := h.loadRecords(ctx, claims, recordFilter{IDs: []uuid.UUID{id}})
	if err != nil || len(records) == 0 {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the dispatched record"})
		return
	}
	c.JSON(http.StatusOK, records[0])
}

// ReceiveRecord godoc
// @Summary Receive a record at the destination branch
// @Description The receiver-side flow: staff at the destination branch confirm the quantity and condition actually received. Any remaining balance can be collected in the same call by sending a payment, which updates the balance and payment status. The record becomes "received"; use POST /api/records/:id/exceptions instead when goods are missing, damaged or wrong.
// @Tags records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Record UUID"
// @Param request body ReceiveRecordRequest true "Receipt details"
// @Success 200 {object} Record
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/records/{id}/receive [post]
func (h *RecordHandler) ReceiveRecord(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	staffID, claims, ok := claimsStaffID(c)
	if !ok {
		return
	}
	if claims.Role == auth.RoleOwner {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "owners cannot receive parcels"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "record id must be a valid UUID"})
		return
	}
	var req ReceiveRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}
	condition := strings.ToLower(strings.TrimSpace(req.Condition))
	if condition == "" {
		condition = "good"
	}
	if !validConditions[condition] {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "condition must be good, damaged or missing"})
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
	if !h.canReceive(claims, record) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "record not found"})
		return
	}
	if record.Status == sqlc.ManifestItemStatusCancelled {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "a cancelled record cannot be received"})
		return
	}
	// A record flagged with an exception before the goods were counted can still
	// be received afterwards; that is how the exception gets closed out.
	if record.Status == sqlc.ManifestItemStatusReceived || (record.Status == sqlc.ManifestItemStatusException && record.ReceivedAt.Valid) {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "this record has already been received"})
		return
	}

	var hasOpenExceptions bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		  SELECT 1 FROM exceptions
		  WHERE manifest_item_id = $1 AND status IN ('open', 'investigating')
		)`, record.ID).Scan(&hasOpenExceptions); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to check the record's exceptions: " + err.Error()})
		return
	}
	receiptStatus := sqlc.ManifestItemStatusReceived
	if hasOpenExceptions {
		receiptStatus = sqlc.ManifestItemStatusException
	}

	receivedQuantity := record.Quantity
	if req.ReceivedQuantity != nil {
		if *req.ReceivedQuantity < 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "received_quantity cannot be negative"})
			return
		}
		receivedQuantity = int32(*req.ReceivedQuantity)
	}

	updated, err := txq.MarkManifestItemReceived(ctx, sqlc.MarkManifestItemReceivedParams{
		ID:               record.ID,
		Status:           receiptStatus,
		ReceivedBy:       uuid.NullUUID{UUID: staffID, Valid: true},
		ReceivedQuantity: sql.NullInt32{Int32: receivedQuantity, Valid: true},
		Condition:        sqlc.NullItemCondition{ItemCondition: sqlc.ItemCondition(condition), Valid: true},
		Notes:            nullIfEmpty(req.Notes),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to receive the record: " + err.Error()})
		return
	}

	if req.Payment != nil {
		if _, err := h.addPayment(ctx, txq, updated, req.Payment.Amount, req.Payment.Method, req.Payment.Reference, req.Payment.Notes, staffID); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
	}

	if err := logActivity(ctx, tx, staffID, "manifest_item", record.ID, "record_received", map[string]any{
		"tracking_number":   record.TrackingNumber,
		"received_quantity": receivedQuantity,
		"condition":         condition,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to log the receipt: " + err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save the receipt: " + err.Error()})
		return
	}

	records, err := h.loadRecords(ctx, claims, recordFilter{IDs: []uuid.UUID{id}})
	if err != nil || len(records) == 0 {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the received record"})
		return
	}
	c.JSON(http.StatusOK, records[0])
}

// CancelRecord godoc
// @Summary Cancel a record
// @Description Cancels a shipment that has not been received yet. Managers and the owner only.
// @Tags records
// @Produce json
// @Security BearerAuth
// @Param id path string true "Record UUID"
// @Success 200 {object} Record
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/records/{id}/cancel [post]
func (h *RecordHandler) CancelRecord(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	staffID, claims, ok := claimsStaffID(c)
	if !ok {
		return
	}
	if claims.Role != auth.RoleManager && !canSeeAllBranches(claims.Role) {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "only a manager or the owner can cancel a record"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "record id must be a valid UUID"})
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
	if !canSeeAllBranches(claims.Role) && !h.canDispatch(claims, record) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "record not found"})
		return
	}
	if record.Status == sqlc.ManifestItemStatusReceived || record.Status == sqlc.ManifestItemStatusException {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "a received record cannot be cancelled"})
		return
	}
	if record.Status == sqlc.ManifestItemStatusCancelled {
		records, err := h.loadRecords(ctx, claims, recordFilter{IDs: []uuid.UUID{id}})
		if err != nil || len(records) == 0 {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the record"})
			return
		}
		c.JSON(http.StatusOK, records[0])
		return
	}

	if _, err := txq.CancelManifestItem(ctx, record.ID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to cancel the record: " + err.Error()})
		return
	}
	if err := logActivity(ctx, tx, staffID, "manifest_item", record.ID, "record_cancelled", map[string]any{
		"tracking_number": record.TrackingNumber,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to log the cancellation: " + err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save the cancellation: " + err.Error()})
		return
	}

	records, err := h.loadRecords(ctx, claims, recordFilter{IDs: []uuid.UUID{id}})
	if err != nil || len(records) == 0 {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load the cancelled record"})
		return
	}
	c.JSON(http.StatusOK, records[0])
}

// resolveOriginBranch picks the branch a new record belongs to. Staff and
// managers always write to their own branch; only the owner may record on
// behalf of another branch.
func (h *RecordHandler) resolveOriginBranch(ctx context.Context, claims *auth.Claims, requested string) (uuid.UUID, error) {
	requested = strings.TrimSpace(requested)
	if requested != "" {
		id, err := uuid.Parse(requested)
		if err != nil {
			return uuid.Nil, errors.New("origin_branch_id must be a valid UUID")
		}
		// Only the owner may record on behalf of another branch.
		if !canSeeAllBranches(claims.Role) && id.String() != claims.BranchID {
			return uuid.Nil, errors.New("you can only record shipments for your own branch")
		}
		return id, nil
	}
	if claims.BranchID == "" {
		return uuid.Nil, errors.New("origin_branch_id is required")
	}
	id, err := uuid.Parse(claims.BranchID)
	if err != nil {
		return uuid.Nil, errors.New("your account is not attached to a valid branch")
	}
	return id, nil
}

// optionalBranch parses and verifies an optional branch id.
func (h *RecordHandler) optionalBranch(ctx context.Context, raw string, field string) (uuid.UUID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return uuid.Nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New(field + " must be a valid UUID")
	}
	branch, err := h.queries.GetBranchByID(ctx, id)
	if err != nil {
		return uuid.Nil, errors.New("destination branch not found")
	}
	if branch.Status != sqlc.RecordStatusActive {
		return uuid.Nil, errors.New("destination branch is inactive")
	}
	return id, nil
}

// trackingNumber returns the supplied number after checking it is free, or
// generates "<branch code>-NNNNN" when none was supplied.
func (h *RecordHandler) trackingNumber(ctx context.Context, q *sqlc.Queries, branchCode string, requested string) (string, error) {
	requested = strings.ToUpper(strings.TrimSpace(requested))
	if requested != "" {
		if !trackingNumberPattern.MatchString(requested) {
			return "", errors.New("tracking_number must be 3-30 characters (letters, digits or dashes)")
		}
		taken, err := q.TrackingNumberExists(ctx, requested)
		if err != nil {
			return "", err
		}
		if taken {
			return "", errors.New("tracking_number " + requested + " is already in use")
		}
		return requested, nil
	}

	prefix := strings.ToUpper(strings.TrimSpace(branchCode))
	if prefix == "" {
		prefix = "LOG"
	}
	for attempt := 0; attempt < 40; attempt++ {
		candidate := fmt.Sprintf("%s-%05d", prefix, 10000+rand.IntN(90000))
		taken, err := q.TrackingNumberExists(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
	}
	return "", errors.New("no free tracking number left for this branch")
}

// canEditRecord reports whether claims may edit a pending record.
func (h *RecordHandler) canEditRecord(claims *auth.Claims, record sqlc.ManifestItem) bool {
	if canSeeAllBranches(claims.Role) {
		return true
	}
	if claims.Role == auth.RoleManager && claims.BranchID == record.OriginBranchID.String() {
		return true
	}
	return claims.Role == auth.RoleStaff && record.RecordedBy.String() == claims.UserID
}

// canDispatch reports whether claims may dispatch a record.
func (h *RecordHandler) canDispatch(claims *auth.Claims, record sqlc.ManifestItem) bool {
	if claims.Role == auth.RoleOwner {
		return false
	}
	if canSeeAllBranches(claims.Role) {
		return true
	}
	return claims.BranchID != "" && record.OriginBranchID.String() == claims.BranchID
}

// canReceive reports whether claims may receive a record at its destination.
func (h *RecordHandler) canReceive(claims *auth.Claims, record sqlc.ManifestItem) bool {
	if claims.Role == auth.RoleOwner {
		return false
	}
	if canSeeAllBranches(claims.Role) {
		return true
	}
	return claims.BranchID != "" && record.DestinationBranchID.String() == claims.BranchID
}

// recordFilter is the filter accepted by the record listing.
type recordFilter struct {
	IDs                 []uuid.UUID
	Query               string
	Statuses            []string
	PaymentStatus       string
	BranchID            uuid.UUID
	OriginBranchID      uuid.UUID
	DestinationBranchID uuid.UUID
	From                *time.Time
	To                  *time.Time
	Outstanding         bool
	Mine                bool
	Limit               int
	Offset              int
}

// knownRecordStatuses bounds the status filter.
var knownRecordStatuses = map[string]bool{
	"pending":    true,
	"in_transit": true,
	"received":   true,
	"exception":  true,
	"cancelled":  true,
}

// parseRecordFilter reads the listing query parameters.
func parseRecordFilter(c *gin.Context) (recordFilter, error) {
	filter := recordFilter{Limit: 50}

	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			if !knownRecordStatuses[s] {
				return filter, errors.New("unknown status: " + s)
			}
			filter.Statuses = append(filter.Statuses, s)
		}
	}
	if raw := strings.TrimSpace(c.Query("payment_status")); raw != "" {
		if raw != string(sqlc.PaymentStatusPending) && raw != string(sqlc.PaymentStatusPartial) && raw != string(sqlc.PaymentStatusPaid) {
			return filter, errors.New("payment_status must be pending, partial or paid")
		}
		filter.PaymentStatus = raw
	}
	for _, spec := range []struct {
		name   string
		target *uuid.UUID
	}{
		{"branch_id", &filter.BranchID},
		{"origin_branch_id", &filter.OriginBranchID},
		{"destination_branch_id", &filter.DestinationBranchID},
	} {
		if raw := strings.TrimSpace(c.Query(spec.name)); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				return filter, errors.New(spec.name + " must be a valid UUID")
			}
			*spec.target = id
		}
	}

	var err error
	if filter.From, err = parseOptionalDate(c.Query("from")); err != nil {
		return filter, errors.New("from must be YYYY-MM-DD")
	}
	if filter.To, err = parseOptionalDate(c.Query("to")); err != nil {
		return filter, errors.New("to must be YYYY-MM-DD")
	}
	if raw := strings.TrimSpace(c.Query("outstanding")); raw != "" {
		filter.Outstanding = raw == "true" || raw == "1"
	}
	if raw := strings.TrimSpace(c.Query("mine")); raw != "" {
		filter.Mine = raw == "true" || raw == "1"
	}
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit < 1 {
			return filter, errors.New("limit must be a positive integer")
		}
		if limit > 200 {
			limit = 200
		}
		filter.Limit = limit
	}
	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		offset, err := strconv.Atoi(raw)
		if err != nil || offset < 0 {
			return filter, errors.New("offset must be zero or positive")
		}
		filter.Offset = offset
	}
	filter.Query = strings.TrimSpace(c.Query("q"))
	return filter, nil
}

// loadRecords runs the record query and attaches payments and exceptions.
func (h *RecordHandler) loadRecords(ctx context.Context, claims *auth.Claims, filter recordFilter) ([]Record, error) {
	where, args := h.recordWhere(claims, filter)

	query := `
		SELECT mi.id::text, mi.tracking_number, COALESCE(mi.serial_number, ''), mi.description,
		       mi.quantity,
		       sender.full_name, COALESCE(sender.phone, ''),
		       receiver.full_name, COALESCE(receiver.phone, ''),
		       mi.origin_branch_id::text, COALESCE(origin.name, ''),
		       mi.destination_branch_id::text, COALESCE(destination.name, ''),
		       mi.amount::float8, mi.amount_paid::float8, COALESCE(mi.balance, 0)::float8,
		       mi.currency, mi.payment_status::text, mi.status::text,
		       mi.recorded_by::text, COALESCE(recorder.full_name, ''),
		       COALESCE(mi.photo_url, ''), COALESCE(mi.notes, ''),
		       mi.dispatched_at,
		       COALESCE(mi.received_by::text, ''), COALESCE(receiver_staff.full_name, ''),
		       mi.received_at, mi.received_quantity, COALESCE(mi.condition::text, ''),
		       mi.created_at, mi.updated_at
		FROM manifest_items mi
		JOIN customers sender ON sender.id = mi.sender_customer_id
		JOIN customers receiver ON receiver.id = mi.receiver_customer_id
		LEFT JOIN branches origin ON origin.id = mi.origin_branch_id
		LEFT JOIN branches destination ON destination.id = mi.destination_branch_id
		LEFT JOIN staff recorder ON recorder.id = mi.recorded_by
		LEFT JOIN staff receiver_staff ON receiver_staff.id = mi.received_by
		WHERE 1 = 1` + where + `
		ORDER BY mi.created_at DESC, mi.tracking_number DESC`

	if len(filter.IDs) == 0 {
		args = append(args, filter.Limit, filter.Offset)
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	}

	rows, err := h.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]Record, 0)
	for rows.Next() {
		var record Record
		var dispatchedAt, receivedAt sql.NullTime
		var receivedQuantity sql.NullInt32
		if err := rows.Scan(
			&record.ID, &record.TrackingNumber, &record.SerialNumber, &record.Description,
			&record.Quantity,
			&record.Sender.FullName, &record.Sender.Phone,
			&record.Receiver.FullName, &record.Receiver.Phone,
			&record.OriginBranchID, &record.OriginBranch,
			&record.DestinationBranchID, &record.DestinationBranch,
			&record.Amount, &record.AmountPaid, &record.Balance,
			&record.Currency, &record.PaymentStatus, &record.Status,
			&record.RecordedBy, &record.RecordedByName,
			&record.PhotoURL, &record.Notes,
			&dispatchedAt,
			&record.ReceivedBy, &record.ReceivedByName,
			&receivedAt, &receivedQuantity, &record.Condition,
			&record.CreatedAt, &record.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if dispatchedAt.Valid {
			record.DispatchedAt = &dispatchedAt.Time
		}
		if receivedAt.Valid {
			record.ReceivedAt = &receivedAt.Time
		}
		if receivedQuantity.Valid {
			quantity := receivedQuantity.Int32
			record.ReceivedQuantity = &quantity
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Payments and exceptions are only attached when a single record (or the
	// records just created) are requested, keeping the list response lean.
	if len(filter.IDs) == 0 {
		return records, nil
	}
	for i := range records {
		id, err := uuid.Parse(records[i].ID)
		if err != nil {
			return nil, err
		}
		payments, err := h.queries.ListPaymentsByItem(ctx, id)
		if err != nil {
			return nil, err
		}
		records[i].Payments = make([]RecordPaymentResponse, 0, len(payments))
		for _, p := range payments {
			records[i].Payments = append(records[i].Payments, RecordPaymentResponse{
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
		exceptions, err := h.queries.ListExceptionsByItem(ctx, id)
		if err != nil {
			return nil, err
		}
		records[i].Exceptions = make([]ExceptionResponse, 0, len(exceptions))
		for _, e := range exceptions {
			response := exceptionResponseFromListRow(e)
			response.TrackingNumber = records[i].TrackingNumber
			records[i].Exceptions = append(records[i].Exceptions, response)
		}
	}
	return records, nil
}

// validateRecordItem checks one item of a create request. It returns an empty
// string when the item is valid.
func validateRecordItem(item CreateRecordItem, hasDefaultDestination bool) string {
	switch {
	case strings.TrimSpace(item.Description) == "":
		return "description is required"
	case item.Quantity < 1:
		return "quantity must be at least 1"
	case strings.TrimSpace(item.Sender.FullName) == "":
		return "sender name is required"
	case strings.TrimSpace(item.Receiver.FullName) == "":
		return "receiver name is required"
	case !hasDefaultDestination && strings.TrimSpace(item.DestinationBranchID) == "":
		return "destination branch is required unless the request has one"
	case item.Amount < 0:
		return "amount cannot be negative"
	case item.AmountPaid < 0:
		return "amount_paid cannot be negative"
	case item.AmountPaid > item.Amount:
		return "amount_paid cannot exceed amount"
	}
	if item.AmountPaid > 0 {
		if method := strings.ToLower(strings.TrimSpace(item.PaymentMethod)); method != "" && !validPaymentMethods[method] {
			return "payment_method must be cash, bank_transfer, pos or other"
		}
	}
	if currency := strings.TrimSpace(item.Currency); currency != "" && len(currency) != 3 {
		return "currency must be a 3-letter code like NGN"
	}
	return ""
}

// resolveCustomer finds a customer by phone, or creates one. Customers are
// shared between branches, so a returning sender or receiver keeps its history.
func resolveCustomer(ctx context.Context, q *sqlc.Queries, p RecordParty) (uuid.UUID, error) {
	name := strings.TrimSpace(p.FullName)
	phone := strings.TrimSpace(p.Phone)
	if name == "" {
		return uuid.Nil, errors.New("name is required")
	}
	if phone == "" {
		created, err := q.CreateCustomer(ctx, sqlc.CreateCustomerParams{
			ID:           uuid.New(),
			FullName:     name,
			CustomerType: sqlc.CustomerTypeIndividual,
		})
		if err != nil {
			return uuid.Nil, err
		}
		return created.ID, nil
	}
	upserted, err := q.UpsertCustomerByPhone(ctx, sqlc.UpsertCustomerByPhoneParams{
		ID:           uuid.New(),
		FullName:     name,
		Phone:        sql.NullString{String: phone, Valid: true},
		CustomerType: sqlc.CustomerTypeIndividual,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return upserted.ID, nil
}

// recordWhere builds the WHERE clause for the record listing. Managers and
// staff are always limited to records touching their own branch.
func (h *RecordHandler) recordWhere(claims *auth.Claims, filter recordFilter) (string, []any) {
	args := make([]any, 0, 12)
	var where strings.Builder

	arg := func(value any) int {
		args = append(args, value)
		return len(args)
	}

	if len(filter.IDs) > 0 {
		ids := make([]string, 0, len(filter.IDs))
		for _, id := range filter.IDs {
			ids = append(ids, id.String())
		}
		n := arg(pq.Array(ids))
		fmt.Fprintf(&where, " AND mi.id::text = ANY($%d)", n)
	}
	if filter.Query != "" {
		n := arg("%" + filter.Query + "%")
		fmt.Fprintf(&where, " AND (mi.tracking_number ILIKE $%d OR COALESCE(mi.serial_number, '') ILIKE $%d OR mi.description ILIKE $%d OR sender.full_name ILIKE $%d OR COALESCE(sender.phone, '') ILIKE $%d OR receiver.full_name ILIKE $%d OR COALESCE(receiver.phone, '') ILIKE $%d)", n, n, n, n, n, n, n)
	}
	if len(filter.Statuses) > 0 {
		n := arg(pq.Array(filter.Statuses))
		fmt.Fprintf(&where, " AND mi.status::text = ANY($%d)", n)
	}
	if filter.PaymentStatus != "" {
		n := arg(filter.PaymentStatus)
		fmt.Fprintf(&where, " AND mi.payment_status::text = $%d", n)
	}
	if filter.BranchID != uuid.Nil {
		n := arg(filter.BranchID.String())
		fmt.Fprintf(&where, " AND (mi.origin_branch_id = $%d::uuid OR mi.destination_branch_id = $%d::uuid)", n, n)
	}
	if filter.OriginBranchID != uuid.Nil {
		n := arg(filter.OriginBranchID.String())
		fmt.Fprintf(&where, " AND mi.origin_branch_id = $%d::uuid", n)
	}
	if filter.DestinationBranchID != uuid.Nil {
		n := arg(filter.DestinationBranchID.String())
		fmt.Fprintf(&where, " AND mi.destination_branch_id = $%d::uuid", n)
	}
	if filter.From != nil {
		n := arg(*filter.From)
		fmt.Fprintf(&where, " AND mi.created_at >= $%d", n)
	}
	if filter.To != nil {
		n := arg(*filter.To)
		fmt.Fprintf(&where, " AND mi.created_at < ($%d::date + 1)", n)
	}
	if filter.Outstanding {
		where.WriteString(" AND mi.status <> 'cancelled' AND COALESCE(mi.balance, 0) > 0")
	}
	if filter.Mine {
		// Counter staff only ever care about the waybills they keyed in.
		if id, err := uuid.Parse(claims.UserID); err == nil {
			n := arg(id.String())
			fmt.Fprintf(&where, " AND mi.recorded_by = $%d::uuid", n)
		} else {
			where.WriteString(" AND FALSE")
		}
	}

	if !canSeeAllBranches(claims.Role) {
		if claims.BranchID == "" {
			where.WriteString(" AND FALSE")
		} else {
			n := arg(claims.BranchID)
			fmt.Fprintf(&where, " AND (mi.origin_branch_id = $%d::uuid OR mi.destination_branch_id = $%d::uuid)", n, n)
		}
	}

	return where.String(), args
}
