package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bol-platform/backend/internal/auth"
	"github.com/bol-platform/backend/internal/config"
	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/ratelimit"
	"github.com/bol-platform/backend/internal/service"
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
	authRoutes.POST("/refresh", api.rateLimit("refresh", 30), api.refresh)
	authRoutes.POST("/logout", api.rateLimit("logout", 30), api.logout)
	authRoutes.POST("/forgot-password", api.rateLimit("forgot-password", 5), api.forgotPassword)
	authRoutes.POST("/reset-password", api.rateLimit("reset-password", 10), api.resetPassword)
	authRoutes.POST("/verify-email", api.rateLimit("verify-email", 10), api.verifyEmail)

	// Company registration is public: it creates the owner account and company
	// in a single transaction.
	v1.POST("/companies", api.rateLimit("company-register", 5), api.registerCompany)

	v1.GET("/public/branding", api.publicBranding)
	v1.GET("/public/manifest", api.publicManifest)
	v1.GET("/public/companies/:subdomain", api.publicCompany)
	v1.GET("/invitations/:token", api.rateLimit("invitation-preview", 30), api.previewInvitation)
	v1.POST("/invitations/:token/accept", api.rateLimit("invitation-accept", 10), api.acceptInvitation)

	authed := v1.Group("", api.requireAuth())
	authed.GET("/users/me", api.getMe)
	authed.PATCH("/users/me", api.updateMe)

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

	company.GET("/branches", api.requirePermission("branches.view"), api.listBranches)
	company.POST("/branches", api.requirePermission("branches.create"), api.createBranch)
	company.GET("/branches/:id", api.requirePermission("branches.view"), api.getBranch)
	company.PATCH("/branches/:id", api.requirePermission("branches.update"), api.updateBranch)
	company.POST("/branches/:id/activate", api.requirePermission("branches.deactivate"), api.setBranchStatus)
	company.POST("/branches/:id/deactivate", api.requirePermission("branches.deactivate"), api.setBranchStatus)
	company.GET("/branches/:id/staff", api.requirePermission("branches.view"), api.listBranchStaff)
	company.PUT("/branches/:id/staff", api.requirePermission("staff.update"), api.setBranchStaff)

	company.GET("/members", api.requirePermission("staff.view"), api.listMembers)
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

	// Rate engine routes
		company.GET("/countries", api.requirePermission("rates.view"), api.listCountries)
		company.POST("/countries", api.requirePermission("rates.create"), api.createCountry)

		company.GET("/states", api.requirePermission("rates.view"), api.listStates)
		company.POST("/states", api.requirePermission("rates.create"), api.createState)

		company.GET("/cities", api.requirePermission("rates.view"), api.listCities)
		company.POST("/cities", api.requirePermission("rates.create"), api.createCity)

		company.GET("/terminals", api.requirePermission("rates.view"), api.listTerminals)
		company.POST("/terminals", api.requirePermission("rates.create"), api.createTerminal)

		company.GET("/shipping-services", api.requirePermission("rates.view"), api.listShippingServices)
		company.POST("/shipping-services", api.requirePermission("rates.create"), api.createShippingService)

		company.GET("/routes", api.requirePermission("rates.view"), api.listCompanyRoutes)
		company.POST("/routes", api.requirePermission("rates.create"), api.createCompanyRoute)

		company.GET("/rate-cards", api.requirePermission("rates.view"), api.listRateCards)
		company.POST("/rate-cards", api.requirePermission("rates.create"), api.createRateCard)
		company.GET("/rate-cards/:id", api.requirePermission("rates.view"), api.getRateCard)

		company.GET("/rate-cards/:id/versions", api.requirePermission("rates.view"), api.listRateCardVersions)
		company.POST("/rate-cards/:id/versions", api.requirePermission("rates.create"), api.createRateCardVersion)

		company.GET("/rate-cards/:id/versions/:version_id/rules", api.requirePermission("rates.view"), api.listPricingRules)
		company.POST("/rate-cards/:id/versions/:version_id/rules", api.requirePermission("rates.create"), api.createPricingRule)

		company.GET("/rate-cards/:id/versions/:version_id/tiers", api.requirePermission("rates.view"), api.listPricingTiers)
		company.POST("/rate-cards/:id/versions/:version_id/tiers", api.requirePermission("rates.create"), api.createPricingTier)

		company.GET("/package-categories", api.requirePermission("rates.view"), api.listPackageCategories)
		company.POST("/package-categories", api.requirePermission("rates.create"), api.createPackageCategory)

		company.GET("/additional-charges", api.requirePermission("rates.view"), api.listAdditionalCharges)
		company.POST("/additional-charges", api.requirePermission("rates.create"), api.createAdditionalCharge)

		company.POST("/quotes", api.requirePermission("rates.create"), api.createShippingQuote)

		return r
}

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type userDTO struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	FullName      string     `json:"full_name"`
	Phone         *string    `json:"phone"`
	AccountType   string     `json:"account_type"`
	PlatformRole  *string    `json:"platform_role"`
	EmailVerified bool       `json:"email_verified"`
	CreatedAt     time.Time  `json:"created_at"`
	LastLoginAt   *time.Time `json:"last_login_at"`
}

func toUserDTO(user db.User) userDTO {
	return userDTO{
		ID:            user.ID,
		Email:         user.Email,
		FullName:      user.FullName,
		Phone:         user.Phone,
		AccountType:   user.AccountType,
		PlatformRole:  user.PlatformRole,
		EmailVerified: user.EmailVerifiedAt != nil,
		CreatedAt:     user.CreatedAt,
		LastLoginAt:   user.LastLoginAt,
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
