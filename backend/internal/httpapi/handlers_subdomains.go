package httpapi

import (
	"encoding/json"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/service"
)

var subdomainPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,28}[a-z0-9]$`)

var reservedSubdomains = map[string]bool{
	"www": true, "api": true, "app": true, "admin": true, "dashboard": true,
	"mail": true, "support": true, "status": true, "blog": true, "docs": true,
	"marketplace": true, "bol": true, "login": true, "signup": true,
	"static": true, "assets": true, "cdn": true, "help": true, "billing": true,
}

// requestSubdomain resolves the left-most label from the Host header. It never
// accepts a company id from the client.
func (a *API) requestSubdomain(c *gin.Context) string {
	host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(c.Request.Host)
	}
	if host == "" {
		return ""
	}
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	} else if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	host = strings.ToLower(strings.Trim(host, "."))
	domain := strings.ToLower(strings.Trim(a.cfg.AppDomain, "."))
	if host == "" || domain == "" {
		return ""
	}
	if host == domain || host == "www."+domain || net.ParseIP(host) != nil {
		return ""
	}
	if !strings.HasSuffix(host, "."+domain) {
		return ""
	}
	label := strings.TrimSuffix(host, "."+domain)
	if label == "" || strings.Contains(label, ".") {
		return ""
	}
	return label
}

func validateSubdomain(value string) (string, string) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "", "Enter a subdomain."
	}
	if !subdomainPattern.MatchString(value) {
		return "", "Use 3–30 lowercase letters, numbers or hyphens, without a leading or trailing hyphen."
	}
	if strings.Contains(value, "--") {
		return "", "Consecutive hyphens are not allowed."
	}
	if reservedSubdomains[value] {
		return "", "That subdomain is reserved by the platform."
	}
	return value, ""
}

func (a *API) getCompanySubdomain(c *gin.Context) {
	tenant, _ := currentTenant(c)
	subdomain, err := a.svc.Q.GetSubdomainByCompany(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"subdomain": nil})
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"subdomain": subdomain.Subdomain})
}

func (a *API) checkSubdomain(c *gin.Context) {
	tenant, _ := currentTenant(c)
	value, reason := validateSubdomain(c.Query("subdomain"))
	if reason != "" {
		c.JSON(http.StatusOK, gin.H{"subdomain": value, "available": false, "reason": reason})
		return
	}

	taken, err := a.svc.Q.SubdomainExists(c.Request.Context(), value)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if taken {
		current, currentErr := a.svc.Q.GetSubdomainByCompany(c.Request.Context(), tenant.CompanyID)
		if currentErr == nil && current.Subdomain == value {
			c.JSON(http.StatusOK, gin.H{"subdomain": value, "available": true})
			return
		}
		c.JSON(http.StatusOK, gin.H{"subdomain": value, "available": false, "reason": "That subdomain is already taken."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"subdomain": value, "available": true})
}

func (a *API) updateSubdomain(c *gin.Context) {
	tenant, _ := currentTenant(c)
	var req struct {
		Subdomain string `json:"subdomain" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Enter a subdomain.")
		return
	}

	value, reason := validateSubdomain(req.Subdomain)
	if reason != "" {
		fail(c, http.StatusUnprocessableEntity, "invalid_subdomain", reason)
		return
	}

	taken, err := a.svc.Q.SubdomainExists(c.Request.Context(), value)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if taken {
		current, currentErr := a.svc.Q.GetSubdomainByCompany(c.Request.Context(), tenant.CompanyID)
		if currentErr != nil || current.Subdomain != value {
			fail(c, http.StatusConflict, "subdomain_taken", "That subdomain is already taken.")
			return
		}
	}

	record, err := a.svc.Q.UpsertSubdomain(c.Request.Context(), db.UpsertSubdomainParams{
		CompanyID: tenant.CompanyID,
		Subdomain: value,
	})
	if err != nil {
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "subdomain_taken", "That subdomain is already taken.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.subdomain_updated",
		EntityType: "company_subdomain",
		EntityID:   &record.ID,
		Metadata:   map[string]any{"subdomain": value},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"subdomain": record.Subdomain})
}

func (a *API) publicCompany(c *gin.Context) {
	value, reason := validateSubdomain(c.Param("subdomain"))
	if reason != "" {
		fail(c, http.StatusNotFound, "company_not_found", "That company page was not found.")
		return
	}
	a.respondPublicCompany(c, value)
}

func (a *API) publicBranding(c *gin.Context) {
	subdomain := a.requestSubdomain(c)
	if subdomain == "" {
		if query := strings.TrimSpace(c.Query("subdomain")); query != "" {
			subdomain, _ = validateSubdomain(query)
		}
	}
	if subdomain == "" {
		c.JSON(http.StatusOK, gin.H{
			"company":   nil,
			"branding":  nil,
			"subdomain": "",
			"manifest": gin.H{
				"name":             "based on logistics (BOL)",
				"short_name":       "BOL",
				"theme_color":      "#0a1220",
				"background_color": "#fbfbf9",
				"icons":            []gin.H{},
			},
		})
		return
	}
	a.respondPublicCompany(c, subdomain)
}

func (a *API) defaultManifestIcons() []gin.H {
	base := strings.TrimRight(a.cfg.AppURL, "/")
	return []gin.H{
		{"src": base + "/icon-192.png", "sizes": "192x192", "type": "image/png", "purpose": "any"},
		{"src": base + "/icon-512.png", "sizes": "512x512", "type": "image/png", "purpose": "any"},
		{"src": base + "/icon-512.png", "sizes": "512x512", "type": "image/png", "purpose": "maskable"},
	}
}

func (a *API) publicManifest(c *gin.Context) {
	subdomain := a.requestSubdomain(c)
	if subdomain == "" {
		if query := strings.TrimSpace(c.Query("subdomain")); query != "" {
			subdomain, _ = validateSubdomain(query)
		}
	}

	manifest := gin.H{
		"id":               "/",
		"name":             "based on logistics (BOL)",
		"short_name":       "BOL",
		"description":      "Logistics operations and marketplace bookings.",
		"start_url":        "/",
		"scope":            "/",
		"display":          "standalone",
		"theme_color":      "#0a1220",
		"background_color": "#fbfbf9",
		"icons":            a.defaultManifestIcons(),
	}

	if subdomain != "" {
		row, err := a.svc.Q.GetPublicCompanyBySubdomain(c.Request.Context(), subdomain)
		if err == nil {
			name := row.PublicName
			shortName := name
			if row.PwaShortName != nil && *row.PwaShortName != "" {
				shortName = *row.PwaShortName
			}
			if len(shortName) > 30 {
				shortName = shortName[:30]
			}
			icons := a.defaultManifestIcons()
			if row.LogoUrl != nil && *row.LogoUrl != "" {
				icons = append([]gin.H{{"src": *row.LogoUrl, "sizes": "any", "purpose": "any"}}, icons...)
			}
			manifest["name"] = name
			manifest["short_name"] = shortName
			manifest["description"] = row.Description
			manifest["theme_color"] = row.PrimaryColor
			manifest["icons"] = icons
			manifest["start_url"] = "/"
		}
	}

	encoded, err := json.Marshal(manifest)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.Header("Cache-Control", "public, max-age=300")
	c.Data(http.StatusOK, "application/manifest+json", encoded)
}

func (a *API) respondPublicCompany(c *gin.Context, subdomain string) {
	row, err := a.svc.Q.GetPublicCompanyBySubdomain(c.Request.Context(), subdomain)
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "company_not_found", "That company page was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	publicName := row.PublicName
	shortName := publicName
	if row.PwaShortName != nil && *row.PwaShortName != "" {
		shortName = *row.PwaShortName
	}
	if len(shortName) > 30 {
		shortName = shortName[:30]
	}

	icons := a.defaultManifestIcons()
	if row.LogoUrl != nil && *row.LogoUrl != "" {
		icons = append([]gin.H{{"src": *row.LogoUrl, "sizes": "any", "purpose": "any"}}, icons...)
	}

	c.JSON(http.StatusOK, gin.H{
		"company": gin.H{
			"id":            row.ID,
			"name":          row.Name,
			"slug":          row.Slug,
			"public_name":   publicName,
			"business_type": row.BusinessType,
			"description":   row.Description,
			"city":          row.City,
			"state":         row.State,
			"country":       row.Country,
			"status":        row.Status,
			"verified_at":   row.VerifiedAt,
		},
		"branding": gin.H{
			"logo_url":        row.LogoUrl,
			"primary_color":   row.PrimaryColor,
			"secondary_color": row.SecondaryColor,
			"favicon_url":     row.FaviconUrl,
			"cover_image_url": row.CoverImageUrl,
			"description":     row.BrandingDescription,
		},
		"subdomain": row.Subdomain,
		"manifest": gin.H{
			"name":             publicName,
			"short_name":       shortName,
			"theme_color":      row.PrimaryColor,
			"background_color": "#fbfbf9",
			"icons":            icons,
		},
	})
}

func (a *API) companySubdomain(c *gin.Context, companyID uuid.UUID) string {
	record, err := a.svc.Q.GetSubdomainByCompany(c.Request.Context(), companyID)
	if err != nil {
		return ""
	}
	return record.Subdomain
}
