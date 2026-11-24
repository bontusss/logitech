package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bol-platform/backend/internal/auth"
	"github.com/bol-platform/backend/internal/config"
	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/ratelimit"
	"github.com/bol-platform/backend/internal/service"
	"github.com/bol-platform/backend/internal/web"
)

const (
	accessCookieName  = "bol_access"
	refreshCookieName = "bol_refresh"
	refreshCookiePath = "/api/v1/auth"
)

type API struct {
	cfg     *config.Config
	svc     *service.Service
	auth    *auth.Manager
	limiter *ratelimit.Limiter
	log     *slog.Logger
}

func New(cfg *config.Config, svc *service.Service, authManager *auth.Manager, limiter *ratelimit.Limiter, log *slog.Logger) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	api := &API{cfg: cfg, svc: svc, auth: authManager, limiter: limiter, log: log}

	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(gin.Recovery(), api.requestLogger(), api.securityHeaders(), api.cors(), api.originGuard())

	_ = os.MkdirAll(cfg.UploadDir, 0o755)
	r.Static("/uploads", cfg.UploadDir)

	r.GET("/health", api.health)

	v1 := r.Group("/api/v1")

	authRoutes := v1.Group("/auth")
	authRoutes.POST("/register", api.rateLimit("register", 10), api.registerCustomer)
	authRoutes.POST("/login", api.rateLimit("login", 10), api.login)
	authRoutes.GET("/google/start", api.rateLimit("google-start", 20), api.googleAuthStart)
	authRoutes.GET("/google/callback", api.rateLimit("google-callback", 30), api.googleAuthCallback)
	authRoutes.POST("/refresh", api.rateLimit("refresh", 30), api.refresh)
	authRoutes.POST("/logout", api.rateLimit("logout", 30), api.logout)
	authRoutes.POST("/forgot-password", api.rateLimit("forgot-password", 5), api.forgotPassword)
	authRoutes.POST("/reset-password", api.rateLimit("reset-password", 10), api.resetPassword)
	authRoutes.POST("/verify-email", api.rateLimit("verify-email", 10), api.verifyEmail)

	// Company registration is public: it creates the owner account and company
	// in a single transaction.
	v1.POST("/companies", api.rateLimit("company-register", 5), api.registerCompany)

	v1.GET("/public/branding", api.publicBranding)
	v1.GET("/public/config", api.publicConfig)
	v1.GET("/public/manifest", api.publicManifest)
	v1.GET("/public/companies/:subdomain", api.publicCompany)
	v1.GET("/public/book/:slug", api.rateLimit("public-booking-page", 60), api.publicBookingPage)
	v1.GET("/public/book/:slug/routes", api.rateLimit("public-booking-routes", 60), api.publicBookingRoutes)
	v1.POST("/public/book/:slug/reviews", api.rateLimit("public-booking-review", 5), api.optionalAuth(), api.submitPublicReview)

	marketplace := v1.Group("/marketplace")
	marketplace.GET("/locations", api.rateLimit("marketplace-locations", 60), api.marketplaceLocations)
	marketplace.GET("/providers", api.rateLimit("marketplace-providers", 60), api.marketplaceProviders)
	marketplace.GET("/providers/:slug", api.rateLimit("marketplace-provider", 60), api.marketplaceProvider)
	marketplace.GET("/providers/:slug/services", api.rateLimit("marketplace-provider-services", 60), api.marketplaceProviderServices)
	marketplace.POST("/search", api.rateLimit("marketplace-search", 30), api.optionalAuth(), api.marketplaceSearch)
	marketplace.GET("/quotes/:reference", api.rateLimit("marketplace-quote", 60), api.marketplaceQuote)
	marketplace.POST("/quotes", api.rateLimit("marketplace-quote-select", 30), api.optionalAuth(), api.selectMarketplaceQuote)
	v1.POST("/public/book/:slug/bookings", api.rateLimit("public-booking-create", 10), api.optionalAuth(), api.createPublicBooking)
	v1.POST("/public/bookings/:reference/claim", api.rateLimit("public-booking-claim", 10), api.optionalAuth(), api.claimPublicBooking)
	v1.GET("/public/riders/:token", api.rateLimit("driver-portal", 60), api.driverPortal)
	v1.POST("/public/riders/:token/location", api.rateLimit("driver-location", 120), api.updateDriverLocation)
	v1.DELETE("/public/riders/:token/location", api.rateLimit("driver-location-stop", 30), api.stopDriverLocation)
	v1.GET("/public/tracking/:code", api.rateLimit("public-tracking", 60), api.publicTracking)
	v1.GET("/invitations/:token", api.rateLimit("invitation-preview", 30), api.previewInvitation)
	v1.POST("/invitations/:token/accept", api.rateLimit("invitation-accept", 10), api.acceptInvitation)

	authed := v1.Group("", api.requireAuth())
	authed.GET("/users/me", api.getMe)
	authed.PATCH("/users/me", api.updateMe)
	authed.POST("/auth/email/resend-verification", api.rateLimit("resend-verification", 3), api.resendEmailVerification)
	authed.POST("/auth/whatsapp/send-code", api.rateLimit("whatsapp-send-code", 5), api.sendWhatsAppCode)
	authed.POST("/auth/whatsapp/verify-code", api.rateLimit("whatsapp-verify-code", 10), api.verifyWhatsAppCode)
	authed.POST("/companies/setup", api.rateLimit("company-setup", 5), api.setupCompany)

	customer := authed.Group("/customer")
	customer.GET("/bookings", api.listCustomerBookings)
	customer.GET("/quotes", api.listCustomerQuotes)
	customer.GET("/quotes/:reference", api.getCustomerQuote)
	customer.POST("/quotes/:reference/select", api.selectCustomerQuote)

	admin := authed.Group("/admin", api.requirePlatformAdmin())
	admin.GET("/overview", api.adminOverview)
	admin.POST("/providers/:id/approve", api.approveProvider)
	admin.POST("/admins", api.requireSuperAdmin(), api.createPlatformAdmin)

	company := authed.Group("/companies/me", api.requireCompany())
	company.GET("", api.getCompany)
	company.PATCH("", api.requirePermission("company.update"), api.updateCompany)
	company.GET("/settings", api.getCompanySettings)
	company.PATCH("/settings", api.requirePermission("company.settings.manage"), api.updateCompanySettings)
	company.PATCH("/branding", api.requirePermission("branding.manage"), api.updateBranding)
	company.POST("/branding/upload", api.requirePermission("branding.manage"), api.uploadBrandingImage)
	company.POST("/submit-verification", api.requirePermission("company.update"), api.submitVerification)

	company.GET("/subdomain", api.requirePermission("subdomain.manage"), api.getCompanySubdomain)
	company.GET("/subdomain/availability", api.requirePermission("subdomain.manage"), api.checkSubdomain)
	company.PUT("/subdomain", api.requirePermission("subdomain.manage"), api.updateSubdomain)

	company.GET("/marketplace", api.requirePermission("company.settings.manage"), api.getCompanyMarketplaceSettings)
	company.PUT("/marketplace", api.requirePermission("company.settings.manage"), api.updateCompanyMarketplaceSettings)

	company.GET("/booking-page", api.requirePermission("bookings.manage"), api.getBookingPage)
	company.GET("/booking-page/availability", api.requirePermission("bookings.manage"), api.checkBookingPageSlug)
	company.PUT("/booking-page", api.requirePermission("bookings.manage"), api.upsertBookingPage)
	company.GET("/booking-routes", api.requirePermission("rates.view"), api.listCompanyBookingRoutes)
	company.POST("/booking-routes", api.requirePermission("rates.create"), api.createCompanyBookingRoute)
	company.PUT("/booking-routes/:id", api.requirePermission("rates.update"), api.updateCompanyBookingRoute)
	company.PATCH("/booking-routes/:id/status", api.requirePermission("rates.update"), api.setCompanyBookingRouteStatus)
	company.POST("/booking-routes/:id/duplicate", api.requirePermission("rates.create"), api.duplicateCompanyBookingRoute)
	company.DELETE("/booking-routes/:id", api.requirePermission("rates.update"), api.deleteCompanyBookingRoute)

	company.GET("/item-categories", api.requirePermission("rates.view"), api.getItemCategoryPolicy)
	company.PUT("/item-categories", api.requirePermission("rates.update"), api.updateItemCategoryPolicy)

	company.GET("/last-mile-areas", api.requirePermission("rates.view"), api.listLastMileAreas)
	company.POST("/last-mile-areas", api.requirePermission("rates.create"), api.createLastMileArea)
	company.GET("/last-mile-areas/:id", api.requirePermission("rates.view"), api.getLastMileArea)
	company.PUT("/last-mile-areas/:id", api.requirePermission("rates.update"), api.updateLastMileArea)
	company.DELETE("/last-mile-areas/:id", api.requirePermission("rates.update"), api.deleteLastMileArea)
	company.PUT("/last-mile-areas/:id/zones", api.requirePermission("rates.update"), api.replaceLastMileZones)
	company.PUT("/last-mile-areas/:id/zone-rates", api.requirePermission("rates.update"), api.replaceZoneRates)
	company.POST("/last-mile-areas/:id/preview", api.requirePermission("rates.view"), api.previewLastMilePrice)

	company.GET("/branches", api.requirePermission("branches.view"), api.listBranches)
	company.POST("/branches", api.requirePermission("branches.create"), api.createBranch)
	company.GET("/branches/:id", api.requirePermission("branches.view"), api.getBranch)
	company.PATCH("/branches/:id", api.requirePermission("branches.update"), api.updateBranch)
	company.POST("/branches/:id/activate", api.requirePermission("branches.deactivate"), api.setBranchStatus)
	company.POST("/branches/:id/deactivate", api.requirePermission("branches.deactivate"), api.setBranchStatus)
	company.GET("/branches/:id/staff", api.requirePermission("branches.view"), api.listBranchStaff)
	company.PUT("/branches/:id/staff", api.requirePermission("staff.update"), api.setBranchStaff)

	company.GET("/members", api.requirePermission("staff.view"), api.listMembers)
	company.GET("/drivers", api.requirePermission("staff.view"), api.listDrivers)
	company.GET("/drivers/live", api.requirePermission("shipments.view"), api.listLiveDrivers)
	company.POST("/drivers", api.requirePermission("staff.invite"), api.createDriver)
	company.PATCH("/drivers/:id", api.requirePermission("staff.update"), api.updateDriver)
	company.PATCH("/drivers/:id/status", api.requirePermission("staff.deactivate"), api.setDriverStatus)
	company.GET("/customers", api.requirePermission("staff.view"), api.listCompanyCustomers)
	company.GET("/customers/:id/bookings", api.requirePermission("staff.view"), api.listCompanyCustomerBookings)
	company.GET("/deliveries", api.requirePermission("shipments.view"), api.listDeliveries)
	company.POST("/deliveries", api.requirePermission("shipments.create"), api.createDelivery)
	company.GET("/requests", api.requirePermission("bookings.manage"), api.listBookingRequests)
	company.POST("/requests/:id/accept", api.requirePermission("bookings.manage"), api.acceptBookingRequest)
	company.POST("/requests/:id/decline", api.requirePermission("bookings.manage"), api.declineBookingRequest)
	company.GET("/deliveries/:id", api.requirePermission("shipments.view"), api.getDelivery)
	company.PUT("/deliveries/:id", api.requirePermission("shipments.update"), api.updateDelivery)
	company.PUT("/deliveries/:id/driver", api.requirePermission("shipments.update"), api.reassignDeliveryDriver)
	company.PATCH("/deliveries/:id/status", api.requirePermission("shipments.update"), api.updateDeliveryStatus)
	company.GET("/deliveries/:id/timeline", api.requirePermission("shipments.view"), api.getDeliveryTimeline)
	company.PATCH("/members/:id", api.requirePermission("staff.update"), api.updateMember)
	company.PUT("/members/:id/branches", api.requirePermission("staff.update"), api.setMemberBranches)
	company.PATCH("/members/:id/role", api.requirePermission("staff.roles.manage"), api.updateMemberRole)
	company.DELETE("/members/:id", api.requirePermission("staff.deactivate"), api.removeMember)

	company.GET("/invitations", api.requirePermission("staff.view"), api.listInvitations)
	company.POST("/invitations", api.requirePermission("staff.invite"), api.inviteStaff)
	company.DELETE("/invitations/:id", api.requirePermission("staff.invite"), api.revokeInvitation)
	company.POST("/invitations/:id/resend", api.requirePermission("staff.invite"), api.resendInvitation)

	company.GET("/roles", api.requirePermission("staff.roles.manage"), api.listRoles)
	company.POST("/roles", api.requirePermission("staff.roles.manage"), api.createRole)
	company.PATCH("/roles/:id", api.requirePermission("staff.roles.manage"), api.updateRole)
	company.DELETE("/roles/:id", api.requirePermission("staff.roles.manage"), api.deleteRole)

	staticHandler := web.Handler()
	r.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if requestPath == "/api" || strings.HasPrefix(requestPath, "/api/") ||
			requestPath == "/health" || strings.HasPrefix(requestPath, "/health/") ||
			requestPath == "/uploads" || strings.HasPrefix(requestPath, "/uploads/") {
			c.String(http.StatusNotFound, "404 page not found")
			return
		}

		staticHandler.ServeHTTP(c.Writer, c.Request)
	})

	return r
}

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type userDTO struct {
	ID               uuid.UUID  `json:"id"`
	Email            string     `json:"email"`
	FullName         string     `json:"full_name"`
	Phone            *string    `json:"phone"`
	WhatsAppVerified bool       `json:"whatsapp_verified"`
	AccountType      string     `json:"account_type"`
	PlatformRole     *string    `json:"platform_role"`
	EmailVerified    bool       `json:"email_verified"`
	CreatedAt        time.Time  `json:"created_at"`
	LastLoginAt      *time.Time `json:"last_login_at"`
}

func toUserDTO(user db.User) userDTO {
	return userDTO{
		ID:               user.ID,
		Email:            user.Email,
		FullName:         user.FullName,
		Phone:            user.Phone,
		WhatsAppVerified: false,
		AccountType:      user.AccountType,
		PlatformRole:     user.PlatformRole,
		EmailVerified:    user.EmailVerifiedAt != nil,
		CreatedAt:        user.CreatedAt,
		LastLoginAt:      user.LastLoginAt,
	}
}

type membershipDTO struct {
	CompanyID        uuid.UUID `json:"company_id"`
	CompanyName      string    `json:"company_name"`
	Slug             string    `json:"slug"`
	Status           string    `json:"status"`
	OnboardingStatus string    `json:"onboarding_status"`
	MemberID         uuid.UUID `json:"member_id"`
	MemberStatus     string    `json:"member_status"`
	Roles            []string  `json:"roles"`
}

func toMembershipDTOs(rows []db.ListUserMembershipsRow) []membershipDTO {
	out := make([]membershipDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, membershipDTO{
			CompanyID:        row.CompanyID,
			CompanyName:      row.Name,
			Slug:             row.Slug,
			Status:           row.Status,
			OnboardingStatus: row.OnboardingStatus,
			MemberID:         row.MemberID,
			MemberStatus:     row.MemberStatus,
			Roles:            nonNilStrings(row.Roles),
		})
	}
	return out
}

type sessionDTO struct {
	User                 userDTO         `json:"user"`
	Companies            []membershipDTO `json:"companies"`
	AccessToken          string          `json:"access_token"`
	ExpiresIn            int             `json:"expires_in"`
	DevVerificationToken string          `json:"dev_verification_token,omitempty"`
}

func toSessionDTO(session service.Session) sessionDTO {
	return sessionDTO{
		User:                 toUserDTO(session.User),
		Companies:            toMembershipDTOs(session.Companies),
		AccessToken:          session.AccessToken,
		ExpiresIn:            int(time.Until(session.AccessExpiresAt).Seconds()),
		DevVerificationToken: session.DevVerificationToken,
	}
}

// ---------------------------------------------------------------------------
// Cookies and responses
// ---------------------------------------------------------------------------

func (a *API) setSessionCookies(c *gin.Context, session service.Session) {
	accessMaxAge := int(time.Until(session.AccessExpiresAt).Seconds())
	if accessMaxAge < 0 {
		accessMaxAge = 0
	}
	refreshMaxAge := int(time.Until(session.RefreshExpiresAt).Seconds())
	if refreshMaxAge < 0 {
		refreshMaxAge = 0
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessCookieName,
		Value:    session.AccessToken,
		Path:     "/",
		MaxAge:   accessMaxAge,
		HttpOnly: true,
		Secure:   a.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshCookieName,
		Value:    session.RefreshToken,
		Path:     refreshCookiePath,
		MaxAge:   refreshMaxAge,
		HttpOnly: true,
		Secure:   a.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *API) clearSessionCookies(c *gin.Context) {
	expired := time.Unix(0, 0)
	http.SetCookie(c.Writer, &http.Cookie{
		Name: accessCookieName, Value: "", Path: "/", MaxAge: -1,
		Expires: expired, HttpOnly: true, Secure: a.cfg.CookieSecure, SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name: refreshCookieName, Value: "", Path: refreshCookiePath, MaxAge: -1,
		Expires: expired, HttpOnly: true, Secure: a.cfg.CookieSecure, SameSite: http.SameSiteLaxMode,
	})
}

func fail(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

func (a *API) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrEmailTaken):
		fail(c, http.StatusConflict, "email_taken", "An account with this email already exists.")
	case errors.Is(err, service.ErrInvalidCredentials):
		fail(c, http.StatusUnauthorized, "invalid_credentials", "Email or password is incorrect.")
	case errors.Is(err, service.ErrGoogleAccountRestricted):
		fail(c, http.StatusForbidden, "google_admin_blocked", "Administrator accounts must sign in with email and password.")
	case errors.Is(err, service.ErrGoogleAccountNotFound):
		fail(c, http.StatusNotFound, "google_account_not_found", "No BOL account is linked to that Google email.")
	case errors.Is(err, service.ErrGoogleAccountConflict):
		fail(c, http.StatusConflict, "google_account_conflict", "That Google account is already linked to another BOL account.")
	case errors.Is(err, service.ErrCompanySetupNotAllowed):
		fail(c, http.StatusForbidden, "company_setup_not_allowed", "Only provider accounts can set up a company.")
	case errors.Is(err, service.ErrCompanyAlreadyExists):
		fail(c, http.StatusConflict, "company_already_exists", "This account already belongs to a company.")
	case errors.Is(err, service.ErrInvalidRefresh):
		a.clearSessionCookies(c)
		fail(c, http.StatusUnauthorized, "invalid_refresh", "Your session has expired. Please sign in again.")
	case errors.Is(err, service.ErrInvalidToken):
		fail(c, http.StatusBadRequest, "invalid_token", "This link is invalid or has expired.")
	case errors.Is(err, service.ErrUserSuspended):
		fail(c, http.StatusForbidden, "account_suspended", "This account is suspended.")
	case errors.Is(err, service.ErrInvalidInvitation):
		fail(c, http.StatusGone, "invitation_invalid", "This invitation is invalid, expired or has already been used.")
	case errors.Is(err, service.ErrInvitationNeedsLogin):
		fail(c, http.StatusConflict, "sign_in_required", "Sign in to the invited account to accept this invitation.")
	case errors.Is(err, service.ErrInvitationEmailMismatch):
		fail(c, http.StatusForbidden, "invitation_email_mismatch", "This invitation was sent to a different email address.")
	case errors.Is(err, service.ErrInvalidInvitationBranch):
		fail(c, http.StatusConflict, "invitation_branch_invalid", "A branch on this invitation is no longer available.")
	case errors.Is(err, service.ErrWhatsAppUnavailable):
		fail(c, http.StatusServiceUnavailable, "whatsapp_unavailable", "WhatsApp verification is not configured. Please try again later.")
	case errors.Is(err, service.ErrPhoneInvalid):
		fail(c, http.StatusUnprocessableEntity, "phone_invalid", "Use an international phone number, such as +2348012345678.")
	case errors.Is(err, service.ErrPhoneCodeCooldown):
		fail(c, http.StatusTooManyRequests, "whatsapp_code_cooldown", "Wait one minute before requesting another code.")
	case errors.Is(err, service.ErrPhoneRateLimited):
		fail(c, http.StatusTooManyRequests, "whatsapp_code_rate_limited", "Too many codes were requested. Try again in an hour.")
	case errors.Is(err, service.ErrPhoneCodeInvalid):
		fail(c, http.StatusBadRequest, "whatsapp_code_invalid", "That WhatsApp verification code is incorrect.")
	case errors.Is(err, service.ErrPhoneCodeExpired):
		fail(c, http.StatusGone, "whatsapp_code_expired", "That WhatsApp verification code has expired. Request a new one.")
	case errors.Is(err, service.ErrInvalidPlatformRole):
		fail(c, http.StatusBadRequest, "invalid_platform_role", "Choose the admin or support role.")
	default:
		a.log.Error("request failed", "error", err, "path", c.FullPath())
		fail(c, http.StatusInternalServerError, "internal_error", "Something went wrong. Please try again.")
	}
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func (a *API) health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	status := http.StatusOK
	dbStatus := "up"
	redisStatus := "disabled"

	if err := a.svc.DB.Ping(ctx); err != nil {
		dbStatus = "down"
		status = http.StatusServiceUnavailable
	}
	if a.svc.Redis != nil {
		redisStatus = "up"
		if err := a.svc.Redis.Ping(ctx).Err(); err != nil {
			redisStatus = "down"
		}
	}

	c.JSON(status, gin.H{
		"status":   map[bool]string{true: "ok", false: "degraded"}[status == http.StatusOK],
		"database": dbStatus,
		"redis":    redisStatus,
	})
}
