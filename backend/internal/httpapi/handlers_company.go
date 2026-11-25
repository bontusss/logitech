package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
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
	Country       string `json:"country" binding:"omitempty,max=2"`
	State         string `json:"state" binding:"required,max=80"`
	City          string `json:"city" binding:"required,max=80"`
	Address       string `json:"address" binding:"required,max=240"`
	ContactName   string `json:"contact_name" binding:"required,min=2,max=120"`
	Password      string `json:"password" binding:"required,min=8,max=128"`
}

type companySetupRequest struct {
	CompanyName   string `json:"company_name" binding:"required,min=2,max=160"`
	BusinessEmail string `json:"business_email" binding:"required,email,max=254"`
	Phone         string `json:"phone" binding:"required,max=30"`
	Country       string `json:"country" binding:"omitempty,max=2"`
	State         string `json:"state" binding:"required,max=80"`
	City          string `json:"city" binding:"required,max=80"`
	Address       string `json:"address" binding:"required,max=240"`
	ContactName   string `json:"contact_name" binding:"required,min=2,max=120"`
}

type companyDTO struct {
	ID               uuid.UUID       `json:"id"`
	Name             string          `json:"name"`
	Slug             string          `json:"slug"`
	BusinessEmail    string          `json:"business_email"`
	Phone            string          `json:"phone"`
	CACBusinessName  *string         `json:"cac_business_name"`
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
	WebsiteURL       *string         `json:"website_url"`
	InstagramURL     *string         `json:"instagram_url"`
	FacebookURL      *string         `json:"facebook_url"`
	TwitterURL       *string         `json:"twitter_url"`
	LinkedInURL      *string         `json:"linkedin_url"`
	YearEstablished  *int32          `json:"year_established"`
	PhotoURLs        []string        `json:"photo_urls"`
}

func toCompanyDTO(company db.Company) companyDTO {
	var yearEstablished *int32
	if company.YearEstablished.Valid {
		year := company.YearEstablished.Int32
		yearEstablished = &year
	}
	return companyDTO{
		ID:               company.ID,
		Name:             company.Name,
		Slug:             company.Slug,
		BusinessEmail:    company.BusinessEmail,
		Phone:            company.Phone,
		CACBusinessName:  company.CacBusinessName,
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
		WebsiteURL:       company.WebsiteUrl,
		InstagramURL:     company.InstagramUrl,
		FacebookURL:      company.FacebookUrl,
		TwitterURL:       company.TwitterUrl,
		LinkedInURL:      company.LinkedinUrl,
		YearEstablished:  yearEstablished,
		PhotoURLs:        nonNilStrings(company.PhotoUrls),
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

// setupCompany completes a provider account created through Google sign-up.
// The account already exists and is signed in; this only creates its company.
func (a *API) setupCompany(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "unauthenticated", "Sign in to continue.")
		return
	}

	var req companySetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the company details you entered.")
		return
	}

	companyID, err := a.svc.SetupCompanyForUser(c.Request.Context(), user.ID, service.CompanyRegistration{
		CompanyName:   req.CompanyName,
		BusinessEmail: req.BusinessEmail,
		Phone:         req.Phone,
		Country:       strings.ToUpper(req.Country),
		State:         req.State,
		City:          req.City,
		Address:       req.Address,
		ContactName:   req.ContactName,
	}, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"company_id": companyID})
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
		Name            *string         `json:"name" binding:"omitempty,min=2,max=160"`
		BusinessEmail   *string         `json:"business_email" binding:"omitempty,email"`
		Phone           *string         `json:"phone" binding:"omitempty,max=30"`
		CACBusinessName *string         `json:"cac_business_name" binding:"omitempty,max=200"`
		Country         *string         `json:"country" binding:"omitempty,max=2"`
		State           *string         `json:"state" binding:"omitempty,max=80"`
		City            *string         `json:"city" binding:"omitempty,max=80"`
		Address         *string         `json:"address" binding:"omitempty,max=240"`
		Description     *string         `json:"description" binding:"omitempty,max=2000"`
		OperatingHours  json.RawMessage `json:"operating_hours"`
		WebsiteURL      *string         `json:"website_url" binding:"omitempty,max=500"`
		InstagramURL    *string         `json:"instagram_url" binding:"omitempty,max=500"`
		FacebookURL     *string         `json:"facebook_url" binding:"omitempty,max=500"`
		TwitterURL      *string         `json:"twitter_url" binding:"omitempty,max=500"`
		LinkedInURL     *string         `json:"linkedin_url" binding:"omitempty,max=500"`
		YearEstablished *int            `json:"year_established"`
		PhotoURLs       []string        `json:"photo_urls"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the company details you entered.")
		return
	}

	name := valueOr(req.Name, current.Name)
	email := valueOr(req.BusinessEmail, current.BusinessEmail)
	phone := valueOr(req.Phone, current.Phone)
	country := strings.ToUpper(valueOr(req.Country, current.Country))
	state := valueOr(req.State, deref(current.State))
	city := valueOr(req.City, deref(current.City))
	address := valueOr(req.Address, deref(current.Address))
	description := valueOr(req.Description, deref(current.Description))
	website := strings.TrimSpace(valueOr(req.WebsiteURL, deref(current.WebsiteUrl)))
	instagram := strings.TrimSpace(valueOr(req.InstagramURL, deref(current.InstagramUrl)))
	facebook := strings.TrimSpace(valueOr(req.FacebookURL, deref(current.FacebookUrl)))
	twitter := strings.TrimSpace(valueOr(req.TwitterURL, deref(current.TwitterUrl)))
	linkedin := strings.TrimSpace(valueOr(req.LinkedInURL, deref(current.LinkedinUrl)))
	for _, link := range []string{website, instagram, facebook, twitter, linkedin} {
		if !validCompanyLink(link) {
			fail(c, http.StatusUnprocessableEntity, "invalid_link", "Website and social links must be full http or https URLs.")
			return
		}
	}

	yearEstablished := int32(0)
	if current.YearEstablished.Valid {
		yearEstablished = current.YearEstablished.Int32
	}
	if req.YearEstablished != nil {
		year := *req.YearEstablished
		if year != 0 && (year < 1800 || year > time.Now().Year()) {
			fail(c, http.StatusUnprocessableEntity, "invalid_year", "Enter a year between 1800 and the current year.")
			return
		}
		yearEstablished = int32(year)
	}

	photos := current.PhotoUrls
	if req.PhotoURLs != nil {
		cleaned, ok := cleanCompanyPhotoURLs(req.PhotoURLs)
		if !ok {
			fail(c, http.StatusUnprocessableEntity, "invalid_photo", "Add up to 8 company photos as https image links.")
			return
		}
		photos = cleaned
	}

	hours := current.OperatingHours
	if len(req.OperatingHours) > 0 {
		var probe map[string]any
		if err := json.Unmarshal(req.OperatingHours, &probe); err != nil || len(req.OperatingHours) > 8192 {
			fail(c, http.StatusUnprocessableEntity, "invalid_hours", "Operating hours must be a JSON object of days.")
			return
		}
		hours = req.OperatingHours
	}

	if name == "" || email == "" || phone == "" || country == "" {
		fail(c, http.StatusBadRequest, "invalid_input", "Name, email, phone and country are required.")
		return
	}

	if req.CACBusinessName != nil && strings.TrimSpace(*req.CACBusinessName) != "" {
		if len(strings.TrimSpace(*req.CACBusinessName)) < 2 {
			fail(c, http.StatusUnprocessableEntity, "invalid_cac_name", "Enter the business name exactly as it appears on your CAC certificate.")
			return
		}
	}

	company, err := a.svc.Q.UpdateCompany(c.Request.Context(), db.UpdateCompanyParams{
		ID:             tenant.CompanyID,
		Name:           strings.TrimSpace(name),
		BusinessEmail:  strings.ToLower(strings.TrimSpace(email)),
		Phone:          strings.TrimSpace(phone),
		Country:        country,
		Column6:        state,
		Column7:        city,
		Column8:        address,
		Column9:        description,
		Column10:       deref(current.CoverImageUrl),
		OperatingHours: hours,
		Column12:       website,
		Column13:       instagram,
		Column14:       facebook,
		Column15:       twitter,
		Column16:       linkedin,
		Column17:       yearEstablished,
		Column18:       nonNilStrings(photos),
	})
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			fail(c, http.StatusConflict, "email_taken", "That business email is already in use.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	if req.CACBusinessName != nil {
		company, err = a.svc.Q.UpdateCompanyCACName(c.Request.Context(), db.UpdateCompanyCACNameParams{
			ID:              tenant.CompanyID,
			CacBusinessName: strings.TrimSpace(*req.CACBusinessName),
		})
		if err != nil {
			a.handleServiceError(c, err)
			return
		}
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
	whatsAppVerified, err := a.svc.IsWhatsAppVerified(c.Request.Context(), owner.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if !whatsAppVerified {
		fail(c, http.StatusUnprocessableEntity, "whatsapp_unverified", "Verify the owner's WhatsApp number before submitting the company for review.")
		return
	}
	if company.State == nil || company.City == nil || company.Address == nil ||
		*company.State == "" || *company.City == "" || *company.Address == "" {
		fail(c, http.StatusUnprocessableEntity, "profile_incomplete", "Add the company location before submitting for review.")
		return
	}
	if company.CacBusinessName == nil || strings.TrimSpace(*company.CacBusinessName) == "" {
		fail(c, http.StatusUnprocessableEntity, "cac_required", "Add the business name registered with the CAC before submitting for review.")
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
	if err := a.svc.CancelProviderReminders(c.Request.Context(), tenant.CompanyID); err != nil {
		a.log.Error("could not cancel provider onboarding reminders", "company_id", tenant.CompanyID, "error", err)
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

func validCompanyLink(value string) bool {
	if value == "" {
		return true
	}
	parsed, err := url.Parse(value)
	return err == nil &&
		(parsed.Scheme == "http" || parsed.Scheme == "https") &&
		parsed.Host != ""
}

func cleanCompanyPhotoURLs(values []string) ([]string, bool) {
	cleaned := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if len(value) > 500 || len(cleaned) >= 8 {
			return nil, false
		}
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return nil, false
		}
		if seen[value] {
			continue
		}
		seen[value] = true
		cleaned = append(cleaned, value)
	}
	return cleaned, true
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
