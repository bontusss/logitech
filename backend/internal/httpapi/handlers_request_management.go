package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/service"
)

func (a *API) listBookingRequests(c *gin.Context) {
	tenant, _ := currentTenant(c)
	rows, err := a.svc.Q.ListCompanyRequests(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]deliveryDTO, 0, len(rows))
	for _, row := range rows {
		result = append(result, deliveryDTO{
			ID: row.ID, SenderName: row.SenderName, SenderPhone: row.SenderPhone,
			ReceiverName: row.ReceiverName, ReceiverPhone: row.ReceiverPhone,
			PickupAddress: row.PickupAddress, DropoffAddress: row.DropoffAddress,
			ItemDescription: row.ItemDescription, DeclaredValueMinor: row.DeclaredValueMinor,
			DeliveryPriceMinor: row.DeliveryPriceMinor, Payer: row.Payer, IsPaid: row.IsPaid,
			PaymentMethod: row.PaymentMethod, AssignedDriverID: row.AssignedDriverID,
			AssignedDriverName: row.AssignedDriverName, PackageCategory: row.PackageCategory,
			WeightKg: numericToFloat64(row.WeightKg), AdditionalInstructions: row.AdditionalInstructions,
			PackagePhotoURL: row.PackagePhotoUrl, PriceIsEstimate: row.PriceIsEstimate,
			BookingSource: row.BookingSource, DeliverySpeed: row.DeliverySpeed,
			DistanceKm: numericToFloat64(row.DistanceKm), Status: row.Status,
			RequestStatus: row.RequestStatus, TrackingCode: row.TrackingCode,
			RiderLinkToken: row.RiderLinkToken, CreatedAt: row.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"requests": result})
}

func (a *API) acceptBookingRequest(c *gin.Context) {
	tenant, _ := currentTenant(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_request", "That request id is invalid.")
		return
	}
	var req struct {
		DriverID string `json:"driver_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "driver_required", "Choose a driver when accepting the request.")
		return
	}
	driverID, err := uuid.Parse(strings.TrimSpace(req.DriverID))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_driver", "Choose a valid active driver.")
		return
	}

	err = pgx.BeginFunc(c.Request.Context(), a.svc.DB, func(tx pgx.Tx) error {
		q := a.svc.Q.WithTx(tx)
		delivery, err := q.AcceptBookingRequest(c.Request.Context(), db.AcceptBookingRequestParams{CompanyID: tenant.CompanyID, ID: id})
		if err != nil {
			return err
		}
		if _, err := tx.Exec(c.Request.Context(), `UPDATE deliveries SET assigned_driver_id=$3, status='assigned', updated_at=now() WHERE id=$1 AND company_id=$2`, id, tenant.CompanyID, driverID); err != nil {
			return err
		}
		if _, err := q.AddDeliveryTimelineEvent(c.Request.Context(), db.AddDeliveryTimelineEventParams{DeliveryID: id, CompanyID: tenant.CompanyID, Status: "assigned", Note: stringPointer("Request accepted and driver assigned"), CreatedBy: actorIDPointer(c)}); err != nil {
			return err
		}
		_ = delivery
		return nil
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusConflict, "request_not_pending", "This request is no longer pending.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	delivery, err := a.svc.Q.GetCompanyDelivery(c.Request.Context(), db.GetCompanyDeliveryParams{CompanyID: tenant.CompanyID, ID: id})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	trackingCode := ""
	if delivery.TrackingCode != nil {
		trackingCode = *delivery.TrackingCode
	}
	a.svc.Audit(c.Request.Context(), service.AuditEntry{ActorID: actorID(c), CompanyID: &tenant.CompanyID, Action: "company.booking_request_accepted", EntityType: "delivery", EntityID: &id, Metadata: map[string]any{"driver_id": driverID, "tracking_code": trackingCode}, IP: c.ClientIP(), UserAgent: c.Request.UserAgent()})
	c.JSON(http.StatusOK, gin.H{"delivery": deliveryAcceptanceDTO(delivery, trackingCode, a.cfg.AppURL)})
}

func stringPointer(v string) *string { return &v }
func actorIDPointer(c *gin.Context) *uuid.UUID {
	id := actorID(c)
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return id
}

func deliveryAcceptanceDTO(row db.GetCompanyDeliveryRow, trackingCode, appURL string) gin.H {
	trackingURL := strings.TrimRight(appURL, "/") + "/track/" + trackingCode
	return gin.H{"id": row.ID, "tracking_code": trackingCode, "tracking_url": trackingURL, "status": "assigned", "request_status": "accepted", "assigned_driver_id": row.AssignedDriverID}
}

func (a *API) declineBookingRequest(c *gin.Context) {
	tenant, _ := currentTenant(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_request", "That request id is invalid.")
		return
	}
	_, err = a.svc.Q.DeclineBookingRequest(c.Request.Context(), db.DeclineBookingRequestParams{CompanyID: tenant.CompanyID, ID: id})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusConflict, "request_not_pending", "This request is no longer pending.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	_, _ = a.svc.Q.AddDeliveryTimelineEvent(c.Request.Context(), db.AddDeliveryTimelineEventParams{DeliveryID: id, CompanyID: tenant.CompanyID, Status: "declined", Note: stringPointer("Booking request declined"), CreatedBy: actorIDPointer(c)})
	c.JSON(http.StatusOK, gin.H{"declined": true})
}

func (a *API) getDelivery(c *gin.Context) {
	tenant, _ := currentTenant(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_delivery", "That delivery id is invalid.")
		return
	}
	row, err := a.svc.Q.GetCompanyDelivery(c.Request.Context(), db.GetCompanyDeliveryParams{CompanyID: tenant.CompanyID, ID: id})
	if err != nil {
		fail(c, http.StatusNotFound, "delivery_not_found", "That delivery was not found.")
		return
	}
	trackingURL := (*string)(nil)
	if row.TrackingCode != nil {
		value := strings.TrimRight(a.cfg.AppURL, "/") + "/track/" + *row.TrackingCode
		trackingURL = &value
	}
	c.JSON(http.StatusOK, gin.H{"delivery": gin.H{"id": row.ID, "sender_name": row.SenderName, "sender_phone": row.SenderPhone, "receiver_name": row.ReceiverName, "receiver_phone": row.ReceiverPhone, "pickup_address": row.PickupAddress, "dropoff_address": row.DropoffAddress, "item_description": row.ItemDescription, "declared_value_minor": row.DeclaredValueMinor, "delivery_price_minor": row.DeliveryPriceMinor, "assigned_driver_id": row.AssignedDriverID, "assigned_driver_name": row.AssignedDriverName, "package_category": row.PackageCategory, "weight_kg": numericToFloat64(row.WeightKg), "additional_instructions": row.AdditionalInstructions, "status": row.Status, "request_status": row.RequestStatus, "tracking_code": row.TrackingCode, "tracking_url": trackingURL, "created_at": row.CreatedAt}})
}

func (a *API) updateDelivery(c *gin.Context) {
	tenant, _ := currentTenant(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_delivery", "That delivery id is invalid.")
		return
	}
	var req struct {
		SenderName             string  `json:"sender_name"`
		SenderPhone            string  `json:"sender_phone"`
		ReceiverName           string  `json:"receiver_name"`
		ReceiverPhone          string  `json:"receiver_phone"`
		PickupAddress          string  `json:"pickup_address"`
		DropoffAddress         string  `json:"dropoff_address"`
		ItemDescription        string  `json:"item_description"`
		DeliveryPriceMinor     int64   `json:"delivery_price_minor"`
		DeclaredValueMinor     int64   `json:"declared_value_minor"`
		PackageCategory        string  `json:"package_category"`
		WeightKg               float64 `json:"weight_kg"`
		AdditionalInstructions string  `json:"additional_instructions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_delivery", "Check the delivery details.")
		return
	}
	updated, err := a.svc.Q.UpdateCompanyDelivery(c.Request.Context(), db.UpdateCompanyDeliveryParams{CompanyID: tenant.CompanyID, ID: id, SenderName: strings.TrimSpace(req.SenderName), SenderPhone: strings.TrimSpace(req.SenderPhone), ReceiverName: strings.TrimSpace(req.ReceiverName), ReceiverPhone: strings.TrimSpace(req.ReceiverPhone), PickupAddress: strings.TrimSpace(req.PickupAddress), DropoffAddress: strings.TrimSpace(req.DropoffAddress), ItemDescription: strings.TrimSpace(req.ItemDescription), DeliveryPriceMinor: req.DeliveryPriceMinor, DeclaredValueMinor: req.DeclaredValueMinor, PackageCategory: strings.TrimSpace(req.PackageCategory), WeightKg: float64ToNumeric(req.WeightKg), AdditionalInstructions: strings.TrimSpace(req.AdditionalInstructions)})
	if err != nil {
		fail(c, http.StatusNotFound, "delivery_not_found", "That delivery was not found.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"delivery": updated})
}

func (a *API) reassignDeliveryDriver(c *gin.Context) {
	tenant, _ := currentTenant(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_delivery", "That delivery id is invalid.")
		return
	}
	var req struct {
		DriverID *string `json:"driver_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_driver", "Choose a valid driver.")
		return
	}
	driverID := uuid.Nil
	if req.DriverID != nil && strings.TrimSpace(*req.DriverID) != "" {
		driverID, err = uuid.Parse(strings.TrimSpace(*req.DriverID))
		if err != nil {
			fail(c, http.StatusBadRequest, "invalid_driver", "Choose a valid driver.")
			return
		}
	}
	updated, err := a.svc.Q.AssignDeliveryDriver(c.Request.Context(), db.AssignDeliveryDriverParams{CompanyID: tenant.CompanyID, ID: id, Column3: driverID})
	if err != nil {
		fail(c, http.StatusNotFound, "delivery_or_driver_not_found", "That delivery or active driver was not found.")
		return
	}
	_, _ = a.svc.Q.AddDeliveryTimelineEvent(c.Request.Context(), db.AddDeliveryTimelineEventParams{DeliveryID: id, CompanyID: tenant.CompanyID, Status: updated.Status, Note: stringPointer("Driver reassigned"), CreatedBy: actorIDPointer(c)})
	c.JSON(http.StatusOK, gin.H{"delivery": updated})
}

func (a *API) updateDeliveryStatus(c *gin.Context) {
	tenant, _ := currentTenant(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_delivery", "That delivery id is invalid.")
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_status", "Choose a valid delivery status.")
		return
	}
	allowed := map[string]bool{"created": true, "assigned": true, "picked_up": true, "in_transit": true, "delivered": true, "cancelled": true}
	if !allowed[req.Status] {
		fail(c, http.StatusBadRequest, "invalid_status", "Choose a valid delivery status.")
		return
	}
	updated, err := a.svc.Q.UpdateCompanyDeliveryStatus(c.Request.Context(), db.UpdateCompanyDeliveryStatusParams{CompanyID: tenant.CompanyID, ID: id, Status: req.Status})
	if err != nil {
		fail(c, http.StatusNotFound, "delivery_not_found", "That delivery was not found.")
		return
	}
	_, _ = a.svc.Q.AddDeliveryTimelineEvent(c.Request.Context(), db.AddDeliveryTimelineEventParams{DeliveryID: id, CompanyID: tenant.CompanyID, Status: req.Status, Note: stringPointer("Status updated manually"), CreatedBy: actorIDPointer(c)})
	c.JSON(http.StatusOK, gin.H{"delivery": updated})
}

func (a *API) getDeliveryTimeline(c *gin.Context) {
	tenant, _ := currentTenant(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_delivery", "That delivery id is invalid.")
		return
	}
	events, err := a.svc.Q.ListDeliveryTimeline(c.Request.Context(), db.ListDeliveryTimelineParams{CompanyID: tenant.CompanyID, DeliveryID: id})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (a *API) publicTracking(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		fail(c, http.StatusNotFound, "tracking_not_found", "That tracking code was not found.")
		return
	}
	var row struct {
		ID             uuid.UUID
		Status         string
		RequestStatus  string
		PickupAddress  string
		DropoffAddress string
		TrackingCode   *string
		UpdatedAt      time.Time
	}
	err := a.svc.DB.QueryRow(c.Request.Context(), `SELECT id,status,request_status,pickup_address,dropoff_address,tracking_code,updated_at FROM deliveries WHERE tracking_code=$1`, code).Scan(&row.ID, &row.Status, &row.RequestStatus, &row.PickupAddress, &row.DropoffAddress, &row.TrackingCode, &row.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "tracking_not_found", "That tracking code was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	events, _ := a.svc.DB.Query(c.Request.Context(), `SELECT status,note,created_at FROM delivery_timeline_events WHERE delivery_id=$1 ORDER BY created_at ASC`, row.ID)
	type event struct {
		Status    string    `json:"status"`
		Note      *string   `json:"note"`
		CreatedAt time.Time `json:"created_at"`
	}
	timeline := make([]event, 0)
	if events != nil {
		defer events.Close()
		for events.Next() {
			var e event
			if events.Scan(&e.Status, &e.Note, &e.CreatedAt) == nil {
				timeline = append(timeline, e)
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"tracking_code": code, "status": row.Status, "request_status": row.RequestStatus, "pickup_address": row.PickupAddress, "dropoff_address": row.DropoffAddress, "updated_at": row.UpdatedAt, "timeline": timeline})
}
