package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bol-platform/backend/internal/service"
)

// adminOverview returns everything the platform administration dashboard needs
// in a single read: marketplace customers, provider companies and staff
// accounts.
func (a *API) adminOverview(c *gin.Context) {
	customers, err := a.svc.Q.ListCustomerAccounts(c.Request.Context())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	providers, err := a.svc.Q.ListProviderCompanies(c.Request.Context())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	admins, err := a.svc.Q.ListPlatformAdmins(c.Request.Context())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customers": customers,
		"providers": providers,
		"admins":    admins,
	})
}

func (a *API) approveProvider(c *gin.Context) {
	user, _ := currentUser(c)
	if user.PlatformRole != nil && *user.PlatformRole == "support" {
		fail(c, http.StatusForbidden, "permission_denied", "Support accounts cannot approve providers.")
		return
	}

	companyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_provider", "That provider id is invalid.")
		return
	}

	company, err := a.svc.Q.ApproveCompany(c.Request.Context(), companyID)
	if err != nil {
		fail(c, http.StatusNotFound, "provider_not_found", "That provider could not be found.")
		return
	}

	if owner, err := a.svc.Q.GetUserByID(c.Request.Context(), company.OwnerUserID); err != nil {
		a.svc.Log.Warn("could not load the provider owner for the approval email", "company_id", company.ID, "error", err)
	} else if err := a.svc.Mailer.SendProviderApproved(c.Request.Context(), owner.Email, owner.FullName, company.Name); err != nil {
		a.svc.Log.Warn("could not send the provider approval email", "company_id", company.ID, "error", err)
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:   actorID(c),
		CompanyID: &company.ID,
		Action:    "admin.provider_approved",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	c.JSON(http.StatusOK, gin.H{"company": toCompanyDTO(company)})
}

func (a *API) createPlatformAdmin(c *gin.Context) {
	var req struct {
		FullName string `json:"full_name" binding:"required,min=2,max=120"`
		Email    string `json:"email" binding:"required,email,max=254"`
		Password string `json:"password" binding:"required,min=8,max=128"`
		Role     string `json:"role" binding:"required,oneof=admin support"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Enter a name, email, password and role for the admin.")
		return
	}

	user, err := a.svc.CreatePlatformAdmin(c.Request.Context(), service.PlatformAdminInput{
		Email:    strings.TrimSpace(req.Email),
		FullName: req.FullName,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:   actorID(c),
		Action:    "admin.created",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Metadata:  map[string]any{"email": user.Email, "role": req.Role},
	})

	c.JSON(http.StatusCreated, gin.H{
		"admin": gin.H{
			"id":            user.ID,
			"email":         user.Email,
			"full_name":     user.FullName,
			"platform_role": user.PlatformRole,
			"status":        user.Status,
			"created_at":    user.CreatedAt,
		},
	})
}
