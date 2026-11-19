package httpapi

import (
	"encoding/json"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/service"
)

var bookingPageSlugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,38}[a-z0-9])?$`)
var bookingPageSlugInvalidChars = regexp.MustCompile(`[^a-z0-9-]+`)
var bookingPageSlugRepeatedHyphens = regexp.MustCompile(`-+`)

func (a *API) publicConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"mapbox_access_token": a.cfg.MapboxAccessToken,
	})
}

func validateBookingPageSlug(value string) (string, string) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "", "Choose the name for your booking page."
	}
	if !bookingPageSlugPattern.MatchString(value) {
		return "", "Use 1–40 lowercase letters, numbers or hyphens, without a leading or trailing hyphen."
	}
	if strings.Contains(value, "--") {
		return "", "Consecutive hyphens are not allowed."
	}
	return value, ""
}

func trimmedPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

type bookingPageDTO struct {
	Slug        string  `json:"slug"`
	Headline    *string `json:"headline"`
	Description *string `json:"description"`
	IsPublished bool    `json:"is_published"`
}

func toBookingPageDTO(page db.BookingPage) bookingPageDTO {
	return bookingPageDTO{
		Slug:        page.Slug,
		Headline:    page.Headline,
		Description: page.Description,
		IsPublished: page.IsPublished,
	}
}

func defaultBookingPageSlug(company db.Company) string {
	name := strings.Fields(company.Name)
	candidates := []string{company.Slug}
	if len(name) > 0 {
		candidates = append([]string{name[0]}, candidates...)
	}

	for _, candidate := range candidates {
		candidate = strings.ToLower(candidate)
		candidate = bookingPageSlugInvalidChars.ReplaceAllString(candidate, "-")
		candidate = strings.Trim(candidate, "-")
		candidate = bookingPageSlugRepeatedHyphens.ReplaceAllString(candidate, "-")
		if len(candidate) > 40 {
			candidate = strings.TrimRight(candidate[:40], "-")
		}
		if slug, reason := validateBookingPageSlug(candidate); reason == "" {
			return slug
		}
	}
	return "company"
}

func (a *API) getBookingPage(c *gin.Context) {
	tenant, _ := currentTenant(c)
	page, err := a.svc.Q.GetBookingPageByCompany(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			company, companyErr := a.svc.Q.GetCompanyByID(c.Request.Context(), tenant.CompanyID)
			if companyErr != nil {
				a.handleServiceError(c, companyErr)
				return
			}
			if company.Status != "verified" {
				c.JSON(http.StatusOK, gin.H{"booking_page": nil})
				return
			}

			baseSlug := defaultBookingPageSlug(company)
			slug := baseSlug
			available := false
			for suffix := 2; suffix <= 1000; suffix++ {
				taken, existsErr := a.svc.Q.BookingPageSlugExists(c.Request.Context(), slug)
				if existsErr != nil {
					a.handleServiceError(c, existsErr)
					return
				}
				if !taken {
					available = true
					break
				}
				number := strconv.Itoa(suffix)
				prefix := baseSlug[:min(len(baseSlug), 40-len(number)-1)]
				slug = strings.TrimRight(prefix, "-") + "-" + number
			}
			if !available {
				fail(c, http.StatusConflict, "booking_page_slug_taken", "Could not choose an available default booking page name.")
				return
			}

			page, err = a.svc.Q.UpsertBookingPage(c.Request.Context(), db.UpsertBookingPageParams{
				CompanyID:   tenant.CompanyID,
				Slug:        slug,
				Headline:    "",
				Description: "",
				IsPublished: true,
			})
			if err != nil {
				if isUniqueViolation(err) {
					fail(c, http.StatusConflict, "booking_page_slug_taken", "The default booking page name was claimed by another company. Please refresh to try again.")
					return
				}
				a.handleServiceError(c, err)
				return
			}
			c.JSON(http.StatusOK, gin.H{"booking_page": toBookingPageDTO(page)})
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking_page": toBookingPageDTO(page)})
}

func (a *API) checkBookingPageSlug(c *gin.Context) {
	tenant, _ := currentTenant(c)
	slug, reason := validateBookingPageSlug(c.Query("slug"))
	if reason != "" {
		c.JSON(http.StatusOK, gin.H{"slug": slug, "available": false, "reason": reason})
		return
	}

	taken, err := a.svc.Q.BookingPageSlugExists(c.Request.Context(), slug)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if taken {
		current, currentErr := a.svc.Q.GetBookingPageByCompany(c.Request.Context(), tenant.CompanyID)
		if currentErr == nil && current.Slug == slug {
			c.JSON(http.StatusOK, gin.H{"slug": slug, "available": true})
			return
		}
		c.JSON(http.StatusOK, gin.H{"slug": slug, "available": false, "reason": "That booking page name is already taken."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"slug": slug, "available": true})
}

func (a *API) upsertBookingPage(c *gin.Context) {
	tenant, _ := currentTenant(c)

	company, err := a.svc.Q.GetCompanyByID(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if company.Status != "verified" {
		fail(c, http.StatusForbidden, "provider_not_approved", "Your booking page becomes available once your company is approved.")
		return
	}

	var req struct {
		Slug        string  `json:"slug" binding:"required,max=40"`
		Headline    *string `json:"headline" binding:"omitempty,max=120"`
		Description *string `json:"description" binding:"omitempty,max=600"`
		IsPublished *bool   `json:"is_published"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the booking page details you entered.")
		return
	}

	slug, reason := validateBookingPageSlug(req.Slug)
	if reason != "" {
		fail(c, http.StatusUnprocessableEntity, "invalid_slug", reason)
		return
	}

	taken, err := a.svc.Q.BookingPageSlugExists(c.Request.Context(), slug)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if taken {
		current, currentErr := a.svc.Q.GetBookingPageByCompany(c.Request.Context(), tenant.CompanyID)
		if currentErr != nil || current.Slug != slug {
			fail(c, http.StatusConflict, "slug_taken", "That booking page name is already taken.")
			return
		}
	}

	published := true
	current, currentErr := a.svc.Q.GetBookingPageByCompany(c.Request.Context(), tenant.CompanyID)
	if currentErr == nil {
		published = current.IsPublished
	}
	if req.IsPublished != nil {
		published = *req.IsPublished
	}

	page, err := a.svc.Q.UpsertBookingPage(c.Request.Context(), db.UpsertBookingPageParams{
		CompanyID:   tenant.CompanyID,
		Slug:        slug,
		Headline:    valueOr(trimmedPtr(req.Headline), ""),
		Description: valueOr(trimmedPtr(req.Description), ""),
		IsPublished: published,
	})
	if err != nil {
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "slug_taken", "That booking page name is already taken.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.booking_page_updated",
		EntityType: "booking_page",
		EntityID:   &page.ID,
		Metadata:   map[string]any{"slug": page.Slug, "is_published": page.IsPublished},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"booking_page": toBookingPageDTO(page)})
}

func (a *API) publicBookingPage(c *gin.Context) {
	slug, reason := validateBookingPageSlug(c.Param("slug"))
	if reason != "" {
		fail(c, http.StatusNotFound, "booking_page_not_found", "That booking page was not found.")
		return
	}

	row, err := a.svc.Q.GetPublicBookingPage(c.Request.Context(), slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "booking_page_not_found", "That booking page was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	var operatingHours any
	if len(row.OperatingHours) > 0 {
		_ = json.Unmarshal(row.OperatingHours, &operatingHours)
	}
	if operatingHours == nil {
		operatingHours = gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{
		"booking_page": gin.H{
			"slug":        row.Slug,
			"headline":    row.Headline,
			"description": row.Description,
		},
		"company": gin.H{
			"id":              row.CompanyID,
			"name":            row.Name,
			"slug":            row.CompanySlug,
			"public_name":     row.PublicName,
			"business_type":   row.BusinessType,
			"business_email":  row.BusinessEmail,
			"phone":           row.Phone,
			"city":            row.City,
			"state":           row.State,
			"country":         row.Country,
			"currency":        row.Currency,
			"operating_hours": operatingHours,
		},
		"branding": gin.H{
			"logo_url":        row.LogoUrl,
			"primary_color":   row.PrimaryColor,
			"secondary_color": row.SecondaryColor,
			"cover_image_url": row.CoverImageUrl,
		},
		"subdomain": row.Subdomain,
	})
}

func (a *API) publicBookingRoutes(c *gin.Context) {
	slug, reason := validateBookingPageSlug(c.Param("slug"))
	if reason != "" {
		fail(c, http.StatusNotFound, "booking_page_not_found", "That booking page was not found.")
		return
	}
	page, err := a.svc.Q.GetPublicBookingPage(c.Request.Context(), slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "booking_page_not_found", "That booking page was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	routes, err := a.svc.Q.ListPublicBookingRoutes(c.Request.Context(), page.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"routes": routes})
}

func (a *API) publicBookingRates(c *gin.Context) {
	slug, reason := validateBookingPageSlug(c.Param("slug"))
	if reason != "" {
		fail(c, http.StatusNotFound, "booking_page_not_found", "That booking page was not found.")
		return
	}
	weight, err := strconv.ParseFloat(c.Query("weight_kg"), 64)
	if err != nil || weight <= 0 || weight > 500 {
		fail(c, http.StatusBadRequest, "invalid_weight", "Enter a package weight between 0.1 and 500 kg.")
		return
	}
	page, err := a.svc.Q.GetPublicBookingPage(c.Request.Context(), slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "booking_page_not_found", "That booking page was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	rates, err := a.svc.Q.ListPublicBookingRates(c.Request.Context(), db.ListPublicBookingRatesParams{
		WeightKg:  float64ToNumeric(weight),
		CompanyID: page.CompanyID,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"rates": rates})
}

func (a *API) createPublicBooking(c *gin.Context) {
	slug, reason := validateBookingPageSlug(c.Param("slug"))
	if reason != "" {
		fail(c, http.StatusNotFound, "booking_page_not_found", "That booking page was not found.")
		return
	}
	page, err := a.svc.Q.GetPublicBookingPage(c.Request.Context(), slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "booking_page_not_found", "That booking page was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	var req struct {
		SenderName             string   `json:"sender_name" binding:"required,min=2,max=120"`
		SenderPhone            string   `json:"sender_phone" binding:"required,min=7,max=30"`
		ReceiverName           string   `json:"receiver_name" binding:"required,min=2,max=120"`
		ReceiverPhone          string   `json:"receiver_phone" binding:"required,min=7,max=30"`
		PickupAddress          string   `json:"pickup_address" binding:"required,min=3,max=500"`
		DropoffAddress         string   `json:"dropoff_address" binding:"required,min=3,max=500"`
		PackageCategory        string   `json:"package_category" binding:"required,oneof=cloth electronics document food jewelry other"`
		DeclaredValueMinor     int64    `json:"declared_value_minor" binding:"required,min=0"`
		WeightKg               float64  `json:"weight_kg" binding:"required,gt=0,lte=500"`
		AdditionalInstructions string   `json:"additional_instructions" binding:"max=1000"`
		PackagePhotoURL        string   `json:"package_photo_url" binding:"omitempty,max=1000"`
		PickupLatitude         *float64 `json:"pickup_latitude"`
		PickupLongitude        *float64 `json:"pickup_longitude"`
		DropoffLatitude        *float64 `json:"dropoff_latitude"`
		DropoffLongitude       *float64 `json:"dropoff_longitude"`
		SelectedRouteID        *string  `json:"selected_route_id"`
		DeliverySpeed          string   `json:"delivery_speed" binding:"omitempty,oneof=standard express same_day"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_booking", "Check the package, route and contact details.")
		return
	}
	if req.DeclaredValueMinor > 10_000_000_000_00 {
		fail(c, http.StatusBadRequest, "invalid_item_value", "The item value is too large.")
		return
	}

	var routeID *uuid.UUID
	deliverySpeed := strings.TrimSpace(req.DeliverySpeed)
	if deliverySpeed == "" {
		deliverySpeed = "standard"
	}
	deliveryPriceMinor := int64(0)
	distanceKM := float64(0)
	priceIsEstimate := true
	if req.SelectedRouteID != nil && strings.TrimSpace(*req.SelectedRouteID) != "" {
		parsedID, parseErr := uuid.Parse(strings.TrimSpace(*req.SelectedRouteID))
		if parseErr != nil {
			fail(c, http.StatusBadRequest, "invalid_route", "Choose a route offered by this provider.")
			return
		}
		selectedRoute, routeErr := a.svc.Q.GetPublicBookingRoute(c.Request.Context(), db.GetPublicBookingRouteParams{
			CompanyID: page.CompanyID,
			ID:        parsedID,
		})
		if routeErr != nil {
			if routeErr == pgx.ErrNoRows {
				fail(c, http.StatusBadRequest, "invalid_route", "That route is no longer available.")
				return
			}
			a.handleServiceError(c, routeErr)
			return
		}
		routeID = &parsedID
		req.PickupAddress = selectedRoute.OriginCity + ", " + selectedRoute.OriginState
		req.DropoffAddress = selectedRoute.DestinationCity + ", " + selectedRoute.DestinationState
		req.PickupLatitude, req.PickupLongitude = nil, nil
		req.DropoffLatitude, req.DropoffLongitude = nil, nil
		distanceKM = numericToFloat64(selectedRoute.DistanceKm)
		var pricePerKm pgtype.Int8
		switch deliverySpeed {
		case "standard":
			pricePerKm = selectedRoute.StandardPricePerKmMinor
		case "express":
			pricePerKm = selectedRoute.ExpressPricePerKmMinor
		case "same_day":
			pricePerKm = selectedRoute.SameDayPricePerKmMinor
		}
		if !pricePerKm.Valid {
			fail(c, http.StatusBadRequest, "service_unavailable", "That delivery speed is not offered on this route.")
			return
		}
		calculatedPrice := int64(math.Ceil(float64(pricePerKm.Int64) * distanceKM))
		if calculatedPrice < selectedRoute.MinimumChargeMinor {
			calculatedPrice = selectedRoute.MinimumChargeMinor
		}
		deliveryPriceMinor = calculatedPrice
		priceIsEstimate = false
	} else if !validCoordinatePair(req.PickupLatitude, req.PickupLongitude) ||
		!validCoordinatePair(req.DropoffLatitude, req.DropoffLongitude) {
		fail(c, http.StatusBadRequest, "locations_not_verified", "Choose pickup and drop-off locations from the map suggestions.")
		return
	}

	photoURL := strings.TrimSpace(req.PackagePhotoURL)
	if photoURL != "" {
		parsedURL, parseErr := url.Parse(photoURL)
		if parseErr != nil || parsedURL.Scheme != "https" || parsedURL.Host != "res.cloudinary.com" {
			fail(c, http.StatusBadRequest, "invalid_package_photo", "Upload the package photo using the booking page photo picker.")
			return
		}
	}

	delivery, err := a.svc.Q.CreatePublicBookingDelivery(c.Request.Context(), db.CreatePublicBookingDeliveryParams{
		CompanyID:              page.CompanyID,
		SenderName:             strings.TrimSpace(req.SenderName),
		SenderPhone:            strings.TrimSpace(req.SenderPhone),
		ReceiverName:           strings.TrimSpace(req.ReceiverName),
		ReceiverPhone:          strings.TrimSpace(req.ReceiverPhone),
		PickupAddress:          strings.TrimSpace(req.PickupAddress),
		DropoffAddress:         strings.TrimSpace(req.DropoffAddress),
		ItemDescription:        req.PackageCategory,
		DeclaredValueMinor:     req.DeclaredValueMinor,
		DeliveryPriceMinor:     deliveryPriceMinor,
		PackageCategory:        req.PackageCategory,
		WeightKg:               float64ToNumeric(req.WeightKg),
		AdditionalInstructions: strings.TrimSpace(req.AdditionalInstructions),
		PackagePhotoUrl:        photoURL,
		PickupLatitude:         optionalNumeric(req.PickupLatitude),
		PickupLongitude:        optionalNumeric(req.PickupLongitude),
		DropoffLatitude:        optionalNumeric(req.DropoffLatitude),
		DropoffLongitude:       optionalNumeric(req.DropoffLongitude),
		SelectedRouteID:        routeID,
		SelectedRateCardID:     nil,
		PriceIsEstimate:        priceIsEstimate,
		DeliverySpeed:          deliverySpeed,
		DistanceKm:             float64ToNumeric(distanceKM),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusBadRequest, "invalid_route", "That route is no longer available.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		CompanyID:  &page.CompanyID,
		Action:     "company.booking_requested",
		EntityType: "delivery",
		EntityID:   &delivery.ID,
		Metadata:   map[string]any{"source": "booking_page"},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusCreated, gin.H{
		"booking": gin.H{
			"reference":            delivery.ID,
			"status":               delivery.Status,
			"delivery_price_minor": delivery.DeliveryPriceMinor,
			"price_is_estimate":    delivery.PriceIsEstimate,
		},
	})
}

func validCoordinatePair(latitude, longitude *float64) bool {
	return latitude != nil && longitude != nil &&
		*latitude >= -90 && *latitude <= 90 &&
		*longitude >= -180 && *longitude <= 180
}

func optionalNumeric(value *float64) pgtype.Numeric {
	if value == nil {
		return pgtype.Numeric{}
	}
	return float64ToNumeric(*value)
}
