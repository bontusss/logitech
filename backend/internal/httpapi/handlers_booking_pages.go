package httpapi

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/service"
)

var bookingPageSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,38}[a-z0-9]$`)

func validateBookingPageSlug(value string) (string, string) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "", "Choose the name for your booking page."
	}
	if !bookingPageSlugPattern.MatchString(value) {
		return "", "Use 3–40 lowercase letters, numbers or hyphens, without a leading or trailing hyphen."
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

func (a *API) getBookingPage(c *gin.Context) {
	tenant, _ := currentTenant(c)
	page, err := a.svc.Q.GetBookingPageByCompany(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"booking_page": nil})
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking_page": toBookingPageDTO(page)})
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
