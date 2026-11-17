package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bol-platform/backend/internal/auth"
	db "github.com/bol-platform/backend/internal/db/sqlc"
)

const (
	ctxUserKey   = "auth.user"
	ctxClaimsKey = "auth.claims"
	ctxTenantKey = "tenant.context"
)

type tenantContext struct {
	CompanyID  uuid.UUID
	MemberID   uuid.UUID
	IsPlatform bool
}

func currentUser(c *gin.Context) (db.User, bool) {
	value, ok := c.Get(ctxUserKey)
	if !ok {
		return db.User{}, false
	}
	user, ok := value.(db.User)
	return user, ok
}

func currentClaims(c *gin.Context) (*auth.Claims, bool) {
	value, ok := c.Get(ctxClaimsKey)
	if !ok {
		return nil, false
	}
	claims, ok := value.(*auth.Claims)
	return claims, ok
}

func currentTenant(c *gin.Context) (tenantContext, bool) {
	value, ok := c.Get(ctxTenantKey)
	if !ok {
		return tenantContext{}, false
	}
	tenant, ok := value.(tenantContext)
	return tenant, ok
}

func accessTokenFromRequest(c *gin.Context) string {
	if header := c.GetHeader("Authorization"); strings.HasPrefix(header, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	}
	if cookie, err := c.Cookie(accessCookieName); err == nil {
		return cookie
	}
	return ""
}

func (a *API) requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := accessTokenFromRequest(c)
		if raw == "" {
			fail(c, http.StatusUnauthorized, "unauthenticated", "Sign in to continue.")
			return
		}

		claims, err := a.auth.Parse(raw)
		if err != nil {
			fail(c, http.StatusUnauthorized, "invalid_access_token", "Your session has expired. Please sign in again.")
			return
		}
		if a.svc.IsAccessTokenRevoked(c.Request.Context(), claims.ID) {
			fail(c, http.StatusUnauthorized, "token_revoked", "This session has been signed out.")
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			fail(c, http.StatusUnauthorized, "invalid_access_token", "Your session is invalid.")
			return
		}
		user, err := a.svc.Q.GetUserByID(c.Request.Context(), userID)
		if err != nil || user.Status != "active" {
			fail(c, http.StatusUnauthorized, "invalid_access_token", "Your session is invalid.")
			return
		}

		c.Set(ctxUserKey, user)
		c.Set(ctxClaimsKey, claims)
		c.Next()
	}
}

// requireCompany resolves the active tenant from X-Company-Id. A company id
// supplied by the client is never trusted on its own: it must match an active
// membership unless the user holds an explicit platform administration role.
func (a *API) requireCompany() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			fail(c, http.StatusUnauthorized, "unauthenticated", "Sign in to continue.")
			return
		}

		header := strings.TrimSpace(c.GetHeader("X-Company-Id"))
		isPlatform := user.PlatformRole != nil &&
			(*user.PlatformRole == "super_admin" || *user.PlatformRole == "admin" || *user.PlatformRole == "support")

		var companyID uuid.UUID
		if header != "" {
			parsed, err := uuid.Parse(header)
			if err != nil {
				fail(c, http.StatusBadRequest, "invalid_company", "The selected company is invalid.")
				return
			}
			companyID = parsed
		} else {
			if isPlatform {
				fail(c, http.StatusBadRequest, "company_required", "Select a company to act on.")
				return
			}
			memberships, err := a.svc.Q.ListUserMemberships(c.Request.Context(), user.ID)
			if err != nil {
				a.handleServiceError(c, err)
				return
			}
			switch len(memberships) {
			case 0:
				fail(c, http.StatusForbidden, "no_company", "This account is not attached to a company.")
				return
			case 1:
				companyID = memberships[0].CompanyID
			default:
				fail(c, http.StatusBadRequest, "company_required", "Choose which company you are working in.")
				return
			}
		}

		company, err := a.svc.Q.GetCompanyByID(c.Request.Context(), companyID)
		if err != nil {
			fail(c, http.StatusForbidden, "company_forbidden", "You do not have access to this company.")
			return
		}

		if isPlatform {
			c.Set(ctxTenantKey, tenantContext{CompanyID: company.ID, IsPlatform: true})
			c.Next()
			return
		}

		member, err := a.svc.Q.GetActiveCompanyMember(c.Request.Context(), db.GetActiveCompanyMemberParams{
			CompanyID: companyID,
			UserID:    user.ID,
		})
		if err != nil {
			// Do not reveal whether the company exists.
			fail(c, http.StatusForbidden, "company_forbidden", "You do not have access to this company.")
			return
		}

		c.Set(ctxTenantKey, tenantContext{CompanyID: company.ID, MemberID: member.ID})
		c.Next()
	}
}

func (a *API) requirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			fail(c, http.StatusUnauthorized, "unauthenticated", "Sign in to continue.")
			return
		}
		tenant, ok := currentTenant(c)
		if !ok {
			fail(c, http.StatusForbidden, "company_required", "Select a company first.")
			return
		}

		if tenant.IsPlatform && user.PlatformRole != nil {
			// Support accounts get read-only platform access.
			if *user.PlatformRole == "support" && c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
				fail(c, http.StatusForbidden, "permission_denied", "Support accounts cannot change company data.")
				return
			}
			c.Next()
			return
		}

		allowed, err := a.svc.Q.UserHasPermission(c.Request.Context(), db.UserHasPermissionParams{
			UserID:    user.ID,
			CompanyID: tenant.CompanyID,
			Key:       permission,
		})
		if err != nil {
			a.handleServiceError(c, err)
			return
		}
		if !allowed {
			fail(c, http.StatusForbidden, "permission_denied", "You do not have permission to do that.")
			return
		}
		c.Next()
	}
}

func (a *API) rateLimit(bucket string, limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		allowed, retryAfter := a.limiter.Allow(c.Request.Context(), bucket+":"+c.ClientIP(), limit, time.Minute)
		if !allowed {
			seconds := int(retryAfter.Seconds()) + 1
			c.Header("Retry-After", strconv.Itoa(seconds))
			fail(c, http.StatusTooManyRequests, "rate_limited", "Too many attempts. Please wait and try again.")
			return
		}
		c.Next()
	}
}

func (a *API) isAllowedOrigin(origin string) bool {
	for _, allowed := range a.cfg.CORSOrigins {
		if strings.EqualFold(strings.TrimRight(origin, "/"), strings.TrimRight(allowed, "/")) {
			return true
		}
	}
	return false
}

func (a *API) cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && a.isAllowedOrigin(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Company-Id")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Max-Age", "600")
			c.Header("Vary", "Origin")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// originGuard is the CSRF layer for cookie-authenticated mutations: a browser
// request carrying a disallowed Origin is rejected before it reaches a handler.
func (a *API) originGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}
		if !a.isAllowedOrigin(origin) {
			fail(c, http.StatusForbidden, "origin_forbidden", "This origin is not allowed to make that request.")
			return
		}
		c.Next()
	}
}

func (a *API) securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

func (a *API) requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		a.log.Info("http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}
