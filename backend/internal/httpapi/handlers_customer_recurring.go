package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	db "github.com/bol-platform/backend/internal/db/sqlc"
)

type customerRecurringShipmentDTO struct {
	ID            uuid.UUID       `json:"id"`
	Label         string          `json:"label"`
	Frequency     string          `json:"frequency"`
	SearchPayload json.RawMessage `json:"search_payload"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type customerRecurringShipmentInput struct {
	Label     string          `json:"label"`
	Frequency string          `json:"frequency"`
	Search    json.RawMessage `json:"search_payload"`
}

func customerRecurringShipmentDTOFromRow(row db.CustomerRecurringShipment) customerRecurringShipmentDTO {
	payload := json.RawMessage(row.SearchPayload)
	if len(payload) == 0 {
		payload = json.RawMessage("{}")
	}
	return customerRecurringShipmentDTO{
		ID:            row.ID,
		Label:         row.Label,
		Frequency:     row.Frequency,
		SearchPayload: payload,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func (a *API) listCustomerRecurringShipments(c *gin.Context) {
	user, _ := currentUser(c)
	rows, err := a.svc.Q.ListCustomerRecurringShipments(c.Request.Context(), user.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]customerRecurringShipmentDTO, 0, len(rows))
	for _, row := range rows {
		result = append(result, customerRecurringShipmentDTOFromRow(row))
	}
	c.JSON(http.StatusOK, gin.H{"recurring_shipments": result})
}

func (a *API) createCustomerRecurringShipment(c *gin.Context) {
	user, _ := currentUser(c)
	var input customerRecurringShipmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, "invalid_recurring_shipment", "Send the shipment details to repeat.")
		return
	}

	input.Label = strings.TrimSpace(input.Label)
	if input.Label == "" {
		input.Label = "Repeating shipment"
	}
	if len([]rune(input.Label)) > 80 {
		fail(c, http.StatusBadRequest, "invalid_recurring_shipment", "Keep the name under 80 characters.")
		return
	}
	if input.Frequency == "" {
		input.Frequency = "monthly"
	}
	if input.Frequency != "weekly" && input.Frequency != "biweekly" && input.Frequency != "monthly" {
		fail(c, http.StatusBadRequest, "invalid_recurring_shipment", "Choose weekly, every two weeks or monthly.")
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(input.Search, &payload); err != nil || payload == nil {
		fail(c, http.StatusBadRequest, "invalid_recurring_shipment", "The saved shipment details are not valid.")
		return
	}
	destinationZone, _ := payload["destination_zone_id"].(string)
	weight, _ := payload["weight_kg"].(float64)
	if strings.TrimSpace(destinationZone) == "" || weight <= 0 {
		fail(c, http.StatusBadRequest, "invalid_recurring_shipment", "Repeat a completed rate search so the route and package are saved.")
		return
	}

	ctx := c.Request.Context()
	existing, err := a.svc.Q.ListCustomerRecurringShipments(ctx, user.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if len(existing) >= 20 {
		fail(c, http.StatusBadRequest, "recurring_shipment_limit", "You can save up to 20 repeating shipments. Remove one to add another.")
		return
	}

	row, err := a.svc.Q.CreateCustomerRecurringShipment(ctx, db.CreateCustomerRecurringShipmentParams{
		CustomerID:    user.ID,
		Label:         input.Label,
		Frequency:     input.Frequency,
		SearchPayload: []byte(input.Search),
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"recurring_shipment": customerRecurringShipmentDTOFromRow(row)})
}

func (a *API) deleteCustomerRecurringShipment(c *gin.Context) {
	user, _ := currentUser(c)
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_recurring_shipment", "That repeating shipment id is invalid.")
		return
	}
	deleted, err := a.svc.Q.DeleteCustomerRecurringShipment(c.Request.Context(), db.DeleteCustomerRecurringShipmentParams{
		ID:         id,
		CustomerID: user.ID,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if deleted == 0 {
		fail(c, http.StatusNotFound, "recurring_shipment_not_found", "That repeating shipment could not be found.")
		return
	}
	c.Status(http.StatusNoContent)
}
