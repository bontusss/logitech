package handler

import (
	"context"
	"database/sql"
	"errors"
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
)

// ManifestHandler handles manual manifest creation.
type ManifestHandler struct {
	cfg     *config.Config
	db      *sql.DB
	queries *sqlc.Queries
}

// NewManifestHandler constructs a ManifestHandler. The pool is required to run
// the manifest + items insert inside a single transaction.
func NewManifestHandler(cfg *config.Config, queries *sqlc.Queries, pool *sql.DB) *ManifestHandler {
	return &ManifestHandler{cfg: cfg, queries: queries, db: pool}
}

// PartyPayload describes a sender or receiver of a manifest item. Customers
// are matched by phone when one is provided, otherwise a new record is created.
type PartyPayload struct {
	FullName string `json:"full_name" example:"Chinedu Okafor"`
	Phone    string `json:"phone,omitempty" example:"08031110001"`
}

// ManifestItemPayload is one line of a manifest.
type ManifestItemPayload struct {
	SerialNumber        string       `json:"serial_number,omitempty" example:"PKG-001"`
	Description         string       `json:"description" example:"Bags of rice"`
	Quantity            int          `json:"quantity" example:"5"`
	Sender              PartyPayload `json:"sender"`
	Receiver            PartyPayload `json:"receiver"`
	DestinationBranchID string       `json:"destination_branch_id" example:"11111111-1111-4111-8111-111111111111"`
	Amount              *float64     `json:"amount,omitempty" example:"15000.50"`
	Currency            string       `json:"currency,omitempty" example:"NGN"`
}

// CreateManifestPayload is the body of POST /api/manifests. ManifestNumber is
// optional: when it is empty the server picks a unique 5-digit number.
// DestinationBranchID is the receiving branch ("to"): when it is set, every
// item is stored with that branch as its destination; when it is empty, each
// item must bring its own destination_branch_id.
type CreateManifestPayload struct {
	ManifestNumber      string                `json:"manifest_number,omitempty" example:"10042"`
	OriginBranchID      string                `json:"origin_branch_id" example:"11111111-1111-4111-8111-111111111111"`
	DestinationBranchID string                `json:"destination_branch_id,omitempty" example:"22222222-2222-4222-8222-222222222222"`
	ManifestDate        string                `json:"manifest_date" example:"2026-10-27"`
	Status              string                `json:"status,omitempty" example:"draft"`
	Items               []ManifestItemPayload `json:"items"`
}

type ManifestItemEditPayload struct {
	ID string `json:"id,omitempty"`
	ManifestItemPayload
}

type UpdateManifestPayload struct {
	ManifestNumber string                    `json:"manifest_number"`
	OriginBranchID string                    `json:"origin_branch_id"`
	ManifestDate   string                    `json:"manifest_date"`
	Status         string                    `json:"status"`
	Items          []ManifestItemEditPayload `json:"items"`
}

type ManifestDetail struct {
	ID             string               `json:"id"`
	ManifestNumber string               `json:"manifest_number"`
	OriginBranchID string               `json:"origin_branch_id"`
	ManifestDate   string               `json:"manifest_date"`
	Status         string               `json:"status"`
	Items          []ManifestItemDetail `json:"items"`
}

type ManifestItemDetail struct {
	ID                  string       `json:"id"`
	SerialNumber        string       `json:"serial_number"`
	Description         string       `json:"description"`
	Quantity            int32        `json:"quantity"`
	Sender              PartyPayload `json:"sender"`
	Receiver            PartyPayload `json:"receiver"`
	DestinationBranchID string       `json:"destination_branch_id"`
	Amount              *float64     `json:"amount,omitempty"`
	Currency            string       `json:"currency"`
}

// CreateManifestResponse is returned after a manifest is saved.
type CreateManifestResponse struct {
	ID             string    `json:"id"`
	ManifestNumber string    `json:"manifest_number"`
	Status         string    `json:"status"`
	OriginBranchID string    `json:"origin_branch_id"`
	ManifestDate   string    `json:"manifest_date"`
	ItemCount      int       `json:"item_count"`
	CreatedAt      time.Time `json:"created_at"`
}

type UpdateManifestResponse struct {
	ID             string `json:"id"`
	ManifestNumber string `json:"manifest_number"`
	Status         string `json:"status"`
	OriginBranchID string `json:"origin_branch_id"`
	ManifestDate   string `json:"manifest_date"`
	ItemCount      int    `json:"item_count"`
}

// ManifestSummary is a row of GET /api/manifests.
type ManifestSummary struct {
	ID             string    `json:"id"`
	ManifestNumber string    `json:"manifest_number"`
	Status         string    `json:"status"`
	OriginBranchID string    `json:"origin_branch_id"`
	CreatedBy      string    `json:"created_by"`
	ManifestDate   string    `json:"manifest_date"`
	ItemCount      int       `json:"item_count"`
	ExceptionCount int       `json:"exception_count"`
	CreatedAt      time.Time `json:"created_at"`
}

// manifestStatuses that a manually created manifest may start with.
var allowedManifestStatuses = map[string]bool{
	"draft":             true,
	"confirmed":         true,
	"awaiting_dispatch": true,
}

// knownManifestStatuses is every value of the manifest_status enum; it bounds
// the statuses filter on GET /api/manifests.
var knownManifestStatuses = map[string]bool{
	"draft":              true,
	"extraction_pending": true,
	"review_required":    true,
	"confirmed":          true,
	"awaiting_dispatch":  true,
	"dispatched":         true,
	"partially_received": true,
	"received":           true,
	"cancelled":          true,
}

// fiveDigitManifestNumber is the shape of every manifest number created by
// this API: exactly five digits, e.g. 10042.
var fiveDigitManifestNumber = regexp.MustCompile(`^[0-9]{5}$`)

// manifestNumberTaken reports whether the number is already used at the branch
// (the schema enforces uniqueness per origin branch).
func (h *ManifestHandler) manifestNumberTaken(ctx context.Context, originBranchID uuid.UUID, number string) (bool, error) {
	var taken bool
	err := h.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM manifests WHERE origin_branch_id = $1 AND manifest_number = $2)`,
		originBranchID, number,
	).Scan(&taken)
	return taken, err
}

// nextManifestNumber returns a random 5-digit number that is free at the branch.
func (h *ManifestHandler) nextManifestNumber(ctx context.Context, originBranchID uuid.UUID) (string, error) {
	for attempt := 0; attempt < 25; attempt++ {
		candidate := strconv.Itoa(10000 + rand.IntN(90000))
		taken, err := h.manifestNumberTaken(ctx, originBranchID, candidate)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
	}
	return "", errors.New("no free 5-digit manifest number left for this branch")
}

// CreateManifest godoc
// @Summary Create a manifest with its items
// @Description Creates a manifest and all of its items in one transaction. When manifest_number is omitted the server generates a unique 5-digit number. When destination_branch_id (the "to" branch) is given, every item is stored with that destination; otherwise each item needs its own destination_branch_id. Sender and receiver customers are matched by phone or created on the fly.
// @Tags manifests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateManifestPayload true "Manifest with items"
// @Success 201 {object} CreateManifestResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/manifests [post]
func (h *ManifestHandler) CreateManifest(c *gin.Context) {
	var req CreateManifestPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}

	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}
	createdBy, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "only staff accounts can create manifests"})
		return
	}

	originBranchID, err := uuid.Parse(strings.TrimSpace(req.OriginBranchID))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "origin_branch_id must be a valid UUID"})
		return
	}
	if _, err := h.queries.GetBranchByID(c.Request.Context(), originBranchID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "origin branch not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to look up origin branch: " + err.Error()})
		return
	}

	// The manifest may carry a single receiving branch ("to"). When it does,
	// every item is stored with that destination; otherwise each item has to
	// name its own.
	destinationBranchID := uuid.Nil
	hasManifestDestination := strings.TrimSpace(req.DestinationBranchID) != ""
	if hasManifestDestination {
		parsed, err := uuid.Parse(strings.TrimSpace(req.DestinationBranchID))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "destination_branch_id must be a valid UUID"})
			return
		}
		if _, err := h.queries.GetBranchByID(c.Request.Context(), parsed); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "destination branch not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to look up destination branch: " + err.Error()})
			return
		}
		destinationBranchID = parsed
	}

	number := strings.TrimSpace(req.ManifestNumber)
	if number != "" && !fiveDigitManifestNumber.MatchString(number) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "manifest_number must be a 5-digit number like 10042"})
		return
	}
	if number == "" {
		generated, err := h.nextManifestNumber(c.Request.Context(), originBranchID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate a manifest number: " + err.Error()})
			return
		}
		number = generated
	} else {
		taken, err := h.manifestNumberTaken(c.Request.Context(), originBranchID, number)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to check the manifest number: " + err.Error()})
			return
		}
		if taken {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "manifest_number " + number + " is already used at this branch"})
			return
		}
	}

	manifestDate, err := parseManifestDate(req.ManifestDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "manifest_date must be YYYY-MM-DD"})
		return
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = string(sqlc.ManifestStatusDraft)
	}
	if !allowedManifestStatuses[status] {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "status must be draft, confirmed or awaiting_dispatch"})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "a manifest needs at least one item"})
		return
	}
	for i, item := range req.Items {
		if err := validateManifestItem(item, hasManifestDestination); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "item " + strconv.Itoa(i+1) + ": " + err.Error()})
			return
		}
	}

	ctx := c.Request.Context()
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to start transaction: " + err.Error()})
		return
	}
	defer func() { _ = tx.Rollback() }()
	txq := h.queries.WithTx(tx)

	manifest, err := txq.CreateManifest(ctx, sqlc.CreateManifestParams{
		ID:             uuid.New(),
		ManifestNumber: number,
		OriginBranchID: originBranchID,
		CreatedBy:      createdBy,
		ManifestDate:   manifestDate,
		Status:         sqlc.ManifestStatus(status),
		// confirmed_by is NOT NULL in the schema; the creator is recorded as
		// the confirmer for manual entries, mirroring the seed data.
		ConfirmedBy: createdBy,
	})
	if err != nil {
		// Backstop for a number that slipped past the pre-check.
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "manifest_number " + number + " is already used at this branch"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to create manifest: " + err.Error()})
		return
	}

	for _, item := range req.Items {
		senderID, err := resolveCustomer(ctx, txq, item.Sender)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save sender: " + err.Error()})
			return
		}
		receiverID, err := resolveCustomer(ctx, txq, item.Receiver)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save receiver: " + err.Error()})
			return
		}

		// A "to" branch on the manifest wins over whatever the item says.
		destBranchID := destinationBranchID
		if !hasManifestDestination {
			parsed, err := uuid.Parse(strings.TrimSpace(item.DestinationBranchID))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "destination_branch_id must be a valid UUID"})
				return
			}
			if _, err := txq.GetBranchByID(ctx, parsed); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					c.JSON(http.StatusBadRequest, ErrorResponse{Error: "destination branch not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to look up destination branch: " + err.Error()})
				return
			}
			destBranchID = parsed
		}

		quantity := item.Quantity
		if quantity < 1 {
			quantity = 1
		}
		currency := strings.TrimSpace(item.Currency)
		if currency == "" {
			currency = "NGN"
		}
		if len(currency) != 3 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "currency must be a 3-letter code like NGN"})
			return
		}

		var amount sql.NullString
		if item.Amount != nil {
			amount = sql.NullString{String: strconv.FormatFloat(*item.Amount, 'f', 2, 64), Valid: true}
		}

		if _, err := txq.CreateManifestItem(ctx, sqlc.CreateManifestItemParams{
			ID:                  uuid.New(),
			ManifestID:          manifest.ID,
			SerialNumber:        nullIfEmpty(item.SerialNumber),
			Description:         strings.TrimSpace(item.Description),
			Quantity:            int32(quantity),
			SenderCustomerID:    senderID,
			ReceiverCustomerID:  receiverID,
			DestinationBranchID: destBranchID,
			Amount:              amount,
			Currency:            strings.ToUpper(currency),
			PaymentStatus:       sqlc.PaymentStatusPending,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create manifest item: " + err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save manifest: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, CreateManifestResponse{
		ID:             manifest.ID.String(),
		ManifestNumber: manifest.ManifestNumber,
		Status:         string(manifest.Status),
		OriginBranchID: manifest.OriginBranchID.String(),
		ManifestDate:   manifest.ManifestDate.Format("2006-01-02"),
		ItemCount:      len(req.Items),
		CreatedAt:      manifest.CreatedAt,
	})
}

// GetManifest godoc
// @Summary Get a manifest
// @Tags manifests
// @Produce json
// @Security BearerAuth
// @Param id path string true "Manifest ID"
// @Success 200 {object} ManifestDetail
// @Failure 404 {object} ErrorResponse
// @Router /api/manifests/{id} [get]
func (h *ManifestHandler) GetManifest(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}

	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}
	manifestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "manifest id must be a valid UUID"})
		return
	}
	manifest, err := h.queries.GetManifestByID(c.Request.Context(), manifestID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "manifest not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load manifest: " + err.Error()})
		return
	}
	if !canViewAllManifests(claims.Role) && manifest.CreatedBy.String() != claims.UserID {
		visible := manifest.Status == sqlc.ManifestStatusReviewRequired
		if !visible {
			if err := h.db.QueryRowContext(c.Request.Context(), `SELECT EXISTS (SELECT 1 FROM exceptions WHERE manifest_id = $1)`, manifestID).Scan(&visible); err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to check manifest exceptions: " + err.Error()})
				return
			}
		}
		if !visible {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "manifest not found"})
			return
		}
	}

	rows, err := h.db.QueryContext(c.Request.Context(), `
		SELECT mi.id, COALESCE(mi.serial_number, ''), mi.description, mi.quantity,
		       sender.full_name, COALESCE(sender.phone, ''),
		       receiver.full_name, COALESCE(receiver.phone, ''),
		       mi.destination_branch_id, mi.amount::float8, mi.currency
		FROM manifest_items mi
		JOIN customers sender ON sender.id = mi.sender_customer_id
		JOIN customers receiver ON receiver.id = mi.receiver_customer_id
		WHERE mi.manifest_id = $1
		ORDER BY mi.created_at ASC`, manifestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load manifest items: " + err.Error()})
		return
	}
	defer rows.Close()

	items := make([]ManifestItemDetail, 0)
	for rows.Next() {
		var item ManifestItemDetail
		var amount sql.NullFloat64
		var destinationID uuid.UUID
		if err := rows.Scan(
			&item.ID, &item.SerialNumber, &item.Description, &item.Quantity,
			&item.Sender.FullName, &item.Sender.Phone,
			&item.Receiver.FullName, &item.Receiver.Phone,
			&destinationID, &amount, &item.Currency,
		); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read manifest items: " + err.Error()})
			return
		}
		item.DestinationBranchID = destinationID.String()
		if amount.Valid {
			item.Amount = &amount.Float64
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read manifest items: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, ManifestDetail{
		ID:             manifest.ID.String(),
		ManifestNumber: manifest.ManifestNumber,
		OriginBranchID: manifest.OriginBranchID.String(),
		ManifestDate:   manifest.ManifestDate.Format("2006-01-02"),
		Status:         string(manifest.Status),
		Items:          items,
	})
}

// UpdateManifest godoc
// @Summary Update an editable manifest and its items
// @Tags manifests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Manifest ID"
// @Param request body UpdateManifestPayload true "Updated manifest"
// @Success 200 {object} UpdateManifestResponse
// @Router /api/manifests/{id} [put]
func (h *ManifestHandler) UpdateManifest(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}
	manifestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "manifest id must be a valid UUID"})
		return
	}
	var req UpdateManifestPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "a manifest needs at least one item"})
		return
	}
	if !fiveDigitManifestNumber.MatchString(strings.TrimSpace(req.ManifestNumber)) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "manifest_number must be a 5-digit number like 10042"})
		return
	}
	manifestDate, err := parseManifestDate(req.ManifestDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "manifest_date must be YYYY-MM-DD"})
		return
	}
	status := strings.TrimSpace(req.Status)
	if !knownManifestStatuses[status] {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "unknown manifest status"})
		return
	}
	originBranchID, err := uuid.Parse(strings.TrimSpace(req.OriginBranchID))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "origin_branch_id must be a valid UUID"})
		return
	}

	tx, err := h.db.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to start transaction: " + err.Error()})
		return
	}
	defer func() { _ = tx.Rollback() }()
	ctx := c.Request.Context()
	var creatorID uuid.UUID
	var currentStatus string
	err = tx.QueryRowContext(ctx, `SELECT created_by, status::text FROM manifests WHERE id = $1 FOR UPDATE`, manifestID).Scan(&creatorID, &currentStatus)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "manifest not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load manifest: " + err.Error()})
		return
	}
	if !canManageAllManifests(claims.Role) && creatorID.String() != claims.UserID {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "manifest not found"})
		return
	}
	if _, err := h.queries.WithTx(tx).GetBranchByID(ctx, originBranchID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "origin branch not found"})
		return
	}
	var numberTaken bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM manifests WHERE origin_branch_id = $1 AND manifest_number = $2 AND id <> $3)`, originBranchID, strings.TrimSpace(req.ManifestNumber), manifestID).Scan(&numberTaken); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to check manifest number: " + err.Error()})
		return
	}
	if numberTaken {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "manifest number is already used at this branch"})
		return
	}

	itemIDs := make(map[uuid.UUID]bool, len(req.Items))
	for index, item := range req.Items {
		if err := validateManifestItem(item.ManifestItemPayload, false); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "item " + strconv.Itoa(index+1) + ": " + err.Error()})
			return
		}
		currency := strings.TrimSpace(item.Currency)
		if currency != "" && len(currency) != 3 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "currency must be a 3-letter code like NGN"})
			return
		}
		destinationID, err := uuid.Parse(strings.TrimSpace(item.DestinationBranchID))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "destination_branch_id must be a valid UUID"})
			return
		}
		if _, err := h.queries.WithTx(tx).GetBranchByID(ctx, destinationID); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "destination branch not found"})
			return
		}
		if item.ID != "" {
			itemID, err := uuid.Parse(item.ID)
			if err != nil || itemIDs[itemID] {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "item ids must be unique valid UUIDs"})
				return
			}
			itemIDs[itemID] = true
		}
	}

	if _, err := tx.ExecContext(ctx, `UPDATE manifests SET manifest_number = $2, origin_branch_id = $3, manifest_date = $4, status = $5, version = version + 1, updated_at = now() WHERE id = $1`, manifestID, strings.TrimSpace(req.ManifestNumber), originBranchID, manifestDate, status); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to update manifest: " + err.Error()})
		return
	}

	for _, item := range req.Items {
		var senderID, receiverID uuid.UUID
		if item.ID != "" {
			itemID, _ := uuid.Parse(item.ID)
			var exists bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM manifest_items WHERE id = $1 AND manifest_id = $2)`, itemID, manifestID).Scan(&exists); err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to verify manifest item: " + err.Error()})
				return
			} else if !exists {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "item does not belong to this manifest"})
				return
			}
		}
		senderID, err = resolveCustomer(ctx, h.queries.WithTx(tx), item.Sender)
		if err == nil {
			receiverID, err = resolveCustomer(ctx, h.queries.WithTx(tx), item.Receiver)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save item contacts: " + err.Error()})
			return
		}

		destinationID, _ := uuid.Parse(strings.TrimSpace(item.DestinationBranchID))
		amount := sql.NullString{}
		if item.Amount != nil {
			amount = sql.NullString{String: strconv.FormatFloat(*item.Amount, 'f', 2, 64), Valid: true}
		}
		currency := strings.ToUpper(strings.TrimSpace(item.Currency))
		if currency == "" {
			currency = "NGN"
		}
		quantity := item.Quantity
		if quantity < 1 {
			quantity = 1
		}
		if item.ID == "" {
			itemID := uuid.New()
			itemIDs[itemID] = true
			_, err = h.queries.WithTx(tx).CreateManifestItem(ctx, sqlc.CreateManifestItemParams{
				ID: itemID, ManifestID: manifestID, SerialNumber: nullIfEmpty(item.SerialNumber),
				Description: strings.TrimSpace(item.Description), Quantity: int32(quantity),
				SenderCustomerID: senderID, ReceiverCustomerID: receiverID,
				DestinationBranchID: destinationID, Amount: amount, Currency: currency,
				PaymentStatus: sqlc.PaymentStatusPending,
			})
		} else {
			itemID, _ := uuid.Parse(item.ID)
			var result sql.Result
			result, err = tx.ExecContext(ctx, `UPDATE manifest_items SET serial_number = $3, description = $4, quantity = $5, sender_customer_id = $6, receiver_customer_id = $7, destination_branch_id = $8, amount = $9, currency = $10, updated_at = now() WHERE id = $1 AND manifest_id = $2`, itemID, manifestID, nullIfEmpty(item.SerialNumber), strings.TrimSpace(item.Description), quantity, senderID, receiverID, destinationID, amount, currency)
			if err == nil {
				var changed int64
				changed, err = result.RowsAffected()
				if err == nil && changed == 0 {
					err = errors.New("item does not belong to this manifest")
				}
			}
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save manifest item: " + err.Error()})
			return
		}
	}

	rows, err := tx.QueryContext(ctx, `SELECT id FROM manifest_items WHERE manifest_id = $1`, manifestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list manifest items: " + err.Error()})
		return
	}
	var removedIDs []uuid.UUID
	for rows.Next() {
		var itemID uuid.UUID
		if err := rows.Scan(&itemID); err != nil {
			rows.Close()
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read manifest items: " + err.Error()})
			return
		}
		if !itemIDs[itemID] {
			removedIDs = append(removedIDs, itemID)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read manifest items: " + err.Error()})
		return
	}
	rows.Close()
	for _, itemID := range removedIDs {
		result, err := tx.ExecContext(ctx, `DELETE FROM manifest_items mi WHERE mi.id = $1 AND NOT EXISTS (SELECT 1 FROM payments p WHERE p.manifest_item_id = mi.id) AND NOT EXISTS (SELECT 1 FROM receipts r WHERE r.manifest_item_id = mi.id) AND NOT EXISTS (SELECT 1 FROM exceptions e WHERE e.manifest_item_id = mi.id)`, itemID)
		if err != nil {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "cannot remove an item with payment, receipt, or exception history"})
			return
		}
		changed, err := result.RowsAffected()
		if err != nil || changed == 0 {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "cannot remove an item with payment, receipt, or exception history"})
			return
		}
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save manifest: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, UpdateManifestResponse{
		ID: manifestID.String(), ManifestNumber: strings.TrimSpace(req.ManifestNumber),
		Status: status, OriginBranchID: originBranchID.String(),
		ManifestDate: manifestDate.Format("2006-01-02"), ItemCount: len(req.Items),
	})
}

func canManageAllManifests(role string) bool {
	return role == auth.RoleManager
}

func canViewAllManifests(role string) bool {
	return role == auth.RoleManager || role == auth.RoleOwner || role == auth.RoleSystemAdmin
}

// ListAttentionManifests returns review-required manifests with at least one exception.
// @Summary List manifests needing attention
// @Tags manifests
// @Produce json
// @Security BearerAuth
// @Success 200 {array} ManifestSummary
// @Router /api/manifests/attention [get]
func (h *ManifestHandler) ListAttentionManifests(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	if _, ok := auth.GetClaims(c); !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(), `
		SELECT m.id, m.manifest_number, m.status::text, m.origin_branch_id,
		       m.created_by, m.manifest_date, m.created_at,
		       (SELECT COUNT(*) FROM manifest_items mi WHERE mi.manifest_id = m.id),
		       (SELECT COUNT(*) FROM exceptions e WHERE e.manifest_id = m.id)
		FROM manifests m
		WHERE m.status = 'review_required'
		   OR EXISTS (SELECT 1 FROM exceptions e WHERE e.manifest_id = m.id)
		ORDER BY m.created_at DESC
		LIMIT 50`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list manifests needing attention: " + err.Error()})
		return
	}
	defer rows.Close()

	results := make([]ManifestSummary, 0)
	for rows.Next() {
		var result ManifestSummary
		var id, originBranchID, createdBy uuid.UUID
		var manifestDate, createdAt time.Time
		if err := rows.Scan(
			&id, &result.ManifestNumber, &result.Status, &originBranchID,
			&createdBy, &manifestDate, &createdAt, &result.ItemCount,
			&result.ExceptionCount,
		); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read manifests needing attention: " + err.Error()})
			return
		}
		result.ID = id.String()
		result.OriginBranchID = originBranchID.String()
		result.CreatedBy = createdBy.String()
		result.ManifestDate = manifestDate.Format("2006-01-02")
		result.CreatedAt = createdAt
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read manifests needing attention: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// ListManifests godoc
// @Summary List manifests
// @Description Returns recent manifests with their item counts, newest first. Staff only ever see manifests they created; managers, owners and the system admin see everyone's. Optionally filter by creator (a staff UUID, or "me") and by statuses (comma-separated).
// @Tags manifests
// @Produce json
// @Security BearerAuth
// @Param created_by query string false "staff UUID, or me"
// @Param statuses query string false "comma-separated manifest statuses, e.g. draft,review_required"
// @Success 200 {array} ManifestSummary
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/manifests [get]
func (h *ManifestHandler) ListManifests(c *gin.Context) {
	if h.queries == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}

	ctx := c.Request.Context()
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	createdBy := strings.TrimSpace(c.Query("created_by"))
	statuses := strings.TrimSpace(c.Query("statuses"))

	// Staff only ever see the manifests they created: for them the creator
	// filter is forced to their own account and any created_by parameter is
	// ignored. Managers, owners and the system admin see everyone's manifests
	// and may pick whose to list.
	canSeeAll := claims.Role == auth.RoleManager || claims.Role == auth.RoleOwner || claims.Role == auth.RoleSystemAdmin
	if !canSeeAll {
		createdBy = claims.UserID
	} else if createdBy == "me" {
		createdBy = claims.UserID
	}

	var rows []sqlc.Manifest
	var err error
	if createdBy == "" && statuses == "" {
		rows, err = h.queries.ListManifests(ctx, sqlc.ListManifestsParams{Limit: 50, Offset: 0})
	} else {
		if createdBy == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "created_by is required when filtering by statuses"})
			return
		}
		creatorID, parseErr := uuid.Parse(createdBy)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "created_by must be a valid UUID or me"})
			return
		}
		cleanStatuses := make([]string, 0, 4)
		for _, s := range strings.Split(statuses, ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			if !knownManifestStatuses[s] {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "unknown status: " + s})
				return
			}
			cleanStatuses = append(cleanStatuses, s)
		}
		if statuses != "" && len(cleanStatuses) == 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "statuses must name at least one manifest status"})
			return
		}
		rows, err = h.queries.ListManifestsByCreator(ctx, sqlc.ListManifestsByCreatorParams{
			CreatedBy: creatorID,
			Statuses:  strings.Join(cleanStatuses, ","),
			Limit:     50,
			Offset:    0,
		})
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list manifests: " + err.Error()})
		return
	}

	results := make([]ManifestSummary, 0, len(rows))
	for _, m := range rows {
		items, err := h.queries.ListManifestItems(ctx, m.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list manifest items: " + err.Error()})
			return
		}
		var exceptionCount int
		if err := h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM exceptions WHERE manifest_id = $1`, m.ID).Scan(&exceptionCount); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to count manifest exceptions: " + err.Error()})
			return
		}
		results = append(results, ManifestSummary{
			ID:             m.ID.String(),
			ManifestNumber: m.ManifestNumber,
			Status:         string(m.Status),
			OriginBranchID: m.OriginBranchID.String(),
			CreatedBy:      m.CreatedBy.String(),
			ManifestDate:   m.ManifestDate.Format("2006-01-02"),
			ItemCount:      len(items),
			ExceptionCount: exceptionCount,
			CreatedAt:      m.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, results)
}

// validateManifestItem checks the fields of a single item payload. The
// destination branch may come from the manifest instead of the item.
func validateManifestItem(item ManifestItemPayload, hasManifestDestination bool) error {
	if strings.TrimSpace(item.Description) == "" {
		return errors.New("description is required")
	}
	if item.Quantity < 0 {
		return errors.New("quantity cannot be negative")
	}
	if strings.TrimSpace(item.Sender.FullName) == "" {
		return errors.New("sender name is required")
	}
	if strings.TrimSpace(item.Receiver.FullName) == "" {
		return errors.New("receiver name is required")
	}
	if !hasManifestDestination && strings.TrimSpace(item.DestinationBranchID) == "" {
		return errors.New("destination branch is required unless the manifest has a \"to\" branch")
	}
	if item.Amount != nil && *item.Amount < 0 {
		return errors.New("amount cannot be negative")
	}
	return nil
}

// resolveCustomer finds a customer by phone or creates a new one.
func resolveCustomer(ctx context.Context, q *sqlc.Queries, p PartyPayload) (uuid.UUID, error) {
	name := strings.TrimSpace(p.FullName)
	phone := strings.TrimSpace(p.Phone)

	if phone != "" {
		existing, err := q.GetCustomerByPhone(ctx, nullIfEmpty(phone))
		if err == nil {
			return existing.ID, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, err
		}
	}

	created, err := q.CreateCustomer(ctx, sqlc.CreateCustomerParams{
		ID:           uuid.New(),
		FullName:     name,
		Phone:        nullIfEmpty(phone),
		CustomerType: sqlc.CustomerTypeIndividual,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return created.ID, nil
}

// parseManifestDate parses a YYYY-MM-DD date, defaulting to today.
func parseManifestDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	return time.Parse("2006-01-02", value)
}
