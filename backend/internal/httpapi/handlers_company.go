package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/service"
)

var (
	hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)
)

type companyRegistrationRequest struct {
	CompanyName   string `json:"company_name" binding:"required,min=2,max=160"`
	BusinessEmail string `json:"business_email" binding:"required,email,max=254"`
	Phone         string `json:"phone" binding:"required,max=30"`
	BusinessType  string `json:"business_type" binding:"required,max=80"`
	Country       string `json:"country" binding:"omitempty,max=2"`
	State         string `json:"state" binding:"required,max=80"`
	City          string `json:"city" binding:"required,max=80"`
	Address       string `json:"address" binding:"required,max=240"`
	ContactName   string `json:"contact_name" binding:"required,min=2,max=120"`
	Password      string `json:"password" binding:"required,min=8,max=128"`
}

type companyDTO struct {
	ID               uuid.UUID       `json:"id"`
	Name             string          `json:"name"`
	Slug             string          `json:"slug"`
	BusinessEmail    string          `json:"business_email"`
	Phone            string          `json:"phone"`
	BusinessType     string          `json:"business_type"`
	Country          string          `json:"country"`
	State            *string         `json:"state"`
	City             *string         `json:"city"`
	Address          *string         `json:"address"`
	Status           string          `json:"status"`
	OnboardingStatus string          `json:"onboarding_status"`
	VerifiedAt       *time.Time      `json:"verified_at"`
	CreatedAt        time.Time       `json:"created_at"`
	Description      *string         `json:"description"`
	CoverImageURL    *string         `json:"cover_image_url"`
	OperatingHours   json.RawMessage `json:"operating_hours"`
}

func toCompanyDTO(company db.Company) companyDTO {
	return companyDTO{
		ID:               company.ID,
		Name:             company.Name,
		Slug:             company.Slug,
		BusinessEmail:    company.BusinessEmail,
		Phone:            company.Phone,
		BusinessType:     company.BusinessType,
		Country:          company.Country,
		State:            company.State,
		City:             company.City,
		Address:          company.Address,
		Status:           company.Status,
		OnboardingStatus: company.OnboardingStatus,
		VerifiedAt:       company.VerifiedAt,
		CreatedAt:        company.CreatedAt,
		Description:      company.Description,
		CoverImageURL:    company.CoverImageUrl,
		OperatingHours:   json.RawMessage(company.OperatingHours),
	}
}

func (a *API) registerCompany(c *gin.Context) {
	var req companyRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the company details and password you entered.")
		return
	}

	session, err := a.svc.RegisterCompany(c.Request.Context(), service.CompanyRegistration{
		CompanyName:   req.CompanyName,
		BusinessEmail: req.BusinessEmail,
		Phone:         req.Phone,
		BusinessType:  req.BusinessType,
		Country:       strings.ToUpper(req.Country),
		State:         req.State,
		City:          req.City,
		Address:       req.Address,
		ContactName:   req.ContactName,
		Password:      req.Password,
	}, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.setSessionCookies(c, session)
	c.JSON(http.StatusCreated, toSessionDTO(session))
}

func (a *API) getCompany(c *gin.Context) {
	tenant, _ := currentTenant(c)
	company, err := a.svc.Q.GetCompanyByID(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	roles := []string{}
	permissions := []string{}
	if tenant.MemberID != uuid.Nil {
		roles, _ = a.svc.Q.ListMemberRoles(c.Request.Context(), tenant.MemberID)
		permissions, _ = a.svc.Q.ListUserPermissions(c.Request.Context(), db.ListUserPermissionsParams{
			UserID:    mustUserID(c),
			CompanyID: tenant.CompanyID,
		})
	}

	settings, _ := a.svc.Q.GetCompanySettings(c.Request.Context(), tenant.CompanyID)
	branding, _ := a.svc.Q.GetCompanyBranding(c.Request.Context(), tenant.CompanyID)

	c.JSON(http.StatusOK, gin.H{
		"company":         toCompanyDTO(company),
		"membership":      gin.H{"id": tenant.MemberID, "roles": nonNilStrings(roles), "permissions": nonNilStrings(permissions)},
		"settings":        settings,
		"branding":        branding,
		"subdomain":       a.companySubdomain(c, tenant.CompanyID),
		"platform_access": tenant.IsPlatform,
	})
}

func (a *API) updateCompany(c *gin.Context) {
	tenant, _ := currentTenant(c)
	current, err := a.svc.Q.GetCompanyByID(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	var req struct {
		Name           *string         `json:"name" binding:"omitempty,min=2,max=160"`
		BusinessEmail  *string         `json:"business_email" binding:"omitempty,email"`
		Phone          *string         `json:"phone" binding:"omitempty,max=30"`
		BusinessType   *string         `json:"business_type" binding:"omitempty,max=80"`
		Country        *string         `json:"country" binding:"omitempty,max=2"`
		State          *string         `json:"state" binding:"omitempty,max=80"`
		City           *string         `json:"city" binding:"omitempty,max=80"`
		Address        *string         `json:"address" binding:"omitempty,max=240"`
		Description    *string         `json:"description" binding:"omitempty,max=2000"`
		OperatingHours json.RawMessage `json:"operating_hours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the company details you entered.")
		return
	}

	name := valueOr(req.Name, current.Name)
	email := valueOr(req.BusinessEmail, current.BusinessEmail)
	phone := valueOr(req.Phone, current.Phone)
	businessType := valueOr(req.BusinessType, current.BusinessType)
	country := strings.ToUpper(valueOr(req.Country, current.Country))
	state := valueOr(req.State, deref(current.State))
	city := valueOr(req.City, deref(current.City))
	address := valueOr(req.Address, deref(current.Address))
	description := valueOr(req.Description, deref(current.Description))
	hours := current.OperatingHours
	if len(req.OperatingHours) > 0 {
		var probe map[string]any
		if err := json.Unmarshal(req.OperatingHours, &probe); err != nil || len(req.OperatingHours) > 8192 {
			fail(c, http.StatusUnprocessableEntity, "invalid_hours", "Operating hours must be a JSON object of days.")
			return
		}
		hours = req.OperatingHours
	}

	if name == "" || email == "" || phone == "" || businessType == "" || country == "" {
		fail(c, http.StatusBadRequest, "invalid_input", "Name, email, phone, business type and country are required.")
		return
	}

	company, err := a.svc.Q.UpdateCompany(c.Request.Context(), db.UpdateCompanyParams{
		ID:             tenant.CompanyID,
		Name:           strings.TrimSpace(name),
		BusinessEmail:  strings.ToLower(strings.TrimSpace(email)),
		Phone:          strings.TrimSpace(phone),
		BusinessType:   strings.TrimSpace(businessType),
		Country:        country,
		Column7:        state,
		Column8:        city,
		Column9:        address,
		Column10:       description,
		Column11:       deref(current.CoverImageUrl),
		OperatingHours: hours,
	})
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			fail(c, http.StatusConflict, "email_taken", "That business email is already in use.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:   actorID(c),
		CompanyID: &tenant.CompanyID,
		Action:    "company.profile_updated",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"company": toCompanyDTO(company)})
}

func (a *API) getCompanySettings(c *gin.Context) {
	tenant, _ := currentTenant(c)
	settings, err := a.svc.Q.GetCompanySettings(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func (a *API) updateCompanySettings(c *gin.Context) {
	tenant, _ := currentTenant(c)
	current, err := a.svc.Q.GetCompanySettings(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	var req struct {
		Timezone           *string `json:"timezone"`
		Currency           *string `json:"currency"`
		WeightUnit         *string `json:"weight_unit"`
		AutoAcceptBookings *bool   `json:"auto_accept_bookings"`
		NotifyEmail        *bool   `json:"notify_email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the settings you entered.")
		return
	}

	timezone := valueOr(req.Timezone, current.Timezone)
	currency := strings.ToUpper(valueOr(req.Currency, current.Currency))
	weightUnit := valueOr(req.WeightUnit, current.WeightUnit)
	autoAccept := current.AutoAcceptBookings
	if req.AutoAcceptBookings != nil {
		autoAccept = *req.AutoAcceptBookings
	}
	notify := current.NotifyEmail
	if req.NotifyEmail != nil {
		notify = *req.NotifyEmail
	}

	if _, err := time.LoadLocation(timezone); err != nil {
		fail(c, http.StatusUnprocessableEntity, "invalid_timezone", "That timezone is not recognised.")
		return
	}
	if !currencyPattern.MatchString(currency) {
		fail(c, http.StatusUnprocessableEntity, "invalid_currency", "Use a three-letter currency code such as NGN.")
		return
	}
	if weightUnit != "kg" && weightUnit != "lb" {
		fail(c, http.StatusUnprocessableEntity, "invalid_weight_unit", "Weight unit must be kg or lb.")
		return
	}

	settings, err := a.svc.Q.UpdateCompanySettings(c.Request.Context(), db.UpdateCompanySettingsParams{
		CompanyID:          tenant.CompanyID,
		Timezone:           timezone,
		Currency:           currency,
		WeightUnit:         weightUnit,
		AutoAcceptBookings: autoAccept,
		NotifyEmail:        notify,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:   actorID(c),
		CompanyID: &tenant.CompanyID,
		Action:    "company.settings_updated",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func (a *API) updateBranding(c *gin.Context) {
	tenant, _ := currentTenant(c)
	current, err := a.svc.Q.GetCompanyBranding(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	var req struct {
		LogoURL        *string `json:"logo_url" binding:"omitempty,url,max=500"`
		PrimaryColor   *string `json:"primary_color"`
		SecondaryColor *string `json:"secondary_color"`
		FaviconURL     *string `json:"favicon_url" binding:"omitempty,url,max=500"`
		CoverImageURL  *string `json:"cover_image_url" binding:"omitempty,url,max=500"`
		PublicName     *string `json:"public_name" binding:"omitempty,max=160"`
		Description    *string `json:"description" binding:"omitempty,max=2000"`
		PWAShortName   *string `json:"pwa_short_name" binding:"omitempty,max=30"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Logo and favicon must be valid URLs.")
		return
	}

	primary := valueOr(req.PrimaryColor, current.PrimaryColor)
	secondary := valueOr(req.SecondaryColor, current.SecondaryColor)
	if !hexColorPattern.MatchString(primary) || !hexColorPattern.MatchString(secondary) {
		fail(c, http.StatusUnprocessableEntity, "invalid_color", "Brand colours must be six-digit hex values.")
		return
	}

	branding, err := a.svc.Q.UpdateCompanyBranding(c.Request.Context(), db.UpdateCompanyBrandingParams{
		CompanyID:      tenant.CompanyID,
		Column2:        valueOr(req.LogoURL, deref(current.LogoUrl)),
		PrimaryColor:   primary,
		SecondaryColor: secondary,
		Column5:        valueOr(req.FaviconURL, deref(current.FaviconUrl)),
		Column6:        valueOr(req.CoverImageURL, deref(current.CoverImageUrl)),
		Column7:        valueOr(req.PublicName, deref(current.PublicName)),
		Column8:        valueOr(req.Description, deref(current.Description)),
		Column9:        valueOr(req.PWAShortName, deref(current.PwaShortName)),
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:   actorID(c),
		CompanyID: &tenant.CompanyID,
		Action:    "company.branding_updated",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"branding": branding})
}

func (a *API) submitVerification(c *gin.Context) {
	tenant, _ := currentTenant(c)
	company, err := a.svc.Q.GetCompanyByID(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if company.Status == "verified" {
		fail(c, http.StatusConflict, "already_verified", "This company is already verified.")
		return
	}

	owner, err := a.svc.Q.GetUserByID(c.Request.Context(), company.OwnerUserID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if owner.EmailVerifiedAt == nil {
		fail(c, http.StatusUnprocessableEntity, "email_unverified", "Verify the owner email before submitting the company for review.")
		return
	}
	if company.State == nil || company.City == nil || company.Address == nil ||
		*company.State == "" || *company.City == "" || *company.Address == "" {
		fail(c, http.StatusUnprocessableEntity, "profile_incomplete", "Add the company location before submitting for review.")
		return
	}

	updated, err := a.svc.Q.UpdateCompanyVerification(c.Request.Context(), db.UpdateCompanyVerificationParams{
		ID:               tenant.CompanyID,
		Status:           "under_review",
		OnboardingStatus: "pending_review",
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:   actorID(c),
		CompanyID: &tenant.CompanyID,
		Action:    "company.verification_submitted",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"company": toCompanyDTO(updated)})
}

func (a *API) listMembers(c *gin.Context) {
	tenant, _ := currentTenant(c)
	members, err := a.svc.Q.ListCompanyMembers(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"members": members})
}

func (a *API) updateMemberRole(c *gin.Context) {
	tenant, _ := currentTenant(c)
	memberID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_member", "That member id is invalid.")
		return
	}

	member, err := a.svc.Q.GetCompanyMemberByID(c.Request.Context(), db.GetCompanyMemberByIDParams{
		CompanyID: tenant.CompanyID,
		ID:        memberID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fail(c, http.StatusNotFound, "member_not_found", "That member was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	roles, _ := a.svc.Q.ListMemberRoles(c.Request.Context(), member.ID)
	if slices.Contains(roles, "owner") {
		fail(c, http.StatusConflict, "owner_immutable", "The company owner's role cannot be changed.")
		return
	}

	var req struct {
		RoleKey string `json:"role_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusUnprocessableEntity, "invalid_role", "Choose a valid assignable role.")
		return
	}
	if member.UserID == mustUserID(c) {
		fail(c, http.StatusConflict, "self_role_change", "You cannot change your own role.")
		return
	}

	role, ok := a.resolveAssignableRole(c, req.RoleKey)
	if !ok {
		return
	}

	err = pgx.BeginFunc(c.Request.Context(), a.svc.DB, func(tx pgx.Tx) error {
		q := a.svc.Q.WithTx(tx)
		if err := q.DeleteMemberRoles(c.Request.Context(), member.ID); err != nil {
			return err
		}
		return q.AddMemberRole(c.Request.Context(), db.AddMemberRoleParams{
			MemberID: member.ID,
			RoleID:   role.ID,
		})
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.member_role_updated",
		EntityType: "company_member",
		EntityID:   &member.ID,
		Metadata:   map[string]any{"role": req.RoleKey},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"member_id": member.ID, "roles": []string{req.RoleKey}})
}

func (a *API) removeMember(c *gin.Context) {
	tenant, _ := currentTenant(c)
	memberID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_member", "That member id is invalid.")
		return
	}

	member, err := a.svc.Q.GetCompanyMemberByID(c.Request.Context(), db.GetCompanyMemberByIDParams{
		CompanyID: tenant.CompanyID,
		ID:        memberID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fail(c, http.StatusNotFound, "member_not_found", "That member was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	roles, _ := a.svc.Q.ListMemberRoles(c.Request.Context(), member.ID)
	if slices.Contains(roles, "owner") {
		fail(c, http.StatusConflict, "owner_immutable", "The company owner cannot be removed.")
		return
	}
	if member.UserID == mustUserID(c) {
		fail(c, http.StatusConflict, "self_remove", "You cannot remove your own membership.")
		return
	}

	if err := a.svc.Q.SoftDeleteMember(c.Request.Context(), db.SoftDeleteMemberParams{
		ID:        member.ID,
		CompanyID: tenant.CompanyID,
	}); err != nil {
		a.handleServiceError(c, err)
		return
	}

	// Removing a member ends their existing sessions and branch assignments.
	// They keep any membership they hold in other companies.
	_ = a.svc.Q.RevokeUserRefreshTokens(c.Request.Context(), member.UserID)
	_ = a.svc.Q.DeleteMemberBranches(c.Request.Context(), member.ID)

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.member_removed",
		EntityType: "company_member",
		EntityID:   &member.ID,
		Metadata:   map[string]any{"user_id": member.UserID},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.Status(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mustUserID(c *gin.Context) uuid.UUID {
	user, _ := currentUser(c)
	return user.ID
}

func actorID(c *gin.Context) *uuid.UUID {
	user, ok := currentUser(c)
	if !ok {
		return nil
	}
	return &user.ID
}

func valueOr[T any](value *T, fallback T) T {
	if value != nil {
		return *value
	}
	return fallback
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
