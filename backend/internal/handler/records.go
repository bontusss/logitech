package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"backend/internal/auth"

	"github.com/gin-gonic/gin"
)

type RecordManifest struct {
	ID             string       `json:"id"`
	ManifestNumber string       `json:"manifest_number"`
	Status         string       `json:"status"`
	ManifestDate   string       `json:"manifest_date"`
	OriginBranch   string       `json:"origin_branch"`
	CreatedAt      time.Time    `json:"created_at"`
	Items          []RecordItem `json:"items"`
}

type RecordItem struct {
	ID                string `json:"id"`
	WaybillNumber     string `json:"waybill_number"`
	Description       string `json:"description"`
	Quantity          int64  `json:"quantity"`
	SenderName        string `json:"sender_name"`
	SenderPhone       string `json:"sender_phone"`
	ReceiverName      string `json:"receiver_name"`
	ReceiverPhone     string `json:"receiver_phone"`
	DestinationBranch string `json:"destination_branch"`
}

type RecordDetail struct {
	RecordManifest
	CompletedAt *time.Time `json:"completed_at"`
}

type RecordReceipt struct {
	ID               string    `json:"id"`
	ReceivedAt       time.Time `json:"received_at"`
	ReceivedQuantity int64     `json:"received_quantity"`
	Condition        string    `json:"condition"`
	ReceivedBy       string    `json:"received_by"`
}

type RecordPayment struct {
	ID          string    `json:"id"`
	Reference   string    `json:"reference"`
	Method      string    `json:"method"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	PaidAt      time.Time `json:"paid_at"`
	CollectedBy string    `json:"collected_by"`
}

type RecordItemDetail struct {
	RecordItem
	Receipts []RecordReceipt `json:"receipts"`
	Payments []RecordPayment `json:"payments"`
}

type RecordDetailResponse struct {
	ID             string             `json:"id"`
	ManifestNumber string             `json:"manifest_number"`
	Status         string             `json:"status"`
	ManifestDate   string             `json:"manifest_date"`
	OriginBranch   string             `json:"origin_branch"`
	CreatedAt      time.Time          `json:"created_at"`
	CompletedAt    *time.Time         `json:"completed_at"`
	Items          []RecordItemDetail `json:"items"`
}

// SearchRecords lists completed manifests, filtering on manifest number, item waybill number, or either party's phone.
// @Summary Search manifests and items
// @Tags manifests
// @Produce json
// @Security BearerAuth
// @Param q query string false "Manifest number, waybill number, or phone"
// @Success 200 {array} RecordManifest
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/records [get]
func (h *ManifestHandler) SearchRecords(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	if _, ok := auth.GetClaims(c); !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	search := strings.TrimSpace(c.Query("q"))
	rows, err := h.db.QueryContext(c.Request.Context(), `
		SELECT m.id::text, m.manifest_number, m.status::text, m.manifest_date,
		       m.created_at, COALESCE(origin.name, ''),
		       mi.id::text, COALESCE(mi.serial_number, ''), COALESCE(mi.description, ''),
		       COALESCE(mi.quantity, 0), COALESCE(sender.full_name, ''),
		       COALESCE(sender.phone, ''), COALESCE(receiver.full_name, ''),
		       COALESCE(receiver.phone, ''), COALESCE(destination.name, '')
		FROM manifests m
		LEFT JOIN branches origin ON origin.id = m.origin_branch_id
		LEFT JOIN manifest_items mi ON mi.manifest_id = m.id
		LEFT JOIN customers sender ON sender.id = mi.sender_customer_id
		LEFT JOIN customers receiver ON receiver.id = mi.receiver_customer_id
		LEFT JOIN branches destination ON destination.id = mi.destination_branch_id
		WHERE m.status = 'received'
		  AND (
		       $1 = ''
		       OR m.manifest_number ILIKE '%' || $1 || '%'
		       OR EXISTS (
				SELECT 1
				FROM manifest_items searched_item
				LEFT JOIN customers searched_sender ON searched_sender.id = searched_item.sender_customer_id
				LEFT JOIN customers searched_receiver ON searched_receiver.id = searched_item.receiver_customer_id
				WHERE searched_item.manifest_id = m.id
				  AND (COALESCE(searched_item.serial_number, '') ILIKE '%' || $1 || '%'
				       OR COALESCE(searched_sender.phone, '') ILIKE '%' || $1 || '%'
				       OR COALESCE(searched_receiver.phone, '') ILIKE '%' || $1 || '%')
		       )
		  )
		ORDER BY m.created_at DESC, mi.created_at ASC`, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to search records: " + err.Error()})
		return
	}
	defer rows.Close()

	results := make([]RecordManifest, 0)
	manifestIndex := make(map[string]int)
	for rows.Next() {
		var manifest RecordManifest
		var manifestDate time.Time
		var itemID, waybillNumber sql.NullString
		var description sql.NullString
		var quantity sql.NullInt64
		var senderName, senderPhone, receiverName, receiverPhone, destinationBranch string
		if err := rows.Scan(
			&manifest.ID, &manifest.ManifestNumber, &manifest.Status, &manifestDate,
			&manifest.CreatedAt, &manifest.OriginBranch, &itemID, &waybillNumber,
			&description, &quantity, &senderName, &senderPhone, &receiverName,
			&receiverPhone, &destinationBranch,
		); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read records: " + err.Error()})
			return
		}
		manifest.ManifestDate = manifestDate.Format("2006-01-02")
		index, exists := manifestIndex[manifest.ID]
		if !exists {
			manifest.Items = make([]RecordItem, 0)
			results = append(results, manifest)
			index = len(results) - 1
			manifestIndex[manifest.ID] = index
		}
		if itemID.Valid {
			results[index].Items = append(results[index].Items, RecordItem{
				ID: itemID.String, WaybillNumber: waybillNumber.String,
				Description: description.String, Quantity: quantity.Int64,
				SenderName: senderName, SenderPhone: senderPhone,
				ReceiverName: receiverName, ReceiverPhone: receiverPhone,
				DestinationBranch: destinationBranch,
			})
		}
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read records: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// GetRecord returns the full receipt and payment history for a received manifest.
// @Summary Get a completed manifest record
// @Tags manifests
// @Produce json
// @Security BearerAuth
// @Param id path string true "Manifest ID"
// @Success 200 {object} RecordDetailResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/records/{id} [get]
func (h *ManifestHandler) GetRecord(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database is unavailable"})
		return
	}
	if _, ok := auth.GetClaims(c); !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}
	manifestID := strings.TrimSpace(c.Param("id"))
	var result RecordDetailResponse
	var manifestDate time.Time
	var completedAt sql.NullTime
	err := h.db.QueryRowContext(c.Request.Context(), `
		SELECT m.id::text, m.manifest_number, m.status::text, m.manifest_date,
		       m.created_at, COALESCE(origin.name, ''),
		       (SELECT MAX(receipt.received_at)
		        FROM receipts receipt
		        JOIN manifest_items mi ON mi.id = receipt.manifest_item_id
		        WHERE mi.manifest_id = m.id)
		FROM manifests m
		LEFT JOIN branches origin ON origin.id = m.origin_branch_id
		WHERE m.id::text = $1 AND m.status = 'received'`, manifestID).Scan(
		&result.ID, &result.ManifestNumber, &result.Status, &manifestDate,
		&result.CreatedAt, &result.OriginBranch, &completedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "completed manifest not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load completed manifest: " + err.Error()})
		return
	}
	result.ManifestDate = manifestDate.Format("2006-01-02")
	if completedAt.Valid {
		result.CompletedAt = &completedAt.Time
	}

	rows, err := h.db.QueryContext(c.Request.Context(), `
		SELECT mi.id::text, COALESCE(mi.serial_number, ''), COALESCE(mi.description, ''),
		       mi.quantity, COALESCE(sender.full_name, ''), COALESCE(sender.phone, ''),
		       COALESCE(receiver.full_name, ''), COALESCE(receiver.phone, ''),
		       COALESCE(destination.name, '')
		FROM manifest_items mi
		LEFT JOIN customers sender ON sender.id = mi.sender_customer_id
		LEFT JOIN customers receiver ON receiver.id = mi.receiver_customer_id
		LEFT JOIN branches destination ON destination.id = mi.destination_branch_id
		WHERE mi.manifest_id::text = $1
		ORDER BY mi.created_at ASC`, manifestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load completed manifest items: " + err.Error()})
		return
	}
	defer rows.Close()
	result.Items = make([]RecordItemDetail, 0)
	for rows.Next() {
		var item RecordItemDetail
		if err := rows.Scan(
			&item.ID, &item.WaybillNumber, &item.Description, &item.Quantity,
			&item.SenderName, &item.SenderPhone, &item.ReceiverName,
			&item.ReceiverPhone, &item.DestinationBranch,
		); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read completed manifest items: " + err.Error()})
			return
		}
		item.Receipts, err = h.listRecordReceipts(c, item.ID)
		if err != nil {
			return
		}
		item.Payments, err = h.listRecordPayments(c, item.ID)
		if err != nil {
			return
		}
		result.Items = append(result.Items, item)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read completed manifest items: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ManifestHandler) listRecordReceipts(c *gin.Context, itemID string) ([]RecordReceipt, error) {
	rows, err := h.db.QueryContext(c.Request.Context(), `
		SELECT receipt.id::text, receipt.received_at, receipt.received_quantity,
		       receipt.condition::text, COALESCE(staff.full_name, '')
		FROM receipts receipt
		LEFT JOIN staff ON staff.id = receipt.received_by
		WHERE receipt.manifest_item_id::text = $1
		ORDER BY receipt.received_at ASC`, itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load receipts: " + err.Error()})
		return nil, err
	}
	defer rows.Close()
	results := make([]RecordReceipt, 0)
	for rows.Next() {
		var receipt RecordReceipt
		if err := rows.Scan(&receipt.ID, &receipt.ReceivedAt, &receipt.ReceivedQuantity, &receipt.Condition, &receipt.ReceivedBy); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read receipts: " + err.Error()})
			return nil, err
		}
		results = append(results, receipt)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read receipts: " + err.Error()})
		return nil, err
	}
	return results, nil
}

func (h *ManifestHandler) listRecordPayments(c *gin.Context, itemID string) ([]RecordPayment, error) {
	rows, err := h.db.QueryContext(c.Request.Context(), `
		SELECT payment.id::text, COALESCE(payment.reference, ''), payment.method::text,
		       payment.amount::float8, payment.currency, payment.paid_at,
		       COALESCE(staff.full_name, '')
		FROM payments payment
		LEFT JOIN staff ON staff.id = payment.recorded_by
		WHERE payment.manifest_item_id::text = $1
		ORDER BY payment.paid_at ASC`, itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load payments: " + err.Error()})
		return nil, err
	}
	defer rows.Close()
	results := make([]RecordPayment, 0)
	for rows.Next() {
		var payment RecordPayment
		if err := rows.Scan(&payment.ID, &payment.Reference, &payment.Method, &payment.Amount, &payment.Currency, &payment.PaidAt, &payment.CollectedBy); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read payments: " + err.Error()})
			return nil, err
		}
		results = append(results, payment)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read payments: " + err.Error()})
		return nil, err
	}
	return results, nil
}
