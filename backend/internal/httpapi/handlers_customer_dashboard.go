package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/bol-platform/backend/internal/db/sqlc"
)

type customerAddressDTO struct {
	ID              uuid.UUID `json:"id"`
	Label           string    `json:"label"`
	AddressType     string    `json:"address_type"`
	ContactName     string    `json:"contact_name"`
	Phone           string    `json:"phone"`
	State           string    `json:"state"`
	City            string    `json:"city"`
	Area            string    `json:"area"`
	StreetAddress   string    `json:"street_address"`
	Directions      string    `json:"directions"`
	Latitude        *float64  `json:"latitude"`
	Longitude       *float64  `json:"longitude"`
	DefaultPickup   bool      `json:"default_pickup"`
	DefaultDelivery bool      `json:"default_delivery"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type customerAddressInput struct {
	Label           string   `json:"label" binding:"required,max=80"`
	AddressType     string   `json:"address_type" binding:"omitempty,oneof=residential business warehouse other"`
	ContactName     string   `json:"contact_name" binding:"omitempty,max=120"`
	Phone           string   `json:"phone" binding:"omitempty,max=30"`
	State           string   `json:"state" binding:"required,max=80"`
	City            string   `json:"city" binding:"required,max=80"`
	Area            string   `json:"area" binding:"omitempty,max=120"`
	StreetAddress   string   `json:"street_address" binding:"required,max=240"`
	Directions      string   `json:"directions" binding:"omitempty,max=500"`
	Latitude        *float64 `json:"latitude"`
	Longitude       *float64 `json:"longitude"`
	DefaultPickup   bool     `json:"default_pickup"`
	DefaultDelivery bool     `json:"default_delivery"`
}

func numericPtr(v pgtype.Numeric) *float64 {
	if !v.Valid {
		return nil
	}
	f := numericToFloat64(v)
	return &f
}

func customerAddressDTOFromRow(row db.CustomerAddress) customerAddressDTO {
	return customerAddressDTO{
		ID:              row.ID,
		Label:           row.Label,
		AddressType:     row.AddressType,
		ContactName:     row.ContactName,
		Phone:           row.Phone,
		State:           row.State,
		City:            row.City,
		Area:            row.Area,
		StreetAddress:   row.StreetAddress,
		Directions:      row.Directions,
		Latitude:        numericPtr(row.Latitude),
		Longitude:       numericPtr(row.Longitude),
		DefaultPickup:   row.DefaultPickup,
		DefaultDelivery: row.DefaultDelivery,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func normalizeAddressInput(input *customerAddressInput) string {
	input.Label = strings.TrimSpace(input.Label)
	input.ContactName = strings.TrimSpace(input.ContactName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.State = strings.TrimSpace(input.State)
	input.City = strings.TrimSpace(input.City)
	input.Area = strings.TrimSpace(input.Area)
	input.StreetAddress = strings.TrimSpace(input.StreetAddress)
	input.Directions = strings.TrimSpace(input.Directions)
	if input.AddressType == "" {
		input.AddressType = "residential"
	}
	if input.Label == "" {
		return "Give this address a short label, such as Home or Office."
	}
	if input.State == "" || input.City == "" {
		return "Choose the state and city for this address."
	}
	if input.StreetAddress == "" {
		return "Enter the street address."
	}
	if (input.Latitude == nil) != (input.Longitude == nil) {
		return "Set both the latitude and longitude, or leave both blank."
	}
	if input.Latitude != nil && (*input.Latitude < -90 || *input.Latitude > 90) {
		return "That latitude is not valid."
	}
	if input.Longitude != nil && (*input.Longitude < -180 || *input.Longitude > 180) {
		return "That longitude is not valid."
	}
	return ""
}

func (a *API) listCustomerAddresses(c *gin.Context) {
	user, _ := currentUser(c)
	rows, err := a.svc.Q.ListCustomerAddresses(c.Request.Context(), user.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]customerAddressDTO, 0, len(rows))
	for _, row := range rows {
		result = append(result, customerAddressDTOFromRow(row))
	}
	c.JSON(http.StatusOK, gin.H{"addresses": result})
}

func (a *API) createCustomerAddress(c *gin.Context) {
	user, _ := currentUser(c)
	var input customerAddressInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, "invalid_address", "Enter a label, state, city and street address.")
		return
	}
	if message := normalizeAddressInput(&input); message != "" {
		fail(c, http.StatusBadRequest, "invalid_address", message)
		return
	}

	ctx := c.Request.Context()
	tx, err := a.svc.DB.Begin(ctx)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	defer tx.Rollback(ctx)
	q := a.svc.Q.WithTx(tx)
	if input.DefaultPickup {
		if err := q.ClearCustomerPickupDefault(ctx, user.ID); err != nil {
			a.handleServiceError(c, err)
			return
		}
	}
	if input.DefaultDelivery {
		if err := q.ClearCustomerDeliveryDefault(ctx, user.ID); err != nil {
			a.handleServiceError(c, err)
			return
		}
	}
	row, err := q.CreateCustomerAddress(ctx, db.CreateCustomerAddressParams{
		CustomerID:      user.ID,
		Label:           input.Label,
		AddressType:     input.AddressType,
		ContactName:     input.ContactName,
		Phone:           input.Phone,
		State:           input.State,
		City:            input.City,
		Area:            input.Area,
		StreetAddress:   input.StreetAddress,
		Directions:      input.Directions,
		Latitude:        optionalNumeric(input.Latitude),
		Longitude:       optionalNumeric(input.Longitude),
		DefaultPickup:   input.DefaultPickup,
		DefaultDelivery: input.DefaultDelivery,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"address": customerAddressDTOFromRow(row)})
}

func (a *API) updateCustomerAddress(c *gin.Context) {
	user, _ := currentUser(c)
	addressID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_address", "That address id is invalid.")
		return
	}
	var input customerAddressInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, "invalid_address", "Enter a label, state, city and street address.")
		return
	}
	if message := normalizeAddressInput(&input); message != "" {
		fail(c, http.StatusBadRequest, "invalid_address", message)
		return
	}

	ctx := c.Request.Context()
	tx, err := a.svc.DB.Begin(ctx)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	defer tx.Rollback(ctx)
	q := a.svc.Q.WithTx(tx)
	if input.DefaultPickup {
		if err := q.ClearCustomerPickupDefault(ctx, user.ID); err != nil {
			a.handleServiceError(c, err)
			return
		}
	}
	if input.DefaultDelivery {
		if err := q.ClearCustomerDeliveryDefault(ctx, user.ID); err != nil {
			a.handleServiceError(c, err)
			return
		}
	}
	row, err := q.UpdateCustomerAddress(ctx, db.UpdateCustomerAddressParams{
		ID:              addressID,
		CustomerID:      user.ID,
		Label:           input.Label,
		AddressType:     input.AddressType,
		ContactName:     input.ContactName,
		Phone:           input.Phone,
		State:           input.State,
		City:            input.City,
		Area:            input.Area,
		StreetAddress:   input.StreetAddress,
		Directions:      input.Directions,
		Latitude:        optionalNumeric(input.Latitude),
		Longitude:       optionalNumeric(input.Longitude),
		DefaultPickup:   input.DefaultPickup,
		DefaultDelivery: input.DefaultDelivery,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "address_not_found", "That saved address was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"address": customerAddressDTOFromRow(row)})
}

func (a *API) deleteCustomerAddress(c *gin.Context) {
	user, _ := currentUser(c)
	addressID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_address", "That address id is invalid.")
		return
	}
	deleted, err := a.svc.Q.DeleteCustomerAddress(c.Request.Context(), db.DeleteCustomerAddressParams{
		ID:         addressID,
		CustomerID: user.ID,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if deleted == 0 {
		fail(c, http.StatusNotFound, "address_not_found", "That saved address was not found.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) listCustomerFavourites(c *gin.Context) {
	user, _ := currentUser(c)
	rows, err := a.svc.Q.ListCustomerFavourites(c.Request.Context(), user.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	providers := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		providers = append(providers, gin.H{
			"id":          row.ID,
			"name":        row.Name,
			"slug":        row.Slug,
			"logo_url":    row.LogoUrl,
			"description": row.Description,
			"rating":      row.Rating,
			"route_count": row.RouteCount,
			"saved_at":    row.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

func (a *API) addCustomerFavourite(c *gin.Context) {
	user, _ := currentUser(c)
	companyID, err := uuid.Parse(strings.TrimSpace(c.Param("companyId")))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_provider", "That provider id is invalid.")
		return
	}
	company, err := a.svc.Q.GetCompanyByID(c.Request.Context(), companyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "provider_not_found", "That logistics provider was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	if company.DeletedAt != nil || company.Status != "verified" {
		fail(c, http.StatusNotFound, "provider_not_found", "That logistics provider was not found.")
		return
	}
	added, err := a.svc.Q.AddCustomerFavourite(c.Request.Context(), db.AddCustomerFavouriteParams{
		CustomerID: user.ID,
		CompanyID:  companyID,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"saved": added > 0})
}

func (a *API) removeCustomerFavourite(c *gin.Context) {
	user, _ := currentUser(c)
	companyID, err := uuid.Parse(strings.TrimSpace(c.Param("companyId")))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_provider", "That provider id is invalid.")
		return
	}
	if _, err := a.svc.Q.RemoveCustomerFavourite(c.Request.Context(), db.RemoveCustomerFavouriteParams{
		CustomerID: user.ID,
		CompanyID:  companyID,
	}); err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) createCustomerIssue(c *gin.Context) {
	user, _ := currentUser(c)
	bookingID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_booking", "That booking reference is invalid.")
		return
	}
	var input struct {
		IssueType   string `json:"issue_type" binding:"required,oneof=delayed missing damaged other"`
		Description string `json:"description" binding:"required,min=5,max=2000"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, "invalid_issue", "Choose what went wrong and add a short description.")
		return
	}
	row, err := a.svc.Q.CreateDeliveryIssue(c.Request.Context(), db.CreateDeliveryIssueParams{
		DeliveryID:  bookingID,
		CustomerID:  user.ID,
		IssueType:   input.IssueType,
		Description: strings.TrimSpace(input.Description),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "booking_not_found", "That booking was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"issue": row})
}

func (a *API) listCustomerIssues(c *gin.Context) {
	user, _ := currentUser(c)
	rows, err := a.svc.Q.ListCustomerDeliveryIssues(c.Request.Context(), user.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	issues := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		issues = append(issues, gin.H{
			"id":              row.ID,
			"delivery_id":     row.DeliveryID,
			"tracking_code":   row.TrackingCode,
			"pickup_address":  row.PickupAddress,
			"dropoff_address": row.DropoffAddress,
			"provider_name":   row.CompanyName,
			"issue_type":      row.IssueType,
			"description":     row.Description,
			"status":          row.Status,
			"created_at":      row.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"issues": issues})
}
