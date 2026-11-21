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

type customerBookingDTO struct {
	ID                 uuid.UUID `json:"id"`
	Status             string    `json:"status"`
	RequestStatus      string    `json:"request_status"`
	TrackingCode       *string   `json:"tracking_code"`
	TrackingURL        *string   `json:"tracking_url"`
	PickupAddress      string    `json:"pickup_address"`
	DropoffAddress     string    `json:"dropoff_address"`
	PackageCategory    string    `json:"package_category"`
	ItemDescription    string    `json:"item_description"`
	WeightKg           float64   `json:"weight_kg"`
	DeliveryPriceMinor int64     `json:"delivery_price_minor"`
	DeclaredValueMinor int64     `json:"declared_value_minor"`
	PriceIsEstimate    bool      `json:"price_is_estimate"`
	DeliverySpeed      string    `json:"delivery_speed"`
	DistanceKm         float64   `json:"distance_km"`
	ProviderID         uuid.UUID `json:"provider_id"`
	ProviderName       string    `json:"provider_name"`
	ProviderLogoURL    *string   `json:"provider_logo_url"`
	Currency           string    `json:"currency"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func customerBookingTrackingURL(code *string, appURL string) *string {
	if code == nil {
		return nil
	}
	value := strings.TrimRight(appURL, "/") + "/track/" + *code
	return &value
}

func (a *API) listCustomerBookings(c *gin.Context) {
	user, _ := currentUser(c)
	rows, err := a.svc.Q.ListCustomerBookings(c.Request.Context(), &user.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]customerBookingDTO, 0, len(rows))
	for _, row := range rows {
		result = append(result, customerBookingDTO{
			ID:                 row.ID,
			Status:             row.Status,
			RequestStatus:      row.RequestStatus,
			TrackingCode:       row.TrackingCode,
			TrackingURL:        customerBookingTrackingURL(row.TrackingCode, a.cfg.AppURL),
			PickupAddress:      row.PickupAddress,
			DropoffAddress:     row.DropoffAddress,
			PackageCategory:    row.PackageCategory,
			ItemDescription:    row.ItemDescription,
			WeightKg:           numericToFloat64(row.WeightKg),
			DeliveryPriceMinor: row.DeliveryPriceMinor,
			DeclaredValueMinor: row.DeclaredValueMinor,
			PriceIsEstimate:    row.PriceIsEstimate,
			DeliverySpeed:      row.DeliverySpeed,
			DistanceKm:         numericToFloat64(row.DistanceKm),
			ProviderID:         row.CompanyID,
			ProviderName:       row.CompanyName,
			ProviderLogoURL:    row.CompanyLogoUrl,
			Currency:           row.Currency,
			CreatedAt:          row.CreatedAt,
			UpdatedAt:          row.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"bookings": result})
}

func publicBookingSummary(row db.GetBookingForClaimRow, appURL string) gin.H {
	summary := gin.H{
		"reference":      row.ID,
		"request_status": row.RequestStatus,
	}
	if row.TrackingCode != nil {
		summary["tracking_code"] = *row.TrackingCode
		summary["tracking_url"] = strings.TrimRight(appURL, "/") + "/track/" + *row.TrackingCode
	}
	return summary
}

// claimPublicBooking links a booking made on a provider's public page to a
// customer account. Anonymous visitors create their account in the same step;
// signed-in customers simply attach the booking to the active session.
func (a *API) claimPublicBooking(c *gin.Context) {
	reference, err := uuid.Parse(strings.TrimSpace(c.Param("reference")))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_booking", "That booking reference is invalid.")
		return
	}

	ctx := c.Request.Context()
	booking, err := a.svc.Q.GetBookingForClaim(ctx, reference)
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "booking_not_found", "That booking was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	if booking.CustomerID != nil {
		user, ok := currentUser(c)
		if ok && user.ID == *booking.CustomerID {
			c.JSON(http.StatusOK, gin.H{
				"linked":  true,
				"booking": publicBookingSummary(booking, a.cfg.AppURL),
			})
			return
		}
		fail(c, http.StatusConflict, "booking_already_claimed", "This booking is already linked to a customer account.")
		return
	}

	user, signedIn := currentUser(c)
	var session service.Session
	if !signedIn {
		var req struct {
			FullName string `json:"full_name" binding:"required,min=2,max=120"`
			Email    string `json:"email" binding:"required,email,max=254"`
			Phone    string `json:"phone" binding:"omitempty,max=30"`
			Password string `json:"password" binding:"required,min=8,max=128"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			fail(c, http.StatusBadRequest, "invalid_input", "Enter your name, email and a password of at least 8 characters.")
			return
		}

		session, err = a.svc.RegisterCustomer(ctx, req.FullName, req.Email, req.Password, req.Phone, c.ClientIP(), c.Request.UserAgent())
		if err != nil {
			a.handleServiceError(c, err)
			return
		}
		a.setSessionCookies(c, session)
		user = session.User
	}

	claimed, err := a.svc.Q.ClaimBookingForCustomer(ctx, db.ClaimBookingForCustomerParams{
		ID:         reference,
		CustomerID: &user.ID,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if claimed == 0 {
		fail(c, http.StatusConflict, "booking_already_claimed", "This booking is already linked to a customer account.")
		return
	}

	a.svc.Audit(ctx, service.AuditEntry{
		ActorID:    &user.ID,
		CompanyID:  &booking.CompanyID,
		Action:     "customer.booking_linked",
		EntityType: "delivery",
		EntityID:   &reference,
		Metadata:   map[string]any{"registered": !signedIn},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	payload := gin.H{
		"linked":  true,
		"booking": publicBookingSummary(booking, a.cfg.AppURL),
	}
	status := http.StatusOK
	if !signedIn {
		status = http.StatusCreated
		sessionDTO := toSessionDTO(session)
		payload["user"] = sessionDTO.User
		payload["companies"] = sessionDTO.Companies
		payload["access_token"] = sessionDTO.AccessToken
		payload["expires_in"] = sessionDTO.ExpiresIn
		if sessionDTO.DevVerificationToken != "" {
			payload["dev_verification_token"] = sessionDTO.DevVerificationToken
		}
	}
	c.JSON(status, payload)
}

func (a *API) listCompanyCustomers(c *gin.Context) {
	tenant, _ := currentTenant(c)
	rows, err := a.svc.Q.ListCompanyCustomers(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"customers": rows})
}

func (a *API) listCompanyCustomerBookings(c *gin.Context) {
	tenant, _ := currentTenant(c)
	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_customer", "That customer id is invalid.")
		return
	}
	rows, err := a.svc.Q.ListCompanyCustomerDeliveries(c.Request.Context(), db.ListCompanyCustomerDeliveriesParams{
		CompanyID:  tenant.CompanyID,
		CustomerID: &customerID,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]deliveryDTO, 0, len(rows))
	for _, row := range rows {
		result = append(result, deliveryDTO{
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
			TrackingURL:            deliveryTrackingURL(row.TrackingCode, a.cfg.AppURL),
			CreatedAt:              row.CreatedAt,
			UpdatedAt:              row.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"deliveries": result})
}
