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

type deliveryDTO struct {
	ID                     uuid.UUID  `json:"id"`
	SenderName             string     `json:"sender_name"`
	SenderPhone            string     `json:"sender_phone"`
	ReceiverName           string     `json:"receiver_name"`
	ReceiverPhone          string     `json:"receiver_phone"`
	PickupAddress          string     `json:"pickup_address"`
	DropoffAddress         string     `json:"dropoff_address"`
	ItemDescription        string     `json:"item_description"`
	DeclaredValueMinor     int64      `json:"declared_value_minor"`
	DeliveryPriceMinor     int64      `json:"delivery_price_minor"`
	Payer                  string     `json:"payer"`
	IsPaid                 bool       `json:"is_paid"`
	PaymentMethod          *string    `json:"payment_method"`
	AssignedDriverID       *uuid.UUID `json:"assigned_driver_id"`
	AssignedDriverName     *string    `json:"assigned_driver_name"`
	PackageCategory        string     `json:"package_category"`
	WeightKg               float64    `json:"weight_kg"`
	AdditionalInstructions string     `json:"additional_instructions"`
	PackagePhotoURL        *string    `json:"package_photo_url"`
	PriceIsEstimate        bool       `json:"price_is_estimate"`
	BookingSource          string     `json:"booking_source"`
	DeliverySpeed          string     `json:"delivery_speed"`
	DistanceKm             float64    `json:"distance_km"`
	Status                 string     `json:"status"`
	RequestStatus          string     `json:"request_status"`
	TrackingCode           *string    `json:"tracking_code"`
	RiderLinkToken         *string    `json:"rider_link_token"`
	TrackingURL            *string    `json:"tracking_url"`
	RiderURL               *string    `json:"rider_url"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

func deliveryTrackingURL(code *string, appURL string) *string {
	if code == nil {
		return nil
	}
	value := strings.TrimRight(appURL, "/") + "/track/" + *code
	return &value
}

func deliveryFromList(row db.ListCompanyDeliveriesRow, appURL string) deliveryDTO {
	return deliveryDTO{
		ID:                     row.ID,
		SenderName:             row.SenderName,
		SenderPhone:            row.SenderPhone,
		ReceiverName:           row.ReceiverName,
		ReceiverPhone:          row.ReceiverPhone,
		PickupAddress:          row.PickupAddress,
		DropoffAddress:         row.DropoffAddress,
		ItemDescription:        row.ItemDescription,
		DeclaredValueMinor:     row.DeclaredValueMinor,
		DeliveryPriceMinor:     row.DeliveryPriceMinor,
		Payer:                  row.Payer,
		IsPaid:                 row.IsPaid,
		PaymentMethod:          row.PaymentMethod,
		AssignedDriverID:       row.AssignedDriverID,
		AssignedDriverName:     row.AssignedDriverName,
		PackageCategory:        row.PackageCategory,
		WeightKg:               numericToFloat64(row.WeightKg),
		AdditionalInstructions: row.AdditionalInstructions,
		PackagePhotoURL:        row.PackagePhotoUrl,
		PriceIsEstimate:        row.PriceIsEstimate,
		BookingSource:          row.BookingSource,
		DeliverySpeed:          row.DeliverySpeed,
		DistanceKm:             numericToFloat64(row.DistanceKm),
		Status:                 row.Status,
		RequestStatus:          row.RequestStatus,
		TrackingCode:           row.TrackingCode,
		RiderLinkToken:         row.RiderLinkToken,
		TrackingURL:            deliveryTrackingURL(row.TrackingCode, appURL),
		CreatedAt:              row.CreatedAt,
		UpdatedAt:              row.UpdatedAt,
	}
}

func (a *API) listDeliveries(c *gin.Context) {
	tenant, _ := currentTenant(c)
	deliveries, err := a.svc.Q.ListCompanyDeliveries(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]deliveryDTO, 0, len(deliveries))
	for _, delivery := range deliveries {
		result = append(result, deliveryFromList(delivery, a.cfg.AppURL))
	}
	c.JSON(http.StatusOK, gin.H{"deliveries": result})
}

func (a *API) createDelivery(c *gin.Context) {
	tenant, _ := currentTenant(c)
	var req struct {
		SenderName         string  `json:"sender_name" binding:"required,min=2,max=120"`
		SenderPhone        string  `json:"sender_phone" binding:"required,min=7,max=30"`
		ReceiverName       string  `json:"receiver_name" binding:"required,min=2,max=120"`
		ReceiverPhone      string  `json:"receiver_phone" binding:"required,min=7,max=30"`
		PickupAddress      string  `json:"pickup_address" binding:"required,min=5,max=500"`
		DropoffAddress     string  `json:"dropoff_address" binding:"required,min=5,max=500"`
		ItemDescription    string  `json:"item_description" binding:"required,min=2,max=500"`
		DeclaredValueMinor *int64  `json:"declared_value_minor" binding:"required"`
		DeliveryPriceMinor *int64  `json:"delivery_price_minor" binding:"required"`
		Payer              string  `json:"payer" binding:"required,oneof=sender receiver"`
		IsPaid             bool    `json:"is_paid"`
		PaymentMethod      string  `json:"payment_method" binding:"omitempty,oneof=cash pos bank_transfer"`
		AssignedDriverID   *string `json:"assigned_driver_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_delivery", "Enter valid sender, receiver, address, item and payment details.")
		return
	}
	if *req.DeclaredValueMinor < 0 || *req.DeliveryPriceMinor < 0 {
		fail(c, http.StatusBadRequest, "invalid_amount", "Declared value and delivery price cannot be negative.")
		return
	}
	if req.IsPaid && req.PaymentMethod == "" {
		fail(c, http.StatusBadRequest, "payment_method_required", "Choose cash, POS or bank transfer for a paid delivery.")
		return
	}
	if !req.IsPaid {
		req.PaymentMethod = ""
	}

	var driverID uuid.UUID
	if req.AssignedDriverID != nil && strings.TrimSpace(*req.AssignedDriverID) != "" {
		parsedID, err := uuid.Parse(strings.TrimSpace(*req.AssignedDriverID))
		if err != nil {
			fail(c, http.StatusBadRequest, "invalid_driver", "Choose a valid active driver.")
			return
		}
		driverID = parsedID
	}
	company, err := a.svc.Q.GetCompanyByID(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if company.Status != "verified" {
		fail(c, http.StatusForbidden, "provider_not_approved", "Deliveries can be created after your company is verified.")
		return
	}

	status := "created"
	if driverID != uuid.Nil {
		status = "assigned"
	}
	createdBy := mustUserID(c)
	delivery, err := a.svc.Q.CreateDelivery(c.Request.Context(), db.CreateDeliveryParams{
		CompanyID:          tenant.CompanyID,
		SenderName:         strings.TrimSpace(req.SenderName),
		SenderPhone:        strings.TrimSpace(req.SenderPhone),
		ReceiverName:       strings.TrimSpace(req.ReceiverName),
		ReceiverPhone:      strings.TrimSpace(req.ReceiverPhone),
		PickupAddress:      strings.TrimSpace(req.PickupAddress),
		DropoffAddress:     strings.TrimSpace(req.DropoffAddress),
		ItemDescription:    strings.TrimSpace(req.ItemDescription),
		DeclaredValueMinor: *req.DeclaredValueMinor,
		DeliveryPriceMinor: *req.DeliveryPriceMinor,
		Payer:              req.Payer,
		IsPaid:             req.IsPaid,
		PaymentMethod:      req.PaymentMethod,
		AssignedDriverID:   driverID,
		Status:             status,
		CreatedBy:          &createdBy,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusBadRequest, "invalid_driver", "The selected driver is unavailable. Choose an active driver from your company.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.delivery_created",
		EntityType: "delivery",
		EntityID:   &delivery.ID,
		Metadata:   map[string]any{"status": delivery.Status, "payer": delivery.Payer, "is_paid": delivery.IsPaid},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusCreated, gin.H{"delivery": deliveryDTO{
		ID:                 delivery.ID,
		SenderName:         delivery.SenderName,
		SenderPhone:        delivery.SenderPhone,
		ReceiverName:       delivery.ReceiverName,
		ReceiverPhone:      delivery.ReceiverPhone,
		PickupAddress:      delivery.PickupAddress,
		DropoffAddress:     delivery.DropoffAddress,
		ItemDescription:    delivery.ItemDescription,
		DeclaredValueMinor: delivery.DeclaredValueMinor,
		DeliveryPriceMinor: delivery.DeliveryPriceMinor,
		Payer:              delivery.Payer,
		IsPaid:             delivery.IsPaid,
		PaymentMethod:      delivery.PaymentMethod,
		AssignedDriverID:   delivery.AssignedDriverID,
		Status:             delivery.Status,
		RequestStatus:      delivery.RequestStatus,
		TrackingCode:       delivery.TrackingCode,
		RiderLinkToken:     delivery.RiderLinkToken,
		CreatedAt:          delivery.CreatedAt,
	}})
}
